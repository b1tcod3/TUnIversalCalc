# TUnIversalCalc

```
▀▛▘▌ ▌   ▜▘              ▜ ▞▀▖   ▜
 ▌ ▌ ▌▛▀▖▐▌ ▌▞▀▖▙▀▖▞▀▘▝▀▖▐ ▌  ▝▀▖▐ ▞▀▖
 ▌ ▌ ▌▌ ▌▐▐▐ ▛▀ ▌  ▝▀▖▞▀▌▐ ▌ ▖▞▀▌▐ ▌ ▖
 ▘ ▝▀ ▘ ▘▀▘▘ ▝▀▘▘  ▀▀ ▝▀▘ ▘▝▀ ▝▀▘ ▘▝▀
```

Calculadora TUI (interfaz de texto) en **Go** con **Bubble Tea**, construida con arquitectura **Elm** y el ecosistema **Charmbracelet**.

## Características

- **Interfaz de paneles divididos** (Grid con Lipgloss):
  - Panel 1 arriba-izquierda: **declaración de objetos** con slash-command (`/ecuacion 2x+2=0`) que se guardan con **auto-nombre** (`e1`, `f1`, `p1`...) al pulsar Enter
  - Panel 2 abajo-izquierda: **operación** sobre un objeto por nombre o sobre el último declarado (`/resolver x`) — **bloqueado** hasta que haya al menos un objeto
  - Panel 3 derecha: vista de salida con el nombre **TUnIversalCalc en arte ASCII** (smblock, en terminales anchas) o la barra de marca compacta, scroll y colores por nivel
- **Objetos nombrados**: cada declaración se registra (`e1`, `f1`, ...) y el comando es *inteligente*: `/resolver x` usa el último objeto, `/resolver e1 x` apunta a uno concreto, y un nombre inexistente da error citando los disponibles
- **LaTeX emulado**: cada declaración/operación muestra la ecuación en formato LaTeX (`2x^{2} + 3x \cdot 2`) junto a su versión Unicode legible (`2·x² + 3·x·2`)
- **Validación en vivo**: el panel 1 previsualiza el auto-nombre, el LaTeX y la validez mientras escribes; los errores aparecen en la salida
- **Plantillas editables** con campos saltables: `/ecu` autocompleta a `/ecuacion ...` y `/res` a `/resolver ...`; `Shift+Tab` recorre los campos (estilo snippets)
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
| `Tab`            | Cambiar entre panel Declaración/Comando (panel 2 solo con objetos declarados) |
| `Shift+Tab`      | Saltar al siguiente campo de plantilla |
| `Enter`          | Insertar plantilla / declarar o evaluar |
| `↑` / `↓`        | Historial de comandos (panel 2)       |
| `PgUp` / `PgDn`  | Scroll del panel de salida            |
| `Esc` / `Ctrl+C` | Salir de la aplicación                |

### Uso (flujo)

1. En el panel 1 declara un objeto con slash-command y pulsa **Enter**: `/ecuacion 2x+2=0`, `/funcion 2x+2` o `/expresion x^2+1`. Se registra con auto-nombre (`e1`, `f1`, `p1`...) y el foco salta al panel 2.
2. En el panel 2 escribe la operación, citando el objeto o dejando que use el último: `/resolver x`, `/resolver e1 x`, `/integral e1`, `/derivar f1`.
3. La salida paso a paso (con LaTeX emulado, Unicode y errores claros) aparece en el panel derecho.

El panel 2 permanece bloqueado (⛔) mientras no haya al menos un objeto declarado.

### Comandos (con autocompletado por prefijo)

Escribe `/` + 3 letras para autocompletar la plantilla y luego `Shift+Tab` para recorrer los campos `?`:

**Panel 1 — declaración** (auto-nombre: `e1`, `f1`, `p1`...):

- `/ecuacion` — `/ecuacion ?` — ecuación con un `=` (ej. `2x+2=0`)
- `/funcion` — `/funcion ?` — función sin `=`
- `/expresion` — `/expresion ?` — expresión sin `=`

**Panel 2 — operación** (el objeto se puede omitir: se usa el último declarado):

- `/resolver` — `/resolver ? ?` — resolver ecuación para la variable (ej. `/resolver x`, `/resolver e1 x`)
- `/integral` — `/integral ?` — integral del objeto
- `/derivar` — `/derivar ?` — derivada del objeto
- `/limite` — `/limite ?` — límite cuando la variable → valor
- `/simplificar` — `/simplificar ?` — simplificar el objeto
- `/matriz` — `/matriz ?` — operaciones con matrices

## Estructura del proyecto

```
.
├── cmd/tuniversalcalc/     # Punto de entrada (bootstrap de Bubble Tea)
└── internal/
    ├── ui/                 # Interfaz: model, update, view, styles, templates, messages
    ├── engine/             # Motor matemático (stub simbólico async + declaraciones)
    └── render/             # Renderizado Unicode y LaTeX emulado (superíndices, ∫, √, π, ∞)
```

## Stack tecnológico

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — framework TUI basado en la arquitectura Elm
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — estilos y composición de la interfaz (grid de paneles)
- [Bubbles](https://github.com/charmbracelet/bubbles) — componentes reutilizables (textinput, viewport)

## Licencia

Pendiente.