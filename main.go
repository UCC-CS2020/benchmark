package main

import (
	_ "github.com/UCCNetworkingSociety/go-ldap"
	"github.com/go-chi/chi"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	http.ListenAndServe(":8080", r)
}