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

// Estilo base de los paneles: borde redondeado. Los paneles 1 y 2 usan
// padding horizontal (1) y vertical 0 para maximizar las filas de input;
// el panel 3 conserva el padding completo.
var basePanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder())

var inputPanel = basePanel.
	Padding(0, 1)

var outputPanel = basePanel.
	Padding(1)

// Variantes estáticas activo/inactivo — nunca se mutan desde View().
var (
	exprPanelInactive = inputPanel.
				BorderForeground(lipgloss.Color(colorInactive))
	exprPanelActive = inputPanel.
			BorderForeground(lipgloss.Color(colorExprActive))

	calcPanelInactive = inputPanel.
				BorderForeground(lipgloss.Color(colorInactive))
	calcPanelActive = inputPanel.
			BorderForeground(lipgloss.Color(colorCalcActive))

	outputPanelStyle = outputPanel.
				BorderForeground(lipgloss.Color(colorOutput))
)

var titleRightStyle = lipgloss.NewStyle().
	Bold(true)

func titleRight(title, short, color string, innerW int) string {
	t := title
	if innerW < visualWidth(title)+2 {
		t = short
	}
	if innerW < visualWidth(short)+2 {
		return ""
	}
	return lipgloss.PlaceHorizontal(
		innerW,
		lipgloss.Right,
		titleRightStyle.Foreground(lipgloss.Color(color)).Render(t),
	)
}

// Banner del nombre de la aplicación (panel derecho)
var bannerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorExprActive)).
	Bold(true)

// Colores por nivel de log
var (
	levelSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	levelErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)
	levelStepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorStep))
	levelInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
)

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorInfo))

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorInactive))

// Marca de la aplicación (panel derecho)
var brandStyle = lipgloss.NewStyle().
	Background(lipgloss.Color(colorExprActive)).
	Foreground(lipgloss.Color("255")).
	Padding(0, 1).
	Bold(true)

// Aviso del panel 2 cuando está bloqueado
var hintStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("244")).
	Italic(true)

// Estilos de la pantalla de ayuda
var (
	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorOutput)).
			Padding(0, 1)
	helpTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorOutput))
	helpFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
)
