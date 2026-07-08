package nginx

// SiteSummary is one entry in Manager.ListVirtualHosts: enough to identify
// and locate a site without re-parsing its config file back into a
// VirtualHost, which this iteration does not attempt.
type SiteSummary struct {
	ServerName string
	FilePath   string
	Enabled    bool
}
