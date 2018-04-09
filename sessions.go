package main

import (
	"fmt"
	"net/http"

	"github.com/UCCNetworkingSociety/go-ldap"
	"github.com/gorilla/sessions"
)

var (
	store = sessions.NewCookieStore([]byte(conf.LDAPKey))
)

func isAlreadyLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		if !session.IsNew {
			if _, err := getUserFromSession(r); err != nil {
				r.Method = "GET"
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		if session.IsNew {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUserFromSession(r *http.Request) (user ldap.User, err error) {
	session, err := store.Get(r, "id")
	if err != nil {
		return
	}

	if val, ok := session.Values["user"]; ok {
		if u, ok := val.(ldap.User); ok {
			return u, nil
		}
		return user, fmt.Errorf("value not a user %v", val)
	}

	return user, fmt.Errorf("no value found")
}
