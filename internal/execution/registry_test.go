package execution_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestRegistryStartsEmpty(t *testing.T) {
	r := execution.NewRegistry()
	if got := r.Strategies(); len(got) != 0 {
		t.Fatalf("Strategies() = %v, want empty", got)
	}
}

func TestRegistryRegisterPreservesOrder(t *testing.T) {
	r := execution.NewRegistry()
	first := &fakeStrategy{name: "first", runtime: models.RuntimeNode}
	second := &fakeStrategy{name: "second", runtime: models.RuntimePython}

	r.Register(first)
	r.Register(second)

	got := r.Strategies()
	if len(got) != 2 {
		t.Fatalf("len(Strategies()) = %d, want 2", len(got))
	}
	if got[0].Metadata().Name != "first" || got[1].Metadata().Name != "second" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestRegistryStrategiesReturnsACopy(t *testing.T) {
	r := execution.NewRegistry()
	r.Register(&fakeStrategy{name: "first", runtime: models.RuntimeNode})

	got := r.Strategies()
	got[0] = &fakeStrategy{name: "mutated", runtime: models.RuntimeGo}

	again := r.Strategies()
	if again[0].Metadata().Name != "first" {
		t.Fatalf("mutating the returned slice affected the registry: %+v", again)
	}
}
