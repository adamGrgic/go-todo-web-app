package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Todo struct {
	Id          int       `json:"id"`
	Task        string    `json:"task"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt time.Time `json:"completedAt"`
}

type TodoHandler struct {
	tmpl *template.Template
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("Jan 2, 2006 at 3:04pm")
}

func NewTodoHandler(templatePaths ...string) *TodoHandler {
	tmpl := template.New("layout.html").Funcs(template.FuncMap{
		"formatDate": formatDate,
	})

	tmpl = template.Must(tmpl.ParseFiles(templatePaths...)) // <-- use 'tmpl.ParseFiles', not 'template.ParseFiles'

	return &TodoHandler{tmpl: tmpl}
}

func (h *TodoHandler) AddTodoHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *TodoHandler) RemoveTodoHandler(w http.ResponseWriter, r *http.Request) {

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

func (h *TodoHandler) CompleteTodoHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *TodoHandler) GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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

	// Define a wrapper struct so your template can access `.Tasks`
	data := struct {
		Tasks []Todo
	}{
		Tasks: todos,
	}

	// fmt.Println("foo")

	err = h.tmpl.ExecuteTemplate(w, "layout.html", data) // ✅ Renders layout + blocks
	if err != nil {
		http.Error(w, "template rendering error", http.StatusInternalServerError)
	}
}
