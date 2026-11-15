package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// visualWidth mide el ancho real en celdas de terminal, ignorando códigos
// ANSI y teniendo en cuenta caracteres anchos (emoji, CJK) y combinados.
func visualWidth(s string) int {
	return ansi.StringWidth(s)
}

// truncateVisual recorta una cadena al ancho visual indicado y añade "…"
// cuando hay recorte. Es segura con secuencias ANSI.
func truncateVisual(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if visualWidth(s) <= maxWidth {
		return s
	}
	return ansi.Truncate(s, maxWidth, "…")
}

// padOrTrimVisual ajusta una cadena a un ancho exacto: la rellena con
// espacios si es más corta y la recorta sin elipsis si es más larga.
func padOrTrimVisual(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := visualWidth(s)
	if w == width {
		return s
	}
	if w > width {
		return ansi.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-w)
}

// clampWidth es el helper histórico de view.go: trunca una cadena plana o
// estilizada para que quepa en maxW columnas, añadiendo "…" al recortar.
func clampWidth(s string, maxW int) string {
	return truncateVisual(s, maxW)
}

// clampLines trunca cada línea de un bloque al ancho visual indicado. Evita
// que Lipgloss envuelva texto dentro de un panel con borde.
func clampLines(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = truncateVisual(l, maxW)
	}
	return strings.Join(lines, "\n")
}

// clipLines recorta un bloque a maxLines líneas. Se usa antes de renderizar
// paneles con Height(): Lipgloss no trunca el contenido que excede el alto.
func clipLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}
