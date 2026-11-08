package rgui

// Page groups a title and the content displayed by a Frame.
func Page(title string, content Node) Node {
	return WithChildren(&pageNode{title: title}, content)
}

type pageNode struct {
	BaseNode
	title string
}
