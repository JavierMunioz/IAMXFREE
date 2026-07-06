package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrApplicationNameRequired is returned by Validate when Name is blank.
	ErrApplicationNameRequired = errors.New("application name is required")
	// ErrApplicationTypeInvalid is returned by Validate when Type is not one
	// of the known ApplicationType values.
	ErrApplicationTypeInvalid = errors.New("application type is invalid")
)

// Application is the central entity IAMXFREE manages: a single deployed
// project, regardless of its stack. Process management, reverse proxy, SSL
// and every other feature are modeled as operations on top of this entity,
// not as separate top-level concepts.
type Application struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        ApplicationType   `json:"type"`
	Framework   Framework         `json:"framework,omitempty"`
	Runtime     Runtime           `json:"runtime,omitempty"`
	Status      ApplicationStatus `json:"status"`

	Source SourceInfo       `json:"source"`
	Config DeploymentConfig `json:"config"`

	// Reserved for future iterations; always present so the on-disk shape
	// does not change when those features are implemented.
	Metrics        RuntimeMetrics     `json:"metrics,omitempty"`
	LastDeployment DeploymentHistory  `json:"last_deployment,omitempty"`
	SSL            SSLCertificate     `json:"ssl,omitempty"`
	ReverseProxy   ReverseProxyConfig `json:"reverse_proxy,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewApplication creates a new Application with a generated ID and
// timestamps. It starts in StatusStopped: registering an application does
// not mean it has been installed, configured or started yet.
func NewApplication(name string, appType ApplicationType) *Application {
	now := time.Now().UTC()
	return &Application{
		ID:        uuid.NewString(),
		Name:      name,
		Type:      appType,
		Status:    StatusStopped,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks the invariants an Application must uphold regardless of
// which service or repository is operating on it.
func (a *Application) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return ErrApplicationNameRequired
	}
	if !a.Type.IsValid() {
		return ErrApplicationTypeInvalid
	}
	return nil
}

// Touch updates UpdatedAt to the current time. Callers should invoke it
// whenever they persist a change to the application.
func (a *Application) Touch() {
	a.UpdatedAt = time.Now().UTC()
}
