package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/calypsodb/calypso"
)

type server struct {
	db *bitcask.Bitcask
}

func newServer(dbpath string) (*server, error) {
	db, err := bitcask.Open(dbpath)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	return &server{db: db}, nil
}

func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]

	err := s.db.Put([]byte(key), []byte(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := s.db.Get([]byte(key))
	if err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Write(value)
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := s.db.Delete([]byte(key))
	if err != nil {
		http.Error(w, "Error deleting key", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Key deleted"))
}

func main() {
	dbpath := "./test"
	srv, err := newServer(dbpath)
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}
	defer srv.db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/set/{key}/{value}", srv.handleSet).Methods("POST")
	r.HandleFunc("/api/get/{key}", srv.handleGet).Methods("GET")
	r.HandleFunc("/api/delete/{key}", srv.handleDelete).Methods("DELETE")

	srvr := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on port 8080...")
	log.Fatal(srvr.ListenAndServe())
}