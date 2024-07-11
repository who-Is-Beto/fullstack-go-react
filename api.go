package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
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

func PermissionDenied(response http.ResponseWriter) {
	WriteJSON(response, http.StatusBadRequest, ApiError{Error: "Permission denied"})
}

func withJWTAuthHandler(HandlerFunc http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		tokenString := request.Header.Get("Authorization")
		token, err := validateJWT(tokenString)

		if err != nil || !token.Valid {
			fmt.Println("Error validating token", err)
			PermissionDenied(response)
			return
		}
		
		userId, err := getId(request)
		if err != nil {
			PermissionDenied(response)
			return
		}
		user, err := storage.GetUserById(userId)
		
		if err != nil {
			WriteJSON(response, http.StatusBadRequest, ApiError{Error: "User not found"})
			return
		}
		
		// Change this to an encrypted password field	
		claims := token.Claims.(jwt.MapClaims) // This is returning a float valyue we need to convert it to a string
		if claims["username"] != user.Username {
			PermissionDenied(response)
			return
		}

		HandlerFunc(response, request)
	}
}

func validateJWT(tokenString string ) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("Unexpected signing method:", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("secret"), nil
	})
}

func generateJWT(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 600000,
		"username": user.Username,
		"email": user.Email,
	}

	// secretKey := os.Getenv("SECRET_JWT_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
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
		return WriteJSON(response, http.StatusBadRequest, ApiError{Error: "Error getting users, please try it later."})
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

	user, err := s.store.GetUserById(id)
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

	user, err := NewAccount(createUserRequest.Username, createUserRequest.Email, createUserRequest.Password)

	if err != nil {
		return err 
	}

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

func (s *APIServer) handleLogin(response http.ResponseWriter, request *http.Request) error {
	if request.Method != http.MethodPost {
		return fmt.Errorf("method %s not supported", request.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return err
	}

	user, err := s.store.GetUserByEmail(req.Email)

	if err != nil {
		return err
	}

	if !user.ComparePassword(req.Password) {
		return WriteJSON(response, http.StatusNotFound, ApiError{Error: "Invalid Credentials"})
	}

	token, err := generateJWT(user)
	if err != nil {
		return err
	}

	return WriteJSON(response, http.StatusOK, LoginResponse{
		Username: user.Username,
		Token: token,
	})
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store: store,
	}
}

func (server *APIServer) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/login", makeHTTPHandleFunc(server.handleLogin))
	router.HandleFunc("/user", makeHTTPHandleFunc(server.handleUser))
	router.HandleFunc("/user/{id}", withJWTAuthHandler(makeHTTPHandleFunc(server.handleGetUser), server.store))

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