package dashboard

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestStatusPresentationCoversEveryKnownStatus(t *testing.T) {
	statuses := []models.ApplicationStatus{
		models.StatusInstalled,
		models.StatusConfigured,
		models.StatusRunning,
		models.StatusStopped,
		models.StatusError,
		models.StatusUpdating,
	}

	for _, status := range statuses {
		icon, color := statusPresentation(status)
		if icon == "" {
			t.Errorf("statusPresentation(%q) returned an empty icon", status)
		}
		if color == nil {
			t.Errorf("statusPresentation(%q) returned a nil color", status)
		}
	}
}

func TestStatusPresentationFallsBackForUnknownStatus(t *testing.T) {
	icon, color := statusPresentation(models.ApplicationStatus("something-new"))
	if icon == "" {
		t.Fatal("expected a fallback icon for an unknown status")
	}
	if color == nil {
		t.Fatal("expected a fallback color for an unknown status")
	}
}
