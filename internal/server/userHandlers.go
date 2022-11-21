package server

import (
	"encoding/json"
	"errors"
	"github.com/dnflash/demo-p1-go-user-management-service/internal/database"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func (s Server) createUserHandler() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Info     string `json:"info"`
	}
	type response struct {
		Success bool `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("createUserHandler: Error decoding JSON, err: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, "username must not be empty", http.StatusBadRequest)
			return
		}
		if req.Password == "" {
			http.Error(w, "password must not be empty", http.StatusBadRequest)
			return
		}
		if req.Role != "user" && req.Role != "admin" {
			http.Error(w, "role should be user or admin", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("createUserHandler: Error generating bcrypt from password, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, err = s.UserDB.InsertUser(r.Context(), database.User{
			Username: req.Username,
			Password: hashedPassword,
			Role:     req.Role,
			Info:     req.Info,
		})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				log.Printf("createUserHandler: Error duplicate key when inserting User, err: %v", err)
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}
			log.Printf("createUserHandler: Error inserting User, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response{Success: true}, http.StatusCreated)
	}
}

func (s Server) getAllUserHandler() http.HandlerFunc {
	type response []database.User
	return func(w http.ResponseWriter, r *http.Request) {
		us, err := s.UserDB.FindAllUsers(r.Context())
		if err != nil {
			log.Printf("getAllUserHandler: Error getting all Users, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response(us), http.StatusOK)
	}
}

func (s Server) getUserHandler() http.HandlerFunc {
	type response database.User
	return func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		if username == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		u, err := s.UserDB.FindUserByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			log.Printf("getUserHandler: Error getting User, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response(u), http.StatusOK)
	}
}

func (s Server) updateUserPasswordHandler() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type response struct {
		Success bool `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("updateUserPasswordHandler: Error decoding JSON, err: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("updateUserPasswordHandler: Error generating bcrypt from password, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := s.UserDB.UpdateUserPassword(r.Context(), req.Username, hashedPassword); err != nil {
			if errors.Is(err, database.ErrNoDocumentsModified) {
				s.writeJsonResponse(w, response{Success: false}, http.StatusOK)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response{Success: true}, http.StatusOK)
	}
}

func (s Server) updateUserInfoHandler() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Info     string `json:"info"`
	}
	type response struct {
		Success bool `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("updateUserInfoHandler: Error decoding JSON, err: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if err := s.UserDB.UpdateUserInfo(r.Context(), req.Username, req.Info); err != nil {
			if errors.Is(err, database.ErrNoDocumentsModified) {
				s.writeJsonResponse(w, response{Success: false}, http.StatusOK)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response{Success: true}, http.StatusOK)
	}
}

func (s Server) updateUserRoleHandler() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	type response struct {
		Success bool `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("updateUserRoleHandler: Error decoding JSON, err: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if req.Role != "user" && req.Role != "admin" {
			http.Error(w, "role should be user or admin", http.StatusBadRequest)
			return
		}

		if err := s.UserDB.UpdateUserRole(r.Context(), req.Username, req.Role); err != nil {
			if errors.Is(err, database.ErrNoDocumentsModified) {
				s.writeJsonResponse(w, response{Success: false}, http.StatusOK)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response{Success: true}, http.StatusOK)
	}
}

func (s Server) deleteUserHandler() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
	}
	type response struct {
		Success bool `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("deleteUserHandler: Error decoding JSON, err: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := s.UserDB.DeleteUserByUsername(r.Context(), req.Username)
		if err != nil {
			if errors.Is(err, database.ErrNoDocumentsModified) {
				s.writeJsonResponse(w, response{Success: false}, http.StatusOK)
				return
			}
			log.Printf("deleteUserHandler: Error deleting user with username: %s, err: %v", req.Username, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.writeJsonResponse(w, response{Success: true}, http.StatusOK)
	}
}
