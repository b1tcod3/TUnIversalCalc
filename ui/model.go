package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -- ESTILOS MINIMALISTAS TIPO BARRAS --
var (
	// Etiqueta principal (TUnIversarlCalc) - Fondo morado, texto blanco
	brandStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		MarginRight(1).
		Bold(true)

	// Etiqueta de símbolo/estado (ej: "❯")
	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")). // Verde
		MarginRight(1).
		Bold(true)

	// Sugerencias inline
	suggestStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Gris oscuro
		Italic(true)

	// Resultado del cálculo (a la derecha o debajo)
	resultStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")). // Amarillo/Naranja
		MarginLeft(2)

	// Estilo para errores o mensajes de sistema
	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginLeft(2)
)

type Model struct {
	TextInput textinput.Model
	Input     string
	Result    string
	Suggest   string
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "ej. /integral x^2"
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 50 // Ajustamos el ancho para que quepa en una línea

	// Ya no necesitamos el prompt interno del textinput, usaremos el nuestro (❯)
	ti.Prompt = ""

	return Model{
		TextInput: ti,
		Result:    "",
		Suggest:   "",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
