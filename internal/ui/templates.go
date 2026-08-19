package ui

import "strings"

// Template es una plantilla editable con marcadores "?" que el usuario
// recorre con Shift+Tab (estilo snippets de VSCode).
type Template struct {
	Trigger    string
	InsertText string
	Desc       string
}

// Registro de plantillas por prefijo: "/int", "/der" completan solos.
var registry = []Template{
	{Trigger: "integral", InsertText: "integral( ?, ?, dx )", Desc: "Integral: argumento, límite inferior, superior o var"},
	{Trigger: "derivar", InsertText: "derivar( ?, ? )", Desc: "Derivada de f(x) respecto a la variable"},
	{Trigger: "limite", InsertText: "limite( x, ?, ? )", Desc: "Límite de f(x) cuando x → valor"},
	{Trigger: "resolver", InsertText: "resolver( ?, ? )", Desc: "Resolver ecuación para la variable"},
	{Trigger: "simplificar", InsertText: "simplificar( ? )", Desc: "Simplificar expresión"},
	{Trigger: "matriz", InsertText: "matriz( [[?, ?],[?, ?]] )", Desc: "Operaciones con matrices"},
}

// MatchTemplate busca una plantilla por prefijo (mínimo 3 caracteres) para
// autocompletar gratis: "/int" → integral.
func MatchTemplate(input string) (Template, bool) {
	s := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(input, "/")))
	if len(s) < 3 {
		return Template{}, false
	}
	for _, t := range registry {
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
