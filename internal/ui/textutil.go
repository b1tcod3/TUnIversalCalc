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
