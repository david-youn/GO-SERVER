package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// making the map
type User struct {
	Name string `json:"name"`
}

// maps id to a user -- local table
var userCache = make(map[int]User)

// making the application thread safe
// blocks all the reading and writing whenever the mutex gets locked
// safe way to sync data in multi-threaded app
var cacheMutex sync.RWMutex

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("server listening to :8080")
	http.ListenAndServe(":8080", mux)
}

func handleRoot(
	w http.ResponseWriter,
	r *http.Request,
) {
	fmt.Fprintf(w, "Hello World")
}

func deleteUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// check if user even exists in database
	if _, ok := userCache[id]; !ok {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	cacheMutex.Lock()
	// deletes key-value pair
	delete(userCache, id)
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func getUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	// can get value of path parameter id
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// locks reading
	cacheMutex.RLock()
	// retrieve user
	user, ok := userCache[id]
	// unlocks reading
	cacheMutex.RUnlock()

	// if user does not exist
	if !ok {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// want to return json representation of user
	// error can occur while converting user struct to valid json representation
	j, err := json.Marshal(user)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// writing the marshalled user to the response writer as a valid json representation
	w.WriteHeader(http.StatusOK)
	w.Write(j)

}

func createUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	// declare empty user struct but don't initialize
	// want to retrieve user data from http request
	var user User
	// creates new decoder based on body in request
	// decode information to our user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if user.Name == "" {
		http.Error(
			w,
			"name is required",
			http.StatusBadRequest,
		)
		return
	}

	// locks mutex
	cacheMutex.Lock()
	// adding user to local database in the next available spot in cache
	userCache[len(userCache)+1] = user
	// unlocks RW access to userCache
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
