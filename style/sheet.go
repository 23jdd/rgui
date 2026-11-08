// Package style converts a supported subset of CSS to rgui styles.
package style

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type State uint8

const (
	Hover State = 1 << iota
	Active
	Focus
	Disabled
	Checked
	Selected
)

// Target is the element and its ancestry as seen by the selector matcher.
type Target struct {
	Type, ID string
	Classes  []string
	State    State
	Parent   *Target
}

type Declaration struct{ Property, Value string }

type Rule struct {
	Selector     string
	Declarations []Declaration
}

// Sheet is immutable after parsing. Rules are returned as independent copies.
type Sheet struct{ rules []compiledRule }

type compound struct {
	kind, id string
	classes  []string
	states   State
	root     bool
}

type compiledRule struct {
	rule     Rule
	selector []compound
	direct   []bool
	score    [3]int
}

// ParseCSS accepts type, class, ID, descendant and direct-child selectors,
// state pseudo-classes, and declarations. Unsupported selectors return errors.
func ParseCSS(css string) (*Sheet, error) {
	clean, err := stripComments(css)
	if err != nil {
		return nil, err
	}
	sheet := &Sheet{}
	for strings.TrimSpace(clean) != "" {
		selector, rest, ok := strings.Cut(clean, "{")
		if !ok {
			return nil, fmt.Errorf("CSS: expected '{' after %q", strings.TrimSpace(clean))
		}
		body, remaining, ok := strings.Cut(rest, "}")
		if !ok || strings.Contains(body, "{") {
			return nil, fmt.Errorf("CSS: invalid block for %q", strings.TrimSpace(selector))
		}
		declarations, err := parseDeclarations(body)
		if err != nil {
			return nil, err
		}
		for _, part := range strings.Split(selector, ",") {
			rule, err := compile(strings.TrimSpace(part), declarations)
			if err != nil {
				return nil, err
			}
			sheet.rules = append(sheet.rules, rule)
		}
		clean = remaining
	}
	return sheet, nil
}

// NewSheet validates structured rules using the same syntax as ParseCSS.
func NewSheet(rules []Rule) (*Sheet, error) {
	return ParseCSS(formatRules(rules))
}

func (s *Sheet) Rules() []Rule {
	if s == nil {
		return nil
	}
	rules := make([]Rule, len(s.rules))
	for i, rule := range s.rules {
		rules[i] = Rule{Selector: rule.rule.Selector, Declarations: append([]Declaration(nil), rule.rule.Declarations...)}
	}
	return rules
}

// CSS serializes the sheet in canonical CSS syntax.
func (s *Sheet) CSS() string { return formatRules(s.Rules()) }

func formatRules(rules []Rule) string {
	var out strings.Builder
	for _, rule := range rules {
		fmt.Fprintf(&out, "%s {\n", rule.Selector)
		for _, declaration := range rule.Declarations {
			fmt.Fprintf(&out, "  %s: %s;\n", declaration.Property, declaration.Value)
		}
		out.WriteString("}\n")
	}
	return out.String()
}

func stripComments(input string) (string, error) {
	var out strings.Builder
	for {
		before, after, ok := strings.Cut(input, "/*")
		out.WriteString(before)
		if !ok {
			return out.String(), nil
		}
		_, rest, ok := strings.Cut(after, "*/")
		if !ok {
			return "", fmt.Errorf("CSS: unterminated comment")
		}
		out.WriteByte(' ')
		input = rest
	}
}

func parseDeclarations(input string) ([]Declaration, error) {
	var result []Declaration
	for _, segment := range strings.Split(input, ";") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		property, value, ok := strings.Cut(segment, ":")
		property, value = strings.TrimSpace(property), strings.TrimSpace(value)
		if !ok || !identifier(property) || value == "" || strings.Contains(value, "!") {
			return nil, fmt.Errorf("CSS: invalid declaration %q", segment)
		}
		if !strings.HasPrefix(property, "--") {
			property = strings.ToLower(property)
		}
		result = append(result, Declaration{Property: property, Value: value})
	}
	return result, nil
}

