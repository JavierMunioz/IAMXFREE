// Package nginx implements the Nginx manager: detecting an Nginx
// installation, and creating, updating, deleting and listing the virtual
// hosts (sites) it serves. It owns the full lifecycle of a site's config
// file — building a typed, in-memory representation, validating it, then
// rendering and writing it through a template — instead of shelling out to
// hand-built strings.
//
// This package depends only on runtimehost (every OS call — reading,
// writing, running nginx itself — goes through it) and models. It never
// imports the TUI, the Dashboard, the Wizard or the Execution Engine, and
// none of those may assume it exists yet; wiring it into the rest of
// IAMXFREE is a separate, later step.
//
// This iteration only supports reverse-proxy virtual hosts (domain ->
// localhost:port). PHP, static files, Apache, load balancing and TLS
// (Certbot/Let's Encrypt) are explicitly out of scope for now.
package nginx
