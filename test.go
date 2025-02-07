package main

import(
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body>Géolocalisation réussie</body></html>")
	}                                                                                                                                                                                                                                                                        )
}