package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View dibuja la interfaz en la terminal
func (m Model) View() string {
	header := titleStyle.Render("TUnIversarlCalc - Motor Matemático")

	// Dibujamos la sugerencia de comando si existe
	suggestionBox := ""
	if m.Suggest != "" {
		suggestionBox = "\nSugerencia: " + cmdStyle.Render(m.Suggest+" (Presiona Enter para completar)")
	}
	// Mostramos tanto la expresión actual como el último resultado
	// EL DETALLE CLAVE:
	// Asegúrate de llamar a m.TextInput.View() para renderizar la caja de texto con su contenido
	body := fmt.Sprintf(
		"\n%s%s\n\nResultado: %s\n\n(Esc: Salir)",
		m.TextInput.View(), // ESTA ES LA LÍNEA QUE DIBUJA EL TEXTO Y EL CURSOR
		suggestionBox,
		m.Result,
	)
	ui := lipgloss.JoinVertical(lipgloss.Left, header, body)

	return appStyle.Render(ui) + "\n"
}
