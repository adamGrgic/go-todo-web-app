package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	Id          uuid.UUID `json:"id"`
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

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func NewTodoHandler(templatePaths ...string) *TodoHandler {
	tmpl := template.New("layout.html").Funcs(template.FuncMap{
		"formatDate": formatDate,
	})

	tmpl = template.Must(tmpl.ParseFiles(templatePaths...)) // <-- use 'tmpl.ParseFiles', not 'template.ParseFiles'

	return &TodoHandler{tmpl: tmpl}
}

func (h *TodoHandler) AddTodoHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	task := r.FormValue("task")
	if task == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	filename := os.Getenv("TODOS_FILE")
	todos, err := loadTodos(filename)
	if err != nil {
		http.Error(w, "Could not load todos", http.StatusInternalServerError)
		return
	}

	newTodo := Todo{
		Id:        uuid.New(),
		Task:      task,
		Done:      false,
		CreatedAt: time.Now(),
	}

	todos = append(todos, newTodo)

	if err := saveTodos(filename, todos); err != nil {
		http.Error(w, "Could not save todo", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loadTodos(filename string) ([]Todo, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func saveTodos(filename string, todos []Todo) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (h *TodoHandler) RemoveTodoHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	filename := os.Getenv("TODOS_FILE")
	todos, err := loadTodos(filename)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	deleted := false
	for i, todo := range todos {
		if todo.Id.String() == id {
			todos = append(todos[:i], todos[i+1:]...)
			deleted = true
			break
		}
	}

	if !deleted {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	if err := saveTodos(filename, todos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TodoHandler) CompleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if !isValidUUID(id) {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	filename := os.Getenv("TODOS_FILE")
	todos, err := loadTodos(filename)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	updated := false
	for i := range todos {
		if todos[i].Id.String() == id {
			todos[i].Done = true
			todos[i].CompletedAt = time.Now()
			updated = true
			break
		}
	}

	if !updated {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	if err := saveTodos(filename, todos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	data := struct {
		Tasks []Todo
	}{
		Tasks: todos,
	}

	err = h.tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "template rendering error", http.StatusInternalServerError)
	}
}
