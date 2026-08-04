package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -- ESTILOS (Lipgloss) --

var (
	// Estilo para la caja principal de la aplicación
	appStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Color violeta/azulado
			Padding(1, 2).
			Width(60)

	// Estilo para título
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // rosa
			Bold(true)

	// Estilo para el Input
	cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // verde
)

// modelo que representa el  estado de la interfaz
type Model struct {
	TextInput textinput.Model // componente robusto de entrada
	Input     string
	Result    string
	Suggest   string // sugerencia de comando
}

// NewModel es el constructor que llamara aplicación
func NewModel() Model {
	// configuramos el componente de entrada de textinput
	ti := textinput.New()
	ti.Placeholder = "Escribe un comando..."
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 50

	return Model{
		TextInput: ti,
		Result:    "",
		Suggest:   "",
	}
}

// Init arranca los comandos iniciales de bubble tea
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
