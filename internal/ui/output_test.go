package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func flattenLogText(s string) string {
	return strings.Join(strings.Fields(ansi.Strip(s)), "")
}

func assertLogWidth(t *testing.T, rendered string, width int) {
	t.Helper()
	for i, line := range strings.Split(ansi.Strip(rendered), "\n") {
		if w := visualWidth(line); w > width {
			t.Errorf("línea %d mide %d columnas (> %d): %q", i+1, w, width, line)
		}
	}
}

func TestRenderLogEntryEnvuelveSinPerderTexto(t *testing.T) {
	content := "Factorización del denominador: (x - 2)*(x + 2)^2 con " +
		strings.Repeat("segmento-largo ", 12) + "final"
	const width = 24

	rendered := renderLogEntry(LevelStep, content, width)
	assertLogWidth(t, rendered, width)

	if got := flattenLogText(rendered); !strings.Contains(got, flattenLogText(content)) {
		t.Fatalf("se perdió texto al envolver:\n%s", ansi.Strip(rendered))
	}
}

func TestRenderLogEntryParteTokenLargo(t *testing.T) {
	content := "inicio " + strings.Repeat("x", 80) + " final"
	const width = 20

	rendered := renderLogEntry(LevelStep, content, width)
	assertLogWidth(t, rendered, width)

	if got := flattenLogText(rendered); !strings.Contains(got, flattenLogText(content)) {
		t.Fatalf("se perdió texto de un token largo:\n%s", ansi.Strip(rendered))
	}
}

func TestRenderLogEntryCaracteresAnchos(t *testing.T) {
	content := "resultado: 漢字😀 · √2 · " + strings.Repeat("αβγ ", 12) + "fin"
	const width = 22

	rendered := renderLogEntry(LevelInfo, content, width)
	assertLogWidth(t, rendered, width)

	if got := flattenLogText(rendered); !strings.Contains(got, flattenLogText(content)) {
		t.Fatalf("se perdió texto con caracteres anchos:\n%s", ansi.Strip(rendered))
	}
}

func TestOutputSeReenvuelveAlRedimensionar(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()

	content := "Paso largo: " + strings.Repeat("segmento ", 30) + "final"
	m.addLog(LevelStep, content)

	before := m.renderLog()
	assertLogWidth(t, before, m.output.Width)
	if !m.output.AtBottom() {
		t.Fatal("la salida debe quedar al final tras addLog")
	}

	m.width, m.height = 40, 16
	m.updateLayout()

	after := m.renderLog()
	assertLogWidth(t, after, m.output.Width)
	if !m.output.AtBottom() {
		t.Fatal("el reajuste debe conservar la salida al final")
	}
	if flattenLogText(before) != flattenLogText(after) {
		t.Fatalf("el reajuste perdió texto:\nantes: %s\ndespués: %s",
			ansi.Strip(before), ansi.Strip(after))
	}
	if strings.Count(after, "\n") < strings.Count(before, "\n") {
		t.Fatal("al reducir el ancho deberían aparecer más líneas")
	}
}

func TestPlainOutputMantieneElTextoOriginal(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 16
	m.updateLayout()

	content := strings.Repeat("texto completo ", 20)
	m.addLog(LevelStep, content)

	if got := m.plainOutput(); !strings.Contains(got, content) {
		t.Fatalf("plainOutput perdió el texto original: %q", got)
	}
}
