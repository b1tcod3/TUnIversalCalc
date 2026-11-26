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
	{Trigger: "desarrollar", InsertText: "/desarrollar ?", Desc: "Desarrollo de productos notables"},
	{Trigger: "factorizar", InsertText: "/factorizar ?", Desc: "Factorización de enteros, polinomios o fracciones"},
	{Trigger: "lagrange", InsertText: "/lagrange ?", Desc: "Interpolación de Lagrange"},
	{Trigger: "limpiar", InsertText: "/limpiar", Desc: "Borra todo el espacio de trabajo (pide confirmación)"},
	{Trigger: "newton", InsertText: "/newton ?", Desc: "Interpolación de Newton (diferencias divididas)"},
	{Trigger: "potencia", InsertText: "/potencia ?", Desc: "Simplificación exacta de potencias"},
	{Trigger: "radical", InsertText: "/radical ?", Desc: "Simplificación exacta de radicales"},
	{Trigger: "radicalizar", InsertText: "/radicalizar ?", Desc: "Racionalización de fracciones con raíces"},
}

// Registro de plantillas de DECLARACIÓN (panel 1). Los objetos se nombran
// solos: /ecuacion 2x+2=0 → e1, /funcion 2x+2 → f1, /expresion x^2 → p1,
// /puntos (1,2) (3,4) → t1.
var declRegistry = []Template{
	{Trigger: "ecuacion", InsertText: "/ecuacion ?", Desc: "Ecuación (auto-e1): ej. 2x+2=0"},
	{Trigger: "expresion", InsertText: "/expresion ?", Desc: "Expresión (auto-p1): ej. x^2+1"},
	{Trigger: "funcion", InsertText: "/funcion ?", Desc: "Función (auto-f1): ej. 2x+2"},
	{Trigger: "limpiar", InsertText: "/limpiar", Desc: "Borra todo el espacio de trabajo (pide confirmación)"},
	{Trigger: "puntos", InsertText: "/puntos ?", Desc: "Puntos (auto-t1): ej. (1,2) (3,4)"},
}

// matchTemplate busca una plantilla por prefijo (mínimo 3 caracteres) para
// autocompletar gratis: "/lag" → lagrange.
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
