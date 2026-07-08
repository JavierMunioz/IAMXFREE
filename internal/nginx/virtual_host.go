package nginx

// VirtualHost is one site Nginx serves: a domain, the port it listens on,
// and the locations that handle requests for it. ServerName also identifies
// the site — its config file on disk is named after it — so two
// VirtualHosts with the same ServerName are the same site.
type VirtualHost struct {
	ServerName    string
	ServerAliases []string
	Listen        int
	Locations     []Location

	// TLS is reserved for a future iteration; always nil today.
	TLS *TLSConfig
}

// FileName is the config file name this VirtualHost is stored under inside
// a Server's SitesAvailableDir (e.g. "example.com.conf").
func (v VirtualHost) FileName() string {
	return v.ServerName + ".conf"
}
