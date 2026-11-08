package rgui

// Expander displays a header and conditionally mounts its content. The
// expanded signal can also be changed by the application.
func Expander(header string, content Node, expanded *Signal[bool]) Node {
	if expanded == nil {
		expanded = NewSignal(false)
	}
	button := componentOf(Button(header, func() { expanded.Set(!expanded.Get()) }))
	button.readText = func() string {
		if expanded.Get() {
			return "v  " + header
		}
		return ">  " + header
	}
	body := newContentSwitch(func() int {
		if expanded.Get() {
			return 0
		}
		return -1
	}, []Node{content})
	return StackPanel(Vertical, 0, button, body)
}

type contentSwitch struct {
	BaseNode
	selectIndex func() int
	contents    []Node
	index       int
}

func newContentSwitch(selectIndex func() int, contents []Node) Node {
	return WithChildren(&contentSwitch{selectIndex: selectIndex, contents: contents, index: -2})
}

func (s *contentSwitch) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}),
		evaluate: func() {
			index := s.selectIndex()
			if index == s.index {
				return
			}
			var children []Node
			if index >= 0 && index < len(s.contents) && s.contents[index] != nil {
				children = []Node{s.contents[index]}
			}
			if err := SetChildren(s.host, children...); err != nil {
				panic(err)
			}
			s.index = index
		},
	}
	update.evaluateTracked()
	return update
}
