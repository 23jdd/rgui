package main

import (
	"fmt"
	"log"

	"rgui"
)

func view() rgui.Node {
	count := rgui.NewSignal(0)
	toolbar := rgui.Grid(
		[]rgui.Track{rgui.Fixed(48)},
		[]rgui.Track{rgui.Auto(), rgui.Star(1), rgui.Fixed(58), rgui.Fixed(58)},
		rgui.Cell{Column: 0, Child: rgui.Text(func() string {
			return fmt.Sprintf("Count: %d", count.Get())
		})},
		rgui.Cell{Column: 2, Child: rgui.Button("-", func() { count.Set(count.Get() - 1) }).
			WithLayout(rgui.Box{Margin: rgui.Insets{Right: 4}})},
		rgui.Cell{Column: 3, Child: rgui.Button("+", func() { count.Set(count.Get() + 1) }).
			WithLayout(rgui.Box{Margin: rgui.Insets{Left: 4}})},
	)
	ratio := rgui.Grid(
		[]rgui.Track{rgui.Fixed(48)},
		[]rgui.Track{rgui.Star(1), rgui.Star(2)},
		rgui.Cell{Column: 0, Child: rgui.Button("1*", func() { count.Set(1) }).
			WithLayout(rgui.Box{Margin: rgui.Insets{Right: 4}})},
		rgui.Cell{Column: 1, Child: rgui.Button("2*", func() { count.Set(2) }).
			WithLayout(rgui.Box{Margin: rgui.Insets{Left: 4}})},
	)
	return rgui.StackPanel(rgui.Vertical, 12, toolbar, ratio)
}

func main() {
	if err := rgui.Run(rgui.WindowConfig{Title: "rgui layout"}, view); err != nil {
		log.Fatal(err)
	}
}
