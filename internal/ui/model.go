package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Focus identifica qué panel tiene el cursor.
type Focus int

const (
	FocusExpr Focus = iota
	FocusCmd
)

func (f Focus) String() string {
	if f == FocusExpr {
		return "Expresión"
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

	exprInput textinput.Model
	cmdInput  textinput.Model
	output    viewport.Model

	focus     Focus
	history   []string
	histIdx   int
	outputLog []LogEntry

	// Variables definidas por el usuario: la expresión del panel 1 se
	// guarda como "f" al pulsar Enter (mapa listo para g, h, ...).
	vars      map[string]string
	exprValid bool
	tabHinted bool
}

func NewModel() Model {
	exprTi := textinput.New()
	exprTi.Placeholder = "ej. x^2 + 3x - 5"
	exprTi.Prompt = "ƒ(x) = "
	exprTi.CharLimit = 200
	exprTi.Focus()

	cmdTi := textinput.New()
	cmdTi.Placeholder = "ej. /integral, /derivar, /limite"
	cmdTi.Prompt = "❯ "
	cmdTi.CharLimit = 200

	output := viewport.New(60, 10)

	m := Model{
		exprInput: exprTi,
		cmdInput:  cmdTi,
		output:    output,
		focus:     FocusExpr,
		histIdx:   -1,
		vars:      map[string]string{},
		outputLog: []LogEntry{
			{Level: LevelInfo, Content: "Bienvenido a TUnIversalCalc"},
			{Level: LevelInfo, Content: "1. Escribe una expresión y pulsa Enter: se guardará como f."},
			{Level: LevelInfo, Content: "2. Úsala en el panel de comandos: /integral f, /derivar f"},
		},
	}
	m.output.SetContent(m.renderLog())
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
