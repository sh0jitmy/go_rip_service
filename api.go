package main

import (
	"encoding/json"
	"net/http"
	//"time"
)

// POST, GET, DELETEのAPIを実装
func routesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var route Route
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		addRoute(route)
		w.WriteHeader(http.StatusCreated)
	case http.MethodGet:
		routes := getRoutes()
		json.NewEncoder(w).Encode(routes)
	case http.MethodDelete:
		var route Route
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		removeRoute(route)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RIP Updateからの経路情報取得API
func ripRoutesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	routes := getRIPRoutes()
	json.NewEncoder(w).Encode(routes)
}
