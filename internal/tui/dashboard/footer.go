package dashboard

// renderFooter shows the keybindings available on the dashboard. Editing,
// deleting and every other per-application action now live in the detail
// view (opened with enter); the dashboard itself is just the list and the
// entry point.
func (m Model) renderFooter() string {
	return footerStyle.Render("a: new application  ·  enter: open  ·  r: refresh  ·  q: quit")
}
