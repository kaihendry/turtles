package main

import (
	"fmt"
	"net/http"
)

var PORT = 8080

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
}
