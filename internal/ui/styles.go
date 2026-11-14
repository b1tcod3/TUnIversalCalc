package ui

import "github.com/charmbracelet/lipgloss"

// Paleta adaptativa: Lipgloss elige el tono claro u oscuro según el fondo
// de la terminal.
var (
	colorExprActive = lipgloss.AdaptiveColor{Light: "#7D4EBD", Dark: "#AF87FF"}
	colorCalcActive = lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#5FD787"}
	colorOutput     = lipgloss.AdaptiveColor{Light: "#C2185B", Dark: "#FF87D7"}
	colorInactive   = lipgloss.AdaptiveColor{Light: "#9E9E9E", Dark: "#585858"}
	colorError      = lipgloss.AdaptiveColor{Light: "#C62828", Dark: "#FF5F5F"}
	colorStep       = lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#FFD75F"}
	colorSuccess    = lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#5FD787"}
	colorInfo       = lipgloss.AdaptiveColor{Light: "#6B6B6B", Dark: "#8A8A8A"}
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
				BorderForeground(colorInactive)
	exprPanelActive = inputPanel.
			BorderForeground(colorExprActive)

	calcPanelInactive = inputPanel.
				BorderForeground(colorInactive)
	calcPanelActive = inputPanel.
			BorderForeground(colorCalcActive)

	outputPanelStyle = outputPanel.
				BorderForeground(colorOutput)
)

var titleRightStyle = lipgloss.NewStyle().
	Bold(true)

func titleRight(title, short string, color lipgloss.TerminalColor, innerW int) string {
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
		titleRightStyle.Foreground(color).Render(t),
	)
}

// Banner del nombre de la aplicación (panel derecho)
var bannerStyle = lipgloss.NewStyle().
	Foreground(colorExprActive).
	Bold(true)

// Colores por nivel de log
var (
	levelSuccessStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	levelErrorStyle   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	levelStepStyle    = lipgloss.NewStyle().Foreground(colorStep)
	levelInfoStyle    = lipgloss.NewStyle().Foreground(colorInfo)
)

var helpStyle = lipgloss.NewStyle().
	Foreground(colorInfo)

var statusStyle = lipgloss.NewStyle().
	Foreground(colorInactive)

// Marca de la aplicación (panel derecho)
var brandStyle = lipgloss.NewStyle().
	Background(colorExprActive).
	Foreground(lipgloss.Color("255")).
	Padding(0, 1).
	Bold(true)

// Aviso del panel 2 cuando está bloqueado
var hintStyle = lipgloss.NewStyle().
	Foreground(colorInfo).
	Italic(true)

// Estilos de la pantalla de ayuda
var (
	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorOutput).
			Padding(0, 1)
	helpTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorOutput)
	helpFooterStyle = lipgloss.NewStyle().Foreground(colorInfo)
)
