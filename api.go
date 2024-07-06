package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error
type APIServer struct {
	listenAddr string
}
type ApiError struct {
	Error string
}

func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)
}

func makeHTTPHandleFunc(callback apiFunc) http.HandlerFunc { // A Decorator that handles requests from the server and returns a handler function
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := callback(writer, request); err != nil {
			WriteJSON(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}


func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (server *APIServer) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/user", makeHTTPHandleFunc(server.handleUser))
	router.HandleFunc("/user/{id}", makeHTTPHandleFunc(server.handleGetUser))

	log.Println("Starting server on: ", server.listenAddr)

	http.ListenAndServe(server.listenAddr, router)
}

func (s *APIServer) handleUser(writer http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case http.MethodGet: 
		return s.handleGetUser(writer, request)
	case http.MethodPost:
		return s.handleCreateUser(writer, request)
	case http.MethodDelete:
		return s.handleDeleteUser(writer, request)
	}
return fmt.Errorf("method %s not supported", request.Method)
}

func (s *APIServer) handleGetUser(writer http.ResponseWriter, request *http.Request) error {
  id := mux.Vars(request)["id"]
	fmt.Println("Getting user with id: ", id)
	return WriteJSON(writer, http.StatusOK, &User{})
}

func (s *APIServer) handleCreateUser(writer http.ResponseWriter, request *http.Request) error {
  return nil
}

func (s *APIServer) handleDeleteUser(writer http.ResponseWriter, request *http.Request) error {
  return nil
}