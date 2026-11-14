# TUnIversalCalc

```
▀▛▘▌ ▌   ▜▘              ▜ ▞▀▖   ▜
 ▌ ▌ ▌▛▀▖▐▌ ▌▞▀▖▙▀▖▞▀▘▝▀▖▐ ▌  ▝▀▖▐ ▞▀▖
 ▌ ▌ ▌▌ ▌▐▐▐ ▛▀ ▌  ▝▀▖▞▀▌▐ ▌ ▖▞▀▌▐ ▌ ▖
 ▘ ▝▀ ▘ ▘▀▘▘ ▝▀▘▘  ▀▀ ▝▀▘ ▘▝▀ ▝▀▘ ▘▝▀
```

Calculadora TUI (interfaz de texto) en **Go** con **Bubble Tea**, construida con arquitectura **Elm** y el ecosistema **Charmbracelet**. Enfocada en **cálculo numérico**: interpolación polinómica.

## Características

- **Interfaz de paneles divididos** (Grid con Lipgloss):
  - Panel 1 arriba-izquierda: **declaración de objetos** con slash-command (`/puntos (1,2) (3,4)`) que se guardan con **auto-nombre** (`e1`, `f1`, `p1`, `t1`...) al pulsar Enter
  - Panel 2 abajo-izquierda: **operación** sobre un objeto por nombre o sobre el último declarado (`/lagrange t1`) — **bloqueado** hasta que haya al menos un objeto
  - Panel 3 derecha: vista de salida con el nombre **TUnIversalCalc en arte ASCII** (smblock, en terminales anchas) o la barra de marca compacta, scroll y colores por nivel
- **Objetos nombrados**: cada declaración se registra (`e1`, `f1`, `t1`, ...) y el comando es *inteligente*: `/lagrange x` usa el último objeto, `/lagrange t1` apunta a uno concreto, y un nombre inexistente da error citando los disponibles
- **LaTeX emulado**: cada declaración/operación muestra la expresión en formato LaTeX (`x^{2} + 3x \cdot 2`) junto a su versión Unicode legible (`2·x² + 3·x·2`)
- **Validación en vivo**: el panel 1 previsualiza el auto-nombre, el LaTeX y la validez mientras escribes; los errores aparecen en la salida
- **Plantillas editables** con campos saltables: paleta de comandos al escribir `/` (lista alfabética con fuzzy match, navegación y Enter para insertar), y `Shift+Tab` recorre los campos (estilo snippets)
- **Renderizado matemático Unicode**: `x^2 → x²`, `sqrt → √`, `pi → π`, `* → ·`, `- → −`, superíndices y subíndices
- Arquitectura Elm real: motor asíncrono vía `tea.Cmd` + `ResultMsg` (la UI no conoce el cálculo)
- Redimensionamiento dinámico (`tea.WindowSizeMsg`), historial de comandos con `↑/↓`, estilos inmutables (sin mutación de globales en `View()`)
- **Interpolación polinómica** (cálculo numérico): declaración de puntos con `/puntos` y dos caminos para el mismo polinomio único — Lagrange y Newton. Newton muestra la tabla completa de diferencias divididas. `symbolic.go` (parser Pratt + derivada/integral/etc. simbólicas) se conserva como base del parser.
- **Sesiones persistentes con autosave**: el trabajo se guarda cada 30 segundos y al salir en `$XDG_DATA_HOME/tuniversalcalc/sessions` (o `~/.local/share/...`). Se puede abrir una sesión concreta por nombre o cambiar entre sesiones con `Ctrl+O`; el selector permite renombrar (`Ctrl+R`), duplicar (`Ctrl+U`) y borrar (`Ctrl+D`).
- **Undo/redo**: `Ctrl+Z` / `Ctrl+Y` con hasta 50 niveles, restaurando declaraciones, historial, foco y salida.
- **Ayuda contextual**: `Ctrl+H` abre una pantalla de ayuda con scroll; `Esc` la cierra.
- **Autocompletado en expresiones**: `Ctrl+Space` sugiere funciones, constantes y objetos declarados; `↑/↓` navega, `Tab`/`Enter` acepta y `Esc` cierra.
- **Evaluación en vivo**: tras una pausa breve, el panel 2 muestra la última línea del resultado de la operación mientras escribes.
- **Tema adaptativo**: los colores se ajustan automáticamente a terminales claras u oscuras.
- **Ratón**: rueda para desplazar la salida y clic para cambiar de panel.
- **Portapapeles**: `F2` copia la salida completa sin códigos ANSI.
- **CLI**: `--session`, `--new`, `--list`, `--no-save`, `--version` y `--help`.

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

