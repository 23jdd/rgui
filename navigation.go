package rgui

// Navigator keeps a history of pages and other nodes. All calls belong to the
// UI thread. Use one Frame per navigator.
type Navigator struct {
	pages   []Node
	index   int
	current *Signal[Node]
}

func NewNavigator(initial Node) *Navigator {
	n := &Navigator{current: NewSignal[Node](nil), index: -1}
	n.Navigate(initial)
	return n
}

func (n *Navigator) Navigate(page Node) {
	if page == nil {
		return
	}
	clear(n.pages[n.index+1:])
	n.pages = append(n.pages[:n.index+1], page)
	n.index++
	n.current.Set(page)
}

func (n *Navigator) Back() bool    { return n.goTo(n.index - 1) }
func (n *Navigator) Forward() bool { return n.goTo(n.index + 1) }

func (n *Navigator) goTo(index int) bool {
	if index < 0 || index >= len(n.pages) {
		return false
	}
	n.index = index
	n.current.Set(n.pages[index])
	return true
}

func (n *Navigator) Title() string {
	if node, ok := n.current.Get().(*component); ok {
		if page, ok := node.custom.(*pageNode); ok {
			return page.title
		}
	}
	return ""
}

// Frame hosts the current page and unmounts it when navigation changes.
func Frame(navigation *Navigator) Node {
	if navigation == nil {
		navigation = NewNavigator(nil)
	}
	return WithChildren(&frameNode{navigation: navigation})
}

type frameNode struct {
	BaseNode
	navigation *Navigator
}

func (f *frameNode) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}), evaluate: func() {
		var children []Node
		if page := f.navigation.current.Get(); page != nil {
			children = []Node{page}
		}
		if err := SetChildren(f.host, children...); err != nil {
			panic(err)
		}
	}}
	update.evaluateTracked()
	return update
}
