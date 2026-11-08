package main

import (
	"fmt"
	"image/color"
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
	header := rgui.StackPanel(rgui.Vertical, 12, toolbar, ratio).
		WithLayout(rgui.Box{Margin: rgui.Insets{Bottom: 16}})
	sidebar := rgui.StackPanel(rgui.Vertical, 8,
		rgui.Text(func() string { return "Dock" }),
		rgui.Button("Reset", func() { count.Set(0) }),
	).WithLayout(rgui.Box{Width: 86, Margin: rgui.Insets{Right: 12}})
	var tags []rgui.Node
	for i := 1; i <= 6; i++ {
		tags = append(tags, rgui.Button(fmt.Sprintf("Item %d", i), func() { count.Set(i) }).
			WithDisabled(i == 6).
			WithLayout(rgui.Box{Width: 104, Height: 36}))
	}
	return rgui.DockPanel(true,
		rgui.Docked{Edge: rgui.DockTop, Child: header},
		rgui.Docked{Edge: rgui.DockLeft, Child: sidebar},
		rgui.Docked{Child: rgui.WrapPanel(rgui.Horizontal, 8, tags...)},
	)
}

func main() {
	config := rgui.WindowConfig{
		Title: "rgui layout", Width: 560, Height: 340,
		MinWidth: 260, MinHeight: 200, MaxWidth: 1000,
		Background: &color.RGBA{R: 25, G: 31, B: 45, A: 255},
	}
	if err := rgui.Run(config, view); err != nil {
		log.Fatal(err)
	}
}
