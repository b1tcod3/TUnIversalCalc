package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// logStyleAndPrefix devuelve el estilo y el prefijo de una entrada del panel
// de salida.
func logStyleAndPrefix(level LogLevel) (lipgloss.Style, string) {
	switch level {
	case LevelSuccess:
		return levelSuccessStyle, "✔ "
	case LevelError:
		return levelErrorStyle, "✗ "
	case LevelStep:
		return levelStepStyle, "⚙ "
	default:
		return levelInfoStyle, "ℹ "
	}
}

// renderLogEntry ajusta una entrada al ancho disponible. El viewport de
// Bubbles no envuelve texto: recorta cada línea horizontalmente. Por eso el
// contenido se envuelve aquí, dejando el prefijo del nivel en la primera
// línea y una sangría equivalente en las continuaciones.
func renderLogEntry(level LogLevel, content string, width int) string {
	style, prefix := logStyleAndPrefix(level)
	prefixWidth := visualWidth(prefix)
	if width < 1 {
		width = 1
	}
	wrapWidth := max(width-prefixWidth, 1)
	indent := strings.Repeat(" ", prefixWidth)

	var b strings.Builder
	first := true
	for logicalLine := range strings.SplitSeq(content, "\n") {
		for _, part := range wrapLogLine(logicalLine, wrapWidth) {
			if !first {
				b.WriteByte('\n')
			}
			if first {
				b.WriteString(style.Render(prefix + part))
				first = false
				continue
			}
			b.WriteString(style.Render(indent + part))
		}
	}
	return b.String()
}

// wrapLogLine envuelve una línea y garantiza que ningún trozo supere width
// columnas. ansi.Wrap respeta palabras y ANSI, pero puede dejar espacios
// finales o una última palabra que sobresale del límite; el troceado final
// evita que el viewport recorte el texto.
func wrapLogLine(line string, width int) []string {
	line = strings.TrimRight(line, " \t")
	if line == "" {
		return []string{""}
	}
	if width < 1 {
		return []string{line}
	}
	var out []string
	for part := range strings.SplitSeq(ansi.Wrap(line, width, " +-*/="), "\n") {
		part = strings.TrimRight(part, " \t")
		out = append(out, hardWrapLogLine(part, width)...)
	}
	return out
}

func hardWrapLogLine(line string, width int) []string {
	line = strings.TrimRight(line, " \t")
	if line == "" {
		return []string{""}
	}
	var out []string
	for visualWidth(line) > width {
		chunk := ansi.Cut(line, 0, width)
		chunkWidth := visualWidth(chunk)
		if chunkWidth == 0 {
			break
		}
		out = append(out, strings.TrimRight(chunk, " \t"))
		line = ansi.Cut(line, chunkWidth, visualWidth(line))
	}
	if line != "" {
		out = append(out, strings.TrimRight(line, " \t"))
	}
	return out
}
