package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type document struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func main() {
	var (
		mu   sync.Mutex
		docs = map[string]document{}
	)

	http.HandleFunc("/api/documents.create", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		var req struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		doc := document{ID: req.Title, Title: req.Title, Text: req.Text}
		mu.Lock()
		docs[doc.ID] = doc
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": doc})
	})

	http.HandleFunc("/api/documents/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/documents/"):]
		mu.Lock()
		doc, ok := docs[id]
		mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": doc})
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
