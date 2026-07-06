// Package jsonstore implements repositories.ApplicationRepository on top of
// one JSON file per application on the local filesystem. It exists to make
// early development possible without a database; when a real database is
// introduced it can be added as a sibling package implementing the same
// interface, and callers will not need to change.
package jsonstore
