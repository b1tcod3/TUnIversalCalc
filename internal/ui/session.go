package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
)

const (
	sessionVersion    = 1
	sessionExt        = ".json"
	maxSessionNameLen = 64
	autosaveInterval  = 30 * time.Second
)

// SessionMeta describe una sesión guardada en disco.
type SessionMeta struct {
	Name     string
	Modified time.Time
}

// SessionData es el contenido serializable de una sesión. No guarda
// resultados calculados: al cargar se reconstruye el log y el usuario puede
// volver a ejecutar sus operaciones.
type SessionData struct {
	Version    int                     `json:"version"`
	DeclInput  string                  `json:"decl_input"`
	CmdInput   string                  `json:"cmd_input"`
	DeclCursor int                     `json:"decl_cursor"`
	CmdCursor  int                     `json:"cmd_cursor"`
	Focus      int                     `json:"focus"`
	History    []string                `json:"history"`
	Output     []LogEntry              `json:"output"`
	Objects    map[string]engine.Decl  `json:"objects"`
	LastObj    string                  `json:"last_obj"`
	Counters   map[engine.DeclType]int `json:"counters"`
}

type autosaveMsg struct{}

type sessionSavedMsg struct {
	name        string
	err         error
	fingerprint string
}

// sessionsDir devuelve el directorio XDG de datos para las sesiones.
func sessionsDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "tuniversalcalc", "sessions")
}

func sessionPath(name string) string {
	return filepath.Join(sessionsDir(), name+sessionExt)
}

func lastSessionPath() string {
	return filepath.Join(sessionsDir(), "last")
}

// sanitizeSessionName convierte el nombre en algo seguro para el sistema de
// archivos. Devuelve "" si no queda nada usable.
func sanitizeSessionName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r < ' ' || strings.ContainsRune(`/\:*?"<>|`, r) {
			return -1
		}
		return r
	}, name)
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.Trim(cleaned, " .")
	if len([]rune(cleaned)) > maxSessionNameLen {
		cleaned = string([]rune(cleaned)[:maxSessionNameLen])
		cleaned = strings.Trim(cleaned, " .")
	}
	return cleaned
}

func SessionExists(name string) bool {
	_, err := os.Stat(sessionPath(name))
	return err == nil
}

// ListSessions devuelve las sesiones ordenadas por fecha de modificación.
func ListSessions() []SessionMeta {
	entries, err := os.ReadDir(sessionsDir())
	if err != nil {
		return nil
	}
	var out []SessionMeta
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), sessionExt) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		out = append(out, SessionMeta{
			Name:     strings.TrimSuffix(entry.Name(), sessionExt),
			Modified: info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Modified.After(out[j].Modified)
	})
	return out
}

// LastSessionName devuelve la última sesión abierta o, si no hay registro,
// la más reciente. El segundo valor indica si existe alguna.
func LastSessionName() (string, bool) {
	if data, err := os.ReadFile(lastSessionPath()); err == nil {
		name := sanitizeSessionName(strings.TrimSpace(string(data)))
		if name != "" && SessionExists(name) {
			return name, true
		}
	}
	sessions := ListSessions()
	if len(sessions) == 0 {
		return "", false
	}
	return sessions[0].Name, true
}

// NewSessionName genera un nombre único basado en la fecha.
func NewSessionName() string {
	base := time.Now().Format("2006-01-02")
	if !SessionExists(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d", base, i)
		if !SessionExists(candidate) {
			return candidate
		}
	}
}

// LoadSession lee y normaliza una sesión del disco.
func LoadSession(name string) (SessionData, error) {
	var data SessionData
	raw, err := os.ReadFile(sessionPath(name))
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, fmt.Errorf("sesión %q corrupta: %w", name, err)
	}
	normalizeSession(&data)
	return data, nil
}

// SaveSession escribe la sesión de forma atómica (temporal + rename).
func SaveSession(name string, data SessionData) error {
	if name == "" {
		return fmt.Errorf("nombre de sesión vacío")
	}
	if data.Version == 0 {
		data.Version = sessionVersion
	}
	normalizeSession(&data)
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if err := atomicWrite(sessionPath(name), payload); err != nil {
		return err
	}
	return atomicWrite(lastSessionPath(), []byte(name+"\n"))
}

// DeleteSession elimina una sesión y limpia el puntero de última sesión.
func DeleteSession(name string) error {
	if err := os.Remove(sessionPath(name)); err != nil {
		return err
	}
	if data, err := os.ReadFile(lastSessionPath()); err == nil {
		if sanitizeSessionName(strings.TrimSpace(string(data))) == name {
			_ = os.Remove(lastSessionPath())
		}
	}
	return nil
}

