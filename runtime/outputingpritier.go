package runtime

import (
	"fmt"
	"sort"
	"strings"
)

// Pretty returns a human-friendly single-line representation of any RuntimeVal.
// It recursively formats arrays and maps, quotes strings, and stabilizes
// map key ordering for deterministic output.
func Pretty(v RuntimeVal) string {
	if v == nil {
		return "null"
	}
	switch t := v.(type) {
	case *NumberVal:
		return fmt.Sprintf("%v", t.Value)
	case *StringVal:
		// Quote strings for clarity
		return fmt.Sprintf("\"%s\"", t.Value)
	case *BooleanVal:
		return fmt.Sprintf("%v", t.Value)
	case Function:
		return "[function]"
	case *ArrayVal:
		parts := make([]string, len(t.Elements))
		for i, el := range t.Elements {
			parts[i] = Pretty(el)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *MapVal:
		keys := make([]string, 0, len(t.Properties))
		for k := range t.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = fmt.Sprintf("\"%s\": %s", k, Pretty(t.Properties[k]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return v.String()
	}
}

// PrettyMultiline returns a multi-line, indented representation suitable for
// deeply nested arrays and maps. Strings are quoted; map keys are sorted.
func PrettyMultiline(v RuntimeVal) string {
	return prettyML(v, 0)
}

func prettyML(v RuntimeVal, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	switch t := v.(type) {
	case nil:
		return indentStr + "null"
	case *NumberVal:
		return indentStr + fmt.Sprintf("%v", t.Value)
	case *StringVal:
		return indentStr + fmt.Sprintf("\"%s\"", t.Value)
	case *BooleanVal:
		return indentStr + fmt.Sprintf("%v", t.Value)
	case Function:
		return indentStr + "[function]"
	case *ArrayVal:
		if len(t.Elements) == 0 {
			return indentStr + "[]"
		}
		var b strings.Builder
		b.WriteString(indentStr)
		b.WriteString("[\n")
		for i, el := range t.Elements {
			b.WriteString(prettyML(el, indent+1))
			if i < len(t.Elements)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(indentStr)
		b.WriteString("]")
		return b.String()
	case *MapVal:
		if len(t.Properties) == 0 {
			return indentStr + "{}"
		}
		keys := make([]string, 0, len(t.Properties))
		for k := range t.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString(indentStr)
		b.WriteString("{\n")
		for i, k := range keys {
			b.WriteString(strings.Repeat("  ", indent+1))
			b.WriteString(fmt.Sprintf("\"%s\": ", k))
			val := t.Properties[k]
			// If nested composite, recurse on next line with its own indent
			switch val.(type) {
			case *ArrayVal, *MapVal:
				b.WriteString("\n")
				b.WriteString(prettyML(val, indent+2))
			default:
				b.WriteString(strings.TrimPrefix(prettyML(val, indent+1), strings.Repeat("  ", indent+1)))
			}
			if i < len(keys)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(indentStr)
		b.WriteString("}")
		return b.String()
	default:
		return indentStr + v.String()
	}
}

// Unescape converts simple escape sequences in plain strings to their
// corresponding characters. Use this when rendering string content to output.
func Unescape(s string) string {
	replacer := strings.NewReplacer("\\r\\n", "\r\n", "\\n", "\n", "\\t", "\t", "\\\\", "\\", "\\\"", "\"")
	return replacer.Replace(s)
}

// formatValue is kept for compatibility with existing code.
// It now delegates to Pretty so all outputs are consistent.
func formatValue(v RuntimeVal) string {
	return Pretty(v)
}
