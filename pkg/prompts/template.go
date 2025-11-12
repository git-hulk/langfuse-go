package prompts

import (
	"fmt"
	"strings"
)

const (
	openingDelimiter = "{{"
	closingDelimiter = "}}"
)

// templateCompiler mirrors the behavior of the TemplateParser in the Python SDK.
// It parses the template sequentially and replaces {{variable}} tokens only when
// a matching entry is provided in the variables map.
type templateCompiler struct {
	template string
}

func newTemplateCompiler(template string) templateCompiler {
	return templateCompiler{template: template}
}

// compile renders the template with the provided variables:
//   - When a placeholder has a matching key, it is replaced with fmt.Sprint(value) (nil -> "").
//   - When a key is missing, the placeholder remains untouched in the result.
//   - Whitespace around the placeholder name is ignored.
func (t templateCompiler) compile(variables map[string]any) string {
	if len(variables) == 0 {
		return t.template
	}

	var builder strings.Builder
	cursor := 0

	for cursor < len(t.template) {
		openIdx := strings.Index(t.template[cursor:], openingDelimiter)
		if openIdx == -1 {
			builder.WriteString(t.template[cursor:])
			break
		}
		openIdx += cursor

		closeIdx := strings.Index(t.template[openIdx+len(openingDelimiter):], closingDelimiter)
		if closeIdx == -1 {
			builder.WriteString(t.template[cursor:])
			break
		}
		closeIdx += openIdx + len(openingDelimiter)

		builder.WriteString(t.template[cursor:openIdx])

		rawName := t.template[openIdx+len(openingDelimiter) : closeIdx]
		varName := strings.TrimSpace(rawName)
		fullPlaceholder := t.template[openIdx : closeIdx+len(closingDelimiter)]

		if value, ok := variables[varName]; ok {
			if value == nil {
				builder.WriteString("")
			} else {
				builder.WriteString(fmt.Sprint(value))
			}
		} else {
			builder.WriteString(fullPlaceholder)
		}

		cursor = closeIdx + len(closingDelimiter)
	}

	return builder.String()
}
