# TUnIversalCalc

Calculadora TUI (interfaz de texto) en **Go** con **Bubble Tea**, construida con arquitectura **Elm** y el ecosistema **Charmbracelet**.

## Características

- **Interfaz de paneles divididos** (Grid con Lipgloss):
  - Panel 1 arriba-izquierda: expresión matemática (`ƒ(x) = x^2 + 3x - 5`)
  - Panel 2 abajo-izquierda: operación/comando (`/integral`, `/derivar`, ...)
  - Panel 3 derecha: salida paso a paso con scroll y colores por nivel
- **Plantillas editables** con campos saltables: `/int` autocompleta a `integral( ?, ?, dx )` y `Shift+Tab` recorre los campos (estilo snippets)
- **Renderizado matemático Unicode**: `x^2 → x²`, `sqrt → √`, `pi → π`, `* → ·`, `- → −`, superíndices y subíndices
- Arquitectura Elm real: motor asíncrono vía `tea.Cmd` + `ResultMsg` (la UI no conoce el cálculo)
- Redimensionamiento dinámico (`tea.WindowSizeMsg`), historial de comandos con `↑/↓`, estilos inmutables (sin mutación de globales en `View()`)
- Motor matemático (`engine`) con interfaz async lista para el parser real (Pratt)

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
go run ./cmd/tuniversalcalc
```

### Controles

| Tecla            | Acción                                |
| ---------------- | ------------------------------------- |
| `Tab`            | Cambiar entre panel Expresión/Comando |
| `Shift+Tab`      | Saltar al siguiente campo de plantilla |
| `Enter`          | Insertar plantilla / evaluar cálculo  |
| `↑` / `↓`        | Historial de comandos (panel 2)       |
| `PgUp` / `PgDn`  | Scroll del panel de salida            |
| `Esc` / `Ctrl+C` | Salir de la aplicación                |

### Comandos (con autocompletado por prefijo)

Escribe `/` + 3 letras para autocompletar la plantilla y luego `Shift+Tab` para recorrer los campos `?`:

- `/integral` — `integral( ?, ?, dx )` — integral definida/indefinida
- `/derivar` — `derivar( ?, ? )` — derivada de f(x) respecto a la variable
- `/limite` — `limite( x, ?, ? )` — límite de f(x) cuando x → valor
- `/resolver` — `resolver( ?, ? )` — resolver ecuación para la variable
- `/simplificar` — `simplificar( ? )` — simplificar expresión
- `/matriz` — `matriz( [[?, ?],[?, ?]] )` — operaciones con matrices

## Estructura del proyecto

```
.
├── cmd/tuniversalcalc/     # Punto de entrada (bootstrap de Bubble Tea)
└── internal/
    ├── ui/                 # Interfaz: model, update, view, styles, templates, messages
    ├── engine/             # Motor matemático (stub simbólico async)
    └── render/             # Renderizado Unicode (superíndices, ∫, √, π, ∞)
```

## Stack tecnológico

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — framework TUI basado en la arquitectura Elm
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — estilos y composición de la interfaz (grid de paneles)
- [Bubbles](https://github.com/charmbracelet/bubbles) — componentes reutilizables (textinput, viewport)

## Licencia

Pendiente.