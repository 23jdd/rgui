package main

import (
	"log"
	"os"
	"path/filepath"

	"rgui"
)

func view() rgui.Node {
	value := rgui.NewSignal("")
	return rgui.StackPanel(rgui.Vertical, 12,
		rgui.Text(func() string { return "Click the field and type with an IME." }),
		rgui.TextBox(value),
		rgui.Text(func() string { return "Committed: " + value.Get() }),
		rgui.Button("Clear", func() { value.Set("") }),
	)
}

func main() {
	config := rgui.WindowConfig{
		Title: "rgui Windows IME", Width: 600,
		FontPath:    filepath.Join(os.Getenv("WINDIR"), "Fonts", "Deng.ttf"),
		ExtraGlyphs: "你好世界中文输入测试拼音候选窗口我爱北京上海广州深圳中华人民共和国",
	}
	if err := rgui.Run(config, view); err != nil {
		log.Fatal(err)
	}
}
