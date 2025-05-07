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
	fmt.Println("Running GOTH Todo App ...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Serve static files under /static/
	http.FileServer(http.Dir("./static"))

	taskHandler := handlers.NewTodoHandler()
	contentHandler := handlers.NewContentHandler()

	port := os.Getenv("PORT")
	http.HandleFunc("/ping", PingHandler)
	http.HandleFunc("/todos/get", taskHandler.GetTodosHandler)
	http.HandleFunc("/todos/add", taskHandler.AddTodoHandler)
	http.HandleFunc("/todos/delete", taskHandler.RemoveTodoHandler)
	http.HandleFunc("/todos/complete", taskHandler.CompleteTodoHandler)

	http.HandleFunc("/home", contentHandler.GetHomePage)

	log.Fatal(http.ListenAndServe(port, nil))

}
