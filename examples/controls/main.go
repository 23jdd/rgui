package main

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"rgui"
	"rgui/style"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed theme.css
var themeCSS string

func styled(node rgui.Node, classes ...string) rgui.Node {
	return rgui.WithStyle(node, rgui.StyleOptions{Classes: classes})
}

func view() rgui.Node {
	status := rgui.NewSignal("READY")
	menu := rgui.Menu(rgui.MenuGroup{Label: "FILE", Items: []rgui.MenuItem{
		{Label: "RESET STATUS", Action: func() { status.Set("READY") }},
	}})
	heading := rgui.WithLayout(rgui.StackPanel(rgui.Vertical, 4,
		styled(rgui.TextBlock(func() string { return "CONTROL GALLERY" }, rgui.Start), "title"),
		styled(rgui.TextBlock(func() string { return "RGUI / SYSTEM MENU" }, rgui.Start), "eyebrow"),
	), rgui.Box{VerticalAlignment: rgui.Center})
	masthead := styled(rgui.Grid([]rgui.Track{rgui.Auto()}, []rgui.Track{rgui.Fixed(80), rgui.Star(1), rgui.Auto()},
		rgui.Cell{Child: &consoleMark{}},
		rgui.Cell{Column: 1, Child: heading},
		rgui.Cell{Column: 2, Child: rgui.WithLayout(menu, rgui.Box{VerticalAlignment: rgui.Center})},
	), "masthead")
	return styled(rgui.DockPanel(true,
		rgui.Docked{Edge: rgui.DockTop, Child: masthead},
		rgui.Docked{Edge: rgui.DockBottom, Child: styled(rgui.DockPanel(true,
			rgui.Docked{Edge: rgui.DockRight, Child: styled(rgui.Text(func() string { return "TAB FOCUS / WHEEL SCROLL" }), "key-hints")},
			rgui.Docked{Child: rgui.StatusBar(styled(rgui.TextBlock(status.Get, rgui.Start), "ready"))},
		), "footer")},
		rgui.Docked{Child: rgui.TabControl([]rgui.TabItem{
			{Header: "INPUTS", Content: inputs()},
			{Header: "DATA", Content: data()},
			{Header: "DOCUMENTS", Content: documents()},
			{Header: "INK", Content: ink()},
		}, nil)},
	), "gallery")
}

func inputs() rgui.Node {
	name := rgui.NewSignal("")
	password := rgui.NewSignal("")
	textBox := rgui.WithStyle(rgui.TextBox(name), rgui.StyleOptions{ID: "name"})
	value := rgui.NewSignal(35.0)
	mode := rgui.NewSignal("one")
	date := rgui.NewSignal(time.Time{})
	fields := rgui.GroupBox("01 / PROFILE", rgui.StackPanel(rgui.Vertical, 10,
		rgui.StackPanel(rgui.Vertical, 4, rgui.Label("_Name", textBox), textBox),
		rgui.StackPanel(rgui.Vertical, 4, rgui.Label("Password", nil), rgui.PasswordBox(password)),
		rgui.DatePicker(date),
		styled(rgui.TextBlock(func() string { return "Name preview: " + name.Get() }, rgui.Start), "caption"),
		styled(rgui.Button("CLEAR PROFILE", func() {
			name.Set("")
			password.Set("")
			date.Set(time.Time{})
		}), "primary"),
	))
	preferences := styled(rgui.GroupBox("02 / PREFERENCES", rgui.StackPanel(rgui.Vertical, 10,
		rgui.CheckBox("Enable option", nil),
		rgui.StackPanel(rgui.Horizontal, 8, rgui.RadioButton("One", "one", mode), rgui.RadioButton("Two", "two", mode)),
		rgui.ComboBox([]string{"First", "Second", "Third"}, nil),
		styled(rgui.TextBlock(func() string { return fmt.Sprintf("Value %.0f", value.Get()) }, rgui.Start), "caption"),
		rgui.Slider(value, 0, 100, 5), rgui.ProgressBar(value, 0, 100),
	)), "side-panel")
	return rgui.ScrollViewer(rgui.Grid([]rgui.Track{rgui.Auto()}, []rgui.Track{rgui.Star(1), rgui.Star(1)},
		rgui.Cell{Child: fields}, rgui.Cell{Column: 1, Child: preferences},
	), nil, false, true)
}

func data() rgui.Node {
	columns := []rgui.GridViewColumn{{Header: "Name"}, {Header: "Value"}}
	return rgui.TabControl([]rgui.TabItem{
		{Header: "Table", Content: rgui.DataGrid(columns, [][]*rgui.Signal[string]{
			{rgui.NewSignal("Alpha"), rgui.NewSignal("12")},
			{rgui.NewSignal("Beta"), rgui.NewSignal("24")},
		})},
		{Header: "Tree", Content: rgui.TreeView([]rgui.TreeItem{
			{Key: "root", Label: "Parent", Children: []rgui.TreeItem{{Key: "a", Label: "Child A"}, {Key: "b", Label: "Child B"}}},
		}, nil)},
		{Header: "1,000 rows", Content: rgui.VirtualizingStackPanel(1000, 36, nil, func(i int) rgui.Node {
			return rgui.Text(func() string { return fmt.Sprintf("Row %d", i+1) })
		})},
	}, nil)
}

func documents() rgui.Node {
	return rgui.TabControl([]rgui.TabItem{
		{Header: "Reader", Content: rgui.FlowDocumentReader([]string{
			"Resize this window to see the text wrap and repaginate.",
			"Switch to continuous view to scroll the same paragraphs.",
		})},
		{Header: "Editor", Content: rgui.RichTextBox(rgui.NewSignal([]rgui.TextRun{
			{Text: "Rich text editor\n", Style: rgui.RichTextStyle{Bold: true}},
			{Text: "Select text and try Ctrl+B or Ctrl+U."},
		}))},
	}, nil)
}

func ink() rgui.Node {
	strokes := rgui.NewSignal([]rgui.InkStroke(nil))
	return rgui.DockPanel(true,
		rgui.Docked{Edge: rgui.DockTop, Child: rgui.ToolBar(rgui.Button("Clear", func() { strokes.Set(nil) }))},
		rgui.Docked{Child: rgui.InkCanvas(strokes, rl.SkyBlue, 3)},
	)
}

func main() {
	config, err := galleryConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := rgui.Run(config, view); err != nil {
		log.Fatal(err)
	}
}

func galleryConfig() (rgui.WindowConfig, error) {
	sheet, err := style.ParseCSS(themeCSS)
	if err != nil {
		return rgui.WindowConfig{}, err
	}
	return rgui.WindowConfig{Title: "rgui control gallery", Width: 840, Height: 560, MinWidth: 680, MinHeight: 450,
		RetainedRendering: true, StyleSheet: sheet}, nil
}
