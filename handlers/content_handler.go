package handlers

import "net/http"

type ContentHandler struct{}

func NewContentHandler() *ContentHandler {
	return &ContentHandler{}
}

func (h *ContentHandler) GetHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/templates/layout.html")
}
