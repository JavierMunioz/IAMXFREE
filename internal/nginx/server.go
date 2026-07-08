package nginx

// Server describes an Nginx installation detected on the host: whether it
// is present at all, what version it reports, and where its configuration
// lives. A Server with Available == false is a normal, expected outcome —
// Nginx simply is not installed — never an error.
type Server struct {
	Available  bool
	BinaryPath string
	Version    string

	// ConfigRoot is the directory the rest of the paths below live under
	// (e.g. "/etc/nginx"). It is empty when it could not be determined.
	ConfigRoot string

	// MainConfigFile is the top-level config Nginx loads on start (e.g.
	// "/etc/nginx/nginx.conf").
	MainConfigFile string

	// SitesAvailableDir and SitesEnabledDir are where virtual host files
	// are stored and, respectively, symlinked into to take effect — the
	// Debian/Ubuntu layout this manager targets.
	SitesAvailableDir string
	SitesEnabledDir   string

	// Notes explicitly states what could not be determined about this
	// installation, instead of a Server silently omitting it.
	Notes []string
}
