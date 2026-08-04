package main

import (
    "log"
    "os"
    "tuniversalcalc/ui"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // tea.withAltScreen() abre la app en un pantalla altenativa
    // para no ensuciar el historial de tu terminal al salir.
    p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        log.Fatalf("Error crítico al iniciar TUnIversalCalc: %v", err)
        os.Exit(1)
    }
}

