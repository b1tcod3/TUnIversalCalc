package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Evaluate es la interfaz asíncrona del motor: recibe el objeto declarado
// (sobre el que ya resolvió la UI), el comando del panel 2 y la variable,
// y devuelve los pasos del cálculo para el panel de salida.
//
// El motor es simbólico: parsea el objeto con el parser Pratt real y las
// operaciones (derivar, simplificar, resolver, integral, límite) trabajan
// sobre el AST y muestran las reglas aplicadas. /matriz y las formas no
// soportadas devuelven stubs explícitos.
func Evaluate(decl Decl, cmd, varName string) ([]string, error) {
	if decl.Nombre == "" {
		return nil, errors.New("objeto vacío: declara uno en el panel 1 (/ecuacion 2x+2=0)")
	}
	if varName == "" {
		varName = "x"
	}

	op := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cmd, "/")))
	op = strings.NewReplacer("(", "", ")", "", ",", "").Replace(op)
	op = strings.TrimSpace(op)
	if op == "" {
		return nil, errors.New("comando vacío en el panel 2")
	}

	pretty, latex := decl.Pretty, decl.Latex
	steps := []string{
		"Objeto: " + decl.Nombre + "  —  " + latex,
	}

	// Una variable simbólica se representa con una sola letra. Si el
	// comando trajo algo raro como variable (p.ej. /limite f1 2), se
	// usa x.
	v := byte('x')
	if len(varName) == 1 && isLetter(varName[0]) {
		v = varName[0]
	}

	switch {
	case strings.HasPrefix(op, "integral"):
		expr, err := parseContent(decl)
		if err != nil {
			return nil, err
		}
		steps = append(steps, "Operación: integral indefinida respecto a "+string(v))
		res, sir, err := Integrate(expr, v)
		if err != nil {
			return nil, err
		}
		steps = append(steps, sir...)
		steps = append(steps,
			"∫ ("+pretty+") d"+string(v)+"  →  "+res.Ascii()+" + C")

	case strings.HasPrefix(op, "deriv"):
		expr, err := parseContent(decl)
		if err != nil {
			return nil, err
		}
		steps = append(steps, "Operación: derivada respecto a "+string(v))
		res, dr, err := Derive(expr, v)
		if err != nil {
			return nil, err
		}
		steps = append(steps, dr...)
		steps = append(steps,
			"d/d"+string(v)+" ("+pretty+")  →  "+decl.Nombre+"′("+string(v)+") = "+res.Ascii())

	case strings.HasPrefix(op, "limit"):
		expr, err := parseContent(decl)
		if err != nil {
			return nil, err
		}
		target, err := limitTarget(cmd)
		if err != nil {
			return nil, err
		}
		steps = append(steps, "Operación: límite")
		res, lr, err := Limit(expr, v, target)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lr...)
		steps = append(steps,
			"lim "+string(v)+"→"+target.String()+" ("+pretty+")  →  "+res.Ascii())

	case strings.HasPrefix(op, "resolver"):
		steps = append(steps, "Operación: resolver ecuación")
		lhs, rhs, err := parseDeclEquation(decl)
		if err != nil {
			return nil, err
		}
		sr, err := Solve(lhs, rhs, v)
		if err != nil {
			return nil, err
		}
		steps = append(steps, sr...)

	case strings.HasPrefix(op, "simplificar"):
		expr, err := parseContent(decl)
		if err != nil {
			return nil, err
		}
		steps = append(steps, "Operación: simplificar")
		res, ps, err := Simplify(expr)
		if err != nil {
			return nil, err
		}
		steps = append(steps, ps...)
		steps = append(steps, "Resultado: "+res.Ascii())

	case strings.HasPrefix(op, "matriz"):
		steps = append(steps,
			"Operación: matriz",
			"El parser simbólico aún no soporta matrices: declara un sistema como /ecuacion o usa una operación escalar",
		)

	default:
		name := op
		if f := strings.Fields(op); len(f) > 0 {
			name = f[0]
		}
		return nil, fmt.Errorf(
			"operación no reconocida: %q — usa /resolver, /integral, /derivar, /limite, /simplificar o /matriz",
			name,
		)
	}

	return steps, nil
}

// parseContent valida y parsea el contenido del objeto (función o expresión).
func parseContent(decl Decl) (Expr, error) {
	if decl.Tipo == DeclEcuacion {
		return nil, errors.New("el objeto es una ecuación: usa /resolver para despejar la variable")
	}
	return ParseExpr(decl.Raw)
}

// parseDeclEquation valida y parsea una ecuación para su resolución.
func parseDeclEquation(decl Decl) (Expr, Expr, error) {
	if decl.Tipo != DeclEcuacion {
		expr, err := ParseExpr(decl.Raw)
		if err != nil {
			return nil, nil, err
		}
		return expr, nil, nil
	}
	return ParseEquation(decl.Raw)
}

// limitTarget extrae el punto del comando /limite buscando el primer token
// numérico o "inf" tras la operación: /limite f1 2, /limite f1 x 2,
// /limite f1 x inf. Por defecto x→0.
func limitTarget(cmd string) (Value, error) {
	norm := strings.NewReplacer("(", "", ")", "", ",", "").Replace(cmd)
	toks := strings.Fields(norm)
	for _, t := range toks[2:] { // /limite objeto? var? valor
		switch t {
		case "inf", "infty":
			return Value{IsInf: true, Sign: 1}, nil
		case "-inf":
			return Value{IsInf: true, Sign: -1}, nil
		}
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return Value{Val: f}, nil
		}
	}
	return Value{Val: 0}, nil
}
