package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
  "log"

	"github.com/gorilla/mux"
	"github.com/calypsodb/calypso"
	"github.com/stretchr/testify/assert"
)

type MockServer struct {
	router *mux.Router
	db     *bitcask.Bitcask
}

func NewMockServer(db *bitcask.Bitcask) *MockServer {
	return &MockServer{
		router: mux.NewRouter(),
		db:     db,
	}
}

func (ms *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ms.router.ServeHTTP(w, r)
}

func TestSetAPI(t *testing.T) {
	db, err := bitcask.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mockServer := NewMockServer(db)

	req, err := http.NewRequest("POST", "/api/set/testkey/testvalue", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mockServer.router.HandleFunc("/api/set/{key}/{value}", mockServer.handleSetAPI)

	mockServer.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "status code should be 200")

	expectedResponse := "OK"
	assert.Equal(t, expectedResponse, rr.Body.String(), "response body should match")

	if t.Failed() {
		t.Logf("FAIL: TestSetAPI")
    log.Println("FAIL: TestSetAPI")
	} else {
		t.Logf("PASS: TestSetAPI")
    log.Println("PASS: TestSetAPI")
	}
}

func (ms *MockServer) handleSetAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]

	err := ms.db.Put([]byte(key), []byte(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}

func TestGetAPI(t *testing.T) {
	db, err := bitcask.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mockServer := NewMockServer(db)

	req, err := http.NewRequest("GET", "/api/get/testkey", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mockServer.router.HandleFunc("/api/get/{key}", mockServer.handleGetAPI)

	mockServer.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "status code should be 200")

	var response struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "testkey", response.Key, "response key should match")
	assert.Equal(t, "testvalue", response.Value, "response value should match")

	if t.Failed() {
		t.Logf("FAIL: TestGetAPI")
	} else {
		t.Logf("PASS: TestGetAPI")
	}
}

func (ms *MockServer) handleGetAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := ms.db.Get([]byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   key,
		Value: string(value),
	}

	json.NewEncoder(w).Encode(response)
}
