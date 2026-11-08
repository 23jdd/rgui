package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestThumbReportsDeltaAndCaptureLoss(t *testing.T) {
	var delta rl.Vector2
	var canceled bool
	root := componentOf(Thumb(DragCallbacks{Moved: func(value rl.Vector2) { delta = value }, Completed: func(_ rl.Vector2, c bool) { canceled = c }}))
	root.mount(make(subscriptions))
	defer root.unmount()
	var router inputRouter
	router.pointer(PointerDown, root, rl.Vector2{X: 3, Y: 4}, rl.Vector2{})
	router.pointer(PointerMove, nil, rl.Vector2{X: 20, Y: 2}, rl.Vector2{})
	router.cancel()
	if delta != (rl.Vector2{X: 17, Y: -2}) || !canceled {
		t.Fatalf("delta=%+v canceled=%v", delta, canceled)
	}
}
