package detail

import "testing"

func TestDeploymentPlanKeyEmitsOpenDeploymentPlanMsg(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	_, cmd := m.handleKey(keyMsg("p"))
	if cmd == nil {
		t.Fatal("expected pressing p to return a command")
	}
	msg, ok := cmd().(OpenDeploymentPlanMsg)
	if !ok {
		t.Fatalf("expected OpenDeploymentPlanMsg, got %T", cmd())
	}
	if msg.AppID != app.ID {
		t.Errorf("AppID = %q, want %q", msg.AppID, app.ID)
	}
}
