package handler

import (
	"fmt"
	"goweb01/model"
	"goweb01/service"
	"net/http"
	"text/template"
	"time"
)

var tmpl = template.Must(template.ParseGlob("static/templates/*.html"))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("jwt")
	isLoggedIn := false
	if err == nil {
		_, err := service.GetClaimsFromJWT(c.Value)
		if err == nil {
			isLoggedIn = true
		}
	}
	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
	}
	tmpl.ExecuteTemplate(w, "index", data)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "home", nil)
}

func RenderLoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "login", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	u, err := model.GetUserByUsername(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	p := r.FormValue("password")
	if !service.CheckPasswordHash(p, u.Password) {
		http.Error(w, "wrong password", http.StatusNotAcceptable)
	}
	t, err := service.GenerateJWT(u.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    t,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
	w.Header().Set("HX-Redirect", "/home")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "successful logout")
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	if !model.ValidateUser(username, password, email) {
		http.Error(w, "invalid user fields", http.StatusNotAcceptable)
		return
	}
	if user, _ := model.GetUserByUsername(username); user.Username == username || user.Email == email {
		http.Error(w, "user already registered", http.StatusNotAcceptable)
		return
	}
	hp, err := service.HashPassword(password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	err = model.AddUser(username, hp, email)
	if err != nil {
		http.Error(w, "error adding user in database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "user registered successfully")
}

func RenderRegisterHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "register", nil)
}

func GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {

}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")
	c, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	claims, err := service.GetClaimsFromJWT(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username := claims["username"].(string)
	err = model.CreatePost(username, title, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Post successfully created")
}

func RenderCreatePostHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "create-post", nil)
}

func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := model.GetAllPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "explore", posts)
	w.WriteHeader(http.StatusOK)
}
