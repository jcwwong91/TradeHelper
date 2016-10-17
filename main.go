package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	t "tradeHelper/tracker"

	"github.com/gorilla/mux"
)

var(
	port = flag.Int("port", 80, "The port to serve off of")
	web = flag.String("web", "/web", "The directory of static files for the web to serve" )
	tracker *t.Tracker
)



type handlerError struct {
	err error
	code int
}

type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// here we could do some prep work before calling the handler if we wanted to

	// call the actual handler
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.err)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.err), err.code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, "Internal server error. Check the logs.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

func getStock(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	ticker := mux.Vars(r)["ticker"]
	if ticker == "" {
		return nil, &handlerError{fmt.Errorf("No stock specified"), 400}
	}
	stock, err := tracker.GetStockConfig(ticker)
	if err != nil {
		return nil, &handlerError{err, 500}
	}
	return stock, nil
}

func addStock(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	ticker := mux.Vars(r)["ticker"]
	if ticker == "" {
		return nil, &handlerError{fmt.Errorf("No stock specified"), 400}
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &handlerError{err, 500}
	}

	tracker.TrackStock(ticker)
	var payload struct {
		// TODO: Extend with internal
	}

	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, &handlerError{err, 500}
	}

	log.Println("Tracking", ticker, "with", payload, "settings")
	return payload, nil
}

func deleteStock(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	ticker := mux.Vars(r)["ticker"]
	if ticker == "" {
		return nil, &handlerError{fmt.Errorf("No stock specified"), 400}
	}
	err := tracker.StopTrackingStock(ticker)
	if err != nil {
		return nil, &handlerError{err, 500}
	}
	return "Stock Removed", nil
}

func main() {

	flag.Parse()

	tracker = t.NewTracker()

	// handle all requests by serving a file of the same name
	fs := http.Dir(*web)
	fileHandler := http.FileServer(fs)

	// setup routes
	router := mux.NewRouter()
	router.Handle("/", http.RedirectHandler("/static/", 302))
	router.Handle("/stocks/{ticker}", handler(getStock)).Methods("GET")
	router.Handle("/stocks/{ticker}", handler(addStock)).Methods("POST")
	router.Handle("/stocks/{ticker}", handler(deleteStock)).Methods("DELETE")
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
