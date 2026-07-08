package models

import "testing"

func TestDeploymentStrategyIsValid(t *testing.T) {
	valid := []DeploymentStrategy{DeploymentStrategyStandard, DeploymentStrategyZeroDowntime}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q.IsValid() = false, want true", s)
		}
	}

	invalid := []DeploymentStrategy{"", "blue_green", "canary"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("%q.IsValid() = true, want false", s)
		}
	}
}
