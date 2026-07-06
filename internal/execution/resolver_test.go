package execution_test

import (
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestResolveWithEmptyRegistryFails(t *testing.T) {
	resolver := execution.NewResolver(execution.NewRegistry())
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode

	if _, err := resolver.Resolve(app); !errors.Is(err, execution.ErrNoStrategyFound) {
		t.Fatalf("Resolve() error = %v, want %v", err, execution.ErrNoStrategyFound)
	}
}

func TestResolveReturnsMatchingStrategy(t *testing.T) {
	registry := execution.NewRegistry()
	node := &fakeStrategy{name: "node", runtime: models.RuntimeNode}
	python := &fakeStrategy{name: "python", runtime: models.RuntimePython}
	registry.Register(node)
	registry.Register(python)

	resolver := execution.NewResolver(registry)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimePython

	got, err := resolver.Resolve(app)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Metadata().Name != "python" {
		t.Fatalf("Resolve() = %q, want %q", got.Metadata().Name, "python")
	}
}

func TestResolveFirstRegisteredMatchWins(t *testing.T) {
	registry := execution.NewRegistry()
	first := &fakeStrategy{name: "first", runtime: models.RuntimeNode}
	second := &fakeStrategy{name: "second", runtime: models.RuntimeNode}
	registry.Register(first)
	registry.Register(second)

	resolver := execution.NewResolver(registry)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode

	got, err := resolver.Resolve(app)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Metadata().Name != "first" {
		t.Fatalf("Resolve() = %q, want %q (first registered)", got.Metadata().Name, "first")
	}
}

func TestResolveNoMatchFails(t *testing.T) {
	registry := execution.NewRegistry()
	registry.Register(&fakeStrategy{name: "node", runtime: models.RuntimeNode})

	resolver := execution.NewResolver(registry)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeRust

	if _, err := resolver.Resolve(app); !errors.Is(err, execution.ErrNoStrategyFound) {
		t.Fatalf("Resolve() error = %v, want %v", err, execution.ErrNoStrategyFound)
	}
}
