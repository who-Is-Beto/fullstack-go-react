package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error
type APIServer struct {
	listenAddr string
	store Storage
}
type ApiError struct {
	Error string `json:"error"`
}

func WriteJSON(r http.ResponseWriter, status int, value any) error {
	r.WriteHeader(status)
	r.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(r).Encode(value)
}

func makeHTTPHandleFunc(callback apiFunc) http.HandlerFunc { // A Decorator that handles requests from the server and returns a handler function
	return func(response http.ResponseWriter, request *http.Request) {
		if err := callback(response, request); err != nil {
			WriteJSON(response, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func (s *APIServer) handleUser(response http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case http.MethodGet: 
		return s.handleGetUsers(response)
	case http.MethodPost:
		return s.handleCreateUser(response, request)
	case http.MethodDelete:
		return s.handleDeleteUser(response, request)
	}
return fmt.Errorf("method %s not supported", request.Method)
}

func (s *APIServer) handleGetUsers(response http.ResponseWriter) error {
	users, err := s.store.GetUsers()
	if err != nil {
		return err
	}
	return WriteJSON(response, http.StatusOK, users)
}

func (s *APIServer) handleGetUser(response http.ResponseWriter, request *http.Request) error {
if request.Method == http.MethodDelete {
		return s.handleDeleteUser(response, request)
}
	id, err := getId(request)
	if err != nil {
		return err
	}

	user, err := s.store.GetUser(id)
	if err != nil {
		return err
	}
	return WriteJSON(response, http.StatusOK, user)
}

func (s *APIServer) handleCreateUser(response http.ResponseWriter, request *http.Request) error {
	createUserRequest := new(createUserRequest)
	if err := json.NewDecoder(request.Body).Decode(createUserRequest); err != nil {
		return err
	}

	user := NewAccount(createUserRequest.Username)

	if err := s.store.CreateUser(user); err != nil {
		return err
	}

	return WriteJSON(response, http.StatusCreated, user)
}

func (s *APIServer) handleDeleteUser(response http.ResponseWriter, request *http.Request) error {
	id, err := getId(request)
	if err != nil {
		return err
	}

	if err := s.store.DeleteUser(id); err != nil {
		return err
	}

	return WriteJSON(response, http.StatusOK, map[string]int{"deleted": id})
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store: store,
	}
}

func (server *APIServer) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/user", makeHTTPHandleFunc(server.handleUser))
	router.HandleFunc("/user/{id}", makeHTTPHandleFunc(server.handleGetUser))

	log.Println("Starting server on: ", server.listenAddr)

	http.ListenAndServe(server.listenAddr, router)
}

func getId(request *http.Request) (int, error) {
	idString := mux.Vars(request)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return id, fmt.Errorf("invalid id %s", idString)
	}
	return id, nil
}