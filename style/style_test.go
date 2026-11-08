package style

import (
	"image/color"
	"reflect"
	"testing"
)

func parse(t *testing.T, css string) *Sheet {
	t.Helper()
	sheet, err := ParseCSS(css)
	if err != nil {
		t.Fatal(err)
	}
	return sheet
}

func TestStylesheetRoundTrip(t *testing.T) {
	sheet := parse(t, `/* theme */ .accent, button:hover { COLOR: var(--ink); padding: 1px 2px; }`)
	rules := sheet.Rules()
	structured, err := NewSheet(rules)
	if err != nil {
		t.Fatal(err)
	}
	if got := parse(t, structured.CSS()).Rules(); !reflect.DeepEqual(got, rules) {
		t.Fatalf("round trip = %#v, want %#v", got, rules)
	}
	rules[0].Declarations[0].Value = "red"
	if sheet.Rules()[0].Declarations[0].Value != "var(--ink)" || len(sheet.Rules()) != 2 {
		t.Fatal("returned rules must be independent and comma selectors must compile separately")
	}
	for _, invalid := range []string{`button > {}`, `button:not(.primary) {}`, `button { color }`, `/* unclosed`} {
		if _, err := ParseCSS(invalid); err == nil {
			t.Errorf("accepted %q", invalid)
		}
	}
}

func TestCascadeAndVariables(t *testing.T) {
	base := parse(t, `
:root { --ink: #123; --a: var(--b); --b: var(--a); color: var(--ink); }
#save { border-color: red; }
`)
	author := parse(t, `
.panel button { border-color: blue; padding: 1px 2px; }
.panel > button { border-color: black; }
button.primary:hover { border-color: var(--missing, var(--ink)); }
button.primary:hover { border-color: var(--a, #456); }
button { background: rgba(0, 20, 40, .5); }
button { padding-left: 18px; }
.panel button { padding-right: 5px; }
`)
	window := &Target{Type: "window"}
	panel := &Target{Type: "groupbox", Classes: []string{"panel"}, Parent: window}
	spacer := &Target{Type: "stackpanel", Parent: panel}
	button := &Target{Type: "Button", ID: "save", Classes: []string{"primary"}, State: Hover, Parent: spacer}
	values := Resolve(button, base, author)
	for property, want := range map[string]string{
		"color": "#123", "border-color": "#456", "padding-left": "2px", "padding-right": "5px", "padding-top": "1px",
	} {
		if got := values.Get(property); got != want {
			t.Errorf("%s = %q, want %q", property, got, want)
		}
	}
	if got := values.Color("background-color", color.RGBA{}); got != (color.RGBA{G: 20, B: 40, A: 128}) {
		t.Errorf("background = %+v", got)
	}
	button.State = 0
	if got := Resolve(button, base, author).Get("border-color"); got != "blue" {
		t.Errorf("descendant border = %q", got)
	}
	button.Parent = panel
	if got := Resolve(button, base, author).Get("border-color"); got != "black" {
		t.Errorf("direct child border = %q", got)
	}
}

func TestLengthAndBorderConversions(t *testing.T) {
	sheet := parse(t, `:root { color: #abc; --spacing: 4px 8px 12px; }
button { border: 2px solid currentColor; margin: var(--spacing); margin-left: 5px; }`)
	values := Resolve(&Target{Type: "button", Parent: &Target{Type: "window"}}, sheet)
	if got := values.Color("border-color", color.RGBA{}); got != (color.RGBA{R: 170, G: 187, B: 204, A: 255}) {
		t.Fatalf("border-color = %+v", got)
	}
	if values.Length("border-width", -1) != 2 || values.Length("margin-right", -1) != 8 || values.Length("margin-left", -1) != 5 {
		t.Fatalf("length conversion: %#v", values)
	}
	if _, ok := Length("20%"); ok {
		t.Fatal("unexpected support for percentage lengths")
	}
}
