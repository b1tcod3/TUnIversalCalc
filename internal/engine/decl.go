package engine

import (
	"fmt"
	"strings"

	"tuniversalcalc/internal/render"
)

// DeclType clasifica los objetos declarados en el panel 1.
type DeclType int

const (
	DeclEcuacion DeclType = iota
	DeclFuncion
	DeclExpresion
)

func (t DeclType) String() string {
	switch t {
	case DeclEcuacion:
		return "ecuación"
	case DeclFuncion:
		return "función"
	}
	return "expresión"
}

// Prefix es el prefijo del nombre automático: e1, f1, p1...
func (t DeclType) Prefix() string {
	switch t {
	case DeclEcuacion:
		return "e"
	case DeclFuncion:
		return "f"
	}
	return "p"
}

// Decl es un objeto matemático declarado por el usuario. Nombre lo asigna
// la UI automáticamente (auto-nombre: e1, f1, p1...) al registrarlo.
type Decl struct {
	Tipo   DeclType
	Nombre string
	Raw    string // contenido tal como se escribió (con "=" para ecuaciones)
	Latex  string // emulación LaTeX para la salida
	Pretty string // versión Unicode legible
}

// ParseDecl valida y clasifica una declaración del panel 1. Siempre debe
// llevar slash-command y devuelve un error "inteligente" si no lo lleva
// o el comando no existe. El nombre se asigna después (auto-nombre).
func ParseDecl(raw string) (Decl, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Decl{}, fmt.Errorf("declaración vacía: usa /ecuacion, /funcion o /expresion")
	}
	if !strings.HasPrefix(s, "/") {
		return Decl{}, fmt.Errorf("el panel 1 es de declaración: usa /ecuacion, /funcion o /expresion")
	}

	fields := strings.Fields(s)
	cmd := strings.ToLower(strings.TrimPrefix(fields[0], "/"))
	content := strings.TrimSpace(s[len(fields[0]):])
	if content == "" {
		return Decl{}, fmt.Errorf("comando %s sin contenido: ej. /ecuacion 2x+2=0", cmd)
	}

	d := Decl{Raw: content}
	switch {
	case strings.HasPrefix(cmd, "ecuac"):
		d.Tipo = DeclEcuacion
		if err := validEquation(content); err != nil {
			return Decl{}, err
		}
	case strings.HasPrefix(cmd, "func"):
		d.Tipo = DeclFuncion
		if strings.Contains(content, "=") {
			return Decl{}, fmt.Errorf("una función no lleva '=': /funcion 2x+2")
		}
		if err := Validate(content); err != nil {
			return Decl{}, err
		}
	case strings.HasPrefix(cmd, "expres"):
		d.Tipo = DeclExpresion
		if strings.Contains(content, "=") {
			return Decl{}, fmt.Errorf("una expresión no lleva '=': /expresion x^2+1")
		}
		if err := Validate(content); err != nil {
			return Decl{}, err
		}
	default:
		return Decl{}, fmt.Errorf("comando no reconocido: %q — usa /ecuacion, /funcion o /expresion", cmd)
	}

	d.Latex = render.ToLatex(content)
	d.Pretty = render.Pretty(content)
	return d, nil
}

// validEquation exige exactamente un '=' con ambos lados válidos.
func validEquation(content string) error {
	sides := strings.Split(content, "=")
	if len(sides) != 2 {
		return fmt.Errorf("una ecuación necesita exactamente un '=': /ecuacion 2x+2=0")
	}
	if err := Validate(sides[0]); err != nil {
		return fmt.Errorf("lado izquierdo inválido: %v", err)
	}
	if err := Validate(sides[1]); err != nil {
		return fmt.Errorf("lado derecho inválido: %v", err)
	}
	return nil
}
