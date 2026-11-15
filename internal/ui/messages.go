package ui

import (
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

// ResultMsg transporta el resultado del motor al bucle principal. ReqID
// identifica la petición que lo originó: si el usuario ya lanzó otra
// operación, el resultado viejo se descarta para no pisar el nuevo.
type ResultMsg struct {
	ReqID int
	Steps []string
	Err   error
}

// evaluateCmd ejecuta el motor de forma asíncrona (patrón Elm): devuelve un
// comando que corre en otra goroutine y envía ResultMsg cuando termina.
func evaluateCmd(decl engine.Decl, cmd, varName string, reqID int) tea.Cmd {
	return func() tea.Msg {
		steps, err := engine.Evaluate(decl, cmd, varName)
		return ResultMsg{ReqID: reqID, Steps: steps, Err: err}
	}
}
