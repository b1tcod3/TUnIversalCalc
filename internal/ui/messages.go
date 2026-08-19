package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
)

// ResultMsg transporta el resultado del motor al bucle principal.
type ResultMsg struct {
	Steps []string
	Err   error
}

// evaluateCmd ejecuta el motor de forma asíncrona (patrón Elm): devuelve un
// comando que corre en otra goroutine y envía ResultMsg cuando termina.
func evaluateCmd(expr, cmd string) tea.Cmd {
	return func() tea.Msg {
		steps, err := engine.Evaluate(expr, cmd)
		return ResultMsg{Steps: steps, Err: err}
	}
}
