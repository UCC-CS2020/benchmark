package main

import (
	"fmt"
	html "html/template"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/UCCNetworkingSociety/go-ldap"
	"github.com/go-chi/chi"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

var (
	conf          config
	homeTemplate  = html.Must(html.ParseFiles("templates/wrapper.html", "templates/home.html"))
	loginTemplate = html.Must(html.ParseFiles("templates/wrapper.html", "templates/login.html"))
	fileTemplate  = html.Must(html.ParseFiles("templates/wrapper.html", "templates/file.html"))
)

type config struct {
	CookieHost string `toml:"cookie_host"`
	LDAPKey    string `toml:"LDAP_Key"`
	LDAPHost   string `toml:"LDAP_Host"`
	LDAPUser   string `toml:"LDAP_User"`
	LDAPBaseDN string `toml:"LDAP_BaseDN"`
}

func init() {
	store.Options = &sessions.Options{
		Domain:   "127.0.0.1",
		MaxAge:   60 * 10,
		HttpOnly: true,
		Path:     "/",
	}
}

func loadConfig() error {
	if _, err := toml.DecodeFile("settings.conf", &conf); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func home(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/html")
	if err := homeTemplate.ExecuteTemplate(w, "main", nil); err != nil {
		fmt.Println(err)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/html")
	if err := loginTemplate.ExecuteTemplate(w, "main", nil); err != nil {
		fmt.Println(err)
	}
}

func loginSubmit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()
	user := r.PostFormValue("user")
	pass := r.PostFormValue("pass")
	u, err := ldap.GetUserFromLDAP(user, pass, conf.LDAPBaseDN, conf.LDAPUser, conf.LDAPKey, conf.LDAPHost)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	session, err := store.New(r, "id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	session.Values["user"] = u

	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/upload", http.StatusFound)
}

func file(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/html")
	if err := fileTemplate.ExecuteTemplate(w, "main", nil); err != nil {
		fmt.Println(err)
	}
}

func fileSubmit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()
	file, _, err := r.FormFile("file")
	if err != nil {
		return
	}
	defer file.Close()

	var b []byte
	file.Read(b)
	fmt.Fprint(w, b)
}

func main() {
	if loadConfig() != nil {
		return
	}

	r := chi.NewRouter()

	r.HandleFunc("/", home)

	r.Route("/login", func(r chi.Router) {
		r.HandleFunc("/", login)
		r.With(isAlreadyLoggedIn).Post("/post", loginSubmit)
	})

	r.Route("/upload", func(r chi.Router) {
		r.Use(isLoggedIn)
		r.Get("/", file)
		r.Post("/submit", fileSubmit)
	})

	fmt.Println("listening on http://localhost:8080")
	http.ListenAndServe(":8080", context.ClearHandler(r))
}
