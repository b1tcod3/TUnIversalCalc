package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"tuniversalcalc/internal/render"
)

// --- Interpolación polinómica (cálculo numérico) ---
//
// Dado un conjunto de n+1 puntos distintos, existe un único polinomio de
// grado <= n que pasa por todos ellos. Lagrange y Newton son caminos
// algorítmicos distintos que producen ese mismo polinomio. Estas funciones
// devuelven cadenas de pasos pensadas para el panel de salida de la UI.

// Lagrange interpolates the given points with the method of Lagrange and
// returns step-by-step output plus the expanded polynomial coefficients.
// Los coeficientes se presentan de menor a mayor grado.
func Lagrange(points []Point) ([]string, error) {
	if err := validatePoints(points); err != nil {
		return nil, err
	}
	n := len(points)

	steps := []string{"Operación: interpolación de Lagrange"}
	steps = append(steps, "Datos: "+formatPoints(points))

	// Coeficientes del polinomio final (formato P(x) = a0 + a1 x + ... + an x^n).
	coeff := make([]float64, n)
	bases := make([]string, n)

	for i, p := range points {
		// Producto (x - x_j)/(x_i - x_j) para todo j != i.
		base := make([]float64, 1)
		base[0] = 1
		var terms []string
		for j, pj := range points {
			if j == i {
				continue
			}
			den := p.X - pj.X
			base = polyMul(base, []float64{-pj.X, 1})
			terms = append(terms, fmt.Sprintf("(x − %s)/(%s − %s)",
				formatNum(pj.X), formatNum(p.X), formatNum(pj.X)))
			// dividir cada coeficiente por el denominador
			for k := range base {
				base[k] /= den
			}
		}
		bases[i] = strings.Join(terms, " · ")
		steps = append(steps, fmt.Sprintf("L%d(x) = %s", i, bases[i]))

		// acumular y_i * L_i
		for k := range base {
			coeff[k] += p.Y * base[k]
		}
	}

	steps = append(steps, "Combinación: P(x) = Σ yi·Li(x)")
	steps = append(steps, "Polinomio (expandido): P(x) = "+formatPoly(coeff))
	return steps, nil
}

// Newton interpolates the given points with divided differences and returns
// the full divided-differences table plus the resulting polynomial.
func Newton(points []Point) ([]string, error) {
	if err := validatePoints(points); err != nil {
		return nil, err
	}
	n := len(points)

	steps := []string{"Operación: interpolación de Newton (diferencias divididas)"}
	steps = append(steps, "Datos: "+formatPoints(points))

	// Tabla triangular: tabla[i][j] = f[x_i, ..., x_{i+j}]
	table := make([][]float64, n)
	for i := range points {
		table[i] = make([]float64, n-i)
		table[i][0] = points[i].Y
	}
	for j := 1; j < n; j++ {
		for i := range n - j {
			table[i][j] = (table[i+1][j-1] - table[i][j-1]) /
				(points[i+j].X - points[i].X) //nolint:gosec // i < n-j acota ambos índices
		}
	}

	steps = append(steps, formatDDTable(points, table))

	// Coeficientes del polinomio expandido a partir de la diagonal superior.
	coeff := make([]float64, 1)
	coeff[0] = 0
	// Forma anidada: P(x) = cc[0] + cc[1] + ...  donde cada cc[k] multiplica
	// (x-x0)(x-x1)...(x-x_{k-1}). Construimos acumulando con Horner.
	terms := make([]string, n)
	for k := range n {
		label := fmt.Sprintf("f[x0..x%d]", k)
		if k == 0 {
			label = "f[x0]"
		}
		terms[k] = fmt.Sprintf("%s = %s", label, formatNum(table[0][k]))
	}
	steps = append(steps, "Coeficientes (diagonal de la tabla): "+strings.Join(terms, ", "))

	// Polinomio en forma de Newton (base nodal).
	var nb strings.Builder
	for k := range n {
		if k > 0 {
			nb.WriteString(" + ")
		}
		nb.WriteString(formatNum(table[0][k]))
		for _, pj := range points[:k] {
			nb.WriteString("(x − " + formatNum(pj.X) + ")")
		}
	}
	steps = append(steps, "Forma de Newton: P(x) = "+nb.String())

	// Expandir la forma de Newton a coeficientes canónicos.
	expanded := make([]float64, 1)
	expanded[0] = 0
	acc := []float64{1} // acumula (x-x0)(x-x1)...(x-x_{k-1})
	for k, pk := range points {
		term := polyScale(acc, table[0][k])
		expanded = polyAdd(expanded, term)
		if k < n-1 {
			acc = polyMul(acc, []float64{-pk.X, 1})
		}
	}
	steps = append(steps, "Polinomio (expandido): P(x) = "+formatPoly(expanded))
	return steps, nil
}

