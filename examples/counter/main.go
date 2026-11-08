package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"rgui"
)

func View() rgui.Node {
	count := rgui.NewSignal(0)

	return rgui.Stack(
		rgui.Text(func() string {
			return fmt.Sprintf("点击次数：%d", count.Get())
		}),
		rgui.Button("加一", func() {
			count.Set(count.Get() + 1)
		}),
	)
}

func main() {
	if err := rgui.Run(rgui.WindowConfig{
		Title:       "rgui counter",
		FontPath:    filepath.Join(os.Getenv("WINDIR"), "Fonts", "Deng.ttf"),
		ExtraGlyphs: "点击次数：加一",
	}, View); err != nil {
		log.Fatal(err)
	}
}
