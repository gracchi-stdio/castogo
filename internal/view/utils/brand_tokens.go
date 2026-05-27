package utils

import (
	"html"
	"strings"
)

// RenderBrandTokens escapes input and renders color tokens like [red:TEXT] or a red arrow [\u2B05].
func RenderBrandTokens(value string) string {
	if value == "" {
		return ""
	}

	colorClasses := map[string]string{
		"red":     "text-primary",
		"primary": "text-primary",
		"muted":   "text-muted-foreground",
		"accent":  "text-accent-foreground",
	}

	builder := strings.Builder{}
	for i := 0; i < len(value); {
		open := strings.IndexByte(value[i:], '[')
		if open == -1 {
			builder.WriteString(html.EscapeString(value[i:]))
			break
		}
		open += i
		if open > i {
			builder.WriteString(html.EscapeString(value[i:open]))
		}
		close := strings.IndexByte(value[open+1:], ']')
		if close == -1 {
			builder.WriteString(html.EscapeString(value[open:]))
			break
		}
		close += open + 1
		token := value[open+1 : close]
		if token == "\u2B05" {
			builder.WriteString("<span class=\"text-primary\">")
			builder.WriteRune('\u2B05')
			builder.WriteString("</span>")
			i = close + 1
			continue
		}
		if parts := strings.SplitN(token, ":", 2); len(parts) == 2 {
			color := strings.TrimSpace(parts[0])
			text := strings.TrimSpace(parts[1])
			if className, ok := colorClasses[color]; ok {
				builder.WriteString("<span class=\"")
				builder.WriteString(className)
				builder.WriteString("\">")
				builder.WriteString(html.EscapeString(text))
				builder.WriteString("</span>")
				i = close + 1
				continue
			}
		}
		builder.WriteString(html.EscapeString(value[open : close+1]))
		i = close + 1
	}

	return builder.String()
}
