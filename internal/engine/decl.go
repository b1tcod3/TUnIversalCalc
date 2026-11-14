package engine

import (
	"fmt"
	"strconv"
	"strings"

	"tuniversalcalc/internal/render"
)

// DeclType clasifica los objetos declarados en el panel 1.
type DeclType int

const (
	DeclEcuacion DeclType = iota
	DeclFuncion
	DeclExpresion
	DeclPuntos
)

func (t DeclType) String() string {
	switch t {
	case DeclEcuacion:
		return "ecuación"
	case DeclFuncion:
		return "función"
	case DeclPuntos:
		return "puntos"
	}
	return "expresión"
}

// Prefix es el prefijo del nombre automático: e1, f1, p1, t1...
func (t DeclType) Prefix() string {
	switch t {
	case DeclEcuacion:
		return "e"
	case DeclFuncion:
		return "f"
	case DeclPuntos:
		return "t"
	}
	return "p"
}

// Point representa un par de coordenadas para interpolación.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Decl es un objeto matemático declarado por el usuario. Nombre lo asigna
// la UI automáticamente (auto-nombre: e1, f1, p1, t1...) al registrarlo.
type Decl struct {
	Tipo   DeclType `json:"tipo"`
	Nombre string   `json:"nombre"`
	Raw    string   `json:"raw"`    // contenido tal como se escribió (con "=" para ecuaciones)
	Latex  string   `json:"latex"`  // emulación LaTeX para la salida
	Pretty string   `json:"pretty"` // versión Unicode legible
	Points []Point  `json:"points"` // puntos para DeclPuntos
}

// ParseDecl valida y clasifica una declaración del panel 1. Acepta
// slash-commands y, como atajo, un entero literal suelto. Devuelve un error
// "inteligente" si no reconoce la entrada. El nombre se asigna después
// (auto-nombre).
func ParseDecl(raw string) (Decl, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Decl{}, fmt.Errorf("declaración vacía: usa /ecuacion, /funcion o /expresion")
	}
	// Un entero literal suelto se considera una expresión: es la vía más
	// directa para factorizar (120 → p1 = 120).
	if n, ok := ParseIntegerLiteral(s); ok {
		return IntegerDecl(n), nil
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
		if _, err := ParseExpr(content); err != nil {
			return Decl{}, err
		}
	case strings.HasPrefix(cmd, "expres"):
		d.Tipo = DeclExpresion
		if strings.Contains(content, "=") {
			return Decl{}, fmt.Errorf("una expresión no lleva '=': /expresion x^2+1")
		}
		if _, err := ParseExpr(content); err != nil {
			return Decl{}, err
		}
	case strings.HasPrefix(cmd, "punt"):
		d.Tipo = DeclPuntos
		points, err := parsePoints(content)
		if err != nil {
			return Decl{}, err
		}
		d.Points = points
	default:
		return Decl{}, fmt.Errorf("comando no reconocido: %q — usa /ecuacion, /funcion, /expresion o /puntos", cmd)
	}

	d.Latex = render.ToLatex(content)
	d.Pretty = render.Pretty(content)
	return d, nil
}

// parsePoints interpreta una lista de puntos en formato "(x,y)" separados
// por espacios: /puntos (1,2) (3,4) (5,6). Se exige al menos 2 puntos y
// abscisas distintas (con abscisas repetidas no existe polinomio único).
func parsePoints(content string) ([]Point, error) {
	groups := splitPointGroups(content)
	if len(groups) < 2 {
		return nil, fmt.Errorf("se necesitan al menos 2 puntos: ej. /puntos (1,2) (3,4)")
	}
	pts := make([]Point, 0, len(groups))
	xs := map[float64]bool{}
	for _, g := range groups {
		x, y, err := parsePoint(g)
		if err != nil {
			return nil, err
		}
		if xs[x] {
			return nil, fmt.Errorf("abscisa repetida x=%g: los puntos deben tener x distintas", x)
		}
		xs[x] = true
		pts = append(pts, Point{X: x, Y: y})
	}
	return pts, nil
}

// splitPointGroups separa el contenido en grupos "(x,y)". Cada grupo envuelve
// un punto con paréntesis; los que no, se agrupan por pares de números.
func splitPointGroups(content string) []string {
	s := strings.TrimSpace(content)
	if s == "" {
		return nil
	}
	var groups []string
	var cur strings.Builder
	flush := func() {
		rest := strings.Fields(cur.String())
		cur.Reset()
		for _, r := range rest {
			groups = append(groups, r)
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '(' {
			flush()
			depth := 1
			j := i + 1
			for j < len(s) && depth > 0 {
				if s[j] == '(' {
					depth++
				} else if s[j] == ')' {
					depth--
				}
				j++
			}
			groups = append(groups, strings.TrimSpace(s[i:j]))
			i = j - 1
			continue
		}
		cur.WriteByte(c)
	}
	flush()
	return groups
}

// parsePoint extrae (x,y) de un grupo; acepta "1,2", "(1,2)", "1 , 2".
func parsePoint(g string) (float64, float64, error) {
	inner := strings.TrimSpace(strings.Trim(g, "()"))
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("punto mal formado: %q — usa (x,y)", g)
	}
	x, errX := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, errY := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if errX != nil || errY != nil {
		return 0, 0, fmt.Errorf("punto mal formado: %q — usa números como (1,2)", g)
	}
	return x, y, nil
}

// validEquation valida la ecuación con el parser simbólico real.
func validEquation(content string) error {
	if _, _, err := ParseEquation(content); err != nil {
		return err
	}
	return nil
}
