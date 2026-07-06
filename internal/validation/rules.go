package validation

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Integer rejects a value that cannot be parsed as a base-10 integer.
func Integer() Validator {
	return func(value string) error {
		if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("%q is not a valid integer", value)
		}
		return nil
	}
}

// Port rejects a value that is not an integer in the 1-65535 range.
func Port() Validator {
	return func(value string) error {
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("%q is not a valid port", value)
		}
		if n < 1 || n > 65535 {
			return fmt.Errorf("port %d is out of range (1-65535)", n)
		}
		return nil
	}
}

// Path rejects a value that is blank or contains a NUL byte. It does not
// check the filesystem: the path may not exist yet (e.g. before a git clone
// runs), and that is not this validator's concern.
func Path() Validator {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return fmt.Errorf("this field is required")
		}
		if strings.ContainsRune(trimmed, 0) {
			return fmt.Errorf("path contains an invalid character")
		}
		return nil
	}
}

// URL rejects a value that is not an absolute URL with a scheme and host.
func URL() Validator {
	return func(value string) error {
		u, err := url.ParseRequestURI(strings.TrimSpace(value))
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("%q is not a valid URL", value)
		}
		return nil
	}
}

var gitSCPLikePattern = regexp.MustCompile(`^[\w.\-]+@[\w.\-]+:[\w./\-]+(\.git)?/?$`)

// GitRepository rejects a value that is neither an absolute URL
// (https://, http://, ssh://, git://) nor an scp-like git remote
// (git@host:path/repo.git).
func GitRepository() Validator {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return fmt.Errorf("this field is required")
		}

		if u, err := url.ParseRequestURI(trimmed); err == nil && u.Scheme != "" && u.Host != "" {
			return nil
		}
		if gitSCPLikePattern.MatchString(trimmed) {
			return nil
		}
		return fmt.Errorf("%q is not a valid git repository URL", value)
	}
}

var domainPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// Domain rejects a value that is not a syntactically valid domain name
// (e.g. "example.com", "sub.example.co").
func Domain() Validator {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if !domainPattern.MatchString(trimmed) {
			return fmt.Errorf("%q is not a valid domain", value)
		}
		return nil
	}
}

var hostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Hostname rejects a value that is not a syntactically valid hostname. Unlike
// Domain, a single label (e.g. "localhost", "vps-01") is accepted.
func Hostname() Validator {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if !hostnamePattern.MatchString(trimmed) {
			return fmt.Errorf("%q is not a valid hostname", value)
		}
		return nil
	}
}

// IP rejects a value that is not a valid IPv4 or IPv6 address.
func IP() Validator {
	return func(value string) error {
		if net.ParseIP(strings.TrimSpace(value)) == nil {
			return fmt.Errorf("%q is not a valid IP address", value)
		}
		return nil
	}
}
