package nginx

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/*.conf.tmpl
var templatesFS embed.FS

// locationTemplates maps a LocationKind to the template that renders it.
// Adding a new kind (PHP, static files, ...) means adding a template file
// under templates/ and one entry here — nothing else in this package
// changes.
var locationTemplates = map[LocationKind]string{
	LocationKindReverseProxy: "location_reverse_proxy.conf.tmpl",
}

var renderTemplates *template.Template

func init() {
	root := template.New("nginx")
	root.Funcs(template.FuncMap{"renderLocation": renderLocation})
	tmpl, err := root.ParseFS(templatesFS, "templates/*.conf.tmpl")
	if err != nil {
		// The template set is embedded and fixed at build time — a parse
		// failure here means the templates themselves are broken, a bug
		// caught immediately by any test or build, never a runtime
		// condition a caller can hit.
		panic(fmt.Sprintf("nginx: embedded templates failed to parse: %v", err))
	}
	renderTemplates = tmpl
}

// renderLocation renders a single Location using the template registered
// for its Kind.
func renderLocation(loc Location) (string, error) {
	name, ok := locationTemplates[loc.Kind]
	if !ok {
		return "", fmt.Errorf("nginx: no template registered for location kind %q", loc.Kind)
	}
	var buf strings.Builder
	if err := renderTemplates.ExecuteTemplate(&buf, name, loc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Render turns vh into the Nginx config text for its virtual host — the
// only place in this package that produces config text, so nothing else
// builds a file by concatenating strings.
func Render(vh VirtualHost) (string, error) {
	var buf strings.Builder
	if err := renderTemplates.ExecuteTemplate(&buf, "server.conf.tmpl", vh); err != nil {
		return "", err
	}
	return buf.String(), nil
}
