package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"personal-web/connection"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var Data = map[string]interface{}{
	"Title": "Personal Web",
}

type MetaData struct {
	Title     string
	IsLogin   bool
	Username  string
	FlashData string
}

var Data1 = MetaData{}

type User struct {
	Id       int
	Name1    string
	Email    string
	Password string
}

type Project struct {
	Id           int
	Name         string
	Post_date    string
	Start_date   time.Time
	End_date     time.Time
	Technologies string
	Author       string
	Description  string
	Image        string
	IsLogin      bool
}

// var Projects = []Project{
// 	{
// 		Title:     "Testing Dummy Data 0",
// 		Post_date: time.Now().String(),
// 		Author:    "Deas Aditya",
// 		Content:   "Halo ini testing dan ini merupakan dummy data",
// 	},
// }

func main() {
	router := mux.NewRouter()
	connection.DatabaseConnect()

	router.PathPrefix("/img").Handler(http.StripPrefix("/img", http.FileServer(http.Dir("./img"))))
	router.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))

	router.HandleFunc("/hello", helloworld).Methods("GET")
	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/project", project).Methods("GET")
	router.HandleFunc("/add-project", addproject).Methods("POST")
	router.HandleFunc("/contact-me", getContact).Methods("GET")
	router.HandleFunc("/Project-detail/{id}", ProjectDetail).Methods("GET")
	router.HandleFunc("/delete-project/{id}", deleteproject).Methods("GET")
	router.HandleFunc("/register", formRegister).Methods("GET")
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/login", formLogin).Methods("GET")
	router.HandleFunc("/login", login).Methods("POST")

	fmt.Println("server running on port 5000")
	http.ListenAndServe("localhost:5000", router)
}

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "aplication/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world"))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	// parsing template html
	var tmpl, err = template.ParseFiles("views/index.html")
	// error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data1.IsLogin = false
	} else {
		Data1.IsLogin = session.Values["IsLogin"].(bool)
		Data1.Username = session.Values["title"].(string)
	}

	rows, _ := connection.Conn.Query(context.Background(), "SELECT tb_projects.id, name, start_date, end_date, description, image, tb_user. title as author FROM tb_projects LEFT JOIN tb_user on tb_projects.author_id = tb_user.id ORDER BY id DESC")

	var result []Project
	for rows.Next() {
		var each = Project{}

		var err = rows.Scan(&each.Id, &each.Name, &each.Start_date, &each.End_date, &each.Description, &each.Image, &each.Author)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		result = append(result, each)
	}

	resp := map[string]interface{}{
		"Data":     Data1,
		"Projects": result,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	// parsing template html
	var tmpl, err = template.ParseFiles("views/Add My Project.html")
	// error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data1.IsLogin = false
	} else {
		Data1.IsLogin = session.Values["IsLogin"].(bool)
		Data1.Username = session.Values["title"].(string)
	}

	fm := session.Flashes("Message")

	var flashes []string

	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
	Data1.FlashData = strings.Join(flashes, "")
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func getContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	// parsing template html
	var tmpl, err = template.ParseFiles("views/contact-me.html")
	// error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	// parsing template html
	var tmpl, err = template.ParseFiles("views/detail.html")
	// error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ProjectDetail := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies, image FROM tb_projects WHERE id=$1", id).Scan(&ProjectDetail.Id, &ProjectDetail.Name, &ProjectDetail.Start_date, &ProjectDetail.End_date, &ProjectDetail.Description, &ProjectDetail.Technologies, &ProjectDetail.Image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	ProjectDetail.Author = "Deas Aditya"
	resp := map[string]interface{}{
		"Data":    Data,
		"Project": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func addproject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	ProjectName := r.PostForm.Get("Project Name")
	Description := r.PostForm.Get("Description")

	var Image string
	Image = "ImageByVariable.png"
	// NodeJS := r.PostForm.Get("Node.Js")
	// ReactJS := r.PostForm.Get("React.Js")
	// NextJS := r.PostForm.Get("Next.Js")
	// Typescript := r.PostForm.Get("Typescript")
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	author := session.Values["ID"].(int)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(name, description, image, author_id) VALUES ($1, $2, $3, $4)", ProjectName, Description, Image, author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	fmt.Println(ProjectName)
	fmt.Println(Description)
	// fmt.Println(NodeJS)
	// fmt.Println(ReactJS)
	// fmt.Println(NextJS)
	// fmt.Println(Typescript)
	// fmt.Println("Start Date : " + r.PostForm.Get("Start Date"))
	// fmt.Println("End Date : " + r.PostForm.Get("End Date"))
	// fmt.Println("Node.Js : " + r.PostForm.Get("Node.Js"))
	// fmt.Println("React.Js : " + r.PostForm.Get("React.Js"))
	// fmt.Println("Next.Js : " + r.PostForm.Get("Next.Js"))
	// fmt.Println("Typescript : " + r.PostForm.Get("Typescript"))

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteproject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	// Projects = append(Projects[:id], Projects[id+1:]...)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	temp, err := template.ParseFiles("views/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	temp.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("title")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	fmt.Println(title)
	fmt.Println(email)
	fmt.Println(password)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	fmt.Println(passwordHash)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_user(title, email, password) VALUES ($1, $2, $3);", title, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	temp, err := template.ParseFiles("views/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	temp.Execute(w, nil)
}
func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT Id, email, title, password FROM tb_user WHERE email=$1", email).Scan(&user.Id, &user.Email, &user.Name1, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = true
	session.Values["title"] = user.Name1
	session.Values["ID"] = user.Id
	session.Options.MaxAge = 10800

	session.AddFlash("Login succes", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
