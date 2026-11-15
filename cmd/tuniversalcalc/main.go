package main

import (
	"flag"
	"fmt"
	"os"
	"tuniversalcalc/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev" // se fija con -ldflags "-X main.version=..."

func main() {
	session := flag.String("session", "", "abrir o crear una sesión por nombre")
	flag.StringVar(session, "s", "", "alias de --session")
	newSession := flag.Bool("new", false, "empezar una sesión nueva")
	flag.BoolVar(newSession, "n", false, "alias de --new")
	listSessions := flag.Bool("list", false, "listar las sesiones guardadas y salir")
	flag.BoolVar(listSessions, "l", false, "alias de --list")
	showVersion := flag.Bool("version", false, "mostrar la versión y salir")
	flag.BoolVar(showVersion, "v", false, "alias de --version")
	noSave := flag.Bool("no-save", false, "no cargar ni guardar sesión")
	flag.Parse()

	if *showVersion {
		fmt.Println("tuniversalcalc", version)
		return
	}

	if *listSessions {
		for _, s := range ui.ListSessions() {
			fmt.Printf("%-30s %s\n", s.Name, s.Modified.Format("2006-01-02 15:04"))
		}
		return
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
			fmt.Fprintf(os.Stderr, "aviso: no se pudo abrir la sesión %q: %v\n", name, err)
		}
	}

	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error crítico al iniciar TUnIversalCalc: %v\n", err)
		os.Exit(1)
	}
}
