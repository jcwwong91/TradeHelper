package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
)

var(
	port = flag.Int("port", 80, "The port to serve off of")
	web = flag.String("web", "/web", "The directory of static files for the web to serve" )
)

func main() {

	flag.Parse()

	// handle all requests by serving a file of the same name
	fs := http.Dir(*web)
	fileHandler := http.FileServer(fs)

	// setup routes
	router := mux.NewRouter()
	router.Handle("/", http.RedirectHandler("/static/", 302))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	http.Handle("/", router)

	go func() {
		addr := fmt.Sprintf(":%d", *port)
		if http.ListenAndServe(addr, nil) != nil {
			log.Fatalf("Failed to start webserver")
		}
	}()

	sigChan := make(chan os.Signal)
	defer close(sigChan)
	signal.Notify(sigChan, os.Interrupt)

	s := <-sigChan
	log.Printf("Recieved signal '%s', shutting down", s)

	//TODO: Anykind of cleanup once the server is shutting down

}
