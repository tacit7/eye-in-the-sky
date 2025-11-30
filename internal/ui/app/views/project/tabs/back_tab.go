package tabs

// RenderBack renders the back button (placeholder)
func RenderBack(ctx *DataContext, styles OverviewStyles) string {
	return styles.Primary.Render("← Back to Overview")
}
