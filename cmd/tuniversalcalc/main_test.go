package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tuniversalcalc/internal/ui"
)

// stubTUI sustituye el programa interactivo y devuelve el modelo que recibió.
func stubTUI(t *testing.T, err error) **ui.Model {
	t.Helper()
	old := runTUI
	t.Cleanup(func() { runTUI = old })

	var got *ui.Model
	runTUI = func(m *ui.Model) error {
		got = m
		return err
	}
	return &got
}

func TestRunVersion(t *testing.T) {
	old := version
	t.Cleanup(func() { version = old })
	version = "1.2.3"

	var out, errOut bytes.Buffer
	if code := run([]string{"--version"}, &out, &errOut); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if got := out.String(); got != "tuniversalcalc 1.2.3\n" {
		t.Fatalf("salida = %q", got)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr inesperado: %q", errOut.String())
	}
}

func TestRunList(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := ui.SaveSession("alfa", ui.SessionData{}); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	var out, errOut bytes.Buffer
	if code := run([]string{"--list"}, &out, &errOut); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if !strings.Contains(out.String(), "alfa") {
		t.Fatalf("la salida no lista la sesión: %q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--help"}, &out, &errOut); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if !strings.Contains(errOut.String(), "Usage of tuniversalcalc") {
		t.Fatalf("stderr sin usage: %q", errOut.String())
	}
	if out.Len() != 0 {
		t.Fatalf("stdout inesperado: %q", out.String())
	}
}

func TestRunFlagDesconocido(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--no-existe"}, &out, &errOut); code != 2 {
		t.Fatalf("código = %d, se esperaba 2", code)
	}
	if !strings.Contains(errOut.String(), "flag provided but not defined") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunNoSave(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := ui.SaveSession("previa", ui.SessionData{}); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	got := stubTUI(t, nil)

	if code := run([]string{"--no-save"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if *got == nil {
		t.Fatal("no se arrancó la TUI")
	}
	if (*got).SessionName() != "" {
		t.Fatalf("--no-save cargó una sesión: %q", (*got).SessionName())
	}
}

func TestRunSession(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	got := stubTUI(t, nil)

	if code := run([]string{"--session", "trabajo"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if *got == nil || (*got).SessionName() != "trabajo" {
		t.Fatalf("sesión abierta inesperada: %+v", *got)
	}
}

func TestRunSessionCorrupta(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	dir := filepath.Join(data, "tuniversalcalc", "sessions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rota.json"), []byte("{no-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := stubTUI(t, nil)

	var errOut bytes.Buffer
	if code := run([]string{"--session", "rota"}, &bytes.Buffer{}, &errOut); code != 0 {
		t.Fatalf("código = %d, se esperaba 0", code)
	}
	if !strings.Contains(errOut.String(), "aviso") {
		t.Fatalf("falta el aviso de sesión corrupta: %q", errOut.String())
	}
	if *got == nil || (*got).SessionName() != "" {
		t.Fatalf("la sesión corrupta no se deshabilitó: %+v", *got)
	}
}

func TestRunErrorTUI(t *testing.T) {
	stubTUI(t, errors.New("boom"))

	var errOut bytes.Buffer
	if code := run([]string{"--no-save"}, &bytes.Buffer{}, &errOut); code != 1 {
		t.Fatalf("código = %d, se esperaba 1", code)
	}
	if !strings.Contains(errOut.String(), "boom") {
		t.Fatalf("stderr sin el error de la TUI: %q", errOut.String())
	}
}