// RenameSession renombra una sesión existente y actualiza el puntero de
// última sesión si era la activa.
func RenameSession(oldName, newName string) error {
	oldName = sanitizeSessionName(oldName)
	newName = sanitizeSessionName(newName)
	if oldName == "" || newName == "" {
		return fmt.Errorf("nombre de sesión vacío o inválido")
	}
	if oldName == newName {
		return nil
	}
	if !SessionExists(oldName) {
		return fmt.Errorf("la sesión %q no existe", oldName)
	}
	if SessionExists(newName) {
		return fmt.Errorf("la sesión %q ya existe", newName)
	}
	if err := os.Rename(sessionPath(oldName), sessionPath(newName)); err != nil {
		return err
	}
	if data, err := os.ReadFile(lastSessionPath()); err == nil {
		if sanitizeSessionName(strings.TrimSpace(string(data))) == oldName {
			_ = atomicWrite(lastSessionPath(), []byte(newName+"\n"))
		}
	}
	return nil
}

// DuplicateSession copia una sesión con un nombre único y devuelve el nuevo.
func DuplicateSession(name string) (string, error) {
	payload, err := os.ReadFile(sessionPath(name))
	if err != nil {
		return "", err
	}
	copyName := uniqueSessionName(name + " copia")
	if err := atomicWrite(sessionPath(copyName), payload); err != nil {
		return "", err
	}
	return copyName, nil
}

func uniqueSessionName(base string) string {
	if !SessionExists(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d", base, i)
		if !SessionExists(candidate) {
			return candidate
		}
	}
}

