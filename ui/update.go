package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

// OJO: Aquí no usamos puntero (m *Model), porque tea.Model exige valores,
// pero sí nos aseguramos de asignar los valores al componente interno.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.Suggest != "" {
				// Autocompleta
				m.TextInput.SetValue(m.Suggest + " ")
				m.TextInput.CursorEnd()
				m.Suggest = ""
			} else {
				// Simula la ejecución
				m.Result = "Evaluando: " + m.TextInput.Value()
				m.TextInput.SetValue("")
			}
			// ¡Importante! Retornamos aquí para evitar procesar el Enter en el componente
			return m, nil
		}
	}

	// 1. Pasa el mensaje al TextInput y GUARDA EL RESULTADO en m.TextInput
	m.TextInput, cmd = m.TextInput.Update(msg)

	// 2. Lee el valor actualizado
	currentText := m.TextInput.Value()
	m.Suggest = ""

	if strings.HasPrefix(currentText, "/") {
		if strings.HasPrefix("/integral", currentText) && currentText != "/integral" {
			m.Suggest = "/integral"
		} else if strings.HasPrefix("/matriz", currentText) && currentText != "/matriz" {
			m.Suggest = "/matriz"
		} else if strings.HasPrefix("/derivada", currentText) && currentText != "/derivada" {
			m.Suggest = "/derivada"
		}
	}

	// Retorna la copia modificada `m`
	return m, cmd
}
