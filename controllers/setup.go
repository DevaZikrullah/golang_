package controllers

import (
	"net/http"
	"test/middleware"

	"github.com/gorilla/mux"
)

func New() http.Handler {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/quests", GetAllQuests).Methods("GET")
	api.HandleFunc("/quest/{id}", GetQuest).Methods("GET")
	api.HandleFunc("/quest", CreateQuest).Methods("POST")
	api.HandleFunc("/quest/{id}", UpdateQuest).Methods("PUT")
	api.HandleFunc("/quest/{id}", DeleteQuest).Methods("DELETE")
	api.HandleFunc("/get-info", GetInfo).Methods("GET")
	api.HandleFunc("/quest-complete", QuestComplete).Methods("POST")

	users := router.PathPrefix("/users").Subrouter()
	users.HandleFunc("/register", Register).Methods("POST")
	users.HandleFunc("/login", Login).Methods("POST")

	router.HandleFunc("/login", LoginHTML).Methods("GET")
	return router
}