### Sesiones y CLI

```bash
go run ./cmd/tuniversalcalc                 # reanuda la última sesión
go run ./cmd/tuniversalcalc --new           # empieza una sesión nueva
go run ./cmd/tuniversalcalc --session "tesis"  # abre o crea "tesis"
go run ./cmd/tuniversalcalc --list         # lista las sesiones guardadas
go run ./cmd/tuniversalcalc --no-save      # modo efímero, sin persistencia
go run ./cmd/tuniversalcalc --version
```

Las sesiones se guardan como JSON en `$XDG_DATA_HOME/tuniversalcalc/sessions/`
(o `~/.local/share/tuniversalcalc/sessions/` si `XDG_DATA_HOME` no está definido).
El guardado es atómico y automático cada 30 segundos; también se guarda al salir.
La barra de estado muestra la sesión activa y un `*` cuando hay cambios sin guardar.

En el selector de sesiones (`Ctrl+O`): `↑/↓` navega, `Enter` abre, `Ctrl+R`
renombra, `Ctrl+U` duplica, `Ctrl+D` borra (confirmación con `y`) y `Esc` cierra.

### Controles

| Tecla            | Acción                                |
| ---------------- | ------------------------------------- |
| `Tab`            | Cambiar entre panel Declaración/Comando (panel 2 solo con objetos declarados). Con la paleta abierta: bajar en la lista |
| `Shift+Tab`      | Saltar al siguiente campo de plantilla. Con la paleta abierta: subir en la lista |
| `↑` / `↓`        | Historial de comandos (panel 2). Con la paleta abierta: mover la selección |
| `Enter`          | Insertar plantilla / declarar o evaluar. Con la paleta abierta: insertar la seleccionada |
| `Ctrl+Z` / `Ctrl+Y` | Deshacer / rehacer |
| `Ctrl+H`         | Abrir/cerrar la ayuda contextual |
| `Ctrl+O`         | Abrir el selector de sesiones |
| `Ctrl+R` / `Ctrl+U` / `Ctrl+D` | En el selector: renombrar / duplicar / borrar |
| `Ctrl+Space`     | Autocompletar funciones, constantes u objetos |
| `Ctrl+S`         | Guardar la sesión ahora |
| `F2`             | Copiar la salida al portapapeles |
| `PgUp` / `PgDn`  | Scroll del panel de salida |
| Rueda del ratón  | Scroll del panel de salida |
| Clic izquierdo   | Cambiar de panel |
| `Esc`            | Cerrar la paleta/ayuda (si está abierta) o salir |
| `Esc` / `Ctrl+C` | Salir de la aplicación                |

### Uso (flujo)

1. En el panel 1 declara un objeto con slash-command y pulsa **Enter**: `/puntos (1,2) (3,4) (5,6)`, `/funcion 2x+2`, `/ecuacion 2x+2=0` o `/expresion x^2+1`. Se registra con auto-nombre (`t1`, `f1`, `e1`, `p1`...) y el foco salta al panel 2.
2. En el panel 2 escribe la operación de interpolación, citando el objeto o dejando que use el último: `/lagrange`, `/lagrange t1`, `/newton t1`.
3. La salida paso a paso (con LaTeX emulado, Unicode y errores claros) aparece en el panel derecho. Ejemplo real: `/puntos (1,2) (3,4) (5,6)` → `/lagrange t1` muestra los polinomios base `L0`, `L1`, `L2` y termina con `P(x) = x + 1`; `/newton t1` muestra la tabla completa de diferencias divididas y llega al mismo `P(x) = x + 1` (unicidad).

