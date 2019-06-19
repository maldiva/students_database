package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tpl *template.Template

func init() {

	databaseUser := os.Getenv("POSTGRES_USER")
	databasePassword := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DB")
	databaseHost := os.Getenv("POSTGRESS_HOST")
	sqlConnection := "postgres://" + databaseUser + ":" + databasePassword + "@" + databaseHost + "/" + database + "?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", sqlConnection)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")

	tpl = template.Must(template.ParseGlob("./src/templates/*.gohtml"))
}

// export fields to templates
// fields changed to uppercase
type Student struct {
	ID         int
	FirstName  string
	LastName   string
	Department string
	GPA        float32
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/students", studentsIndex)
	http.HandleFunc("/students/show", studentsShow)
	http.HandleFunc("/students/add", studentsCreateForm)
	http.HandleFunc("/students/add/process", studentsCreateProcess)
	http.HandleFunc("/students/update", studentsUpdateForm)
	http.HandleFunc("/students/update/process", studentsUpdateProcess)
	http.HandleFunc("/students/delete/process", studentsDeleteProcess)
	http.ListenAndServe(":80", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

func studentsIndex(w http.ResponseWriter, r *http.Request) {
	// students
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM students")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	students := make([]Student, 0)
	for rows.Next() {
		student := Student{}
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Department, &student.GPA) // order matters

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		students = append(students, student)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "students.gohtml", students)
}

func studentsShow(w http.ResponseWriter, r *http.Request) {
	// students/show
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	ID := r.FormValue("id")
	if ID == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM students WHERE ID = $1", ID)

	student := Student{}
	err := row.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Department, &student.GPA)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "show.gohtml", student)
}

func studentsCreateForm(w http.ResponseWriter, r *http.Request) {
	// students/add
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func studentsCreateProcess(w http.ResponseWriter, r *http.Request) {
	// students/add/process
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	student := Student{}
	student.FirstName = r.FormValue("firstname")
	student.LastName = r.FormValue("lastname")
	student.Department = r.FormValue("department")
	GPA := r.FormValue("gpa")

	// validate form values
	if student.FirstName == "" || student.LastName == "" || student.Department == "" || GPA == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// convert form values
	f64, err := strconv.ParseFloat(GPA, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Please hit back and enter a correct number for the GPA", http.StatusNotAcceptable)
		return
	}
	student.GPA = float32(f64)

	// insert values
	assignedID := 0
	err = db.QueryRow("INSERT INTO students (firstname, lastname, department, gpa) VALUES ($1, $2, $3, $4) RETURNING id", student.FirstName, student.LastName, student.Department, student.GPA).Scan(&assignedID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	student.ID = int(assignedID)

	// confirm insertion
	tpl.ExecuteTemplate(w, "created.gohtml", student)
}

func studentsUpdateForm(w http.ResponseWriter, r *http.Request) {
	// students/update
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	ID := r.FormValue("id")
	if ID == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM students WHERE id = $1", ID)

	student := Student{}
	err := row.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Department, &student.GPA)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", student)
}

func studentsUpdateProcess(w http.ResponseWriter, r *http.Request) {
	// students/update/process
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	student := Student{}
	ID := r.FormValue("id")
	student.FirstName = r.FormValue("firstname")
	student.LastName = r.FormValue("lastname")
	student.Department = r.FormValue("department")
	GPA := r.FormValue("gpa")

	// validate form values
	if student.FirstName == "" || student.LastName == "" || student.Department == "" || GPA == "" || ID == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// convert form values
	f64, err := strconv.ParseFloat(GPA, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Please hit back and enter a correct number for the GPA", http.StatusNotAcceptable)
		return
	}
	student.GPA = float32(f64)

	id, err := strconv.ParseInt(ID, 10, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Please hit back and enter a correct meaning for the ID number", http.StatusNotAcceptable)
		return
	}
	student.ID = int(id)

	// insert values
	_, err = db.Exec("UPDATE students SET firstname=$1, lastname=$2, department=$3, gpa=$4 WHERE id=$5;", student.FirstName, student.LastName, student.Department, student.GPA, student.ID)
	if err != nil {
		http.Error(w, http.StatusText(505), 505)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "updated.gohtml", student)
}

func studentsDeleteProcess(w http.ResponseWriter, r *http.Request) {
	// students/delete/process
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	ID := r.FormValue("id")
	if ID == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// delete book
	_, err := db.Exec("DELETE FROM students WHERE id=$1;", ID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}
