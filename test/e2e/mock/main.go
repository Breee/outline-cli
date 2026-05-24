package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type document struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	CollectionID     string `json:"collectionId,omitempty"`
	ParentDocumentID string `json:"parentDocumentId,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
	URL              string `json:"url,omitempty"`
}

type collection struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	URLId string `json:"urlId"`
}

func main() {
	var (
		mu          sync.Mutex
		docs        = map[string]document{}
		collections = map[string]collection{
			"col-1": {ID: "col-1", Name: "col-1", URLId: "col-1"},
		}
		nextDocID int
	)

	requireAuth := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, `{"ok":false,"error":"authentication required"}`, http.StatusUnauthorized)
			return false
		}
		return true
	}

	http.HandleFunc("/api/auth.info", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"data": map[string]any{
				"user": map[string]any{"name": "E2E User", "email": "e2e@test.local"},
				"team": map[string]any{"name": "E2E Team"},
			},
		})
	})

	http.HandleFunc("/api/collections.list", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		mu.Lock()
		cols := make([]collection, 0, len(collections))
		for _, c := range collections {
			cols = append(cols, c)
		}
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": cols})
	})

	http.HandleFunc("/api/collections.create", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		col := collection{ID: fmt.Sprintf("col-%s", req.Name), Name: req.Name, URLId: req.Name}
		collections[col.ID] = col
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": col})
	})

	http.HandleFunc("/api/documents.create", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			Title            string `json:"title"`
			Text             string `json:"text"`
			CollectionID     string `json:"collectionId"`
			ParentDocumentID string `json:"parentDocumentId"`
			Publish          bool   `json:"publish"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		nextDocID++
		doc := document{
			ID:               fmt.Sprintf("doc-%d", nextDocID),
			Title:            req.Title,
			Text:             req.Text,
			CollectionID:     req.CollectionID,
			ParentDocumentID: req.ParentDocumentID,
			UpdatedAt:        "2026-01-01T00:00:00.000Z",
			URL:              fmt.Sprintf("/doc/%s-%d", strings.ReplaceAll(strings.ToLower(req.Title), " ", "-"), nextDocID),
		}
		docs[doc.ID] = doc
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": doc})
	})

	http.HandleFunc("/api/documents.update", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Text  string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		doc, ok := docs[req.ID]
		if !ok {
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "not found"})
			return
		}
		doc.Title = req.Title
		doc.Text = req.Text
		docs[req.ID] = doc
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": doc})
	})

	http.HandleFunc("/api/documents.info", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		doc, ok := docs[req.ID]
		mu.Unlock()
		if !ok {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": doc})
	})

	http.HandleFunc("/api/documents.search", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			Query        string `json:"query"`
			CollectionID string `json:"collectionId"`
			Limit        int    `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		var results []map[string]any
		for _, doc := range docs {
			if req.CollectionID != "" && doc.CollectionID != req.CollectionID {
				continue
			}
			if req.Query != "" && !strings.Contains(strings.ToLower(doc.Title), strings.ToLower(req.Query)) &&
				!strings.Contains(strings.ToLower(doc.Text), strings.ToLower(req.Query)) {
				continue
			}
			results = append(results, map[string]any{
				"document": doc,
				"context":  doc.Text,
			})
			if req.Limit > 0 && len(results) >= req.Limit {
				break
			}
		}
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":         true,
			"data":       results,
			"pagination": map[string]any{"total": len(results)},
		})
	})

	http.HandleFunc("/api/documents.list", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}
		var req struct {
			CollectionID string `json:"collectionId"`
			Limit        int    `json:"limit"`
			Offset       int    `json:"offset"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		var results []document
		for _, doc := range docs {
			if req.CollectionID != "" && doc.CollectionID != req.CollectionID {
				continue
			}
			results = append(results, doc)
		}
		mu.Unlock()
		// Apply offset/limit.
		if req.Offset >= len(results) {
			results = nil
		} else {
			results = results[req.Offset:]
			if req.Limit > 0 && len(results) > req.Limit {
				results = results[:req.Limit]
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": results})
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("mock outline server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