func atomicWrite(path string, payload []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	cleanup := func() {
		temp.Close()
		os.Remove(tempName)
	}
	if _, err := temp.Write(payload); err != nil {
		cleanup()
		return err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := temp.Close(); err != nil {
		os.Remove(tempName)
		return err
	}
	if err := os.Chmod(tempName, 0o644); err != nil {
		os.Remove(tempName)
		return err
	}
	if err := os.Rename(tempName, path); err != nil {
		os.Remove(tempName)
		return err
	}
	return nil
}

func normalizeSession(data *SessionData) {
	if data.Version == 0 {
		data.Version = sessionVersion
	}
	if data.Objects == nil {
		data.Objects = map[string]engine.Decl{}
	}
	if data.Counters == nil {
		data.Counters = map[engine.DeclType]int{}
	}
	for name, d := range data.Objects {
		n := suffixNumber(name, d.Tipo.Prefix())
		if n > data.Counters[d.Tipo] {
			data.Counters[d.Tipo] = n
		}
	}
	if data.History == nil {
		data.History = []string{}
	}
	if data.Output == nil {
		data.Output = []LogEntry{}
	}
	if data.Focus < int(FocusDecl) || data.Focus > int(FocusCmd) {
		data.Focus = int(FocusDecl)
	}
	if data.Focus == int(FocusCmd) && len(data.Objects) == 0 {
		data.Focus = int(FocusDecl)
	}
	if data.LastObj == "" || !hasObject(data.Objects, data.LastObj) {
		data.LastObj = lastObjectName(data.Objects)
	}
}

func suffixNumber(name, prefix string) int {
	if !strings.HasPrefix(name, prefix) {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimPrefix(name, prefix))
	if err != nil || n < 1 {
		return 0
	}
	return n
}

func hasObject(objs map[string]engine.Decl, name string) bool {
	_, ok := objs[name]
	return ok
}

func lastObjectName(objs map[string]engine.Decl) string {
	if len(objs) == 0 {
		return ""
	}
	names := make([]string, 0, len(objs))
	for name := range objs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names[len(names)-1]
}

// sessionData crea una copia serializable del estado actual.
func (m *Model) sessionData() SessionData {
	data := SessionData{
		Version:    sessionVersion,
		DeclInput:  m.declInput.Value(),
		CmdInput:   m.cmdInput.Value(),
		DeclCursor: m.declInput.Position(),
		CmdCursor:  m.cmdInput.Position(),
		Focus:      int(m.focus),
		History:    append([]string(nil), m.history...),
		Output:     append([]LogEntry(nil), m.outputLog...),
		Objects:    make(map[string]engine.Decl, len(m.objs)),
		LastObj:    m.lastObj,
		Counters:   make(map[engine.DeclType]int, len(m.counters)),
	}
	for name, d := range m.objs {
		cp := d
		if d.Points != nil {
			cp.Points = append([]engine.Point(nil), d.Points...)
		}
		data.Objects[name] = cp
	}
	for t, n := range m.counters {
		data.Counters[t] = n
	}
	return data
}

// applySession reemplaza el estado actual por el de una sesión cargada.
func (m *Model) applySession(name string, data SessionData) {
	normalizeSession(&data)
	m.sessionName = name
	m.sessionEnabled = true

	m.declInput.SetValue(data.DeclInput)
	m.declInput.SetCursor(data.DeclCursor)
	m.cmdInput.SetValue(data.CmdInput)
	m.cmdInput.SetCursor(data.CmdCursor)

	m.focus = Focus(data.Focus)
	m.declInput.Blur()
	m.cmdInput.Blur()
	if m.focus == FocusDecl {
		m.declInput.Focus()
	} else {
		m.cmdInput.Focus()
	}

	m.objs = data.Objects
	m.lastObj = data.LastObj
	m.counters = data.Counters
	m.declValid = len(m.objs) > 0
	m.tabHinted = false
	m.history = append([]string(nil), data.History...)
	m.histIdx = len(m.history)
	m.outputLog = append([]LogEntry(nil), data.Output...)
	if len(m.outputLog) == 0 {
		m.outputLog = []LogEntry{
			{Level: LevelInfo, Content: "Sesión cargada: " + name},
		}
	}

	m.undoStack = nil
	m.redoStack = nil
	m.reqSeq = 0
	m.pendingReq = 0
	m.helpOpen = false
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.output.SetContent(m.renderLog())
	m.output.GotoBottom()
	m.dirty = false
}

// OpenSession activa una sesión y carga su contenido si existe.
func (m *Model) OpenSession(name string) error {
	clean := sanitizeSessionName(name)
	if clean == "" {
		return fmt.Errorf("nombre de sesión vacío o inválido")
	}
	m.sessionName = clean
	m.sessionEnabled = true
	if !SessionExists(clean) {
		m.dirty = false
		return nil
	}
	data, err := LoadSession(clean)
	if err != nil {
		return err
	}
	m.applySession(clean, data)
	return nil
}

// SessionName devuelve el nombre de la sesión activa ("" si es efímera).
func (m *Model) SessionName() string {
	return m.sessionName
}

// DisableSession desactiva la persistencia (por ejemplo, si la sesión está
// corrupta y no queremos sobrescribirla por accidente).
func (m *Model) DisableSession() {
	m.sessionEnabled = false
	m.sessionName = ""
	m.dirty = false
}

// autosaveTick programa el siguiente guardado automático.
func (m *Model) autosaveTick() tea.Cmd {
	return tea.Tick(autosaveInterval, func(time.Time) tea.Msg {
		return autosaveMsg{}
	})
}

// saveSessionCmd guarda en segundo plano una copia del estado actual.
func (m *Model) saveSessionCmd() tea.Cmd {
	if !m.sessionEnabled || m.sessionName == "" {
		return nil
	}
	name := m.sessionName
	data := m.sessionData()
	fingerprint := sessionFingerprint(data)
	return func() tea.Msg {
		return sessionSavedMsg{name: name, err: SaveSession(name, data), fingerprint: fingerprint}
	}
}

// sessionFingerprint permite saber si el estado cambió mientras un guardado
// asíncrono estaba en vuelo.
func sessionFingerprint(data SessionData) string {
	payload, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(payload)
}

// saveSessionNow guarda de forma síncrona; se usa al salir para no competir
// con tea.Quit.
func (m *Model) saveSessionNow() {
	if !m.sessionEnabled || m.sessionName == "" {
		return
	}
	if err := SaveSession(m.sessionName, m.sessionData()); err != nil {
		m.addLog(LevelError, "No se pudo guardar la sesión: "+err.Error())
		return
	}
	m.dirty = false
}

// manualSave guarda bajo demanda y avisa si no hay sesión activa.
func (m *Model) manualSave() tea.Cmd {
	if !m.sessionEnabled || m.sessionName == "" {
		m.addLog(LevelInfo, "Sesión no habilitada: inicia sin --no-save para guardar")
		return nil
	}
	return m.saveSessionCmd()
}

// quit guarda y termina el programa.
func (m *Model) quit() tea.Cmd {
	m.saveSessionNow()
	return tea.Quit
}
