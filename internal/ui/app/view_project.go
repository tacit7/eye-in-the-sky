package app

import (
	"fmt"
	"strings"
)

// renderProjectTab renders the detected project information with sections
func (m *Model) renderProjectTab() string {
	var b strings.Builder

	if m.projectInfo == nil {
		b.WriteString(m.styles.Subtle.Render("  Not in a git repository"))
		b.WriteString("\n")
		return b.String()
	}

	// Project header
	b.WriteString(m.styles.Title.Render("Project: "))
	if m.projectInfo.Owner != "" && m.projectInfo.RepoName != "" {
		b.WriteString(m.styles.Text.Render(fmt.Sprintf("%s/%s", m.projectInfo.Owner, m.projectInfo.RepoName)))
	} else if m.projectInfo.RepoName != "" {
		b.WriteString(m.styles.Text.Render(m.projectInfo.RepoName))
	} else {
		b.WriteString(m.styles.Text.Render("(unknown)"))
	}
	b.WriteString("\n")

	// Project metadata
	if m.projectInfo.Branch != "" {
		b.WriteString(m.styles.Primary.Render("  Branch: "))
		b.WriteString(m.styles.Text.Render(m.projectInfo.Branch))
		b.WriteString("  ")
	}
	if m.projectInfo.Commit != "" {
		b.WriteString(m.styles.Primary.Render("Commit: "))
		b.WriteString(m.styles.Text.Render(m.projectInfo.Commit))
	}
	b.WriteString("\n")

	// Root path
	if m.projectInfo.GitRootPath != "" {
		b.WriteString(m.styles.Subtle.Render("  " + truncate(m.projectInfo.GitRootPath, 70)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Section navigation
	b.WriteString(m.renderProjectSectionNav())
	b.WriteString("\n")

	// Active section content
	switch m.projectSelectedSection {
	case 0: // TaskWarrior tasks
		b.WriteString(m.renderProjectTasks())
	case 1: // CLAUDE.md
		b.WriteString(m.renderProjectClaudeMD())
	case 2: // .md files
		b.WriteString(m.renderProjectMarkdownFiles())
	}

	return b.String()
}

// renderProjectSectionNav renders the section navigation indicator
func (m *Model) renderProjectSectionNav() string {
	sections := []string{"[1] Tasks", "[2] CLAUDE.md", "[3] Markdown"}
	var nav string

	for i, sec := range sections {
		if i == m.projectSelectedSection {
			nav += m.styles.Primary.Reverse(true).Render(" " + sec + " ")
		} else {
			nav += m.styles.Text.Render(" " + sec + " ")
		}
	}

	return "  " + nav
}

// renderProjectTasks renders the open TaskWarrior tasks section
func (m *Model) renderProjectTasks() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Open Tasks"))
	b.WriteString("\n")

	if len(m.projectTasks) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No open tasks"))
		b.WriteString("\n")
		return b.String()
	}

	for i, task := range m.projectTasks {
		selected := i == m.projectTasksIndex
		line := fmt.Sprintf("  %s  %s", task.Priority, truncate(task.Description, 60))

		if selected {
			b.WriteString(m.styles.Primary.Reverse(true).Render(line))
		} else {
			b.WriteString(m.styles.Text.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderProjectClaudeMD renders the CLAUDE.md content
func (m *Model) renderProjectClaudeMD() string {
	var b strings.Builder

	if m.claudeMDContent == "" {
		b.WriteString(m.styles.Subtle.Render("  CLAUDE.md not found"))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(m.styles.Title.Render("CLAUDE.md"))
	b.WriteString("\n")

	// Render first ~20 lines
	lines := strings.Split(m.claudeMDContent, "\n")
	maxLines := 20
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	for i := 0; i < maxLines; i++ {
		line := lines[i]
		if len(line) > 70 {
			line = truncate(line, 70)
		}
		b.WriteString(m.styles.Text.Render(line))
		b.WriteString("\n")
	}

	if len(lines) > maxLines {
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... (%d more lines)", len(lines)-maxLines)))
		b.WriteString("\n")
	}

	return b.String()
}

// renderProjectMarkdownFiles renders the list of markdown files
func (m *Model) renderProjectMarkdownFiles() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Markdown Files"))
	b.WriteString("\n")

	if len(m.projectMDFiles) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No markdown files found"))
		b.WriteString("\n")
		return b.String()
	}

	for i, file := range m.projectMDFiles {
		selected := i == m.projectMDFilesIndex
		line := fmt.Sprintf("  %s", file.Name)

		if selected {
			b.WriteString(m.styles.Primary.Reverse(true).Render(line))
			b.WriteString("\n")

			// Show preview of selected file
			preview := strings.Split(file.Content, "\n")
			previewLines := 5
			if len(preview) < previewLines {
				previewLines = len(preview)
			}

			for j := 0; j < previewLines; j++ {
				pline := preview[j]
				if len(pline) > 68 {
					pline = truncate(pline, 68)
				}
				b.WriteString(m.styles.Subtle.Render("    " + pline))
				b.WriteString("\n")
			}

			if len(preview) > previewLines {
				b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("    ... (%d more lines)", len(preview)-previewLines)))
				b.WriteString("\n")
			}
		} else {
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	return b.String()
}