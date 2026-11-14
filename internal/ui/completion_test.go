package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCompletionObjetos(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m) // deja el foco en el panel 2

	m.cmdInput.SetValue("/lagrange t")
	m.cmdInput.SetCursor(len([]rune("/lagrange t")))
	m.openCompletions()
	if !m.completion.Open {
		t.Fatal("el autocompletado debe abrirse con un prefijo de objeto")
	}
	if len(m.completion.Items) != 1 || m.completion.Items[0].Text != "t1" {
		t.Fatalf("sugerencias = %+v", m.completion.Items)
	}
	m.acceptCompletion()
	if got := m.cmdInput.Value(); got != "/lagrange t1" {
		t.Fatalf("valor tras aceptar = %q", got)
	}
	if m.completion.Open {
		t.Fatal("el popup debe cerrarse al aceptar")
	}
}

func TestCompletionFunciones(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/funcion sq")
	m.declInput.SetCursor(len([]rune("/funcion sq")))
	m.openCompletions()
	if !m.completion.Open {
		t.Fatal("el autocompletado debe abrirse para funciones")
	}
	m.acceptCompletion()
	if got := m.declInput.Value(); got != "/funcion sqrt" {
		t.Fatalf("valor tras aceptar = %q", got)
	}
}

func TestCompletionCtrlSpace(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/lagrange ")
	m.cmdInput.SetCursor(len([]rune("/lagrange ")))

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlAt}); !handled {
		t.Fatal("Ctrl+Space debe manejarse")
	}
	if !m.completion.Open {
		t.Fatal("Ctrl+Space debe abrir el autocompletado")
	}
	if len(m.completion.Items) == 0 {
		t.Fatal("debe haber sugerencias")
	}
}

func TestCompletionFiltraAlEscribir(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/lagrange ")
	m.cmdInput.SetCursor(len([]rune("/lagrange ")))
	m.openCompletions()
	if !m.completion.Open {
		t.Fatal("debe abrirse")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	if _, handled := m.handleCompletionKey(msg); !handled {
		t.Fatal("la tecla debe manejarse")
	}
	if got := m.cmdInput.Value(); got != "/lagrange t" {
		t.Fatalf("input = %q", got)
	}
	if len(m.completion.Items) != 1 || m.completion.Items[0].Text != "t1" {
		t.Fatalf("sugerencias = %+v", m.completion.Items)
	}
}

func TestCompletionPanel2Bloqueado(t *testing.T) {
	m := newFlowModel()
	m.focus = FocusCmd
	m.openCompletions()
	if m.completion.Open {
		t.Fatal("no debe abrirse sin objetos declarados")
	}
}

func TestCompletionViewCabe(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/lagrange ")
	m.cmdInput.SetCursor(len([]rune("/lagrange ")))
	m.openCompletions()
	if !strings.Contains(m.View(), "t1") {
		t.Fatal("la vista debe mostrar las sugerencias")
	}
	for _, size := range [][2]int{{100, 28}, {80, 24}, {60, 20}} {
		m.width, m.height = size[0], size[1]
		m.updateLayout()
		checkView(t, &m, size, "autocompletado")
	}
}
