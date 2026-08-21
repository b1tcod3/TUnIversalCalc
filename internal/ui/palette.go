package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Palette es el selector de comandos que aparece al escribir "/" en un
// panel. Hace fuzzy match sobre la lista del registro del panel (orden
// alfabético) y una fila resaltada se navega con flechas y se inserta
// con Enter.
type Palette struct {
	Open   bool
	List   []Template
	Sel    int
	Off    int // desplazamiento de scroll en la ventana visible
	Max    int // filas visibles según el layout
	Anchor Focus
}

// fragmentOf devuelve el texto tras "/" (minúsculas, sin espacios en los
// extremos) o "" si el valor no empieza por "/".
func fragmentOf(v string) string {
	if !strings.HasPrefix(v, "/") {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "/")))
}

// buildPalette filtra la lista por fuzzy match del fragmento conservando el
// orden alfabético original del registro. Un fragmento vacío devuelve todo.
func buildPalette(list []Template, frag string) []Template {
	if frag == "" {
		return list
	}
	out := make([]Template, 0, len(list))
	for _, t := range list {
		if fuzzy.MatchFold(frag, t.Trigger) {
			out = append(out, t)
		}
	}
	return out
}

// refreshPalette reajusta la paleta según el input enfocado. La abre si el
// valor empieza por "/" (con al menos un match) y la cierra si el fragmento
// deja de coincidir con nada o si coincide exactamente con un trigger
// (caso en el que el usuario escribe la orden completa a mano).
func (m *Model) refreshPalette() {
	wasOpen := m.palette.Open

	var list []Template
	var raw string
	switch m.focus {
	case FocusDecl:
		m.palette.Anchor = FocusDecl
		list = declRegistry
		raw = m.declInput.Value()
	default:
		m.palette.Anchor = FocusCmd
		list = opsRegistry
		raw = m.cmdInput.Value()
	}

	if !strings.HasPrefix(raw, "/") {
		m.closePalette()
		return
	}
	frag := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "/")))
	if frag == "" {
		m.palette.List = buildPalette(list, "")
		m.palette.Open = len(m.palette.List) > 0
	} else {
		for i := range list {
			if strings.ToLower(list[i].Trigger) == frag {
				m.closePalette() // coincidencia exacta: el usuario escribe la orden a mano
				return
			}
		}
		m.palette.List = buildPalette(list, frag)
		m.palette.Open = len(m.palette.List) > 0
	}

	if !wasOpen && m.palette.Open {
		m.palette.Sel, m.palette.Off = 0, 0
	}
	m.clampPalette()
}

// closePalette cierra la paleta conservando Max (lo fija updateLayout).
// Resetear el struct completo borraría Max y la paleta no renderizaría.
func (m *Model) closePalette() {
	m.palette.Open = false
	m.palette.List = nil
	m.palette.Sel, m.palette.Off = 0, 0
}

// clampPalette mantiene la selección y la ventana visible en rango.
func (m *Model) clampPalette() {
	p := &m.palette
	if !p.Open || len(p.List) == 0 {
		p.Sel, p.Off = 0, 0
		return
	}
	if p.Sel >= len(p.List) {
		p.Sel = len(p.List) - 1
	}
	if p.Sel < p.Off {
		p.Off = p.Sel
	}
	if p.Sel >= p.Off+p.Max {
		p.Off = p.Sel - p.Max + 1
	}
}

func (m *Model) paletteUp() {
	if !m.palette.Open || len(m.palette.List) == 0 {
		return
	}
	if m.palette.Sel > 0 {
		m.palette.Sel--
	}
	m.clampPalette()
}

func (m *Model) paletteDown() {
	if !m.palette.Open || len(m.palette.List) == 0 {
		return
	}
	if m.palette.Sel < len(m.palette.List)-1 {
		m.palette.Sel++
	}
	m.clampPalette()
}

// paletteSelect inserta la plantilla seleccionada y cierra la paleta.
func (m *Model) paletteSelect() {
	tpl := m.palette.List[m.palette.Sel]
	m.closePalette()
	val, pos := JumpNextField(tpl.InsertText, 0)
	if m.focus == FocusDecl {
		m.declInput.SetValue(val)
		m.declInput.SetCursor(pos)
	} else {
		m.cmdInput.SetValue(val)
		m.cmdInput.SetCursor(pos)
	}
	m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
}

// paletteView pinta las filas visibles de la paleta (sin salto final):
// trigger resaltado, plantilla y descripción truncadas al ancho interior.
func (m Model) paletteView() string {
	if !m.palette.Open || len(m.palette.List) == 0 {
		return ""
	}
	innerW := m.leftWidth() - chromeInW - chromeInW
	var lines []string
	from := m.palette.Off
	to := from + m.palette.Max
	if to > len(m.palette.List) {
		to = len(m.palette.List)
	}
	for i := from; i < to; i++ {
		t := m.palette.List[i]
		line := clampWidth(t.Trigger+"  "+t.InsertText+"  — "+t.Desc, innerW)
		if i == m.palette.Sel {
			line = paletteSelStyle.Render(line)
		} else {
			line = paletteRowStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// Estilos de la paleta
var (
	paletteRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
	paletteSelStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			Bold(true)
)
