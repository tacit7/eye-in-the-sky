package parser

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/models"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
)

// Parser handles parallel JSONL file parsing
type Parser struct {
	workers       int
	dedupeMap     map[string]bool
	dedupeMutex   sync.Mutex
}

// ParseResult contains the result of parsing a file
type ParseResult struct {
	Entries []db.UsageEntryRow
	Error   error
}

// New creates a new Parser with specified worker count
func New(workerCount int) *Parser {
	if workerCount <= 0 {
		workerCount = 4 // Default to 4 workers
	}
	return &Parser{
		workers:   workerCount,
		dedupeMap: make(map[string]bool),
	}
}

// ParseFiles parses JSONL files using a worker pool
func (p *Parser) ParseFiles(files []FileInfo) ([]db.UsageEntryRow, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to parse")
	}

	// Create channels for work distribution and results
	fileChan := make(chan FileInfo, len(files))
	resultChan := make(chan ParseResult, len(files))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go p.parseWorker(&wg, fileChan, resultChan)
	}

	// Send work to channel
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Collect results
	var allEntries []db.UsageEntryRow
	var errs []error
	received := 0

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		received++
		if result.Error != nil {
			errs = append(errs, result.Error)
		}
		allEntries = append(allEntries, result.Entries...)
	}

	// Return all entries even if some files had errors
	return allEntries, nil
}

// parseWorker processes files from the channel
func (p *Parser) parseWorker(wg *sync.WaitGroup, fileChan chan FileInfo, resultChan chan ParseResult) {
	defer wg.Done()

	for file := range fileChan {
		entries, err := p.parseFile(file)
		resultChan <- ParseResult{
			Entries: entries,
			Error:   err,
		}
	}
}

// parseFile parses a single JSONL file
func (p *Parser) parseFile(file FileInfo) ([]db.UsageEntryRow, error) {
	f, err := os.Open(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file.Path, err)
	}
	defer f.Close()

	var entries []db.UsageEntryRow
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Parse JSON
		var entry models.UsageEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Silently skip malformed JSON
			continue
		}

		// Validate and convert to database row
		row := p.convertToDBRow(entry, file.Project)
		if row == nil {
			continue
		}

		// Check for duplicate
		if p.isDuplicate(row.UniqueHash) {
			continue
		}

		entries = append(entries, *row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file %s: %w", file.Path, err)
	}

	return entries, nil
}

// convertToDBRow converts a UsageEntry to a database row
func (p *Parser) convertToDBRow(entry models.UsageEntry, project string) *db.UsageEntryRow {
	// Validate required fields
	if entry.SessionID == "" || entry.Message.ID == "" || entry.RequestID == "" {
		return nil
	}

	// Create unique hash from messageId + requestId
	hash := createHash(entry.Message.ID + entry.RequestID)

	// Parse timestamp
	timestamp := entry.Timestamp.Format(time.RFC3339)

	return &db.UsageEntryRow{
		SessionID:           entry.SessionID,
		Timestamp:           timestamp,
		Project:             project,
		Model:               entry.Message.Model,
		InputTokens:         entry.Message.Usage.InputTokens,
		OutputTokens:        entry.Message.Usage.OutputTokens,
		CacheCreationTokens: entry.Message.Usage.CacheCreationInputTokens,
		CacheReadTokens:     entry.Message.Usage.CacheReadInputTokens,
		TotalCost:           entry.CostUSD,
		MessageID:           entry.Message.ID,
		RequestID:           entry.RequestID,
		UniqueHash:          hash,
	}
}

// isDuplicate checks if a hash has been seen before
func (p *Parser) isDuplicate(hash string) bool {
	p.dedupeMutex.Lock()
	defer p.dedupeMutex.Unlock()

	if p.dedupeMap[hash] {
		return true
	}
	p.dedupeMap[hash] = true
	return false
}

// createHash creates a SHA256 hash of the input string
func createHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}
