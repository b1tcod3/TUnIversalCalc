package ui

import (
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// completionItem es una sugerencia dentro de una expresión.
type completionItem struct {
	Text string
	Desc string
}

// completion es el popup de autocompletado de funciones, constantes y
// objetos. Se abre con Ctrl+Space y filtra por el prefijo de la palabra
// bajo el cursor.
type completion struct {
	Open   bool
	Anchor Focus
	Items  []completionItem
	Sel    int
	Off    int
	Max    int
}

var builtinCompletions = []completionItem{
	{Text: "sqrt", Desc: "raíz cuadrada"},
	{Text: "sin", Desc: "seno"},
	{Text: "cos", Desc: "coseno"},
	{Text: "abs", Desc: "valor absoluto"},
	{Text: "pi", Desc: "constante π"},
	{Text: "inf", Desc: "infinito"},
}

func (m *Model) completionCandidates() []completionItem {
	items := append([]completionItem(nil), builtinCompletions...)
	names := make([]string, 0, len(m.objs))
	for name := range m.objs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		items = append(items, completionItem{Text: name, Desc: "objeto declarado"})
	}
	return items
}

func (m *Model) focusedInput() *textinput.Model {
	if m.focus == FocusDecl {
		return &m.declInput
	}
	return &m.cmdInput
}

// wordBounds devuelve el rango (en runas) de la palabra bajo el cursor.
func wordBounds(s string, cursor int) (start, end int) {
	runes := []rune(s)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	start = cursor
	for start > 0 && isWordRune(runes[start-1]) {
		start--
	}
	end = cursor
	for end < len(runes) && isWordRune(runes[end]) {
		end++
	}
	return start, end
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// openCompletions abre el popup para la palabra actual. No se abre si el
// cursor está dentro del comando inicial (eso lo cubre la paleta "/").
func (m *Model) openCompletions() {
	if m.focus == FocusCmd && !m.declValid {
		return
	}
	ti := m.focusedInput()
	val := ti.Value()
	runes := []rune(val)
	cursor := ti.Position()
	if cursor > len(runes) {
		cursor = len(runes)
	}
	if strings.HasPrefix(val, "/") && !strings.Contains(string(runes[:cursor]), " ") {
		m.refreshPalette()
		return
	}
	start, end := wordBounds(val, cursor)
	word := string(runes[start:end])
	if word != "" && !containsLetter(word) {
		return
	}
	items := m.filterCompletions(word)
	if len(items) == 0 {
		return
	}
	m.closePalette()
	m.completion = completion{
		Open:   true,
		Anchor: m.focus,
		Items:  items,
		Sel:    0,
		Off:    0,
		Max:    m.completion.Max,
	}
	m.clampCompletion()
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func (m *Model) filterCompletions(prefix string) []completionItem {
	lower := strings.ToLower(prefix)
	out := make([]completionItem, 0, len(builtinCompletions))
	for _, item := range m.completionCandidates() {
		if lower == "" || strings.HasPrefix(strings.ToLower(item.Text), lower) {
			if strings.EqualFold(item.Text, prefix) {
				continue
			}
			out = append(out, item)
		}
	}
	return out
}

func (m *Model) closeCompletion() {
	m.completion.Open = false
	m.completion.Items = nil
	m.completion.Sel = 0
	m.completion.Off = 0
}

func (m *Model) clampCompletion() {
	c := &m.completion
	if !c.Open || len(c.Items) == 0 {
		c.Sel, c.Off = 0, 0
		return
	}
	if c.Sel >= len(c.Items) {
		c.Sel = len(c.Items) - 1
	}
	if c.Sel < 0 {
		c.Sel = 0
	}
	if c.Max < 1 {
		c.Max = 1
	}
	if c.Sel < c.Off {
		c.Off = c.Sel
	}
	if c.Sel >= c.Off+c.Max {
		c.Off = c.Sel - c.Max + 1
	}
}

// handleCompletionKey procesa las teclas con el popup abierto.
func (m *Model) handleCompletionKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyEsc:
		m.closeCompletion()
		return nil, true
	case tea.KeyUp, tea.KeyShiftTab:
		if m.completion.Sel > 0 {
			m.completion.Sel--
		}
		m.clampCompletion()
		return nil, true
	case tea.KeyDown, tea.KeyTab:
		if m.completion.Sel < len(m.completion.Items)-1 {
			m.completion.Sel++
		}
		m.clampCompletion()
		return nil, true
	case tea.KeyEnter:
		m.acceptCompletion()
		return nil, true
	case tea.KeyCtrlAt:
		m.closeCompletion()
		return nil, true
	}

	// Cualquier otra tecla se envía al input y se refiltra.
	var cmd tea.Cmd
	if m.focus == FocusDecl {
		m.declInput, cmd = m.declInput.Update(msg)
	} else {
		m.cmdInput, cmd = m.cmdInput.Update(msg)
	}
	m.dirty = true
	m.refreshPalette()
	if m.palette.Open {
		m.closeCompletion()
		return cmd, true
	}
	m.refilterCompletion()
	if m.focus == FocusCmd {
		if previewCmd := m.schedulePreview(); previewCmd != nil {
			cmd = tea.Batch(cmd, previewCmd)
		}
	}
	return cmd, true
}

func (m *Model) refilterCompletion() {
	ti := m.focusedInput()
	start, end := wordBounds(ti.Value(), ti.Position())
	word := string([]rune(ti.Value())[start:end])
	items := m.filterCompletions(word)
	if len(items) == 0 {
		m.closeCompletion()
		return
	}
	m.completion.Items = items
	m.completion.Sel = 0
	m.completion.Off = 0
	m.clampCompletion()
}

func (m *Model) acceptCompletion() {
	if !m.completion.Open || len(m.completion.Items) == 0 {
		return
	}
	item := m.completion.Items[m.completion.Sel]
	ti := m.focusedInput()
	runes := []rune(ti.Value())
	start, end := wordBounds(ti.Value(), ti.Position())
	out := make([]rune, 0, len(runes)-(end-start)+len(item.Text))
	out = append(out, runes[:start]...)
	out = append(out, []rune(item.Text)...)
	out = append(out, runes[end:]...)
	m.pushUndo()
	ti.SetValue(string(out))
	ti.SetCursor(start + len([]rune(item.Text)))
	m.closeCompletion()
	m.refreshPalette()
	m.dirty = true
}

// completionView pinta las filas visibles del popup.
func (m *Model) completionView() string {
	c := m.completion
	if !c.Open || len(c.Items) == 0 {
		return ""
	}
	innerW := m.leftWidth() - chromeInW - chromeInW
	from := c.Off
	if from < 0 {
		from = 0
	}
	to := from + c.Max
	if to > len(c.Items) {
		to = len(c.Items)
	}
	var lines []string
	for i := from; i < to; i++ {
		item := c.Items[i]
		line := truncateVisual(item.Text+"  — "+item.Desc, innerW)
		if i == c.Sel {
			line = paletteSelStyle.Render(line)
		} else {
			line = paletteRowStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
