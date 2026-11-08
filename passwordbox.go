package rgui

// PasswordBox edits masked text. It permits paste and blocks copy and cut.
// The application owns the signal and decides how the password is stored.
func PasswordBox(value *Signal[string]) Node {
	box := TextBox(value).(*textBox)
	box.masked = true
	return box
}
