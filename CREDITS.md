# Créditos y atribuciones

## Proyecto

**TUnIversalCalc** — calculadora TUI en Go.
Repositorio: <https://github.com/b1tcod3/TUnIversalCalc>

## Inspiración

La experiencia de usuario de este proyecto se inspiró en ideas de
**[nascTUI](https://github.com/parnoldx/nascTUI)**, de
[@parnoldx](https://github.com/parnoldx), publicado bajo **GPL-2.0**.

Concretamente se tomaron como referencia conceptual:

- la persistencia de sesiones con autosave;
- el sistema de undo/redo;
- la ayuda contextual con scroll;
- el autocompletado de funciones y variables;
- el soporte de ratón y el pegado multilínea.

**Este repositorio no incorpora código fuente de nascTUI.** Las ideas se
reimplementaron de forma independiente para `tuniversalcalc`. Si en el futuro
se copia código de nascTUI o se enlaza `libqalculate`, el proyecto completo
deberá distribuirse bajo GPL-2.0 y conservar los avisos de copyright
correspondientes; una sección de créditos no sustituye a la licencia.

## Dependencias

| Componente | Uso en el proyecto | Autor | Licencia |
| --- | --- | --- | --- |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Runtime TUI y arquitectura Elm | Charmbracelet, Inc. | MIT |
| [Lipgloss](https://github.com/charmbracelet/lipgloss) | Estilos, paneles y layout | Charmbracelet, Inc. | MIT |
| [Bubbles](https://github.com/charmbracelet/bubbles) | `textinput`, `viewport` y componentes reutilizables | Charmbracelet, Inc. | MIT |
| [x/ansi](https://github.com/charmbracelet/x) | Ancho y recorte de texto con ANSI | Charmbracelet, Inc. | MIT |
| [fuzzysearch](https://github.com/lithammer/fuzzysearch) | Coincidencia difusa de la paleta | Peter Lithammer | MIT |
| [atotto/clipboard](https://github.com/atotto/clipboard) | Copiado al portapapeles | Ato Araki | BSD-3-Clause |
| [Go](https://go.dev) | Lenguaje y toolchain | The Go Authors | BSD-3-Clause |

Los textos completos de las licencias están en los repositorios enlazados y
en el módulo correspondiente de `go.sum`.

## Herramientas

- El banner ASCII se generó con [`toilet`](https://github.com/cacalabs/toilet)
  usando la fuente `smblock`.
- Las pruebas de integración de la TUI usan las utilidades de Bubble Tea. La
  incorporación de `teatest` para pruebas de extremo a extremo queda como
  trabajo futuro.
