package models

// ApplicationDraft holds the raw, temporary data captured while a user is
// registering a new application (e.g. through the TUI wizard). It has no
// structural relationship to Application — no shared nested types, no
// partial Application held inside it — so that the capture step (and
// whatever UI drives it) never depends on Application's shape. The only
// connection between the two is ToApplication, the single conversion point
// that runs once, when the capture is complete.
type ApplicationDraft struct {
	Name      string
	Type      ApplicationType
	Framework Framework
	Runtime   Runtime
	Path      string
	Port      int
	Domain    string
	RepoURL   string
}

// ToApplication converts the draft into a persistable Application. This is
// the only place a draft turns into a real entity.
func (d ApplicationDraft) ToApplication() *Application {
	app := NewApplication(d.Name, d.Type)
	app.Framework = d.Framework
	app.Runtime = d.Runtime
	app.Source = SourceInfo{
		LocalPath:     d.Path,
		RepositoryURL: d.RepoURL,
	}
	app.Config = DeploymentConfig{
		InternalPort: d.Port,
		Domain:       d.Domain,
	}
	return app
}
