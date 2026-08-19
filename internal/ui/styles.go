package ui

import "github.com/charmbracelet/lipgloss"

// Paleta de colores por panel
const (
	colorExprActive = "62"  // Morado
	colorCalcActive = "42"  // Verde
	colorOutput     = "205" // Rosa
	colorInactive   = "240" // Gris oscuro
	colorError      = "196" // Rojo
	colorStep       = "220" // Amarillo
	colorSuccess    = "42"
	colorInfo       = "244"
)

// Estilo base de los paneles: borde redondeado + padding (sin dimensiones,
// se aplican por render según el tamaño de la terminal).
var basePanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1)

// Variantes estáticas activo/inactivo — nunca se mutan desde View().
var (
	exprPanelInactive = basePanel.Copy().
				BorderForeground(lipgloss.Color(colorInactive))
	exprPanelActive = basePanel.Copy().
			BorderForeground(lipgloss.Color(colorExprActive))

	calcPanelInactive = basePanel.Copy().
				BorderForeground(lipgloss.Color(colorInactive))
	calcPanelActive = basePanel.Copy().
			BorderForeground(lipgloss.Color(colorCalcActive))

	outputPanelStyle = basePanel.Copy().
				BorderForeground(lipgloss.Color(colorOutput))
)

var panelHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorInactive)).
	Bold(true)

func panelHeader(title string) string {
	return panelHeaderStyle.Render(title)
}

// Colores por nivel de log
var (
	levelSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	levelErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)
	levelStepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorStep))
	levelInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
)

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorInfo)).
	MarginTop(1)

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorInactive))
