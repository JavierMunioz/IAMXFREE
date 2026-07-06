// Package repositories persists and retrieves IAMXFREE's own state (registered
// applications, deployment history, configuration snapshots). It abstracts
// the storage backend so it can move from flat files to something else later
// without touching services.
package repositories
