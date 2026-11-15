package ui

import (
	"strings"
	"testing"
)

func TestPreviewLagrange(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/lagrange t1")
	m.cmdInput.SetCursor(len([]rune("/lagrange t1")))

	if cmd := m.schedulePreview(); cmd == nil {
		t.Fatal("schedulePreview debe devolver un tick")
	}
	seq := m.preview.seq
	cmd := m.handlePreviewTick(previewTickMsg{Seq: seq})
	if cmd == nil {
		t.Fatal("el tick debe lanzar la evaluación")
	}
	msg, ok := cmd().(previewResultMsg)
	if !ok || msg.Err != nil {
		t.Fatalf("resultado de preview inesperado: %+v", msg)
	}
	m.applyPreviewResult(msg)
	if !strings.Contains(m.preview.text, "P(x) = x + 1") {
		t.Fatalf("preview = %q", m.preview.text)
	}
	for _, size := range [][2]int{{100, 28}, {80, 24}, {60, 20}} {
		m.width, m.height = size[0], size[1]
		m.updateLayout()
		checkView(t, &m, size, "preview")
	}
}

func TestPreviewIgnoraSecuenciaVieja(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.preview.seq = 5
	if cmd := m.handlePreviewTick(previewTickMsg{Seq: 4}); cmd != nil {
		t.Fatal("un tick viejo no debe lanzar evaluación")
	}
	m.applyPreviewResult(previewResultMsg{Seq: 4, Steps: []string{"viejo"}})
	if m.preview.text != "" {
		t.Fatalf("preview viejo aplicado: %q", m.preview.text)
	}
}

func TestPreviewSeLimpiaAlEjecutar(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/lagrange t1")
	m.cmdInput.SetCursor(len([]rune("/lagrange t1")))
	m.preview.text = "temporal"
	m.handleEnter()
	if m.preview.text != "" {
		t.Fatalf("preview no se limpió: %q", m.preview.text)
	}
}

func TestPreviewErrorSeMuestra(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.preview.err = testError("no soportado")
	view := m.View()
	if !strings.Contains(view, "no soportado") {
		t.Fatal("el error de preview debe aparecer en la vista")
	}
}

type testError string

func (e testError) Error() string { return string(e) }
