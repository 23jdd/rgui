package style

import (
	"image/color"
	"math"
	"strconv"
	"strings"
)

// Values contains resolved declarations, including inherited custom properties.
type Values map[string]string

// Resolve applies sheets in order. A later sheet has higher priority than an
// earlier sheet. Specificity and source order apply within each sheet.
func Resolve(target *Target, sheets ...*Sheet) Values {
	if target == nil {
		return Values{}
	}
	values := make(Values)
	if target.Parent != nil {
		parent := Resolve(target.Parent, sheets...)
		for key := range parent {
			if strings.HasPrefix(key, "--") || key == "color" || key == "font-size" || key == "letter-spacing" {
				values[key] = parent.Get(key)
			}
		}
	}
	matched := make([][]compiledRule, len(sheets))
	for i, sheet := range sheets {
		matched[i] = sheet.matching(target)
		values.apply(matched[i], true)
	}
	for _, rules := range matched {
		values.apply(rules, false)
	}
	return values
}

func (v Values) apply(rules []compiledRule, custom bool) {
	winners := make(map[string]winner)
	for _, rule := range rules {
		for _, decl := range rule.rule.Declarations {
			if strings.HasPrefix(decl.Property, "--") != custom {
				continue
			}
			for _, item := range v.expandDeclaration(decl) {
				if old, ok := winners[item.Property]; !ok || stronger(rule.score, old.score) {
					winners[item.Property] = winner{value: item.Value, score: rule.score}
				}
			}
		}
	}
	for key, value := range winners {
		v[key] = value.value
	}
}

// Get resolves CSS var() references, including fallback values.
func (v Values) Get(property string) string { return v.resolve(v[property]) }

func (v Values) resolve(value string) string {
	if !strings.Contains(value, "var(") {
		return value
	}
	return v.expand(value, make(map[string]bool))
}

func (v Values) expandDeclaration(d Declaration) []Declaration {
	switch d.Property {
	case "background":
		value := v.resolve(d.Value)
		if _, ok := Color(value); ok || strings.EqualFold(value, "currentColor") {
			return []Declaration{d, {Property: "background-color", Value: value}}
		}
	case "padding", "margin":
		parts := strings.Fields(v.resolve(d.Value))
		if len(parts) == 0 || len(parts) > 4 {
			break
		}
		for _, part := range parts {
			if _, ok := Length(part); !ok {
				return []Declaration{d}
			}
		}
		top, right, bottom, left := parts[0], parts[0], parts[0], parts[0]
		if len(parts) >= 2 {
			right, left = parts[1], parts[1]
		}
		if len(parts) >= 3 {
			bottom = parts[2]
		}
		if len(parts) == 4 {
			left = parts[3]
		}
		return []Declaration{d, {d.Property + "-top", top}, {d.Property + "-right", right},
			{d.Property + "-bottom", bottom}, {d.Property + "-left", left}}
	case "border":
		// This subset accepts none or a pixel width, solid and a color.
		value := v.resolve(d.Value)
		if value == "none" {
			return []Declaration{d, {"border-style", "none"}, {"border-width", "0"}}
		}
		width, rest, ok := strings.Cut(value, " ")
		if _, valid := Length(width); ok && valid {
			kind, tint, _ := strings.Cut(strings.TrimSpace(rest), " ")
			tint = strings.TrimSpace(tint)
			if tint == "" {
				tint = "currentColor"
			}
			if _, valid := Color(tint); kind == "solid" && (valid || strings.EqualFold(tint, "currentColor")) {
				return []Declaration{d, {"border-style", "solid"}, {"border-width", width}, {"border-color", tint}}
			}
		}
	}
	return []Declaration{d}
}

func (v Values) expand(value string, seen map[string]bool) string {
	for {
		i := strings.Index(value, "var(")
		if i < 0 {
			return value
		}
		depth, end := 1, i+4
		for end < len(value) && depth > 0 {
			switch value[end] {
			case '(':
				depth++
			case ')':
				depth--
			}
			end++
		}
		if depth != 0 {
			return ""
		}
		key, fallback, _ := strings.Cut(value[i+4:end-1], ",")
		key = strings.TrimSpace(key)
		replacement := ""
		if strings.HasPrefix(key, "--") && !seen[key] {
			seen[key] = true
			replacement = v.expand(v[key], seen)
			delete(seen, key)
		}
		if replacement == "" {
			replacement = v.expand(strings.TrimSpace(fallback), seen)
		}
		if replacement == "" {
			return ""
		}
		value = value[:i] + replacement + value[end:]
	}
}

func (v Values) Length(property string, fallback float32) float32 {
	if value, ok := Length(v.Get(property)); ok {
		return value
	}
	return fallback
}

func (v Values) Color(property string, fallback color.RGBA) color.RGBA {
	value := v.Get(property)
	if property != "color" && strings.EqualFold(value, "currentColor") {
		value = v.Get("color")
	}
	if value, ok := Color(value); ok {
		return value
	}
	return fallback
}

var namedColors = map[string]color.RGBA{
	"transparent": {}, "black": {A: 255}, "white": {R: 255, G: 255, B: 255, A: 255},
	"red": {R: 255, A: 255}, "green": {G: 128, A: 255}, "blue": {B: 255, A: 255},
	"gray": {R: 128, G: 128, B: 128, A: 255}, "grey": {R: 128, G: 128, B: 128, A: 255},
	"rebeccapurple": {R: 102, G: 51, B: 153, A: 255},
}

// Color converts hex colors, rgb()/rgba(), and a small set of common names.
func Color(input string) (color.RGBA, bool) {
	input = strings.TrimSpace(strings.ToLower(input))
	if value, ok := namedColors[input]; ok {
		return value, true
	}
	if hex, ok := strings.CutPrefix(input, "#"); ok {
		if len(hex) == 3 || len(hex) == 4 {
			var expanded strings.Builder
			for _, digit := range hex {
				expanded.WriteRune(digit)
				expanded.WriteRune(digit)
			}
			hex = expanded.String()
		}
		if len(hex) == 6 {
			hex += "ff"
		}
		if len(hex) != 8 {
			return color.RGBA{}, false
		}
		value, err := strconv.ParseUint(hex, 16, 32)
		return color.RGBA{R: uint8(value >> 24), G: uint8(value >> 16), B: uint8(value >> 8), A: uint8(value)}, err == nil
	}
	name, body, ok := strings.Cut(input, "(")
	if !ok || !strings.HasSuffix(body, ")") || (name != "rgb" && name != "rgba") {
		return color.RGBA{}, false
	}
	fields := strings.Fields(strings.NewReplacer(",", " ", "/", " ").Replace(strings.TrimSuffix(body, ")")))
	if len(fields) < 3 || len(fields) > 4 {
		return color.RGBA{}, false
	}
	channels := [4]uint8{0, 0, 0, 255}
	for i, field := range fields {
		percent := strings.HasSuffix(field, "%")
		value, err := strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return color.RGBA{}, false
		}
		if percent {
			value = value / 100 * 255
		} else if i == 3 {
			value *= 255
		}
		channels[i] = uint8(math.Round(max(0, min(255, value))))
	}
	return color.RGBA{R: channels[0], G: channels[1], B: channels[2], A: channels[3]}, true
}
