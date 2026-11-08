package rgui

// BulletDecorator places a bullet beside content, aligned with its top edge.
func BulletDecorator(bullet, content Node, gap float32) Node {
	marker := panel(overlayLayout{}, []Node{bullet})
	marker.Margin.Right = max(0, gap)
	marker.VerticalAlignment = Start
	return DockPanel(true, Docked{Edge: DockLeft, Child: marker}, Docked{Child: content})
}
