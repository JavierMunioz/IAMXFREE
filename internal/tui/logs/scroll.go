package logs

// currentTop reports the absolute index (into buffer.all()) of the
// topmost visible line, given a total of total retained lines.
func (m Model) currentTop(total int) int {
	maxTop := total - m.visibleHeight()
	if maxTop < 0 {
		maxTop = 0
	}
	if m.followLive {
		return maxTop
	}

	top := m.topIndex
	if top > maxTop {
		top = maxTop
	}
	if top < 0 {
		top = 0
	}
	return top
}

// scrollBy moves the visible window by delta lines (negative scrolls
// toward older lines, positive toward newer). Scrolling back down to the
// bottom re-enables followLive.
func (m Model) scrollBy(delta int) Model {
	total := m.buffer.len()
	top := m.currentTop(total) + delta

	maxTop := total - m.visibleHeight()
	if maxTop < 0 {
		maxTop = 0
	}
	if top < 0 {
		top = 0
	}
	if top > maxTop {
		top = maxTop
	}

	m.topIndex = top
	m.followLive = top >= maxTop
	return m
}

// scrollToStart jumps to the oldest retained line and disables followLive.
func (m Model) scrollToStart() Model {
	m.topIndex = 0
	m.followLive = false
	return m
}

// scrollToEnd jumps to the newest line and re-enables followLive, so new
// incoming lines keep scrolling into view.
func (m Model) scrollToEnd() Model {
	m.followLive = true
	return m
}