// validatePoints exige puntos no vacíos y con abscisas estrictamente
// crecientes o al menos todas distintas en valor.
func validatePoints(pts []Point) error {
	if len(pts) < 2 {
		return errors.New("se necesitan al menos 2 puntos para interpolar")
	}
	seen := map[float64]bool{}
	for _, p := range pts {
		if seen[p.X] {
			return fmt.Errorf("abscisa repetida x=%s: los puntos deben tener x distintas", formatNum(p.X))
		}
		seen[p.X] = true
	}
	return nil
}

// formatPoints renderiza la lista de puntos como "(x0, y0) (x1, y1) ...".
func formatPoints(pts []Point) string {
	parts := make([]string, 0, len(pts))
	for _, p := range pts {
		parts = append(parts, fmt.Sprintf("(%s, %s)", formatNum(p.X), formatNum(p.Y)))
	}
	return strings.Join(parts, " ")
}

// formatDDTable renderiza la tabla triangular de diferencias divididas.
func formatDDTable(points []Point, table [][]float64) string {
	n := len(points)
	// Encabezado de niveles
	cols := []string{"x", "f[x]"}
	for j := 1; j < n; j++ {
		cols = append(cols, fmt.Sprintf("D.%d", j))
	}
	var b strings.Builder
	b.WriteString("Tabla de diferencias divididas:")
	b.WriteString("\n  " + padRow(cols))
	for i := range n {
		row := []string{formatNum(points[i].X), formatNum(table[i][0])}
		for j := 1; j <= i; j++ {
			row = append(row, formatNum(table[i-j][j]))
		}
		b.WriteString("\n  " + padRow(row))
	}
	return b.String()
}

// padRow alinea cada celda a la izquierda con ancho mínimo constante.
func padRow(cells []string) string {
	w := 12
	if len(cells) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cells))
	for _, c := range cells {
		parts = append(parts, padCell(c, w))
	}
	return strings.TrimRight(strings.Join(parts, " "), " ")
}

func padCell(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// --- Operaciones con polinomios (coeficientes de menor a mayor grado) ---

// polyMul multiplica dos polinomios en base x.
func polyMul(a, b []float64) []float64 {
	res := make([]float64, len(a)+len(b)-1)
	for i, ai := range a {
		for j, bj := range b {
			res[i+j] += ai * bj
		}
	}
	return res
}

// polyAdd suma dos polinomios (mantiene el mayor grado).
func polyAdd(a, b []float64) []float64 {
	n := max(len(b), len(a))
	res := make([]float64, n)
	copy(res, a)
	for i := range b {
		res[i] += b[i]
	}
	return res
}

// polyScale multiplica un polinomio por un escalar.
func polyScale(a []float64, s float64) []float64 {
	res := make([]float64, len(a))
	for i := range a {
		res[i] = a[i] * s
	}
	return res
}

// formatPoly convierte los coeficientes (menor a mayor grado) a una cadena
// legible tipo "3 − 2x + x²" (Unicode) y su variante Latex emulada no se
// genera aquí: la UI decide cómo presentarlo.
func formatPoly(c []float64) string {
	if len(c) == 0 {
		return "0"
	}
	var parts []string
	for d, v := range slices.Backward(c) {
		if approxZero(v) {
			continue
		}
		sign := "+"
		if v < 0 {
			sign = "−"
			v = -v
		}
		coefStr := formatNum(v)
		var term string
		switch d {
		case 0:
			term = coefStr
		case 1:
			term = coefStr + "x"
			if coefStr == "1" {
				term = "x"
			}
		default:
			term = coefStr + "x" + supExponent(d)
			if coefStr == "1" {
				term = "x" + supExponent(d)
			}
		}
		prefix := sign + " "
		if len(parts) == 0 {
			prefix = ""
			if sign == "−" {
				prefix = "−"
			}
		}
		parts = append(parts, prefix+term)
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " ")
}

// supExponent devuelve el exponente entero d como superíndice Unicode.
func supExponent(d int) string {
	return render.ToSuperscript(strconv.Itoa(d))
}

// approxZero considera despreciables los coeficientes muy cercanos a cero.
func approxZero(v float64) bool {
	return v > -1e-9 && v < 1e-9
}
