package dashboard

// renderFooter shows the keybindings available in the current mode. Some
// actions (edit, delete) are listed even though they are not implemented
// yet — pressing them tells the user so via the status line instead of
// doing nothing silently.
func (m Model) renderFooter() string {
	if m.mode == viewDetail {
		return footerStyle.Render("esc: back  ·  q: quit")
	}
	return footerStyle.Render("a: new application  ·  enter: open  ·  e: edit  ·  d: delete  ·  r: refresh  ·  q: quit")
}
