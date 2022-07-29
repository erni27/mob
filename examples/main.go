package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erni27/mob"
)

type UserGetter interface {
	Get(context.Context, DummyUserRequest) (DummyUserResponse, error)
}

type DummyUserRequest struct {
	ID string
}

type DummyUserResponse struct {
	Name  string
	Email string
}

func GetUserHandler(u UserGetter) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var dureq DummyUserRequest
		_ = json.NewDecoder(req.Body).Decode(&dureq)
		res, _ := u.Get(req.Context(), dureq)
		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(res)
	}
}

func GetUser(rw http.ResponseWriter, req *http.Request) {
	var dureq DummyUserRequest
	_ = json.NewDecoder(req.Body).Decode(&dureq)
	res, _ := mob.Send[DummyUserRequest, DummyUserResponse](req.Context(), dureq)
	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode(res)
}

func main() {
	var handler http.HandlerFunc = GetUser
	http.ListenAndServe(":8080", handler)
}
