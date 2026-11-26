# TUnIversalCalc

```
▀▛▘▌ ▌   ▜▘              ▜ ▞▀▖   ▜
 ▌ ▌ ▌▛▀▖▐▌ ▌▞▀▖▙▀▖▞▀▘▝▀▖▐ ▌  ▝▀▖▐ ▞▀▖
 ▌ ▌ ▌▌ ▌▐▐▐ ▛▀ ▌  ▝▀▖▞▀▌▐ ▌ ▖▞▀▌▐ ▌ ▖
 ▘ ▝▀ ▘ ▘▀▘▘ ▝▀▘▘  ▀▀ ▝▀▘ ▘▝▀ ▝▀▘ ▘▝▀
```

Calculadora TUI (interfaz de texto) en **Go** con **Bubble Tea**, construida con arquitectura **Elm** y el ecosistema **Charmbracelet**. Enfocada en **cálculo numérico** (interpolación polinómica) y **álgebra exacta acotada** (factorización, potencias y radicales con fracciones).

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
- **Factorización de enteros, polinomios y fracciones algebraicas**: al declarar `/expresion 120` o escribir `120` directamente en el panel 1 se detecta el entero y se muestra automáticamente su descomposición en factores primos. `/factorizar` también factoriza polinomios (factor común, diferencia de cuadrados, suma/diferencia de cubos, trinomios, agrupación y raíces racionales) y expresiones racionales: combina numerador y denominador, factoriza ambos y cancela factores comunes. Ejemplos: `4x^2-25 → (2x-5)(2x+5)`, `2x^2+5x-12 → (2x-3)(x+4)`, `(x^2+3x+2)/(x^2-x-2) → (x+2)/(x-2)`, `x^2/(x^2-4)-(x+1)/(x+2) → 1/(x-2)`. Para enteros usa rueda `6k±1`, Miller-Rabin determinista y Pollard Rho.
- **Completar el cuadrado con `/factorizar -cuadrado`**: reescribe una cuadrática `ax^2+bx+c` como `a(x-h)^2+k`, muestra los pasos y el vértice. Ejemplos: `x^2+x+1 → (x + 1/2)^2 + 3/4`, `2x^2-12x+11 → 2*(x - 3)^2 - 7`.
- **Álgebra exacta acotada**: `/potencia` y `/radical` simplifican productos, cocientes, potencias y raíces con fracciones exactas (`math/big.Rat`), sin pérdida de precisión. Soportan variables de varias letras y raíces n-ésimas con `raiz(n, expr)` o `x^(1/n)`. Ejemplos: `(3a^3b^3)(4ab^2)^2 → 48·a⁵·b⁷`, `sqrt(200)-sqrt(32) → 6·√2`, `raiz(3,16) → 2·∛2`, `((3x^(3/2)y^3)/(x^2y^(-1/2)))^-2 → x/(9y^7)`.
- **Racionalización con `/radicalizar`**: multiplica por el conjugado cuando el numerador o el denominador es un binomio con raíces cuadradas, simplifica el resultado y cancela factores comunes. Ejemplos: `sqrt(10)/(sqrt(5)-2) → 5·√2 + 2·√10`, `(sqrt(4+h)-2)/h → 1/(√(4 + h) + 2)`.
- **Comando `/limpiar` y atajo `Ctrl+L`**: borran objetos, historial, salida, inputs y undo/redo de la sesión actual. Funcionan desde cualquier panel y piden confirmación `s/N`; al confirmar, la sesión se guarda vacía y se conserva su nombre.
- **Desarrollo de productos notables**: `/desarrollar` expande productos y potencias enteras de sumas y muestra la fórmula reconocida: cuadrado/cubo de un binomio, suma por diferencia, producto de binomios, cuadrado de un trinomio y binomio de Newton.
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
- [golangci-lint](https://golangci-lint.run/) v2 o superior (solo para `make lint`, `make lint-fix` y `make fmt`)

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
| `Ctrl+L`         | Abrir la confirmación de `/limpiar` |
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

### Álgebra exacta

`/potencia` y `/radical` trabajan sobre la expresión completa. Declara primero:

```text
/expresion (3a^3b^3)(4ab^2)^2
/potencia
→ 48*a^5*b^7

/expresion sqrt(200)-sqrt(32)
/radical
→ 6*sqrt(2)

/expresion ((3x^(3/2)y^3)/(x^2y^(-1/2)))^-2
/potencia
→ x/(9*y^7)
```

`/radical` simplifica las raíces que ya contiene la expresión; para extraer una raíz nueva, declárala con `raiz(expr)` o `raiz(n, expr)`. La raíz cuadrada se escribe `raiz(expr)` o `sqrt(expr)`. Las raíces n-ésimas se escriben `raiz(n, expr)` o `x^(1/n)`:

```text
/expresion raiz(3, 16)
/radical
→ 2*raiz(3, 2)   (Unicode: 2·∛2)
```

`/radicalizar` se usa cuando la fracción tiene un binomio con raíces en el
numerador o en el denominador. Multiplica por el conjugado, simplifica y
cancela los factores comunes:

```text
/expresion sqrt(10)/(sqrt(5)-2)
/radicalizar
→ 5*sqrt(2) + 2*sqrt(10)   (Unicode: 5·√2 + 2·√10)

/expresion (sqrt(4+h)-2)/h
/radicalizar
→ 1/(sqrt(4 + h) + 2)   (Unicode: 1/(√(4 + h) + 2))
```

El resultado puede seguir conteniendo raíces; lo importante es que el
binomio original se ha racionalizado y que los factores comunes se han
cancelado.

Para desarrollar productos notables:

```text
/expresion (a+b)^2
/desarrollar
→ a^2 + 2*a*b + b^2   (Unicode: a² + 2·a·b + b²)

/expresion (a+b)(a-b)
/desarrollar
→ a^2 - b^2

/expresion (a+b)^3
/desarrollar
→ a^3 + 3*a^2*b + 3*a*b^2 + b^3
```

La factorización usa `/factorizar` sobre la expresión declarada. Para polinomios:

```text
/expresion 4x^2-25
/factorizar
→ (2*x - 5)*(2*x + 5)

/expresion x^3-3x^2-4x+12
/factorizar
→ (x - 3)*(x - 2)*(x + 2)

/expresion x^3y-4xy
/factorizar
→ x*y*(x - 2)*(x + 2)
```

Para expresiones racionales, `/factorizar` combina las operaciones, factoriza
el numerador y el denominador y cancela los factores comunes:

```text
/expresion (x^2+3x+2)/(x^2-x-2)
/factorizar
→ (x + 2)/(x - 2)

/expresion ((2x^2-x-1)/(x^2-9))*((x+3)/(2x+1))
/factorizar
→ (x - 1)/(x - 3)

/expresion x^2/(x^2-4)-(x+1)/(x+2)
/factorizar
→ 1/(x - 2)

/expresion (y/x-x/y)/(1/y-1/x)
/factorizar
→ -x - y
```

Para completar el cuadrado de una cuadrática, añade la bandera
`-cuadrado`:

```text
/expresion x^2+x+1
/factorizar -cuadrado
→ (x + 1/2)^2 + 3/4

/expresion 2x^2-12x+11
/factorizar -cuadrado
→ 2*(x - 3)^2 - 7
```

La salida muestra los coeficientes, la forma mónica cuando el coeficiente
principal no es 1, el término que se suma y resta, la forma completada y el
vértice. Solo se admiten polinomios de grado 2 en una variable con
coeficientes racionales.

El orden de los factores puede variar; el producto es equivalente.
La cancelación se apoya en los patrones de factorización soportados; un factor
común que no encaje en ellos puede quedar sin cancelar.

El motor exacto no usa `float64`: los coeficientes y exponentes son fracciones de `math/big.Rat`. No soporta funciones trascendentes (`sin`, `cos`, `ln`), constantes irracionales (`pi`, `inf`) ni números complejos. `/potencia`, `/radical` y `/desarrollar` no admiten división por polinomios que no sean monomios; `/radicalizar` admite fracciones con binomios de raíces cuadradas y `/factorizar` trata expresiones racionales con exponentes enteros. `/factorizar -cuadrado` requiere una cuadrática univariada con coeficientes racionales. Para simplificar raíces pares se asume que las variables son positivas (`sqrt(x^2) = x`).

Las variables pueden tener una letra (`x`) o tres o más letras (`alpha`, `beta`, `tasa`). Por compatibilidad con la notación algebraica, dos letras seguidas siguen siendo un producto implícito: `ab = a*b`. Para multiplicar nombres largos usa `*` o un espacio (`alpha*beta`); antes de un paréntesis escribe `alpha*(beta+1)`. `raiz(expr)` es la raíz cuadrada y `raiz(n, expr)` la raíz n-ésima.

### Limpiar el espacio de trabajo

El comando global `/limpiar` borra los objetos, el historial, la salida, los inputs y el undo/redo de la sesión actual. Funciona desde cualquier panel y pide confirmación:

```text
/limpiar
¿Limpiar todo?  s = confirmar · n/Esc = cancelar
```

Al confirmar se conserva el nombre de la sesión y se guarda vacía inmediatamente. Esta acción no se puede deshacer con `Ctrl+Z`; si necesitas conservar el trabajo anterior, cancela y usa `Ctrl+O` para duplicar la sesión antes de limpiar.

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
- `/factorizar` — `/factorizar ?` — factorización de enteros (`/factorizar 120`), polinomios o expresiones racionales (`/factorizar p1`), con factorización del numerador y del denominador y cancelación de factores comunes
- `/factorizar -cuadrado` — completa el cuadrado de una cuadrática univariada y muestra el vértice
- `/desarrollar` — `/desarrollar ?` — desarrollo de productos notables y potencias enteras de sumas
- `/potencia` — `/potencia ?` — simplificación exacta de productos, cocientes y potencias de una expresión o función
- `/radical` — `/radical ?` — simplificación exacta de raíces y exponentes fraccionarios de una expresión o función
- `/radicalizar` — `/radicalizar ?` — racionalización por conjugado de fracciones con binomios de raíces cuadradas
- `/limpiar` — `/limpiar` — comando global: borra objetos, historial, salida, inputs y undo/redo de la sesión actual (confirmación `s/N`)

Ambos métodos de interpolación producen el mismo polinomio único de grado ≤ n para n+1 puntos.

La factorización también se ejecuta sola al declarar una expresión que sea un entero literal:

```text
/expresion 120
120
→ 120 = 2³ · 3 · 5
```

## Calidad

La configuración de `golangci-lint` vive en `.golangci.yml` y cubre corrección,
estilo, complejidad, seguridad y modernización del código. El CI la ejecuta en
cada push y pull request.

```bash
make lint      # analiza todo el proyecto
make lint-fix  # aplica las correcciones automáticas
make fmt       # aplica gofumpt y goimports
```

## Estructura del proyecto

```
.
├── cmd/tuniversalcalc/     # Punto de entrada (flags, sesión y bootstrap de Bubble Tea)
└── internal/
    ├── ui/                 # Interfaz: model, update, view, styles, templates, messages,
    │                       # session, sessionpicker, undo, help, completion, preview,
    │                       # textutil, mouse, clipboard, palette
    ├── engine/             # Declaraciones + interpolación + parser simbólico + motor exacto (exact_*.go)
    └── render/             # Renderizado Unicode y LaTeX emulado (superíndices, raíces, ∫, π, ∞)
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