El panel 2 permanece bloqueado (⛔) mientras no haya al menos un objeto declarado.

### Comandos (con autocompletado por prefijo)

**Paleta de comandos**: escribe `/` y aparece la lista de comandos del panel en **orden alfabético**, filtrada por **fuzzy match** (`/lag` → `lagrange`, `/new` → `newton`). Navega con `↑`/`↓` (o `Shift+Tab`/`Tab`), confirma con `Enter` (inserta la plantilla y coloca el cursor en el primer campo) y cierra con `Esc`. La lista usa scroll si no cabe en el panel. Si escribes el comando completo a mano (o con argumentos), la paleta se cierra sola.

**Panel 1 — declaración** (auto-nombre: `e1`, `f1`, `p1`, `t1`...):

- `/ecuacion` — `/ecuacion ?` — ecuación con un `=` (ej. `2x+2=0`)
- `/expresion` — `/expresion ?` — expresión sin `=`
- `/funcion` — `/funcion ?` — función sin `=`
- `/puntos` — `/puntos ?` — puntos para interpolar (ej. `/puntos (1,2) (3,4) (5,6)` o `/puntos 1,2 3,4`)

**Panel 2 — operación** (el objeto se puede omitir: se usa el último declarado):

- `/lagrange` — `/lagrange ?` — polinomio interpolador por el método de Lagrange (bases `L_i(x)` + suma ponderada)
- `/newton` — `/newton ?` — polinomio interpolador por diferencias divididas (tabla completa + forma de Newton)

Ambos producen el mismo polinomio único de grado ≤ n para n+1 puntos.

## Estructura del proyecto

```
.
├── cmd/tuniversalcalc/     # Punto de entrada (flags, sesión y bootstrap de Bubble Tea)
└── internal/
    ├── ui/                 # Interfaz: model, update, view, styles, templates, messages,
    │                       # session, sessionpicker, undo, help, completion, preview,
    │                       # textutil, mouse, clipboard, palette
    ├── engine/             # Declaraciones + interpolación (Lagrange/Newton) + parser simbólico (symbolic.go)
    └── render/             # Renderizado Unicode y LaTeX emulado (superíndices, ∫, √, π, ∞)
```

## Stack tecnológico

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — framework TUI basado en la arquitectura Elm
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — estilos y composición de la interfaz (grid de paneles)
- [Bubbles](https://github.com/charmbracelet/bubbles) — componentes reutilizables (textinput, viewport)
- [x/ansi](https://github.com/charmbracelet/x) — medición y recorte de texto con ANSI y caracteres anchos
- [atotto/clipboard](https://github.com/atotto/clipboard) — copiado al portapapeles

## Créditos

- **Inspiración de UX:** [nascTUI](https://github.com/parnoldx/nascTUI), de [@parnoldx](https://github.com/parnoldx), bajo GPL-2.0. Las ideas de sesiones persistentes, autosave, undo/redo, ayuda contextual, autocompletado, evaluación en vivo y soporte de ratón se reimplementaron para este proyecto; **no se copió código** de nascTUI. El modo hoja multilínea y el backend libqalculate quedan como trabajo futuro.
- **Bubble Tea**, **Lipgloss**, **Bubbles** y **x/ansi** — [Charmbracelet](https://github.com/charmbracelet).
- **fuzzysearch** — [lithammer/fuzzysearch](https://github.com/lithammer/fuzzysearch).
- **Go** — [go.dev](https://go.dev).
- **Banner ASCII** — generado con [`toilet`](https://github.com/cacalabs/toilet) usando la fuente `smblock`.

Ver [CREDITS.md](CREDITS.md) para el detalle de licencias y atribuciones.

## Licencia

Pendiente de decisión. Este repositorio no incorpora código de nascTUI; si en el futuro se copia código GPL-2.0, el proyecto completo deberá distribuirse bajo GPL-2.0.