# TUnIversalCalc

Calculadora TUI (interfaz de texto) en **Go** con **Bubble Tea**, construida con arquitectura **Elm** y el ecosistema **Charmbracelet**.

## Características

- Interfaz de terminal interactiva con estilo propio (bordes y colores vía Lipgloss)
- Entrada de comandos con **autocompletado** de sugerencias (`/integral`, `/derivada`, `/matriz`)
- Arquitectura Elm: estado centralizado (`Model`), mensajes (`Msg`) y renderizado declarativo
- Modo de pantalla alterna (`AltScreen`) que no ensucia el historial de tu terminal
- Motor matemático (`engine`) preparado para el analizador (lexer/parser) de expresiones

## Requisitos

- Go 1.26 o superior

## Instalación

```bash
git clone https://github.com/b1tcod3/TUnIversalCalc.git
cd TUnIversalCalc
go mod download
```

## Uso

```bash
go run main.go
```

### Controles

| Tecla        | Acción                          |
| ------------ | ------------------------------- |
| `Enter`      | Ejecuta / autocompleta comando  |
| `Tab`        | Autocompleta sugerencia         |
| `Esc` / `Ctrl+C` | Salir de la aplicación      |

### Comandos

Escribe `/` seguido del nombre para ver las sugerencias disponibles:

- `/integral` — cálculo de integrales
- `/derivada` — cálculo de derivadas
- `/matriz` — operaciones con matrices

## Estructura del proyecto

```
.
├── main.go         # Punto de entrada (bootstrap de Bubble Tea)
├── engine/         # Motor matemático (evaluación de expresiones)
└── ui/             # Interfaz de usuario (Model, Update, View)
```

## Stack tecnológico

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — framework TUI basado en la arquitectura Elm
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — estilos y composición de la interfaz
- [Bubbles](https://github.com/charmbracelet/bubbles) — componentes reutilizables (textinput)

## Licencia

Pendiente.
