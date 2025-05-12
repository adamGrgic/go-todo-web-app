package main

import (
	"fmt"
	"goth-todo/handlers"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")

	fmt.Printf("Starting up GOTH Todo App on port %s...", port)

	// Serve static files under /static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	taskHandler := handlers.NewTodoHandler("static/templates/layout.html", "static/templates/tasks.html")
	// contentHandler := handlers.NewContentHandler()

	http.HandleFunc("/ping", PingHandler)

	http.HandleFunc("/", taskHandler.GetTodosHandler)
	http.HandleFunc("/todos/add", taskHandler.AddTodoHandler)
	http.HandleFunc("/todos/delete", taskHandler.RemoveTodoHandler)
	http.HandleFunc("/todos/complete", taskHandler.CompleteTodoHandler)

	http.HandleFunc("/home", taskHandler.GetTodosHandler)

	log.Fatal(http.ListenAndServe(port, nil))

}
