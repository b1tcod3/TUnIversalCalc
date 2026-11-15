package ui

import (
	"maps"
	"tuniversalcalc/internal/engine"

	"github.com/charmbracelet/bubbles/textinput"
)

const maxUndoLevels = 50

// uiSnapshot guarda el estado editable de la aplicación para undo/redo.
// No incluye layout, sesión ni los stacks de undo, que sobreviven al
// viaje de ida y vuelta.
type uiSnapshot struct {
	DeclValue  string
	CmdValue   string
	DeclCursor int
	CmdCursor  int

	Focus Focus

	History []string
	HistIdx int

	OutputLog []LogEntry

	Objects  map[string]engine.Decl
	LastObj  string
	Counters map[engine.DeclType]int

	DeclValid bool
	TabHinted bool
}

// snapshot crea una copia profunda del estado editable.
func (m *Model) snapshot() uiSnapshot {
	s := uiSnapshot{
		DeclValue:  m.declInput.Value(),
		CmdValue:   m.cmdInput.Value(),
		DeclCursor: m.declInput.Position(),
		CmdCursor:  m.cmdInput.Position(),
		Focus:      m.focus,
		History:    append([]string(nil), m.history...),
		HistIdx:    m.histIdx,
		OutputLog:  append([]LogEntry(nil), m.outputLog...),
		Objects:    make(map[string]engine.Decl, len(m.objs)),
		LastObj:    m.lastObj,
		Counters:   make(map[engine.DeclType]int, len(m.counters)),
		DeclValid:  m.declValid,
		TabHinted:  m.tabHinted,
	}
	for name, d := range m.objs {
		cp := d
		if d.Points != nil {
			cp.Points = append([]engine.Point(nil), d.Points...)
		}
		s.Objects[name] = cp
	}
	maps.Copy(s.Counters, m.counters)
	return s
}

// restore aplica un snapshot al modelo.
func (m *Model) restore(s uiSnapshot) {
	m.declInput.SetValue(s.DeclValue)
	m.setCursor(&m.declInput, s.DeclCursor)
	m.cmdInput.SetValue(s.CmdValue)
	m.setCursor(&m.cmdInput, s.CmdCursor)

	m.focus = s.Focus
	m.declInput.Blur()
	m.cmdInput.Blur()
	if m.focus == FocusDecl {
		m.declInput.Focus()
	} else {
		m.cmdInput.Focus()
	}

	m.history = append([]string(nil), s.History...)
	m.histIdx = s.HistIdx
	m.outputLog = append([]LogEntry(nil), s.OutputLog...)

	m.objs = s.Objects
	m.lastObj = s.LastObj
	m.counters = s.Counters
	m.declValid = s.DeclValid
	m.tabHinted = s.TabHinted

	m.helpOpen = false
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.invalidatePending()
	m.output.SetContent(m.renderLog())
	m.output.GotoBottom()
	m.dirty = true
}

// setCursor fija el cursor; textinput lo acota al rango válido.
func (m *Model) setCursor(ti *textinput.Model, pos int) {
	ti.SetCursor(pos)
}

// pushUndo guarda el estado actual antes de una acción que lo modifica.
func (m *Model) pushUndo() {
	m.undoStack = append(m.undoStack, m.snapshot())
	if len(m.undoStack) > maxUndoLevels {
		m.undoStack = m.undoStack[len(m.undoStack)-maxUndoLevels:]
	}
	m.redoStack = nil
}

// undo deshace la última acción y devuelve si había algo que deshacer.
func (m *Model) undo() bool {
	if len(m.undoStack) == 0 {
		return false
	}
	m.redoStack = append(m.redoStack, m.snapshot())
	if len(m.redoStack) > maxUndoLevels {
		m.redoStack = m.redoStack[len(m.redoStack)-maxUndoLevels:]
	}
	last := len(m.undoStack) - 1
	s := m.undoStack[last]
	m.undoStack = m.undoStack[:last]
	m.restore(s)
	return true
}

// redo reaplica la última acción deshecha.
func (m *Model) redo() bool {
	if len(m.redoStack) == 0 {
		return false
	}
	m.undoStack = append(m.undoStack, m.snapshot())
	if len(m.undoStack) > maxUndoLevels {
		m.undoStack = m.undoStack[len(m.undoStack)-maxUndoLevels:]
	}
	last := len(m.redoStack) - 1
	s := m.redoStack[last]
	m.redoStack = m.redoStack[:last]
	m.restore(s)
	return true
}
