package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Todo struct {
	Id          int       `json:"id"`
	Task        string    `json:"task"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func AddTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain") // use text/plain since we're returning a simple string

	var todos_current []Todo
	filename := os.Getenv("TODOS_FILE")

	var todos_new []Todo
	err := json.NewDecoder(r.Body).Decode(&todos_new)
	if err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for i := range todos_new {
		todos_new[i].CreatedAt = time.Now()
	}

	if _, err := os.Stat(filename); err != nil {
		http.Error(w, "todo file does not exist", http.StatusNotFound)
		return
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "could not read from file", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(bytes, &todos_current); err != nil {
		http.Error(w, "could not parse JSON", http.StatusInternalServerError)
		return
	}

	todos_current = append(todos_current, todos_new...)

	saved_todos, err := json.Marshal(todos_current)
	if err != nil {
		http.Error(w, "could not marshal JSON", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(filename, saved_todos, 0644); err != nil {
		http.Error(w, "could not write to JSON file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// func RemoveTodoHandler(w http.ResponseWriter, r *http.Request) {

// }

// func CompleteTodoHandler(w http.ResponseWriter, r *http.Request) {

// }

func GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var todos []Todo
	filename := os.Getenv("TODOS_FILE")

	if _, err := os.Stat(filename); err != nil {
		http.Error(w, "todo file does not exist", http.StatusNotFound)
		return
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "could not read from file", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(bytes, &todos); err != nil {
		http.Error(w, "could not parse JSON", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		http.Error(w, "could not encode JSON", http.StatusInternalServerError)
		return
	}
}

func main() {
	fmt.Println("Creating Server ...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(time.Now())

	// create a location or server where readings can be retrieved from
	port := os.Getenv("PORT")
	http.HandleFunc("/ping", PingHandler)
	http.HandleFunc("/todos/get", GetTodosHandler)
	http.HandleFunc("/todos/add", AddTodoHandler)

	log.Fatal(http.ListenAndServe(port, nil))

}
