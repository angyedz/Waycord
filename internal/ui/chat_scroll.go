package ui

func messageHistoryOverflows(contentHeight, viewportHeight int) bool {
	return viewportHeight > 0 && contentHeight > viewportHeight
}

func (a *App) scrollMessagesToLatest() {
	a.messageList.ScrollToEnd = true
	a.messageList.Position.BeforeEnd = false
}
