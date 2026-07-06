package planner

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
)

func TestRegistryStartsEmpty(t *testing.T) {
	r := NewRegistry()
	if got := r.Planners(); len(got) != 0 {
		t.Fatalf("Planners() = %v, want empty", got)
	}
}

func TestRegistryRegisterPreservesOrder(t *testing.T) {
	r := NewRegistry()
	first := &fakePlanner{name: "first", projectType: inspection.ProjectTypeNode}
	second := &fakePlanner{name: "second", projectType: inspection.ProjectTypePython}

	r.Register(first)
	r.Register(second)

	got := r.Planners()
	if len(got) != 2 {
		t.Fatalf("len(Planners()) = %d, want 2", len(got))
	}
	if got[0].Name() != "first" || got[1].Name() != "second" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestRegistryPlannersReturnsACopy(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakePlanner{name: "first", projectType: inspection.ProjectTypeNode})

	got := r.Planners()
	got[0] = &fakePlanner{name: "mutated", projectType: inspection.ProjectTypeGo}

	again := r.Planners()
	if again[0].Name() != "first" {
		t.Fatalf("mutating the returned slice affected the registry: %+v", again)
	}
}
