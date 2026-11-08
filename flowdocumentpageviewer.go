package rgui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"rgui/draw"
)

// FlowDocumentPageViewer wraps paragraphs and paginates them using the
// current viewport. Index holds the zero-based requested page.
func FlowDocumentPageViewer(paragraphs []string, index *Signal[int]) Node {
	if index == nil {
		index = NewSignal(0)
	}
	page := &flowPage{text: strings.Join(paragraphs, "\n\n"), index: index, count: NewSignal(1)}
	page.Padding = UniformInsets(8)
	return documentControls(index, page.count.Get, page)
}

type flowPage struct {
	BaseNode
	text         string
	index, count *Signal[int]
	measureText  func(string) rl.Vector2
	lines        []string
	perPage      int
}

func (p *flowPage) Measure(ctx MeasureContext) rl.Vector2 {
	p.measureText = ctx.MeasureText
	return rl.Vector2{X: min(360, ctx.Available.X), Y: min(240, ctx.Available.Y)}
}

func (p *flowPage) bind(pending subscriptions) *binding {
	return bindVisual(&p.BaseNode, p.index, pending)
}

func (p *flowPage) Place(bounds rl.Rectangle) {
	p.BaseNode.Place(bounds)
	p.lines = draw.WrapLines(p.text, max(1, bounds.Width), func(line string) float32 { return p.measureText(line).X })
	p.perPage = max(1, int(bounds.Height/fontSize))
	count := max(1, (len(p.lines)+p.perPage-1)/p.perPage)
	if p.count.Get() != count {
		p.count.Set(count)
	}
}

func (p *flowPage) Primitives(ctx PaintContext) []Node {
	bounds := p.Bounds()
	first := documentIndex(p.index.Get(), p.count.Get()) * p.perPage
	end := min(len(p.lines), first+p.perPage)
	var nodes []Node
	for i := first; i < end; i++ {
		nodes = append(nodes, &TextNode{Position: rl.Vector2{X: bounds.X, Y: bounds.Y + float32(i-first)*fontSize},
			Text: p.lines[i], Style: textStyle(ctx.Font)})
	}
	return nodes
}
