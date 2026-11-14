package ui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestTeatestFlujoLagrange(t *testing.T) {
	m := NewModel()
	tm := teatest.NewTestModel(t, &m, teatest.WithInitialTermSize(100, 28))

	tm.Type("/puntos (1,2) (3,4) (5,6)")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.Type("/lagrange t1")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Operación: interpolación de Lagrange")) &&
			bytes.Contains(b, []byte("P(x) = x + 1"))
	}, teatest.WithDuration(5*time.Second), teatest.WithCheckInterval(25*time.Millisecond))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

func TestGoldenAyuda(t *testing.T) {
	m := NewModel()
	m.width, m.height = 80, 24
	m.updateLayout()
	m.toggleHelp()
	golden.RequireEqual(t, []byte(ansi.Strip(m.helpView())))
}
