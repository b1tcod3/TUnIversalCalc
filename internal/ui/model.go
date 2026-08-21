package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
)

// Focus identifica qué panel tiene el cursor.
type Focus int

const (
	FocusDecl Focus = iota
	FocusCmd
)

func (f Focus) String() string {
	if f == FocusDecl {
		return "Declaración"
	}
	return "Comando"
}

// LogLevel clasifica cada entrada del historial de salida.
type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelSuccess
	LevelError
	LevelStep
)

// LogEntry es una línea tipada del panel de salida.
type LogEntry struct {
	Level   LogLevel
	Content string
}

const maxLogEntries = 200

type Model struct {
	width  int
	height int

	declInput textinput.Model
	cmdInput  textinput.Model
	output    viewport.Model

	focus     Focus
	history   []string
	histIdx   int
	outputLog []LogEntry

	// Objetos declarados por el usuario: el panel 1 (siempre con
	// slash-command) registra objetos con auto-nombre (e1, f1, p1...)
	// y el panel 2 opera sobre ellos por nombre o sobre el último.
	objs      map[string]engine.Decl
	lastObj   string
	counters  map[engine.DeclType]int
	declValid bool
	tabHinted bool

	// Selector de comandos al escribir "/" (paleta de auto-completado).
	palette Palette
}

func NewModel() Model {
	declTi := textinput.New()
	declTi.Placeholder = "ej. /ecuacion 2x+2=0, /funcion 2x+2"
	declTi.Prompt = "❯ "
	declTi.CharLimit = 200
	declTi.Focus()

	cmdTi := textinput.New()
	cmdTi.Placeholder = "ej. /resolver x, /integral, /derivar"
	cmdTi.Prompt = "❯ "
	cmdTi.CharLimit = 200

	output := viewport.New(60, 10)

	m := Model{
		declInput: declTi,
		cmdInput:  cmdTi,
		output:    output,
		focus:     FocusDecl,
		histIdx:   -1,
		objs:      map[string]engine.Decl{},
		counters:  map[engine.DeclType]int{},
		outputLog: []LogEntry{
			{Level: LevelInfo, Content: "Bienvenido a TUnIversalCalc"},
			{Level: LevelInfo, Content: "1. Declara un objeto: /ecuacion 2x+2=0, /funcion 2x+2 o /expresion x^2+1"},
			{Level: LevelInfo, Content: "2. Opera por nombre o sobre el último objeto: /resolver x, /integral, /derivar"},
		},
	}
	m.output.SetContent(m.renderLog())
	return m
}

func (Model) Init() tea.Cmd {
	return textinput.Blink
}
