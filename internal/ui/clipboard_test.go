package ui

import (
	"errors"
	"strings"
	"testing"
)

func TestCopyOutput(t *testing.T) {
	old := writeClipboard
	t.Cleanup(func() { writeClipboard = old })

	var copied string
	writeClipboard = func(s string) error {
		copied = s
		return nil
	}

	m := newFlowModel()
	m.outputLog = []LogEntry{
		{Level: LevelInfo, Content: "primera línea"},
		{Level: LevelSuccess, Content: "segunda línea"},
	}
	m.copyOutput()

	if copied != "primera línea\nsegunda línea" {
		t.Fatalf("texto copiado = %q", copied)
	}
	if strings.Contains(copied, "\x1b[") {
		t.Fatalf("el texto copiado contiene ANSI: %q", copied)
	}
	if !strings.Contains(m.plainOutput(), "Salida copiada al portapapeles (2 líneas)") {
		t.Fatalf("falta el aviso de copia: %q", m.plainOutput())
	}
}

func TestCopyOutputError(t *testing.T) {
	old := writeClipboard
	t.Cleanup(func() { writeClipboard = old })
	writeClipboard = func(string) error { return errors.New("sin portapapeles") }

	m := newFlowModel()
	m.copyOutput()
	if !strings.Contains(m.plainOutput(), "No se pudo copiar al portapapeles") {
		t.Fatalf("falta el error de copia: %q", m.plainOutput())
	}
}

func TestCopyOutputVacio(t *testing.T) {
	old := writeClipboard
	t.Cleanup(func() { writeClipboard = old })
	writeClipboard = func(string) error {
		t.Fatal("no debe escribirse en el portapapeles")
		return nil
	}

	m := newFlowModel()
	m.outputLog = nil
	m.copyOutput()
	if !strings.Contains(m.plainOutput(), "No hay salida para copiar") {
		t.Fatalf("falta el aviso de salida vacía: %q", m.plainOutput())
	}
}
