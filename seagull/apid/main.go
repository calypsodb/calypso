package main

import (
	"log"
	"os"
	"net/http"
	"path"
	"time"
	"html/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/calypsodb/calypso"
)

var templates = template.Must(template.ParseGlob("../../dolphin/templates/*.html"))

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

func (s *server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	keysSlice := s.db.Keys()

	var keys [][]byte
	for key := range keysSlice {
		keys = append(keys, key)
	}

	type KeyValue struct {
		Key   string
		Value string
	}

	var data struct {
		KeyValueList []KeyValue
	}
	data.KeyValueList = make([]KeyValue, 0)

	for _, key := range keys {
		value, err := s.db.Get(key)
		if err != nil {
			http.Error(w, "Error retrieving value", http.StatusInternalServerError)
			return
		}
		kv := KeyValue{Key: string(key), Value: string(value)}
		data.KeyValueList = append(data.KeyValueList, kv)
	}

	err := templates.ExecuteTemplate(w, "browse.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	usr, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}
	dbpath := path.Join(usr, "cdb_test")

	srv, err := newServer(dbpath)
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}
	defer srv.db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/set/{key}/{value}", srv.handleSet).Methods("POST")
	r.HandleFunc("/api/get/{key}", srv.handleGet).Methods("GET")
	r.HandleFunc("/api/delete/{key}", srv.handleDelete).Methods("DELETE")
	r.HandleFunc("/api/browse", srv.handleBrowse).Methods("GET")

	srvr := &http.Server{
		Addr:		 ":7777",
		Handler:	  handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedOrigins([]string{"http://localhost:8080"}),
			handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
		)(r),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on port 7777...")
	log.Fatal(srvr.ListenAndServe())
}