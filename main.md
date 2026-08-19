# main.go

Es el punto de entrada de la arquitectura Elm.

El archivo vive en `cmd/tuniversalcalc/main.go` (estructura estándar Go: el código
de la aplicación va en `internal/` y el binario se construye desde `cmd/`).

Crea un archivo llamado main.go en cmd/tuniversalcalc/ de tu proyecto tuniversalcalc y pega el siguiente código. Este es el boilerplate funcional que dibuja la caja, captura las teclas y maneja el estado de forma reactiva.

Código Base (main.go)

Go

Esta dividida en varias secciones:

1. Package Main
```
package main
```

para lanzar la función principal main()

2. Bloque de imports

```
import (
    "log"
    "os"
    "tuniversalcalc/ui"
    tea "github.com/charmbracelet/bubbletea"
)
```

- log: Paquete estándar para registro de errores. Aquí se usa log.Fatalf para imprimir un mensaje y terminar el programa.

- os: Paquete estándar para interactuar con el sistema operativo. Se usa os.Exit(1) para salir con un código de error.

- "tuniversalcalc/internal/ui": Paquete propio del proyecto.

- tea "github.com/charmbracelet/bubbletea": La librería Bubble Tea, importada con el alias tea. Proporciona el runtime que ejecuta el bucle principal de la arquitectura Elm.

3. Función Main

Es el corazón del programa. Realiza dos tareas: crear el programa Bubble Tea y ejecutarlo.

4. Inicializar BubleTea

```
p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
```

- tea.NewProgram crea una nueva instancia del runtime de Bubble Tea. Su primer argumento es el modelo inicial (el estado de la aplicación), que debe cumplir la interfaz tea.Model (métodos Init, Update y View).
ui.NewModel() llama a una función constructora definida en el paquete ui que devuelve ese modelo con sus valores por defecto (pantalla actual, números, etc.).

- El segundo argumento es tea.WithAltScreen(), una opción de configuración. Activa el modo de pantalla alternativa del terminal. Esto significa que la app se ejecutará en un buffer separado; al salir, el terminal vuelve a su estado original sin dejar rastros de la interfaz en el historial de scroll.

- La variable p es un puntero a tea.Program, el objeto que controla el ciclo de vida.

5. Ejecutar BubleTea

```
if _, err := p.Run(); err != nil { ... }
```

- p.Run() inicia el bucle principal de eventos. Es una llamada bloqueante: se queda escuchando entradas del teclado, temporizadores y comandos hasta que se recibe un mensaje tea.Quit. Retorna el modelo final (que aquí ignoramos con _) y un posible error.

- Si err != nil, significa que algo falló durante la ejecución (problemas de terminal, etc.). Entonces:

    - log.Fatalf(...) imprime un mensaje de error con formato y llama a os.Exit(1) automáticamente.

    - Aunque log.Fatalf ya termina el programa, el código incluye explícitamente os.Exit(1) por claridad o por si se elimina el log. Es redundante pero inofensivo.




Ejecutar la aplicación

Una vez guardado el archivo, compila y ejecuta el programa directamente desde tu terminal con:
Bash
```bash
go run main.go
```
