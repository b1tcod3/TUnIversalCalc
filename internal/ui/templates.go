package ui

import "strings"

// Template es una plantilla editable con marcadores "?" que el usuario
// recorre con Shift+Tab (estilo snippets de VSCode).
type Template struct {
	Trigger    string
	InsertText string
	Desc       string
}

// Registro de plantillas de OPERACIÓN (panel 2). Los objetos no hacen
// falta citarlos: si se omite el nombre, se opera sobre el último.
var opsRegistry = []Template{
	{Trigger: "resolver", InsertText: "/resolver ? ?", Desc: "Resolver ecuación: objeto (opcional) y variable"},
	{Trigger: "integral", InsertText: "/integral ?", Desc: "Integral del objeto indicado (o el último)"},
	{Trigger: "derivar", InsertText: "/derivar ?", Desc: "Derivada del objeto respecto a la variable"},
	{Trigger: "limite", InsertText: "/limite ?", Desc: "Límite cuando la variable → valor"},
	{Trigger: "simplificar", InsertText: "/simplificar ?", Desc: "Simplificar la expresión del objeto"},
	{Trigger: "matriz", InsertText: "/matriz ?", Desc: "Operaciones con matrices"},
}

// Registro de plantillas de DECLARACIÓN (panel 1). Los objetos se nombran
// solos: /ecuacion 2x+2=0 → e1, /funcion 2x+2 → f1, /expresion x^2 → p1.
var declRegistry = []Template{
	{Trigger: "ecuacion", InsertText: "/ecuacion ?", Desc: "Ecuación (auto-e1): ej. 2x+2=0"},
	{Trigger: "funcion", InsertText: "/funcion ?", Desc: "Función (auto-f1): ej. 2x+2"},
	{Trigger: "expresion", InsertText: "/expresion ?", Desc: "Expresión (auto-p1): ej. x^2+1"},
}

// matchTemplate busca una plantilla por prefijo (mínimo 3 caracteres) para
// autocompletar gratis: "/int" → integral.
func matchTemplate(list []Template, input string) (Template, bool) {
	s := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(input, "/")))
	if len(s) < 3 {
		return Template{}, false
	}
	for _, t := range list {
		if strings.HasPrefix(t.Trigger, s) {
			return t, true
		}
	}
	return Template{}, false
}

// MatchTemplate busca en el registro de operaciones (panel 2).
func MatchTemplate(input string) (Template, bool) {
	return matchTemplate(opsRegistry, input)
}

// HasFields indica si la plantilla aún tiene marcadores "?" por completar.
func HasFields(val string) bool {
	return strings.Contains(val, "?")
}

// JumpNextField elimina el siguiente "?" a partir de la posición del cursor
// (con wrap al inicio si no hay más adelante) y devuelve el nuevo valor con
// la posición donde colocar el cursor. No guarda estado: escanea el valor
// actual, por lo que es robusto ante ediciones y borrados.
func JumpNextField(val string, cursor int) (string, int) {
	idx := strings.Index(val[cursor:], "?")
	if idx < 0 {
		idx = strings.Index(val, "?")
		if idx < 0 {
			return val, cursor
		}
		return val[:idx] + val[idx+1:], idx
	}
	pos := cursor + idx
	return val[:pos] + val[pos+1:], pos
}