package tabs

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderFiles renders the files tab with file browser detection
func RenderFiles(ctx *DataContext, styles OverviewStyles) string {
	if ctx == nil || ctx.Project == nil {
		return theme.TextMuted.Render("\n  No project selected\n")
	}

	var content strings.Builder

	// Header
	content.WriteString(styles.SectionTitle.Render("\uf07c Project Files") + "\n\n") // nf-fa-folder_open

	// Project path
	projectPath := ""
	if ctx.Project.Path != nil {
		projectPath = *ctx.Project.Path
	}

	if projectPath == "" {
		content.WriteString(theme.TextMuted.Render("No project path available\n"))
		return content.String()
	}

	content.WriteString(styles.Label.Render("Project Path: "))
	content.WriteString(styles.Value.Render(projectPath) + "\n\n")

	// Detect available file browsers
	hasRanger := commandExists("ranger")
	hasLf := commandExists("lf")

	if !hasRanger && !hasLf {
		content.WriteString(styles.Warning.Render("\uf071 No file browser detected\n\n")) // nf-fa-exclamation_triangle
		content.WriteString(theme.TextMuted.Render("Install ranger or lf for file browsing:\n"))
		content.WriteString(theme.TextMuted.Render("  brew install ranger\n"))
		content.WriteString(theme.TextMuted.Render("  brew install lf\n"))
		return content.String()
	}

	// Show available browsers
	content.WriteString(styles.Success.Render("\uf00c File browsers available:\n\n")) // nf-fa-check

	if hasRanger {
		content.WriteString(fmt.Sprintf("  %s ranger\n", styles.Primary.Render("•")))
		content.WriteString(fmt.Sprintf("    %s\n", theme.TextMuted.Render("Launch: ranger "+projectPath)))
	}

	if hasLf {
		content.WriteString(fmt.Sprintf("  %s lf\n", styles.Primary.Render("•")))
		content.WriteString(fmt.Sprintf("    %s\n", theme.TextMuted.Render("Launch: lf "+projectPath)))
	}

	content.WriteString("\n")
	content.WriteString(theme.TextMuted.Render("Note: Launch file browser from a separate terminal to browse project files\n"))

	return content.String()
}

// commandExists checks if a command is available in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
