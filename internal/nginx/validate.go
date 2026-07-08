package nginx

import (
	"fmt"
	"regexp"
	"strings"
)

var hostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// ValidateVirtualHost checks vh's in-memory representation for problems
// that would make it a broken or nonsensical site — before it is ever
// rendered or written to disk. It never touches the filesystem or runs
// nginx; that check (against the config actually on disk) is
// Manager.ValidateConfig.
func ValidateVirtualHost(vh VirtualHost) ValidationResult {
	var errs []string

	if strings.TrimSpace(vh.ServerName) == "" {
		errs = append(errs, "server_name is required")
	} else if !hostnamePattern.MatchString(vh.ServerName) {
		errs = append(errs, fmt.Sprintf("server_name %q is not a valid hostname", vh.ServerName))
	}

	for _, alias := range vh.ServerAliases {
		if !hostnamePattern.MatchString(alias) {
			errs = append(errs, fmt.Sprintf("server alias %q is not a valid hostname", alias))
		}
	}

	if vh.Listen < 1 || vh.Listen > 65535 {
		errs = append(errs, fmt.Sprintf("listen port %d is out of range (1-65535)", vh.Listen))
	}

	if len(vh.Locations) == 0 {
		errs = append(errs, "at least one location is required")
	}

	seenPaths := make(map[string]bool, len(vh.Locations))
	for _, loc := range vh.Locations {
		errs = append(errs, validateLocation(loc, seenPaths)...)
	}

	return ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
		Output: strings.Join(errs, "\n"),
	}
}

func validateLocation(loc Location, seenPaths map[string]bool) []string {
	var errs []string

	if !strings.HasPrefix(loc.Path, "/") {
		errs = append(errs, fmt.Sprintf("location path %q must start with \"/\"", loc.Path))
	}
	if seenPaths[loc.Path] {
		errs = append(errs, fmt.Sprintf("duplicate location path %q", loc.Path))
	}
	seenPaths[loc.Path] = true

	switch loc.Kind {
	case LocationKindReverseProxy:
		errs = append(errs, validateReverseProxyLocation(loc)...)
	default:
		errs = append(errs, fmt.Sprintf("location %q has unsupported kind %q", loc.Path, loc.Kind))
	}

	return errs
}

func validateReverseProxyLocation(loc Location) []string {
	if loc.Upstream == nil {
		return []string{fmt.Sprintf("location %q is a reverse proxy but has no upstream", loc.Path)}
	}

	var errs []string
	if strings.TrimSpace(loc.Upstream.Host) == "" {
		errs = append(errs, fmt.Sprintf("location %q upstream host is required", loc.Path))
	}
	if loc.Upstream.Port < 1 || loc.Upstream.Port > 65535 {
		errs = append(errs, fmt.Sprintf("location %q upstream port %d is out of range (1-65535)", loc.Path, loc.Upstream.Port))
	}
	return errs
}
