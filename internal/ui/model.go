package ui

import (
	"tuniversalcalc/internal/engine"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	Level   LogLevel `json:"level"`
	Content string   `json:"content"`
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

	// Autocompletado de funciones/constantes/objetos (Ctrl+Space).
	completion completion

	// Petición asíncrona más reciente: los resultados con un ReqID viejo
	// se descartan para no pisar una operación nueva.
	reqSeq     int
	pendingReq int

	// Undo/redo de la sesión (snapshots acotados).
	undoStack []uiSnapshot
	redoStack []uiSnapshot

	// Persistencia: sesión activa y si hay cambios sin guardar.
	sessionName    string
	sessionEnabled bool
	dirty          bool

	// Ayuda contextual (Ctrl+H).
	helpOpen     bool
	helpViewport viewport.Model

	// Selector de sesiones (Ctrl+O).
	picker sessionPicker

	// Previsualización en vivo del comando (debounce).
	preview previewState
}

func NewModel() Model {
	declTi := textinput.New()
	declTi.Placeholder = "ej. 120, /puntos (1,2) (3,4), /funcion 2x+2"
	declTi.Prompt = "❯ "
	declTi.CharLimit = 200
	declTi.Focus()

	cmdTi := textinput.New()
	cmdTi.Placeholder = "ej. /lagrange t1, /newton"
	cmdTi.Prompt = "❯ "
	cmdTi.CharLimit = 200

	output := viewport.New(60, 10)
	helpVp := viewport.New(0, 0)

	m := Model{
		declInput:    declTi,
		cmdInput:     cmdTi,
		output:       output,
		helpViewport: helpVp,
		focus:        FocusDecl,
		histIdx:      -1,
		picker:       newSessionPicker(),
		objs:         map[string]engine.Decl{},
		counters:     map[engine.DeclType]int{},
		outputLog: []LogEntry{
			{Level: LevelInfo, Content: "Bienvenido a TUnIversalCalc"},
			{Level: LevelInfo, Content: "1. Declara un entero o un objeto: 120, /puntos (1,2) (3,4), /funcion 2x+2"},
			{Level: LevelInfo, Content: "2. Interpola por nombre o sobre el último objeto: /lagrange t1, /newton"},
			{Level: LevelInfo, Content: "Atajos: Ctrl+H ayuda · Ctrl+Z/Ctrl+Y deshacer/rehacer · Ctrl+S guardar"},
		},
	}
	m.output.SetContent(m.renderLog())
	return m
}

func (m *Model) Init() tea.Cmd {
	if m.sessionEnabled {
		return tea.Batch(textinput.Blink, m.autosaveTick())
	}
	return textinput.Blink
}
