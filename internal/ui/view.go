package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	// Antes del primer WindowSizeMsg usamos dimensiones por defecto
	if m.width == 0 || m.height == 0 {
		m.width, m.height = 100, 28
		m.updateLayout()
	}

	// --- PANEL 1: EXPRESIÓN ---
	exprStyle := exprPanelInactive
	if m.focus == FocusExpr {
		exprStyle = exprPanelActive
	}
	exprStyle = exprStyle.Copy().Width(m.leftWidth()).Height(m.leftHeight())
	panel1 := exprStyle.Render(panelHeader(" 1. Expresión Matemática ") + "\n\n" + m.exprInput.View())

	// --- PANEL 2: CÁLCULO ---
	calcStyle := calcPanelInactive
	if m.focus == FocusCmd {
		calcStyle = calcPanelActive
	}
	calcStyle = calcStyle.Copy().Width(m.leftWidth()).Height(m.leftHeight())
	panel2 := calcStyle.Render(panelHeader(" 2. Operación/Comando ") + "\n\n" + m.cmdInput.View())

	// --- PANEL 3: SALIDA ---
	panel3 := outputPanelStyle.Copy().
		Width(m.rightWidth()).
		Height(m.rightHeight()).
		Render(m.output.View())

	// --- ENSAMBLADO DE LA CUADRÍCULA ---
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, panel1, panel2)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, panel3)

	help := helpStyle.Render("[Tab] Panel • [Shift+Tab] Campo • [Enter] Evaluar • [↑↓] Historial • [PgUp/PgDn] Scroll • [Esc] Salir")

	preview := ""
	if v := strings.TrimSpace(m.exprInput.Value()); v != "" {
		preview = " | Expr: " + v
	}

	status := statusStyle.Render(fmt.Sprintf("Modo: %s%s | Historial: %d | Motor: stub simbólico", m.focus, preview, len(m.history)))

	return "\n" + grid + "\n" + help + "\n" + status + "\n"
}
