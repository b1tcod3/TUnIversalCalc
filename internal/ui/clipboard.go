package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

// writeClipboard permite sustituir el portapapeles del sistema en tests.
var writeClipboard = clipboard.WriteAll

// copyOutput copia al portapapeles la salida completa sin códigos ANSI.
func (m *Model) copyOutput() {
	text := strings.TrimSpace(m.plainOutput())
	if text == "" {
		m.addLog(LevelInfo, "No hay salida para copiar")
		return
	}
	if err := writeClipboard(text); err != nil {
		m.addLog(LevelError, "No se pudo copiar al portapapeles: "+err.Error())
		return
	}
	lines := strings.Count(text, "\n") + 1
	m.addLog(LevelSuccess, fmt.Sprintf("Salida copiada al portapapeles (%d líneas)", lines))
}

// plainOutput devuelve el log sin estilos, listo para copiar o exportar.
func (m *Model) plainOutput() string {
	var b strings.Builder
	for i, e := range m.outputLog {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(e.Content)
	}
	return b.String()
}
