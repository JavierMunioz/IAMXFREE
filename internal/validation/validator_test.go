package validation_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/validation"
)

func check(t *testing.T, v validation.Validator, value string, wantErr bool) {
	t.Helper()
	err := v(value)
	if wantErr && err == nil {
		t.Errorf("value %q: expected an error, got nil", value)
	}
	if !wantErr && err != nil {
		t.Errorf("value %q: unexpected error: %v", value, err)
	}
}

func TestRequired(t *testing.T) {
	v := validation.Required()
	check(t, v, "hello", false)
	check(t, v, "", true)
	check(t, v, "   ", true)
}

func TestOptional(t *testing.T) {
	v := validation.Optional(validation.Port())
	check(t, v, "", false)
	check(t, v, "3000", false)
	check(t, v, "not-a-port", true)
}

func TestAll(t *testing.T) {
	v := validation.All(validation.Required(), validation.Port())
	check(t, v, "", true)
	check(t, v, "abc", true)
	check(t, v, "8080", false)
}

func TestInteger(t *testing.T) {
	v := validation.Integer()
	check(t, v, "42", false)
	check(t, v, "-7", false)
	check(t, v, "3.14", true)
	check(t, v, "abc", true)
}

func TestPort(t *testing.T) {
	v := validation.Port()
	check(t, v, "1", false)
	check(t, v, "65535", false)
	check(t, v, "0", true)
	check(t, v, "65536", true)
	check(t, v, "abc", true)
}

func TestPath(t *testing.T) {
	v := validation.Path()
	check(t, v, "/srv/apps/my-api", false)
	check(t, v, "", true)
	check(t, v, "   ", true)
}

func TestURL(t *testing.T) {
	v := validation.URL()
	check(t, v, "https://example.com", false)
	check(t, v, "http://example.com/path", false)
	check(t, v, "not a url", true)
	check(t, v, "example.com", true)
}

func TestGitRepository(t *testing.T) {
	v := validation.GitRepository()
	check(t, v, "https://github.com/user/repo.git", false)
	check(t, v, "git@github.com:user/repo.git", false)
	check(t, v, "ssh://git@host/path/repo.git", false)
	check(t, v, "not a repo", true)
	check(t, v, "", true)
}

func TestDomain(t *testing.T) {
	v := validation.Domain()
	check(t, v, "example.com", false)
	check(t, v, "sub.example.co", false)
	check(t, v, "localhost", true)
	check(t, v, "-bad.com", true)
}

func TestHostname(t *testing.T) {
	v := validation.Hostname()
	check(t, v, "localhost", false)
	check(t, v, "vps-01", false)
	check(t, v, "example.com", false)
	check(t, v, "bad host", true)
}

func TestIP(t *testing.T) {
	v := validation.IP()
	check(t, v, "127.0.0.1", false)
	check(t, v, "::1", false)
	check(t, v, "999.999.999.999", true)
	check(t, v, "not-an-ip", true)
}
