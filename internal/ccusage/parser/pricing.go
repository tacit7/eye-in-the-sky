package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// LiteLLMPricing represents the pricing data from LiteLLM
type LiteLLMPricing struct {
	InputCostPerToken                float64 `json:"input_cost_per_token"`
	OutputCostPerToken               float64 `json:"output_cost_per_token"`
	CacheCreationInputTokenCost      float64 `json:"cache_creation_input_token_cost"`
	CacheReadInputTokenCost          float64 `json:"cache_read_input_token_cost"`
	InputCostPerTokenAbove200k       float64 `json:"input_cost_per_token_above_200k_tokens"`
	OutputCostPerTokenAbove200k      float64 `json:"output_cost_per_token_above_200k_tokens"`
	CacheCreationAbove200k           float64 `json:"cache_creation_input_token_cost_above_200k_tokens"`
	CacheReadAbove200k               float64 `json:"cache_read_input_token_cost_above_200k_tokens"`
}

// PricingCache holds cached pricing data
type PricingCache struct {
	mu       sync.RWMutex
	data     map[string]LiteLLMPricing
	lastFetch time.Time
}

const liteLLMPricingURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"
const cacheTTL = 24 * time.Hour

var pricingCache = &PricingCache{
	data: make(map[string]LiteLLMPricing),
}

// fetchPricingFromLiteLLM fetches pricing data from LiteLLM GitHub
func fetchPricingFromLiteLLM() (map[string]LiteLLMPricing, error) {
	log.Println("[PRICING] Fetching pricing data from LiteLLM...")
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(liteLLMPricingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pricing: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing fetch failed with status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read pricing response: %w", err)
	}
	
	var data map[string]LiteLLMPricing
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse pricing JSON: %w", err)
	}
	
	log.Printf("[PRICING] Successfully fetched pricing for %d models", len(data))
	return data, nil
}

// getCachedPricing returns pricing data from cache, fetching if necessary
func getCachedPricing() (map[string]LiteLLMPricing, error) {
	pricingCache.mu.RLock()
	if len(pricingCache.data) > 0 && time.Since(pricingCache.lastFetch) < cacheTTL {
		defer pricingCache.mu.RUnlock()
		return pricingCache.data, nil
	}
	pricingCache.mu.RUnlock()
	
	// Fetch fresh data
	data, err := fetchPricingFromLiteLLM()
	if err != nil {
		log.Printf("[PRICING] Error fetching: %v, using fallback", err)
		// Return cached data if available, even if stale
		pricingCache.mu.RLock()
		defer pricingCache.mu.RUnlock()
		if len(pricingCache.data) > 0 {
			return pricingCache.data, nil
		}
		// No cache available, return error
		return nil, err
	}
	
	// Update cache
	pricingCache.mu.Lock()
	pricingCache.data = data
	pricingCache.lastFetch = time.Now()
	pricingCache.mu.Unlock()
	
	return data, nil
}

// getModelPricing returns pricing for a specific model
func getModelPricing(model string) *LiteLLMPricing {
	pricing, err := getCachedPricing()
	if err != nil {
		log.Printf("[PRICING] Failed to get pricing, using fallback for model %s", model)
		return nil
	}
	
	// Try direct match
	if p, ok := pricing[model]; ok {
		return &p
	}
	
	// Try with anthropic/ prefix
	if p, ok := pricing["anthropic/"+model]; ok {
		return &p
	}
	
	log.Printf("[PRICING] Model %s not found in LiteLLM pricing, using fallback", model)
	return nil
}

// calculateCostFromLiteLLM calculates cost using LiteLLM pricing data
func calculateCostFromLiteLLM(model string, inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens int) float64 {
	pricing := getModelPricing(model)
	if pricing == nil {
		// Fall back to hardcoded pricing
		return calculateCostFallback(model, inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens)
	}
	
	cost := 0.0
	
	// Calculate input cost
	if inputTokens > 200000 {
		aboveTokens := inputTokens - 200000
		belowTokens := 200000
		cost += float64(belowTokens) * pricing.InputCostPerToken
		cost += float64(aboveTokens) * pricing.InputCostPerTokenAbove200k
	} else {
		cost += float64(inputTokens) * pricing.InputCostPerToken
	}
	
	// Calculate output cost
	if outputTokens > 200000 {
		aboveTokens := outputTokens - 200000
		belowTokens := 200000
		cost += float64(belowTokens) * pricing.OutputCostPerToken
		cost += float64(aboveTokens) * pricing.OutputCostPerTokenAbove200k
	} else {
		cost += float64(outputTokens) * pricing.OutputCostPerToken
	}
	
	// Calculate cache creation cost
	if cacheCreationTokens > 200000 {
		aboveTokens := cacheCreationTokens - 200000
		belowTokens := 200000
		cost += float64(belowTokens) * pricing.CacheCreationInputTokenCost
		cost += float64(aboveTokens) * pricing.CacheCreationAbove200k
	} else {
		cost += float64(cacheCreationTokens) * pricing.CacheCreationInputTokenCost
	}
	
	// Calculate cache read cost
	if cacheReadTokens > 200000 {
		aboveTokens := cacheReadTokens - 200000
		belowTokens := 200000
		cost += float64(belowTokens) * pricing.CacheReadInputTokenCost
		cost += float64(aboveTokens) * pricing.CacheReadAbove200k
	} else {
		cost += float64(cacheReadTokens) * pricing.CacheReadInputTokenCost
	}
	
	return cost
}

// calculateCostFallback uses hardcoded pricing as fallback
func calculateCostFallback(model string, inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens int) float64 {
	// Pricing per 1M tokens
	var inputPrice, outputPrice, cacheWritePrice, cacheReadPrice float64
	
	// Model pricing (in USD per 1M tokens)
	switch model {
	case "claude-opus-4-20250514", "claude-opus-4-1-20250805":
		inputPrice = 15.0
		outputPrice = 75.0
		cacheWritePrice = 18.75 // 25% of output price
		cacheReadPrice = 1.50   // 2% of output price
	case "claude-sonnet-4-20250514", "claude-sonnet-4-1-20250805":
		inputPrice = 3.0
		outputPrice = 15.0
		cacheWritePrice = 3.75   // 25% of output price
		cacheReadPrice = 0.30    // 2% of output price
	case "claude-haiku-4-5-20251001":
		inputPrice = 0.80
		outputPrice = 4.0
		cacheWritePrice = 1.0    // 25% of output price
		cacheReadPrice = 0.08    // 2% of output price
	default:
		// Fallback to Sonnet 4 pricing if model not recognized
		inputPrice = 3.0
		outputPrice = 15.0
		cacheWritePrice = 3.75
		cacheReadPrice = 0.30
	}
	
	// Calculate total cost
	inputCost := float64(inputTokens) * inputPrice / 1_000_000
	outputCost := float64(outputTokens) * outputPrice / 1_000_000
	cacheWriteCost := float64(cacheCreationTokens) * cacheWritePrice / 1_000_000
	cacheReadCost := float64(cacheReadTokens) * cacheReadPrice / 1_000_000
	
	return inputCost + outputCost + cacheWriteCost + cacheReadCost
}
