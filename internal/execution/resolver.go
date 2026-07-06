package execution

import (
	"errors"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// ErrNoStrategyFound is returned when no registered strategy can handle an
// application.
var ErrNoStrategyFound = errors.New("execution: no strategy can handle this application")

// Resolver decides which registered Strategy manages a given Application.
// It contains no technology-specific logic itself — it only asks each
// registered Strategy whether it can handle the application, in
// registration order, and returns the first one that says yes. Supporting a
// new technology never means touching Resolver; it means registering one
// more Strategy.
type Resolver struct {
	registry *Registry
}

// NewResolver builds a Resolver backed by registry.
func NewResolver(registry *Registry) *Resolver {
	return &Resolver{registry: registry}
}

// Resolve returns the first registered Strategy that can handle app, or
// ErrNoStrategyFound if none can.
func (r *Resolver) Resolve(app *models.Application) (Strategy, error) {
	for _, strategy := range r.registry.Strategies() {
		if strategy.CanHandle(app) {
			return strategy, nil
		}
	}
	return nil, ErrNoStrategyFound
}
