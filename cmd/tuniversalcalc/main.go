package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"tuniversalcalc/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev" // se fija con -ldflags "-X main.version=..."

// runTUI permite sustituir el programa interactivo en los tests.
var runTUI = func(m *ui.Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run ejecuta la CLI y devuelve el código de salida. Las salidas tempranas
// (--version, --list, --help) no arrancan la TUI.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("tuniversalcalc", flag.ContinueOnError)
	flags.SetOutput(stderr)

	session := flags.String("session", "", "abrir o crear una sesión por nombre")
	flags.StringVar(session, "s", "", "alias de --session")
	newSession := flags.Bool("new", false, "empezar una sesión nueva")
	flags.BoolVar(newSession, "n", false, "alias de --new")
	listSessions := flags.Bool("list", false, "listar las sesiones guardadas y salir")
	flags.BoolVar(listSessions, "l", false, "alias de --list")
	showVersion := flags.Bool("version", false, "mostrar la versión y salir")
	flags.BoolVar(showVersion, "v", false, "alias de --version")
	noSave := flags.Bool("no-save", false, "no cargar ni guardar sesión")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *showVersion {
		_, _ = fmt.Fprintln(stdout, "tuniversalcalc", version)
		return 0
	}

	if *listSessions {
		for _, s := range ui.ListSessions() {
			_, _ = fmt.Fprintf(stdout, "%-30s %s\n", s.Name, s.Modified.Format("2006-01-02 15:04"))
		}
		return 0
	}

	name := *session
	if *newSession || name == "" {
		if last, ok := ui.LastSessionName(); ok && !*newSession {
			name = last
		} else {
			name = ui.NewSessionName()
		}
	}

	m := ui.NewModel()
	if !*noSave {
		if err := m.OpenSession(name); err != nil {
			m.DisableSession()
			_, _ = fmt.Fprintf(stderr, "aviso: no se pudo abrir la sesión %q: %v\n", name, err)
		}
	}

	if err := runTUI(&m); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error crítico al iniciar TUnIversalCalc: %v\n", err)
		return 1
	}
	return 0
}
