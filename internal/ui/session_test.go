package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tuniversalcalc/internal/engine"
)

func withTempSessions(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
}

func TestSanitizeSessionName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"  mi sesión  ", "mi sesión"},
		{"a/b\\c:d", "abcd"},
		{"...oculta", "oculta"},
		{"nombre*con?raros", "nombreconraros"},
	}
	for _, c := range cases {
		if got := sanitizeSessionName(c.in); got != c.want {
			t.Errorf("sanitizeSessionName(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}

func TestSessionRoundtrip(t *testing.T) {
	withTempSessions(t)

	m := newFlowModel()
	m.declInput.SetValue("/puntos (1,2) (3,4) (5,6)")
	m.handleEnter()
	m.cmdInput.SetValue("/lagrange t1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("la operación debe producir un comando")
	}
	m.applyResult(cmd().(ResultMsg))

	if err := SaveSession("trabajo", m.sessionData()); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	if !SessionExists("trabajo") {
		t.Fatal("la sesión no existe tras guardar")
	}

	data, err := LoadSession("trabajo")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if len(data.Objects) != 1 || data.Objects["t1"].Tipo != engine.DeclPuntos {
		t.Fatalf("objetos no persistidos: %+v", data.Objects)
	}
	if data.LastObj != "t1" || len(data.Objects["t1"].Points) != 3 {
		t.Fatalf("último objeto/puntos incorrectos: %+v", data)
	}
	if len(data.History) != 1 || data.History[0] != "/lagrange t1" {
		t.Fatalf("historial no persistido: %+v", data.History)
	}
	if data.Counters[engine.DeclPuntos] != 1 {
		t.Fatalf("contadores no persistidos: %+v", data.Counters)
	}

	m2 := newFlowModel()
	m2.applySession("trabajo", data)
	if len(m2.objs) != 1 || !m2.declValid || m2.lastObj != "t1" {
		t.Fatalf("applySession incorrecto: objs=%v last=%q valid=%v", m2.objs, m2.lastObj, m2.declValid)
	}
	if m2.SessionName() != "trabajo" || m2.dirty {
		t.Fatalf("estado de sesión incorrecto: name=%q dirty=%v", m2.SessionName(), m2.dirty)
	}
	if !strings.Contains(m2.plainOutput(), "Listo") {
		t.Fatalf("salida no restaurada: %q", m2.plainOutput())
	}
}

func TestSessionListAndLast(t *testing.T) {
	withTempSessions(t)

	for _, name := range []string{"ayer", "hoy"} {
		data := SessionData{Objects: map[string]engine.Decl{}, Counters: map[engine.DeclType]int{}}
		if err := SaveSession(name, data); err != nil {
			t.Fatalf("SaveSession(%q): %v", name, err)
		}
	}
	// Forzar que "ayer" sea más reciente para comprobar el orden por mtime.
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(sessionPath("ayer"), past, past); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	sessions := ListSessions()
	if len(sessions) != 2 || sessions[0].Name != "hoy" {
		t.Fatalf("orden de sesiones incorrecto: %+v", sessions)
	}
	if last, ok := LastSessionName(); !ok || last != "hoy" {
		t.Fatalf("última sesión = %q, %v", last, ok)
	}

	if err := DeleteSession("hoy"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if SessionExists("hoy") {
		t.Fatal("la sesión eliminada sigue existiendo")
	}
}

func TestSessionCorrupta(t *testing.T) {
	withTempSessions(t)
	if err := os.MkdirAll(sessionsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir(), "rota.json"), []byte("{no-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSession("rota"); err == nil {
		t.Fatal("una sesión corrupta debe devolver error")
	}
}

func TestSessionNombresUnicos(t *testing.T) {
	withTempSessions(t)
	name := NewSessionName()
	if name == "" {
		t.Fatal("NewSessionName vacío")
	}
	if err := SaveSession(name, SessionData{}); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	if again := NewSessionName(); again == name {
		t.Fatalf("NewSessionName repitió %q", name)
	}
}

func TestAutosaveNoPierdeCambiosEnVuelo(t *testing.T) {
	withTempSessions(t)
	m := newFlowModel()
	m.sessionName = "vuelo"
	m.sessionEnabled = true
	m.dirty = true

	cmd := m.saveSessionCmd()
	if cmd == nil {
		t.Fatal("se esperaba comando de guardado")
	}
	// El usuario escribe mientras el guardado está en vuelo.
	m.addLog(LevelInfo, "cambio posterior")
	msg := cmd()
	if _, _ = m.Update(msg); !m.dirty {
		t.Fatal("dirty no debe limpiarse si hubo cambios después del snapshot")
	}
}
