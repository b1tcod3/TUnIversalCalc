package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	// 1. Etiqueta de la aplicación: [ TUnIversarlCalc ]
	brand := brandStyle.Render("TUnIversarlCalc")

	// 2. El símbolo de inicio de línea (verde)
	prompt := promptStyle.Render("❯")

	// 3. La caja de texto interactiva
	inputLine := m.TextInput.View()

	// 4. La sugerencia inline (gris) pegada al input
	suggest := ""
	if m.Suggest != "" {
		suggest = suggestStyle.Render("→ " + m.Suggest)
	}

	// 5. El resultado a la derecha
	resultStr := ""
	if m.Result != "" {
		resultStr = resultStyle.Render(m.Result)
	}

	// -- CONSTRUCCIÓN FINAL HORIZONTAL --
	// Juntamos todo en una sola línea principal
	mainLine := lipgloss.JoinHorizontal(lipgloss.Top, brand, prompt, inputLine, suggest, resultStr)

	// Le damos un poco de respiro por arriba y abajo
	return "\n" + mainLine + "\n\n" + infoStyle.Render("Esc/Ctrl+C para salir | Enter para evaluar") + "\n"
}
