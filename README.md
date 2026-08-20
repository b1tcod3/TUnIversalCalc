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
- **Plantillas editables** con campos saltables: paleta de comandos al escribir `/` (lista alfabética con fuzzy match, navegación y Enter para insertar), y `Shift+Tab` recorre los campos (estilo snippets)
- **Renderizado matemático Unicode**: `x^2 → x²`, `sqrt → √`, `pi → π`, `* → ·`, `- → −`, superíndices y subíndices
- Arquitectura Elm real: motor asíncrono vía `tea.Cmd` + `ResultMsg` (la UI no conoce el cálculo)
- Redimensionamiento dinámico (`tea.WindowSizeMsg`), historial de comandos con `↑/↓`, estilos inmutables (sin mutación de globales en `View()`)
- **Motor simbólico real**: parser Pratt (lexer + AST) con precedencias, multiplicación implícita (`2x`, `x^(2y)`), paréntesis y errores con posición. Las operaciones trabajan sobre el árbol y muestran las reglas aplicadas paso a paso
- **Operaciones simbólicas**: derivada (suma, producto, cociente, potencia y cadena para `sqrt/sin/cos/abs`), simplificación (términos semejantes, `x·x → x²`), resolución de ecuaciones lineales y cuadráticas (Δ) y límites por sustitución con L'Hôpital (0/0 o ∞/∞) e infinitos polinómicos. Integrales término a término (polinomios, `sin/cos`, `1/x → ln|x|`, `√x`). Lo no soportado da un stub explícito (p. ej. `/matriz`, `x^x`, `∫x·sin(x) dx`)

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
| `Tab`            | Cambiar entre panel Declaración/Comando (panel 2 solo con objetos declarados). Con la paleta abierta: bajar en la lista |
| `Shift+Tab`      | Saltar al siguiente campo de plantilla. Con la paleta abierta: subir en la lista |
| `↑` / `↓`        | Historial de comandos (panel 2). Con la paleta abierta: mover la selección |
| `Enter`          | Insertar plantilla / declarar o evaluar. Con la paleta abierta: insertar la seleccionada |
| `Esc`            | Cerrar la paleta de comandos (si está abierta) o salir |
| `Esc` / `Ctrl+C` | Salir de la aplicación                |

### Uso (flujo)

1. En el panel 1 declara un objeto con slash-command y pulsa **Enter**: `/ecuacion 2x+2=0`, `/funcion 2x+2` o `/expresion x^2+1`. Se registra con auto-nombre (`e1`, `f1`, `p1`...) y el foco salta al panel 2.
2. En el panel 2 escribe la operación, citando el objeto o dejando que use el último: `/resolver x`, `/resolver e1 x`, `/integral e1`, `/derivar f1`, `/simplificar p1`, `/limite f1 2` o `/limite f1 x inf`.
3. La salida paso a paso (con LaTeX emulado, Unicode y errores claros) aparece en el panel derecho. Ejemplo real: `/derivar f1` con `f1 = x^2+3x-5` muestra las reglas de la suma y la potencia y termina con `f1′(x) = 2x + 3`; `/resolver x` con `2x+2` da `x = -1`.

El panel 2 permanece bloqueado (⛔) mientras no haya al menos un objeto declarado.

### Comandos (con autocompletado por prefijo)

**Paleta de comandos**: escribe `/` y aparece la lista de comandos del panel en **orden alfabético**, filtrada por **fuzzy match** (`/rslv` → `resolver`, `/fnc` → `funcion`). Navega con `↑`/`↓` (o `Shift+Tab`/`Tab`), confirma con `Enter` (inserta la plantilla y coloca el cursor en el primer campo) y cierra con `Esc`. La lista usa scroll si no cabe en el panel. Si escribes el comando completo a mano (o con argumentos), la paleta se cierra sola.

**Panel 1 — declaración** (auto-nombre: `e1`, `f1`, `p1`...):

- `/ecuacion` — `/ecuacion ?` — ecuación con un `=` (ej. `2x+2=0`)
- `/expresion` — `/expresion ?` — expresión sin `=`
- `/funcion` — `/funcion ?` — función sin `=`

**Panel 2 — operación** (el objeto se puede omitir: se usa el último declarado):

- `/resolver` — `/resolver ? ?` — resolver ecuación para la variable (ej. `/resolver x`, `/resolver e1 x`)
- `/integral` — `/integral ?` — integral del objeto
- `/derivar` — `/derivar ?` — derivada del objeto
- `/limite` — `/limite ?` — límite cuando la variable → valor (`/limite f1 2`, `/limite f1 x inf`; sin valor se usa 0)
- `/simplificar` — `/simplificar ?` — simplificar el objeto
- `/matriz` — `/matriz ?` — operaciones con matrices

## Estructura del proyecto

```
.
├── cmd/tuniversalcalc/     # Punto de entrada (bootstrap de Bubble Tea)
└── internal/
    ├── ui/                 # Interfaz: model, update, view, styles, templates, messages
    ├── engine/             # Motor simbólico: lexer, parser Pratt (AST), derivada/simplificar/resolver/integral/límite
    └── render/             # Renderizado Unicode y LaTeX emulado (superíndices, ∫, √, π, ∞)
```

## Stack tecnológico

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — framework TUI basado en la arquitectura Elm
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — estilos y composición de la interfaz (grid de paneles)
- [Bubbles](https://github.com/charmbracelet/bubbles) — componentes reutilizables (textinput, viewport)

## Licencia

Pendiente.