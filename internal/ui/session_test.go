package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
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
	declararPuntosYEjecutarLagrange(t, &m)
	guardarSesionDePrueba(t, &m, "trabajo")
	data := cargarSesionDePrueba(t, "trabajo")
	comprobarDatosSesion(t, data)
	comprobarRestauracionSesion(t, data)
}

func guardarSesionDePrueba(t *testing.T, m *Model, name string) {
	t.Helper()
	if err := SaveSession(name, m.sessionData()); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	if !SessionExists(name) {
		t.Fatal("la sesión no existe tras guardar")
	}
}

func cargarSesionDePrueba(t *testing.T, name string) SessionData {
	t.Helper()
	data, err := LoadSession(name)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	return data
}

func comprobarDatosSesion(t *testing.T, data SessionData) {
	t.Helper()
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
}

func comprobarRestauracionSesion(t *testing.T, data SessionData) {
	t.Helper()
	m := newFlowModel()
	m.applySession("trabajo", data)
	if len(m.objs) != 1 || !m.declValid || m.lastObj != "t1" {
		t.Fatalf("applySession incorrecto: objs=%v last=%q valid=%v", m.objs, m.lastObj, m.declValid)
	}
	if m.SessionName() != "trabajo" || m.dirty {
		t.Fatalf("estado de sesión incorrecto: name=%q dirty=%v", m.SessionName(), m.dirty)
	}
	if !strings.Contains(m.plainOutput(), "Listo") {
		t.Fatalf("salida no restaurada: %q", m.plainOutput())
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

func TestManualSave(t *testing.T) {
	withTempSessions(t)
	m := newFlowModel()
	if err := m.OpenSession("trabajo"); err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	m.addLog(LevelInfo, "cambio pendiente")

	cmd := m.manualSave()
	if cmd == nil {
		t.Fatal("manualSave debe devolver un comando con sesión activa")
	}
	msg, ok := cmd().(sessionSavedMsg)
	if !ok {
		t.Fatalf("mensaje inesperado: %T", msg)
	}
	if msg.err != nil {
		t.Fatalf("no se pudo guardar: %v", msg.err)
	}
	if !SessionExists("trabajo") {
		t.Fatal("la sesión no se guardó en disco")
	}
}

func TestManualSaveSinSesion(t *testing.T) {
	m := newFlowModel()
	if cmd := m.manualSave(); cmd != nil {
		t.Fatal("manualSave no debe guardar sin sesión activa")
	}
	if !strings.Contains(m.plainOutput(), "Sesión no habilitada") {
		t.Fatalf("falta el aviso de sesión no habilitada: %q", m.plainOutput())
	}
}

func TestDisableSession(t *testing.T) {
	withTempSessions(t)
	m := newFlowModel()
	if err := m.OpenSession("trabajo"); err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	m.dirty = true

	m.DisableSession()
	if m.sessionEnabled || m.SessionName() != "" || m.dirty {
		t.Fatalf("sesión no deshabilitada: enabled=%v name=%q dirty=%v",
			m.sessionEnabled, m.SessionName(), m.dirty)
	}
	if cmd := m.manualSave(); cmd != nil {
		t.Fatal("manualSave no debe guardar con la sesión deshabilitada")
	}
}

func TestAutosaveTick(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cmd := autosaveTick()
		if cmd == nil {
			t.Fatal("autosaveTick debe devolver un comando")
		}
		if _, ok := cmd().(autosaveMsg); !ok {
			t.Fatalf("mensaje inesperado: %T", cmd())
		}
	})
}

func TestOpenSessionNombreInvalido(t *testing.T) {
	withTempSessions(t)
	m := newFlowModel()
	if err := m.OpenSession("   "); err == nil {
		t.Fatal("un nombre vacío debe devolver error")
	}
}

func TestErroresDeSesionEnvuelven(t *testing.T) {
	withTempSessions(t)

	_, err := LoadSession("no-existe")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadSession debe envolver ErrNotExist: %v", err)
	}
	if !strings.Contains(err.Error(), "leyendo sesión") {
		t.Fatalf("falta contexto en LoadSession: %v", err)
	}

	if err := DeleteSession("no-existe"); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("DeleteSession debe envolver ErrNotExist: %v", err)
	}

	if _, err := DuplicateSession("no-existe"); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("DuplicateSession debe envolver ErrNotExist: %v", err)
	}
}

func TestSaveSessionErrorEnvuelve(t *testing.T) {
	base := t.TempDir()
	blocker := filepath.Join(base, "bloqueo")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", blocker)

	err := SaveSession("trabajo", SessionData{})
	if err == nil {
		t.Fatal("SaveSession debía fallar con un directorio inválido")
	}
	if !strings.Contains(err.Error(), "guardando sesión") {
		t.Fatalf("falta contexto en SaveSession: %v", err)
	}
}

// TestGuardadoSeRegistraUnaVez protege la regla de manejo único: el comando
// devuelve el error y solo el Update lo registra en el log.
func TestGuardadoSeRegistraUnaVez(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	m := newFlowModel()
	if err := m.OpenSession("trabajo"); err != nil {
		t.Fatalf("OpenSession: %v", err)
	}

	blocker := filepath.Join(data, "bloqueo")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", blocker)

	cmd := m.manualSave()
	if cmd == nil {
		t.Fatal("manualSave debe devolver un comando")
	}
	msg, ok := cmd().(sessionSavedMsg)
	if !ok || msg.err == nil {
		t.Fatalf("se esperaba un sessionSavedMsg con error: %+v", msg)
	}
	if strings.Contains(m.plainOutput(), "No se pudo guardar") {
		t.Fatal("el error no debe registrarse antes de que Update lo maneje")
	}

	m.Update(msg)
	if got := strings.Count(m.plainOutput(), "No se pudo guardar"); got != 1 {
		t.Fatalf("el error se registró %d veces, se esperaba 1", got)
	}
}

func TestSessionCorrupta(t *testing.T) {
	withTempSessions(t)
	if err := os.MkdirAll(sessionsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir(), "rota.json"), []byte("{no-json"), 0o600); err != nil {
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
