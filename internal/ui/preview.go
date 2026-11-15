package ui

import (
	"strings"
	"time"
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

// previewDelay es el debounce de la evaluación en vivo del panel 2.
var previewDelay = 350 * time.Millisecond

// previewState guarda el último resultado de la previsualización.
type previewState struct {
	seq  int
	text string
	err  error
}

type previewTickMsg struct{ Seq int }

type previewResultMsg struct {
	Seq   int
	Steps []string
	Err   error
}

// schedulePreview programa una evaluación del comando actual tras el
// debounce. Cada llamada invalida la anterior.
func (m *Model) schedulePreview() tea.Cmd {
	if m.focus != FocusCmd || !m.declValid {
		m.clearPreview()
		return nil
	}
	m.preview.seq++
	seq := m.preview.seq
	return tea.Tick(previewDelay, func(time.Time) tea.Msg {
		return previewTickMsg{Seq: seq}
	})
}

// handlePreviewTick resuelve el comando y lanza el motor en segundo plano.
func (m *Model) handlePreviewTick(msg previewTickMsg) tea.Cmd {
	if msg.Seq != m.preview.seq {
		return nil
	}
	if m.focus != FocusCmd || !m.declValid {
		m.clearPreview()
		return nil
	}
	val := m.cmdInput.Value()
	decl, varName, ok := m.previewTarget(val)
	if !ok {
		m.preview.text = ""
		m.preview.err = nil
		return nil
	}
	return previewEvalCmd(decl, val, varName, msg.Seq)
}

func previewEvalCmd(decl engine.Decl, cmd, varName string, seq int) tea.Cmd {
	return func() tea.Msg {
		steps, err := engine.Evaluate(decl, cmd, varName)
		return previewResultMsg{Seq: seq, Steps: steps, Err: err}
	}
}

func (m *Model) applyPreviewResult(msg previewResultMsg) {
	if msg.Seq != m.preview.seq {
		return
	}
	if msg.Err != nil {
		m.preview.text = ""
		m.preview.err = msg.Err
		return
	}
	m.preview.err = nil
	m.preview.text = ""
	if len(msg.Steps) > 0 {
		m.preview.text = msg.Steps[len(msg.Steps)-1]
	}
}

// clearPreview invalida cualquier evaluación en vuelo y limpia el texto.
func (m *Model) clearPreview() {
	m.preview.seq++
	m.preview.text = ""
	m.preview.err = nil
}

// previewTarget es como resolveTarget pero sin escribir en el log: se usa
// mientras el usuario escribe y todavía puede haber un comando incompleto.
func (m *Model) previewTarget(cmd string) (engine.Decl, string, bool) {
	toks := splitCommandTokens(normalizeCommandArgs(cmd))
	if d, ok := m.literalIntegerDecl(cmd); ok {
		return d, "x", true
	}
	decl, varName, err := m.resolveTargetTokens(toks)
	if err != nil {
		return engine.Decl{}, "", false
	}
	return decl, varName, true
}

// previewView renderiza la última línea del resultado en vivo.
func (m *Model) previewView() string {
	w := max(m.leftWidth()-panelBorder-inputPadX-2, 10)
	if m.preview.err != nil {
		return levelErrorStyle.Render(clampWidth("✗ "+m.preview.err.Error(), w))
	}
	if m.preview.text == "" {
		return ""
	}
	return levelStepStyle.Render(clampWidth("↳ "+m.preview.text, w))
}

// normalizeCommandArgs limpia el comando para separar operación y argumentos.
func normalizeCommandArgs(cmd string) string {
	return strings.NewReplacer("(", "", ")", "", ",", "").Replace(cmd)
}

func splitCommandTokens(s string) []string {
	return strings.Fields(s)
}
