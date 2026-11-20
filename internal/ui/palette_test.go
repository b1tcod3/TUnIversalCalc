package ui

import (
	"reflect"
	"testing"
	"tuniversalcalc/internal/engine"
)

func TestFragmentOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"2x+2", ""},
		{"/", ""},
		{"/ecuacion", "ecuacion"},
		{"/  ecuacion 2x+2  ", "ecuacion 2x+2"},
		{"/FUNcion", "funcion"},
		{"/resolver x", "resolver x"},
	}
	for _, c := range cases {
		if got := fragmentOf(c.in); got != c.want {
			t.Errorf("fragmentOf(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}

func names(list []Template) []string {
	out := make([]string, len(list))
	for i, t := range list {
		out[i] = t.Trigger
	}
	return out
}

func TestBuildPalette(t *testing.T) {
	// Fragmento vacío devuelve todo, en orden alfabético
	decl := buildPalette(declRegistry, "")
	if !reflect.DeepEqual(names(decl), []string{"ecuacion", "expresion", "funcion", "limpiar", "puntos"}) {
		t.Fatalf("declaraciones no alfabéticas: %v", names(decl))
	}
	ops := buildPalette(opsRegistry, "")
	if !reflect.DeepEqual(names(ops), []string{"desarrollar", "factorizar", "lagrange", "limpiar", "newton", "potencia", "radical"}) {
		t.Fatalf("operaciones no alfabéticas: %v", names(ops))
	}

	// Fuzzy match conservando el orden del registro
	cases := []struct {
		list []Template
		frag string
		want []string
	}{
		{declRegistry, "fnc", []string{"funcion"}},
		{declRegistry, "expre", []string{"expresion"}},
		{declRegistry, "ecu", []string{"ecuacion"}},
		{declRegistry, "punt", []string{"puntos"}},
		{declRegistry, "lim", []string{"limpiar"}},
		{opsRegistry, "des", []string{"desarrollar"}},
		{opsRegistry, "lag", []string{"lagrange"}},
		{opsRegistry, "lim", []string{"limpiar"}},
		{opsRegistry, "newt", []string{"newton"}},
		{opsRegistry, "pot", []string{"potencia"}},
		{opsRegistry, "rad", []string{"radical"}},
		{opsRegistry, "zzz", []string{}},
	}
	for _, c := range cases {
		if got := names(buildPalette(c.list, c.frag)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("buildPalette(%v, %q) = %v, se esperaba %v", names(c.list), c.frag, got, c.want)
		}
	}
}

func TestPaletteMaxPreservadoAlCerrar(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()
	want := m.palette.Max
	if want == 0 {
		t.Fatal("updateLayout debe fijar Max > 0")
	}

	// Escribir texto sin "/" (p.ej. los BlinkMsg del cursor en la app real) no
	// debe borrar Max, o la paleta se abriría sin filas visibles después.
	m.declInput.SetValue("2x+2")
	m.refreshPalette()
	if m.palette.Open {
		t.Fatal("no debe abrirse al escribir texto plano")
	}
	if m.palette.Max != want {
		t.Fatalf("Max se perdió al cerrar: %d, se esperaba %d", m.palette.Max, want)
	}

	// La coincidencia exacta del trigger tampoco debe borrar Max.
	m.declInput.SetValue("/ecuacion")
	m.refreshPalette()
	if m.palette.Max != want {
		t.Fatalf("Max se perdió al escribir el trigger completo: %d, se esperaba %d", m.palette.Max, want)
	}

	// Tras "/" la paleta abre conservando Max y renderiza filas.
	m.declInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open {
		t.Fatal("debe abrirse con '/'")
	}
	if m.palette.Max != want {
		t.Fatalf("Max se perdió al abrir: %d, se esperaba %d", m.palette.Max, want)
	}
	if got := m.paletteView(); got == "" {
		t.Fatal("la paleta no renderiza filas (Max o lista vacíos)")
	}
}

func TestPaletteScroll(t *testing.T) {
	m := NewModel()
	m.palette.Open = true
	m.palette.Max = 2
	m.palette.List = buildPalette(declRegistry, "")
	m.palette.Sel, m.palette.Off = 0, 0

	// La ventana avanza para que la selección sea siempre visible
	m.paletteDown() // sel 1
	if m.palette.Off != 0 {
		t.Fatalf("Off = %d, se esperaba 0", m.palette.Off)
	}
	m.paletteDown() // sel 2 → ya no cabe en la ventana [0..2)
	if m.palette.Sel != 2 || m.palette.Off != 1 {
		t.Fatalf("tras sel=2: sel=%d off=%d", m.palette.Sel, m.palette.Off)
	}
	m.paletteDown() // sel 3
	if m.palette.Off != 2 {
		t.Fatalf("Off = %d, se esperaba 2", m.palette.Off)
	}
	m.paletteUp() // sel 2 → Off no retrocede hasta que sel < Off
	if m.palette.Sel != 2 || m.palette.Off != 2 {
		t.Fatalf("tras subir: sel=%d off=%d", m.palette.Sel, m.palette.Off)
	}
	m.paletteUp() // sel 1 → sale de la ventana [2..4)
	if m.palette.Sel != 1 || m.palette.Off != 1 {
		t.Fatalf("tras subir a 1: sel=%d off=%d", m.palette.Sel, m.palette.Off)
	}
}

func TestPaletteAbreEnPanel1(t *testing.T) {
	m := newFlowModel()

	// Sin "/" no hay paleta
	m.declInput.SetValue("2x+2")
	m.refreshPalette()
	if m.palette.Open {
		t.Fatal("no debe abrirse sin '/'")
	}

	// "/" en panel 1 → todas las declaraciones alfabéticas
	m.declInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open {
		t.Fatal("debe abrirse con '/'")
	}
	if !reflect.DeepEqual(names(m.palette.List), []string{"ecuacion", "expresion", "funcion", "limpiar", "puntos"}) {
		t.Fatalf("lista de declaraciones incorrecta: %v", names(m.palette.List))
	}
	if m.palette.Sel != 0 || m.palette.Anchor != FocusDecl {
		t.Fatalf("selección/ancla inicial incorrecta: %+v", m.palette)
	}
}

func TestPaletteFuzzyEInsercion(t *testing.T) {
	m := newFlowModel()

	// Fuzzy: "/fnc" solo deja funcion
	m.declInput.SetValue("/fnc")
	m.refreshPalette()
	if !reflect.DeepEqual(names(m.palette.List), []string{"funcion"}) {
		t.Fatalf("fuzzy resultó: %v", names(m.palette.List))
	}

	// Enter inserta la plantilla en el panel 1 y cierra
	m.paletteSelect()
	if m.declInput.Value() != "/funcion " {
		t.Fatalf("plantilla no insertada: %q", m.declInput.Value())
	}
	if m.palette.Open {
		t.Fatal("la paleta debe cerrarse al seleccionar")
	}
}

func TestPaletteNavegacionClamp(t *testing.T) {
	m := newFlowModel()

	m.declInput.SetValue("/")
	m.refreshPalette()

	// Navegación con clamp (sin wrap)
	m.paletteDown()
	if m.palette.Sel != 1 {
		t.Fatalf("Tab debe bajar a 1, quedó %d", m.palette.Sel)
	}
	m.paletteDown()
	m.paletteDown()
	m.paletteDown()
	m.paletteDown()
	if m.palette.Sel != 4 {
		t.Fatalf("la selección debe quedarse en el último (4), quedó %d", m.palette.Sel)
	}
	m.paletteUp()
	if m.palette.Sel != 3 {
		t.Fatalf("Shift+Tab debe subir a 3, quedó %d", m.palette.Sel)
	}
	m.paletteUp()
	m.paletteUp()
	m.paletteUp()
	if m.palette.Sel != 0 {
		t.Fatalf("la selección debe quedarse en 0, quedó %d", m.palette.Sel)
	}
}

func TestPaletteCierre(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open {
		t.Fatal("la paleta debe estar abierta")
	}

	// Cerrar la paleta (lo que hace Esc)
	m.closePalette()
	if m.palette.Open {
		t.Fatal("la paleta debe cerrarse")
	}

	// Coincidencia exacta de trigger cierra (se escribe la orden a mano)
	m.declInput.SetValue("/ecuacion")
	m.refreshPalette()
	if m.palette.Open {
		t.Fatal("la paleta debe cerrarse al escribir el trigger completo")
	}

	// Sin ningún match cierra
	m.declInput.SetValue("/zzz")
	m.refreshPalette()
	if m.palette.Open {
		t.Fatal("la paleta debe cerrarse sin coincidencias")
	}
}

func TestPalettePanel2Operaciones(t *testing.T) {
	m := newFlowModel()
	m.focus = FocusCmd
	m.cmdInput.Blur()
	m.declInput.Blur()
	m.cmdInput.Focus()

	// Panel 2: "/lag" → lagrange; Enter inserta plantilla de operación
	m.cmdInput.SetValue("/lag")
	m.refreshPalette()
	if !m.palette.Open || !reflect.DeepEqual(names(m.palette.List), []string{"lagrange"}) {
		t.Fatalf("paleta panel 2 incorrecta: open=%v %v", m.palette.Open, names(m.palette.List))
	}
	m.paletteSelect()
	if m.cmdInput.Value() != "/lagrange " {
		t.Fatalf("plantilla de operación no insertada: %q", m.cmdInput.Value())
	}

	// Cambiar de panel cierra la paleta
	m.cmdInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open {
		t.Fatal("paleta panel 2 debe estar abierta")
	}
	m.switchPanel()
	if m.palette.Open {
		t.Fatal("cambiar de panel debe cerrar la paleta")
	}
}

func TestPaletteConObjetosRegistrados(t *testing.T) {
	m := newFlowModel()

	// El objeto registrado no interfiere con la paleta
	decl := engine.Decl{Tipo: engine.DeclFuncion, Nombre: "f1", Latex: "2x + 2"}
	m.objs["f1"] = decl
	m.focus = FocusCmd
	m.cmdInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open {
		t.Fatal("la paleta debe abrirse en el panel 2 con objetos registrados")
	}
}
