package main

import "fmt"
import "log"
import "net/http"

func rootHandler(resp http.ResponseWriter, req *http.Request) {
	log.Print("access \"/root\"")
	fmt.Fprintln(resp, "Hello, world!")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	log.Print("start catmgrd")
	log.Fatal(http.ListenAndServe(":10777", mux))
}
