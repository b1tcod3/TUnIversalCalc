package ui

import (
	"fmt"
	"sort"
	"strings"
	"tuniversalcalc/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

const helpBarText = "[Tab] Panel • [Enter] Ejecutar • [Ctrl+O] Sesiones • [Ctrl+L] Limpiar • " +
	"[Ctrl+Z/Y] Undo/Redo • [Ctrl+H] Ayuda • [Ctrl+S] Guardar • [Esc] Salir"

func (m *Model) View() string {
	if view, ok := m.modalView(); ok {
		return view
	}

	lw, rw := m.leftWidth(), m.rightWidth()
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, m.declPanel(lw), m.cmdPanel(lw))
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, m.outputPanel(rw))

	help := helpStyle.Render(clampWidth(helpBarText, m.width))

	// Sin newlines externos: el alto total debe ser exactamente m.height
	// para que la terminal no haga scroll y corte la parte superior.
	return grid + "\n" + help + "\n" + statusStyle.Render(clampWidth(m.statusBar(), m.width))
}

func (m *Model) modalView() (string, bool) {
	switch {
	case m.clearConfirm:
		return m.clearConfirmView(), true
	case m.helpOpen:
		return m.helpView(), true
	case m.picker.open:
		return m.pickerView(), true
	}
	return "", false
}

// declPanel compone el panel 1, incluidos la paleta y el autocompletado.
func (m *Model) declPanel(lw int) string {
	innerL := lw - panelBorder - inputPadX

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
	content := clipLines(
		titleRight(" 1. Declaración ", " Declaración ", colorExprActive, innerL)+"\n\n"+exprBody,
		m.leftHeight()-panelBorder,
	)
	return exprStyle.Render(content)
}

// cmdPanel compone el panel 2, incluido el aviso de panel bloqueado y la
// previsualización en vivo.
func (m *Model) cmdPanel(lw int) string {
	innerL := lw - panelBorder - inputPadX
	paletteInCmd := m.palette.Open && m.palette.Anchor == FocusCmd
	completionInCmd := m.completion.Open && m.completion.Anchor == FocusCmd

	calcContent := m.cmdInput.View()
	switch {
	case paletteInCmd:
		calcContent += "\n\n" + m.paletteView()
	case completionInCmd:
		calcContent += "\n\n" + m.completionView()
	case !m.declValid:
		hint := "⛔ Sin objetos declarados\n   Declara en el panel 1:\n   /puntos (1,2) (3,4) (Enter)"
		calcContent = hintStyle.Render(clampLines(hint, innerL))
	}
	if !paletteInCmd && !completionInCmd {
		if pv := m.previewView(); pv != "" {
			calcContent += "\n\n" + pv
		}
	}

	calcStyle := calcPanelInactive
	if m.focus == FocusCmd && m.declValid {
		calcStyle = calcPanelActive
	}
	calcStyle = calcStyle.Width(lw - panelBorder).Height(m.leftHeight() - panelBorder)
	content := clipLines(
		titleRight(" 2. Operación/Comando ", " Comando ", colorCalcActive, innerL)+"\n\n"+calcContent,
		m.leftHeight()-panelBorder,
	)
	return calcStyle.Render(content)
}

// outputPanel compone el panel derecho con la cabecera y el viewport.
func (m *Model) outputPanel(rw int) string {
	innerR := rw - panelBorder - outputPadX
	content := clipLines(
		m.panel3Header(innerR)+m.output.View(),
		m.rightHeight()-panelBorder-outputPadY,
	)
	return outputPanelStyle.
		Width(rw - panelBorder).
		Height(m.rightHeight() - panelBorder).
		Render(content)
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
	if isClearCommand(raw) {
		return "\n\n" + levelInfoStyle.Render(clampWidth(
			"/limpiar borra todo el espacio de trabajo (pide confirmación)", innerL-2))
	}
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
