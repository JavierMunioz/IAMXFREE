package nginx

import "errors"

var (
	// ErrNotAvailable is returned when the Manager's host has no detected
	// Nginx installation, or its sites-available/sites-enabled directories
	// could not be located.
	ErrNotAvailable = errors.New("nginx: not available")

	// ErrSiteAlreadyExists is returned by CreateVirtualHost when a config
	// file for the given ServerName already exists.
	ErrSiteAlreadyExists = errors.New("nginx: site already exists")

	// ErrSiteNotFound is returned by UpdateVirtualHost and
	// DeleteVirtualHost when no config file exists for the given site.
	ErrSiteNotFound = errors.New("nginx: site not found")
)