func identifier(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool { return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_') }) < 0
}

func compile(selector string, declarations []Declaration) (compiledRule, error) {
	rule := compiledRule{rule: Rule{Selector: selector, Declarations: declarations}}
	tokens := strings.Fields(strings.ReplaceAll(selector, ">", " > "))
	for _, token := range tokens {
		if token == ">" {
			if len(rule.direct) == 0 || rule.direct[len(rule.direct)-1] {
				return rule, fmt.Errorf("CSS: invalid selector %q", selector)
			}
			rule.direct[len(rule.direct)-1] = true
			continue
		}
		item, score, err := parseCompound(token)
		if err != nil {
			return rule, err
		}
		for i := range score {
			rule.score[i] += score[i]
		}
		rule.selector = append(rule.selector, item)
		rule.direct = append(rule.direct, false)
	}
	if len(rule.selector) == 0 || rule.direct[len(rule.direct)-1] {
		return rule, fmt.Errorf("CSS: invalid selector %q", selector)
	}
	return rule, nil
}

func parseCompound(token string) (compound, [3]int, error) {
	var result compound
	var score [3]int
	rest := token
	for rest != "" {
		prefix := byte(0)
		if strings.ContainsRune(".#:", rune(rest[0])) {
			prefix, rest = rest[0], rest[1:]
		}
		end := strings.IndexAny(rest, ".#:")
		if end < 0 {
			end = len(rest)
		}
		value := rest[:end]
		rest = rest[end:]
		if !identifier(value) && !(prefix == 0 && value == "*") {
			return result, score, fmt.Errorf("CSS: unsupported selector %q", token)
		}
		switch prefix {
		case '.':
			result.classes = append(result.classes, value)
			score[1]++
		case '#':
			if result.id != "" {
				return result, score, fmt.Errorf("CSS: duplicate ID in %q", token)
			}
			result.id = value
			score[0]++
		case ':':
			switch strings.ToLower(value) {
			case "root":
				result.root = true
			case "hover":
				result.states |= Hover
			case "active":
				result.states |= Active
			case "focus", "focus-visible":
				result.states |= Focus
			case "disabled":
				result.states |= Disabled
			case "checked":
				result.states |= Checked
			case "selected":
				result.states |= Selected
			default:
				return result, score, fmt.Errorf("CSS: unsupported pseudo-class :%s", value)
			}
			score[1]++
		default:
			if value != "*" {
				result.kind = strings.ToLower(value)
				score[2]++
			}
		}
	}
	return result, score, nil
}

func (c compound) matches(target *Target) bool {
	if target == nil || c.root && target.Parent != nil || c.kind != "" && !strings.EqualFold(c.kind, target.Type) ||
		c.id != "" && c.id != target.ID || target.State&c.states != c.states {
		return false
	}
	for _, class := range c.classes {
		found := false
		for _, candidate := range target.Classes {
			if class == candidate {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (r compiledRule) matches(index int, target *Target) bool {
	if !r.selector[index].matches(target) {
		return false
	}
	if index == 0 {
		return true
	}
	if r.direct[index-1] {
		return r.matches(index-1, target.Parent)
	}
	for ancestor := target.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if r.matches(index-1, ancestor) {
			return true
		}
	}
	return false
}

func stronger(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return true
}

type winner struct {
	value string
	score [3]int
}

func (s *Sheet) matching(target *Target) []compiledRule {
	var out []compiledRule
	if s == nil {
		return out
	}
	for _, rule := range s.rules {
		if rule.matches(len(rule.selector)-1, target) {
			out = append(out, rule)
		}
	}
	return out
}

// Length converts a CSS pixel length to logical pixels. Zero may omit its unit.
func Length(value string) (float32, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	pixels := strings.HasSuffix(value, "px")
	value = strings.TrimSpace(strings.TrimSuffix(value, "px"))
	number, err := strconv.ParseFloat(value, 32)
	return float32(number), err == nil && !math.IsNaN(number) && !math.IsInf(number, 0) && (pixels || number == 0)
}
