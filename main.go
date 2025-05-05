package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
	w.Header().Set("Content-Type", "application/json")

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
	w.Write([]byte("Successfully added tasks"))
}

func RemoveTodoHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var todos_current []Todo
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

	if err := json.Unmarshal(bytes, &todos_current); err != nil {
		http.Error(w, "could not parse JSON", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	idStr := query.Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for i, todo := range todos_current {
		if todo.Id == id {
			todos_current = append(todos_current[:i], todos_current[i+1:]...)

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
			fmt.Fprintf(w, "Deleted todo with ID %d", id)
			return
		}
	}

	http.Error(w, "Todo not found", http.StatusNotFound)
}

func CompleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var todos_current []Todo
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

	if err := json.Unmarshal(bytes, &todos_current); err != nil {
		http.Error(w, "could not parse JSON", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	idStr := query.Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID was invalid or not available", http.StatusBadRequest)
		return
	}

	for i, todo := range todos_current {
		if todo.Id == id {
			todos_current[i].Done = true
			todos_current[i].CompletedAt = time.Now()

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
			fmt.Fprintf(w, "Completed todo with ID %d", id)
			return
		}
	}

	http.Error(w, "Todo not found", http.StatusNotFound)
}

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
	fmt.Println("Running GOTH Todo App ...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// create a location or server where readings can be retrieved from
	port := os.Getenv("PORT")
	http.HandleFunc("/ping", PingHandler)
	http.HandleFunc("/todos/get", GetTodosHandler)
	http.HandleFunc("/todos/add", AddTodoHandler)
	http.HandleFunc("/todos/delete", RemoveTodoHandler)
	http.HandleFunc("/todos/complete", CompleteTodoHandler)

	log.Fatal(http.ListenAndServe(port, nil))

}
