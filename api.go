package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return CORS(func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	})
}

type ApiServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *ApiServer {
	return &ApiServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *ApiServer) Run() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	router := mux.NewRouter()
	// router.HandleFunc("/geolocation", withJWTAuth(makeHTTPHandleFunc(s.handleGeolocationPosition)))

	router.HandleFunc("/user", makeHTTPHandleFunc(s.handleUser))
	router.HandleFunc("/user/{id}", makeHTTPHandleFunc(s.handleUserByID))

	loggingMiddleware := LoggingMiddleware(logger)
	loggedRouter := loggingMiddleware(router)
	if err := http.ListenAndServe(s.listenAddr, loggedRouter); err != nil {
		logger.Error().Err(err)
		os.Exit(1)
	}
}

func (s *ApiServer) handleUser(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetUsers(w, r)
	case "POST":
		return s.handleCreateUser(w, r)
	default:
		return errors.New("method not allowed")
	}
}

func (s *ApiServer) handleUserByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetUser(w, r)
	case "POST":
		return s.handleUpdateUser(w, r)
	case "DELETE":
		return s.handleDeleteUser(w, r)
	default:
		return errors.New("method not allowed")
	}
}

func (s *ApiServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.store.GetUsers()
	if err != nil {
		return err
	}
	WriteJson(w, http.StatusOK, users)
	return nil
}

func (s *ApiServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	createUserReq := &CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		return err
	}
	user := NewUser(createUserReq.Sub)
	id, err := s.store.CreateUser(user)
	if err != nil {
		return err
	}
	if user.Sub != id {
		return err
	}
	user.Sub = id
	return WriteJson(w, http.StatusCreated, id)
}
func (s *ApiServer) handleGetUser(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	user, err := s.store.GetUser(id)
	if err != nil {
		return WriteJson(w, http.StatusNotFound, ApiError{Error: err.Error()})
	}
	WriteJson(w, http.StatusOK, user)
	return nil
}

func (s *ApiServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ApiServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		next(w, r)
	}
}
