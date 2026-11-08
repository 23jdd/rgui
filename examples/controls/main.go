package main

import (
	"fmt"
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"rgui"
)

func view() rgui.Node {
	status := rgui.NewSignal("Ready")
	return rgui.DockPanel(true,
		rgui.Docked{Edge: rgui.DockTop, Child: rgui.Menu(
			rgui.MenuGroup{Label: "File", Items: []rgui.MenuItem{
				{Label: "Reset status", Action: func() { status.Set("Ready") }},
			}},
		)},
		rgui.Docked{Edge: rgui.DockBottom, Child: rgui.StatusBar(rgui.Text(status.Get))},
		rgui.Docked{Child: rgui.TabControl([]rgui.TabItem{
			{Header: "Inputs", Content: inputs()},
			{Header: "Data", Content: data()},
			{Header: "Documents", Content: documents()},
			{Header: "Ink", Content: ink()},
		}, nil)},
	)
}

func inputs() rgui.Node {
	name := rgui.NewSignal("")
	textBox := rgui.TextBox(name)
	value := rgui.NewSignal(35.0)
	mode := rgui.NewSignal("one")
	date := rgui.NewSignal(time.Time{})
	fields := rgui.StackPanel(rgui.Vertical, 8,
		rgui.Label("_Name", textBox), textBox,
		rgui.Text(func() string { return "Name: " + name.Get() }),
		rgui.Label("Password", nil), rgui.PasswordBox(nil),
		rgui.Separator(rgui.Horizontal, 1, rl.Gray),
		rgui.Text(func() string { return fmt.Sprintf("Value %.0f", value.Get()) }),
		rgui.Slider(value, 0, 100, 5), rgui.ProgressBar(value, 0, 100),
		rgui.CheckBox("Enable option", nil),
		rgui.StackPanel(rgui.Horizontal, 8, rgui.RadioButton("One", "one", mode), rgui.RadioButton("Two", "two", mode)),
		rgui.ComboBox([]string{"First", "Second", "Third"}, nil),
		rgui.DatePicker(date),
	)
	return rgui.ScrollViewer(rgui.GroupBox("Signal-backed controls", fields), nil, false, true)
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
	if err := rgui.Run(rgui.WindowConfig{Title: "rgui control gallery", Width: 840, Height: 620, RetainedRendering: true}, view); err != nil {
		log.Fatal(err)
	}
}
