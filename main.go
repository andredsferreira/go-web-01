package main

import (
	"fmt"
	"goweb01/data"
	"goweb01/db"
	"html/template"
	"log"
	"net/http"
)

func main() {
	db.ConnectDatabase()

	files := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", files))

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("POST /register", handleRegister)
	http.HandleFunc("POST /login", handleLogin)
	http.HandleFunc("POST /logout", handleLogout)
	http.HandleFunc("GET /users", handleUsers)
	fmt.Println("server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 3 || len(password) < 3 {
		http.Error(w, "invalid username/password", http.StatusNotAcceptable)
		return
	}
	if user, _ := data.GetUserByUsername(username); user.Username == username {
		http.Error(w, "user already registered", http.StatusNotAcceptable)
		return
	}
	hp, err := hashPassword(password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	err = data.AddUser(username, hp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "user registered successfully.")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := data.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "invalid username", http.StatusUnauthorized)
		return
	}
	if !checkPasswordHash(password, user.HashedPassword) {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}
	token, err := generateJWT(username)
	if err != nil {
		http.Error(w, "error generating JWT token", http.StatusInternalServerError)
		return
	}
	csrfToken, err := generateJWT(username)
	if err != nil {
		http.Error(w, "error generating JWT token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		HttpOnly: false,
	})
	err = user.UpdateUserSession(token, csrfToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintln(w, "successful login")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := Authorize(r); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		HttpOnly: false,
	})
	user, err := data.GetUserByUsername(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	user.ClearUserSession()
	fmt.Fprintln(w, "successful logout")
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	if err := Authorize(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "hello, you are authorized!")
}
