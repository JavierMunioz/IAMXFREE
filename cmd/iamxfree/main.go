// Command iamxfree is the entrypoint binary. It stays intentionally thin and
// delegates all behavior to internal/cli.
package main

import (
	"fmt"
	"os"

	"github.com/JavierMunioz/IAMXFREE/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
