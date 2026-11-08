package rgui

import "testing"

func TestAccessTextHandlesEscapes(t *testing.T) {
	text, index := parseAccessText("a__b _Name_")
	if string(text) != "a_b Name_" || index != 4 {
		t.Fatalf("text=%q index=%d", text, index)
	}
}
