package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s Server) Router() *mux.Router {
	r := mux.NewRouter()

	r.PathPrefix("/docs").Handler(http.StripPrefix("/docs", http.FileServer(http.Dir("docs"))))

	api := r.NewRoute().Subrouter()
	api.Use(s.authMw)
	api.HandleFunc("/user/get", s.getAllUserHandler()).Methods(http.MethodGet)
	api.HandleFunc("/user/get/{username}", s.getUserHandler()).Methods(http.MethodGet)

	adminAPI := api.NewRoute().Subrouter()
	adminAPI.Use(s.adminAccessMw)
	adminAPI.HandleFunc("/user/create", s.createUserHandler()).Methods(http.MethodPost)
	adminAPI.HandleFunc("/user/update-password", s.updateUserPasswordHandler()).Methods(http.MethodPost)
	adminAPI.HandleFunc("/user/update-role", s.updateUserRoleHandler()).Methods(http.MethodPost)
	adminAPI.HandleFunc("/user/update-info", s.updateUserInfoHandler()).Methods(http.MethodPost)
	adminAPI.HandleFunc("/user/delete", s.deleteUserHandler()).Methods(http.MethodPost)

	return r
}
