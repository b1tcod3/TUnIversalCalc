# Iniciando el proyecto

1.Inicializar el proyecto:
Crea el entorno y el módulo.
Abre tu terminal y ejecuta:
```
Bash
mkdir tuniversarlcalc && cd tuniversarlcalc
go mod init tuniversarlcalc
```

2. Instalar el framework TUI: 
Bubble Tea y Lipgloss para estilos.
Añade la dependencia principal de la interfaz y la librería de estilos (para darle bordes y color a la calculadora):

```Bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
```
