// Package managers implements concrete managers for the resources IAMXFREE
// controls on the VPS: OS processes, Nginx, Apache, environment files, etc.
// Each manager owns the lifecycle operations (start/stop/reload/inspect) for
// exactly one kind of resource.
package managers
