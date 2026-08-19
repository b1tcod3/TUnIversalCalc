package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"tuniversalcalc/internal/render"
)

func (m Model) View() string {
	// Antes del primer WindowSizeMsg usamos dimensiones por defecto
	if m.width == 0 || m.height == 0 {
		m.width, m.height = 100, 28
		m.updateLayout()
	}

	lw, rw := m.leftWidth(), m.rightWidth()
	innerL, innerR := lw-chromeInW-chromeInW, rw-chromeOutW-chromeOutW

	// --- PANEL 1: EXPRESIÓN ---
	exprStyle := exprPanelInactive
	if m.focus == FocusExpr {
		exprStyle = exprPanelActive
	}
	exprStyle = exprStyle.Copy().Width(lw - chromeInW).Height(m.leftHeight() - chromeInH)
	exprBody := m.exprInput.View() + m.exprEchoLine()
	panel1 := exprStyle.Render(titleRight(" 1. Expresión Matemática ", " Expresión ", colorExprActive, innerL) + "\n\n" + exprBody)

	// --- PANEL 2: CÁLCULO (bloqueado sin expresión válida) ---
	calcContent := m.cmdInput.View()
	if !m.exprValid {
		calcContent = hintStyle.Render("⛔ Esperando expresión válida\n   Escribe f = expresión en el panel 1\n   y pulsa Enter para guardar.")
	}
	calcStyle := calcPanelInactive
	if m.focus == FocusCmd && m.exprValid {
		calcStyle = calcPanelActive
	}
	calcStyle = calcStyle.Copy().Width(lw - chromeInW).Height(m.leftHeight() - chromeInH)
	panel2 := calcStyle.Render(titleRight(" 2. Operación/Comando ", " Comando ", colorCalcActive, innerL) + "\n\n" + calcContent)

	// --- PANEL 3: VISTA / PASO A PASO (derecha) ---
	panel3 := outputPanelStyle.Copy().
		Width(rw - chromeOutW).
		Height(m.rightHeight() - chromeOutH).
		Render(m.panel3Header(innerR) + m.output.View())

	// --- ENSAMBLADO DE LA CUADRÍCULA ---
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, panel1, panel2)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, panel3)

	help := helpStyle.Render(clampWidth(
		"[Tab] Panel • [Shift+Tab] Campo • [Enter] Guardar/Evaluar • [↑↓] Historial • [PgUp/PgDn] Scroll • [Esc] Salir",
		m.width,
	))

	// Sin newlines externos: el alto total debe ser exactamente m.height
	// para que la terminal no haga scroll y corte la parte superior.
	return grid + "\n" + help + "\n" + statusStyle.Render(clampWidth(m.statusBar(), m.width))
}

// clampWidth trunca una cadena plana (sin ANSI) para que quepa en max
// columnas, añadiendo "…" cuando recorta.
func clampWidth(s string, max int) string {
	if max <= 1 {
		return ""
	}
	if w := len([]rune(s)); w > max {
		return string([]rune(s)[:max-1]) + "…"
	}
	return s
}

// exprEchoLine muestra en vivo la expresión como f: aunque el input recorte
// el texto con scroll, siempre se ve el valor actual y su validez.
func (m Model) exprEchoLine() string {
	raw := strings.TrimSpace(m.exprInput.Value())
	if raw == "" {
		return ""
	}
	mark := "✗"
	markStyle := levelErrorStyle
	if m.exprValid {
		mark = "✔"
		markStyle = levelSuccessStyle
	}
	return "\n\n" + levelInfoStyle.Render("f = "+render.Pretty(raw)+" ") + markStyle.Render(mark)
}

// panel3Header es la parte superior del panel derecho: el banner grande si
// hay ancho, o una marca compacta que nunca envuelve; debajo el título de
// vista a la derecha.
func (m Model) panel3Header(innerW int) string {
	brand := ""
	switch {
	case m.showBanner():
		brand = lipgloss.PlaceHorizontal(innerW, lipgloss.Center, bannerStyle.Render(BANNER))
	case innerW >= 16:
		brand = brandStyle.Render("TUnIversalCalc")
	case innerW >= 8:
		brand = brandStyle.Render("TUC")
	}
	return brand + "\n" + titleRight(" Vista / Paso a Paso ", " Vista ", colorOutput, innerW) + "\n\n"
}

// showBanner indica si el arte ASCII cabe en el panel derecho
// (ancho y alto suficientes para no exprimir el viewport).
func (m Model) showBanner() bool {
	return m.rightWidth() >= bannerWidth+6 &&
		m.rightHeight() >= bannerRows+10
}

// panel3HeaderRows: filas que consume la cabecera del panel derecho.
func (m Model) panel3HeaderRows() int {
	if m.showBanner() {
		return bannerRows + 2 // banner (4) + título (1) + blanco (1)
	}
	return 3 // marca (1) + título (1) + blanco (1)
}

func (m Model) statusBar() string {
	var fStatus string
	raw := strings.TrimSpace(m.exprInput.Value())
	switch {
	case m.exprValid:
		fStatus = "✔ " + raw
	case raw != "":
		fStatus = "✗ " + raw
	default:
		fStatus = "— en espera"
	}

	cmdState := "bloqueado"
	if m.exprValid {
		cmdState = "habilitado"
	}

	return statusStyle.Render(fmt.Sprintf(
		"Modo: %s | f: %s | Comandos: %s | Historial: %d",
		m.focus, fStatus, cmdState, len(m.history),
	))
}
