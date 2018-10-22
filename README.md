# Platform.sh helper library

This is a Go library for accessing the Platform.sh environment.

## Install

Add a dependency on `github.com/platformsh/gohelper` to your application. We recommend downloading it using Go modules in Go 1.11 and later.


## Usage

Example:
```go
package main

import (
	_ "github.com/go-sql-driver/mysql"
	psh "github.com/platformsh/gohelper"
	"net/http"
)

func main() {

	p, err := psh.NewPlatformInfo()

	if err != nil {
		panic("Not in a Platform.sh Environment.")
	}

	// Set up an extremely simple web server response.
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})

    // Note the Port value used here.
	http.ListenAndServe(":"+p.Port, nil)
}
