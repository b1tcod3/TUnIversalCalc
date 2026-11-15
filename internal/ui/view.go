package ui

import (
	"fmt"
	"sort"
	"strings"
	"tuniversalcalc/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string { //nolint:gocyclo // composición de los tres paneles en una sola pasada
	if m.helpOpen {
		return m.helpView()
	}
	if m.picker.open {
		return m.pickerView()
	}

	lw, rw := m.leftWidth(), m.rightWidth()
	innerL, innerR := lw-panelBorder-inputPadX, rw-panelBorder-outputPadX

	// --- PANEL 1: DECLARACIÓN ---
	exprStyle := exprPanelInactive
	if m.focus == FocusDecl {
		exprStyle = exprPanelActive
	}
	exprStyle = exprStyle.Width(lw - panelBorder).Height(m.leftHeight() - panelBorder)
	exprBody := m.declInput.View()
	switch {
	case m.palette.Open && m.palette.Anchor == FocusDecl:
		exprBody += "\n\n" + m.paletteView()
	case m.completion.Open && m.completion.Anchor == FocusDecl:
		exprBody += "\n\n" + m.completionView()
	default:
		exprBody += m.declEchoLine()
	}
	panel1Content := clipLines(
		titleRight(" 1. Declaración ", " Declaración ", colorExprActive, innerL)+"\n\n"+exprBody,
		m.leftHeight()-panelBorder,
	)
	panel1 := exprStyle.Render(panel1Content)

	// --- PANEL 2: CÁLCULO (bloqueado sin objetos declarados) ---
	calcContent := m.cmdInput.View()
	switch {
	case m.palette.Open && m.palette.Anchor == FocusCmd:
		calcContent += "\n\n" + m.paletteView()
	case m.completion.Open && m.completion.Anchor == FocusCmd:
		calcContent += "\n\n" + m.completionView()
	case !m.declValid:
		hint := "⛔ Sin objetos declarados\n   Declara en el panel 1:\n   /puntos (1,2) (3,4) (Enter)"
		calcContent = hintStyle.Render(clampLines(hint, innerL))
	}
	if (!m.palette.Open || m.palette.Anchor != FocusCmd) &&
		(!m.completion.Open || m.completion.Anchor != FocusCmd) {
		if pv := m.previewView(); pv != "" {
			calcContent += "\n\n" + pv
		}
	}
	calcStyle := calcPanelInactive
	if m.focus == FocusCmd && m.declValid {
		calcStyle = calcPanelActive
	}
	calcStyle = calcStyle.Width(lw - panelBorder).Height(m.leftHeight() - panelBorder)
	panel2Content := clipLines(
		titleRight(" 2. Operación/Comando ", " Comando ", colorCalcActive, innerL)+"\n\n"+calcContent,
		m.leftHeight()-panelBorder,
	)
	panel2 := calcStyle.Render(panel2Content)

	// --- PANEL 3: VISTA / PASO A PASO (derecha) ---
	panel3Content := clipLines(
		m.panel3Header(innerR)+m.output.View(),
		m.rightHeight()-panelBorder-outputPadY,
	)
	panel3 := outputPanelStyle.
		Width(rw - panelBorder).
		Height(m.rightHeight() - panelBorder).
		Render(panel3Content)

	// --- ENSAMBLADO DE LA CUADRÍCULA ---
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, panel1, panel2)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, panel3)

	help := helpStyle.Render(clampWidth(
		"[Tab] Panel • [Enter] Ejecutar • [Ctrl+O] Sesiones • [Ctrl+Z/Y] Undo/Redo • [Ctrl+H] Ayuda • [Ctrl+S] Guardar • [Esc] Salir",
		m.width,
	))

	// Sin newlines externos: el alto total debe ser exactamente m.height
	// para que la terminal no haga scroll y corte la parte superior.
	return grid + "\n" + help + "\n" + statusStyle.Render(clampWidth(m.statusBar(), m.width))
}

// declEchoLine muestra en vivo cómo se interpreta la declaración del panel
// 1 (auto-nombre + LaTeX + Unicode) y sus errores, aunque el input recorte
// el texto con scroll. Solo lee estado: no registra ni muta nada.
func (m *Model) declEchoLine() string {
	raw := strings.TrimSpace(m.declInput.Value())
	if raw == "" {
		return ""
	}
	innerL := m.leftWidth() - panelBorder - inputPadX
	d, err := engine.ParseDecl(raw)
	if err != nil {
		return "\n\n" + levelErrorStyle.Render(clampWidth("✗ "+err.Error(), innerL-2))
	}
	name := fmt.Sprintf("%s%d", d.Tipo.Prefix(), m.counters[d.Tipo]+1)
	return "\n\n" +
		levelInfoStyle.Render(clampWidth(name+" = "+d.Latex, innerL-2)) + " " + levelSuccessStyle.Render("✔") +
		"\n" + levelInfoStyle.Render(clampWidth(d.Pretty, innerL-2))
}

// panel3Header es la parte superior del panel derecho: el banner grande si
// hay ancho, o una marca compacta que nunca envuelve; debajo el título de
// vista a la derecha.
func (m *Model) panel3Header(innerW int) string {
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
func (m *Model) showBanner() bool {
	return m.rightWidth() >= bannerWidth+6 &&
		m.rightHeight() >= bannerRows+10
}

// panel3HeaderRows: filas que consume la cabecera del panel derecho.
func (m *Model) panel3HeaderRows() int {
	if m.showBanner() {
		return bannerRows + 2 // banner (4) + título (1) + blanco (1)
	}
	return 3 // marca (1) + título (1) + blanco (1)
}

func (m *Model) statusBar() string {
	objs := m.objList()
	if objs == "" {
		objs = "—"
	}
	cmdState := "bloqueado"
	if m.declValid {
		cmdState = "habilitado"
	}
	session := "—"
	if m.sessionEnabled {
		session = m.sessionName
		if m.dirty {
			session += "*"
		}
	}

	return statusStyle.Render(fmt.Sprintf(
		"Panel: %s | Objetos: %s | Comandos: %s | Historial: %d | Sesión: %s",
		m.focus, objs, cmdState, len(m.history), session,
	))
}

// objList devuelve los nombres de los objetos declarados ordenados.
func (m *Model) objList() string {
	names := make([]string, 0, len(m.objs))
	for n := range m.objs {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
