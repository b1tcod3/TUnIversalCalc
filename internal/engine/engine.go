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

	op := normalizeOp(cmd)
	if op == "" {
		return nil, errors.New("comando vacío en el panel 2")
	}

	// Una variable simbólica se representa con una sola letra. Si el
	// comando trajo algo raro como variable (p.ej. /limite f1 2), se
	// usa x.
	v := byte('x')
	if len(varName) == 1 && isLetter(varName[0]) {
		v = varName[0]
	}

	steps := []string{
		"Objeto: " + decl.Nombre + "  —  " + decl.Latex,
	}
	tail, err := evalOp(decl, op, cmd, v)
	if err != nil {
		return nil, err
	}
	return append(steps, tail...), nil
}

// normalizeOp limpia el comando del panel 2: minúsculas, sin "/" inicial
// ni paréntesis o comas.
func normalizeOp(cmd string) string {
	op := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cmd, "/")))
	op = strings.NewReplacer("(", "", ")", "", ",", "").Replace(op)
	return strings.TrimSpace(op)
}

// evalOp despacha la operación reconocida y devuelve los pasos que siguen
// a la cabecera común del objeto.
func evalOp(decl Decl, op, cmd string, v byte) ([]string, error) {
	switch {
	case strings.HasPrefix(op, "integral"):
		return evalIntegral(decl, v)
	case strings.HasPrefix(op, "deriv"):
		return evalDerivada(decl, v)
	case strings.HasPrefix(op, "limit"):
		return evalLimite(decl, cmd, v)
	case strings.HasPrefix(op, "resolver"):
		return evalResolver(decl, v)
	case strings.HasPrefix(op, "simplificar"):
		return evalSimplificar(decl)
	case strings.HasPrefix(op, "matriz"):
		return []string{
			"Operación: matriz",
			"El parser simbólico aún no soporta matrices: declara un sistema como /ecuacion o usa una operación escalar",
		}, nil
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
}

func evalIntegral(decl Decl, v byte) ([]string, error) {
	expr, err := parseContent(decl)
	if err != nil {
		return nil, err
	}
	steps := []string{"Operación: integral indefinida respecto a " + string(v)}
	res, sir, err := Integrate(expr, v)
	if err != nil {
		return nil, err
	}
	steps = append(steps, sir...)
	steps = append(steps,
		"∫ ("+decl.Pretty+") d"+string(v)+"  →  "+res.Ascii()+" + C")
	return steps, nil
}

func evalDerivada(decl Decl, v byte) ([]string, error) {
	expr, err := parseContent(decl)
	if err != nil {
		return nil, err
	}
	steps := []string{"Operación: derivada respecto a " + string(v)}
	res, dr, err := Derive(expr, v)
	if err != nil {
		return nil, err
	}
	steps = append(steps, dr...)
	steps = append(steps,
		"d/d"+string(v)+" ("+decl.Pretty+")  →  "+decl.Nombre+"′("+string(v)+") = "+res.Ascii())
	return steps, nil
}

func evalLimite(decl Decl, cmd string, v byte) ([]string, error) {
	expr, err := parseContent(decl)
	if err != nil {
		return nil, err
	}
	target, err := limitTarget(cmd)
	if err != nil {
		return nil, err
	}
	steps := []string{"Operación: límite"}
	res, lr, err := Limit(expr, v, target)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lr...)
	steps = append(steps,
		"lim "+string(v)+"→"+target.String()+" ("+decl.Pretty+")  →  "+res.Ascii())
	return steps, nil
}

func evalResolver(decl Decl, v byte) ([]string, error) {
	lhs, rhs, err := parseDeclEquation(decl)
	if err != nil {
		return nil, err
	}
	sr, err := Solve(lhs, rhs, v)
	if err != nil {
		return nil, err
	}
	return append([]string{"Operación: resolver ecuación"}, sr...), nil
}

func evalSimplificar(decl Decl) ([]string, error) {
	expr, err := parseContent(decl)
	if err != nil {
		return nil, err
	}
	res, ps, err := Simplify(expr)
	if err != nil {
		return nil, err
	}
	steps := append([]string{"Operación: simplificar"}, ps...)
	return append(steps, "Resultado: "+res.Ascii()), nil
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
