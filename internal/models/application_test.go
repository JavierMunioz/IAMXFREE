package models_test

import (
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestNewApplicationDefaults(t *testing.T) {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if app.ID == "" {
		t.Fatal("expected a generated ID")
	}
	if app.Status != models.StatusStopped {
		t.Fatalf("Status = %q, want %q", app.Status, models.StatusStopped)
	}
	if app.CreatedAt.IsZero() || app.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps to be set")
	}
}

func TestApplicationValidate(t *testing.T) {
	tests := []struct {
		name    string
		app     *models.Application
		wantErr error
	}{
		{
			name:    "valid application",
			app:     models.NewApplication("my-api", models.ApplicationTypeAPI),
			wantErr: nil,
		},
		{
			name:    "blank name",
			app:     models.NewApplication("   ", models.ApplicationTypeAPI),
			wantErr: models.ErrApplicationNameRequired,
		},
		{
			name:    "unknown type",
			app:     models.NewApplication("my-api", models.ApplicationType("unknown")),
			wantErr: models.ErrApplicationTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.Validate()
			if tt.wantErr == nil && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplicationTypeIsValid(t *testing.T) {
	if !models.ApplicationTypeMicroservice.IsValid() {
		t.Fatal("expected ApplicationTypeMicroservice to be valid")
	}
	if models.ApplicationType("not-a-type").IsValid() {
		t.Fatal("expected unknown type to be invalid")
	}
}

func TestApplicationStatusIsValid(t *testing.T) {
	if !models.StatusRunning.IsValid() {
		t.Fatal("expected StatusRunning to be valid")
	}
	if models.ApplicationStatus("not-a-status").IsValid() {
		t.Fatal("expected unknown status to be invalid")
	}
}
