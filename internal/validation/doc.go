// Package validation provides small, composable validators for raw string
// input. Wizard steps (and anything else that captures free text) depend on
// this package instead of hand-rolling checks, so a new kind of validation
// can be added here without touching any Step implementation.
package validation
