package ui

import (
	"strings"
	"testing"
)

func TestResultadoObsoletoSeIgnora(t *testing.T) {
	m := newFlowModel()
	m.pendingReq = 2

	m.applyResult(ResultMsg{ReqID: 1, Steps: []string{"viejo"}})
	if strings.Contains(m.plainOutput(), "viejo") {
		t.Fatal("un resultado con ReqID viejo no debe aplicarse")
	}

	m.applyResult(ResultMsg{ReqID: 2, Steps: []string{"nuevo"}})
	if !strings.Contains(m.plainOutput(), "nuevo") {
		t.Fatal("el resultado actual debe aplicarse")
	}
	if m.pendingReq != 0 {
		t.Fatalf("pendingReq = %d, se esperaba 0", m.pendingReq)
	}
}

func TestResultadoSinReqIDSeAplica(t *testing.T) {
	m := newFlowModel()
	m.pendingReq = 5
	m.applyResult(ResultMsg{Steps: []string{"compat"}})
	if !strings.Contains(m.plainOutput(), "compat") {
		t.Fatal("los resultados sin ReqID deben seguir aplicándose (tests/compatibilidad)")
	}
}

func TestHandleEnterAsignaReqID(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)
	m.cmdInput.SetValue("/newton t1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("se esperaba comando")
	}
	msg := cmd().(ResultMsg)
	if msg.ReqID == 0 || msg.ReqID != m.pendingReq {
		t.Fatalf("ReqID = %d, pendingReq = %d", msg.ReqID, m.pendingReq)
	}
}
