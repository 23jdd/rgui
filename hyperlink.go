package rgui

// Hyperlink invokes navigate when clicked or activated from the keyboard.
// The application decides how to handle the destination.
func Hyperlink(label string, navigate func()) Node {
	link := componentOf(Button(label, navigate))
	link.link = true
	return link
}
