package outline

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type AuthConfig struct {
	HeaderValue string
}

func NewAuthConfig(apiToken, username, password, oidcToken string) (AuthConfig, error) {
	switch {
	case apiToken != "":
		return AuthConfig{HeaderValue: "Bearer " + apiToken}, nil
	case oidcToken != "":
		return AuthConfig{HeaderValue: "Bearer " + oidcToken}, nil
	case username != "" || password != "":
		if username == "" || password == "" {
			return AuthConfig{}, errors.New("username and password must both be set")
		}
		creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		return AuthConfig{HeaderValue: "Basic " + creds}, nil
	default:
		return AuthConfig{}, errors.New("no auth provided: set --api-token, --oidc-access-token or basic auth flags")
	}
}

type Client struct {
	baseURL string
	auth    AuthConfig
	http    *http.Client
}

type Collection struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	URLId string `json:"urlId"`
	URL   string `json:"url"`
}

type Document struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	ParentDocumentID string `json:"parentDocumentId,omitempty"`
	CollectionID     string `json:"collectionId,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
	URL              string `json:"url,omitempty"`
}

func NewClient(baseURL string, auth AuthConfig, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("server URL is required")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), auth: auth, http: httpClient}, nil
}

// AuthInfo holds user information from the auth.info endpoint.
type AuthInfo struct {
	User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
	Team struct {
		Name string `json:"name"`
	} `json:"team"`
}

// GetAuthInfo checks the current authentication and returns user/team info.
func (c *Client) GetAuthInfo(ctx context.Context) (AuthInfo, error) {
	var response struct {
		OK   bool     `json:"ok"`
		Data AuthInfo `json:"data"`
		Err  string   `json:"error"`
	}
	if err := c.post(ctx, "/api/auth.info", map[string]any{}, &response); err != nil {
		return AuthInfo{}, err
	}
	if !response.OK {
		return AuthInfo{}, errors.New(response.Err)
	}
	return response.Data, nil
}

func (c *Client) ListCollections(ctx context.Context) ([]Collection, error) {
	var response struct {
		OK   bool         `json:"ok"`
		Data []Collection `json:"data"`
		Err  string       `json:"error"`
	}
	if err := c.post(ctx, "/api/collections.list", map[string]any{}, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, errors.New(response.Err)
	}
	return response.Data, nil
}

// ResolveCollectionID resolves a collection identifier (UUID or slug) to a UUID.
func (c *Client) ResolveCollectionID(ctx context.Context, idOrSlug string) (string, error) {
	if isUUID(idOrSlug) {
		return idOrSlug, nil
	}
	collections, err := c.ListCollections(ctx)
	if err != nil {
		return "", fmt.Errorf("listing collections: %w", err)
	}
	for _, col := range collections {
		// Match by urlId, URL slug (e.g. "test-PIVpIacwIa"), name, or UUID.
		slug := strings.TrimPrefix(col.URL, "/collection/")
		if col.URLId == idOrSlug || slug == idOrSlug || col.Name == idOrSlug {
			return col.ID, nil
		}
	}
	return "", fmt.Errorf("collection %q not found", idOrSlug)
}

// CreateCollection creates a new collection with the given name and returns its ID.
func (c *Client) CreateCollection(ctx context.Context, name string) (string, error) {
	var response struct {
		OK   bool       `json:"ok"`
		Data Collection `json:"data"`
		Err  string     `json:"error"`
	}
	payload := map[string]any{"name": name}
	if err := c.post(ctx, "/api/collections.create", payload, &response); err != nil {
		return "", err
	}
	if !response.OK {
		return "", errors.New(response.Err)
	}
	return response.Data.ID, nil
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		}
	}
	return true
}

// SearchDocuments searches for documents by title within a collection.
func (c *Client) SearchDocuments(ctx context.Context, collectionID, title string) ([]Document, error) {
	payload := map[string]any{
		"query":        title,
		"collectionId": collectionID,
	}

	var response struct {
		OK   bool `json:"ok"`
		Data []struct {
			Document Document `json:"document"`
		} `json:"data"`
		Err string `json:"error"`
	}

	if err := c.post(ctx, "/api/documents.search", payload, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, errors.New(response.Err)
	}
	var docs []Document
	for _, d := range response.Data {
		docs = append(docs, d.Document)
	}
	return docs, nil
}

// SearchResult holds a search result with context snippet.
type SearchResult struct {
	Document Document `json:"document"`
	Context  string   `json:"context"`
}

// Search performs a full-text search across documents.
func (c *Client) Search(ctx context.Context, query string, collectionID string, limit int) ([]SearchResult, error) {
	payload := map[string]any{
		"query": query,
	}
	if collectionID != "" {
		payload["collectionId"] = collectionID
	}
	if limit > 0 {
		payload["limit"] = limit
	}

	var response struct {
		OK   bool           `json:"ok"`
		Data []SearchResult `json:"data"`
		Err  string         `json:"error"`
	}

	if err := c.post(ctx, "/api/documents.search", payload, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, errors.New(response.Err)
	}
	return response.Data, nil
}

// ListDocuments lists all documents in a collection.
func (c *Client) ListDocuments(ctx context.Context, collectionID string) ([]Document, error) {
	var allDocs []Document
	offset := 0
	limit := 100

	for {
		payload := map[string]any{
			"collectionId": collectionID,
			"limit":        limit,
			"offset":       offset,
		}

		var response struct {
			OK   bool       `json:"ok"`
			Data []Document `json:"data"`
			Err  string     `json:"error"`
		}

		if err := c.post(ctx, "/api/documents.list", payload, &response); err != nil {
			return nil, err
		}
		if !response.OK {
			return nil, errors.New(response.Err)
		}
		allDocs = append(allDocs, response.Data...)
		if len(response.Data) < limit {
			break
		}
		offset += limit
	}

	return allDocs, nil
}

// ListAllDocuments lists documents, optionally filtered by collection.
// If collectionID is empty, lists documents across all collections.
func (c *Client) ListAllDocuments(ctx context.Context, collectionID string) ([]Document, error) {
	var allDocs []Document
	offset := 0
	limit := 100

	for {
		payload := map[string]any{
			"limit":  limit,
			"offset": offset,
		}
		if collectionID != "" {
			payload["collectionId"] = collectionID
		}

		var response struct {
			OK   bool       `json:"ok"`
			Data []Document `json:"data"`
			Err  string     `json:"error"`
		}

		if err := c.post(ctx, "/api/documents.list", payload, &response); err != nil {
			return nil, err
		}
		if !response.OK {
			return nil, errors.New(response.Err)
		}
		allDocs = append(allDocs, response.Data...)
		if len(response.Data) < limit {
			break
		}
		offset += limit
	}

	return allDocs, nil
}

// GetDocument retrieves a single document by ID.
func (c *Client) GetDocument(ctx context.Context, id string) (Document, error) {
	var response struct {
		OK   bool     `json:"ok"`
		Data Document `json:"data"`
		Err  string   `json:"error"`
	}

	if err := c.post(ctx, "/api/documents.info", map[string]any{"id": id}, &response); err != nil {
		return Document{}, err
	}
	if !response.OK {
		return Document{}, errors.New(response.Err)
	}
	return response.Data, nil
}

// DocumentOptions holds optional fields for create/update.
type DocumentOptions struct {
	Icon             string
	ParentDocumentID string
}

func (c *Client) CreateDocument(ctx context.Context, collectionID, title, text string, publish bool, opts DocumentOptions) (Document, error) {
	payload := map[string]any{
		"collectionId": collectionID,
		"title":        title,
		"text":         text,
		"publish":      publish,
	}
	if opts.ParentDocumentID != "" {
		payload["parentDocumentId"] = opts.ParentDocumentID
	}
	if opts.Icon != "" {
		payload["icon"] = opts.Icon
	}

	var response struct {
		OK   bool     `json:"ok"`
		Data Document `json:"data"`
		Err  string   `json:"error"`
	}

	if err := c.post(ctx, "/api/documents.create", payload, &response); err != nil {
		return Document{}, err
	}
	if !response.OK {
		if response.Err == "" {
			response.Err = "outline API request failed"
		}
		return Document{}, errors.New(response.Err)
	}
	return response.Data, nil
}

// UpdateDocument updates an existing document.
func (c *Client) UpdateDocument(ctx context.Context, docID, title, text string, opts DocumentOptions) (Document, error) {
	payload := map[string]any{
		"id":    docID,
		"title": title,
		"text":  text,
	}
	if opts.Icon != "" {
		payload["icon"] = opts.Icon
	}
	if opts.ParentDocumentID != "" {
		payload["parentDocumentId"] = opts.ParentDocumentID
	}

	var response struct {
		OK   bool     `json:"ok"`
		Data Document `json:"data"`
		Err  string   `json:"error"`
	}

	if err := c.post(ctx, "/api/documents.update", payload, &response); err != nil {
		return Document{}, err
	}
	if !response.OK {
		if response.Err == "" {
			response.Err = "outline API request failed"
		}
		return Document{}, errors.New(response.Err)
	}
	return response.Data, nil
}

// UploadAttachment uploads a local file and returns the URL to use in markdown.
func (c *Client) UploadAttachment(ctx context.Context, filePath string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", filePath, err)
	}

	name := filepath.Base(filePath)
	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Step 1: Create attachment metadata.
	var createResp struct {
		OK   bool `json:"ok"`
		Data struct {
			UploadURL string            `json:"uploadUrl"`
			Form      map[string]string `json:"form"`
			Attachment struct {
				URL string `json:"url"`
			} `json:"attachment"`
		} `json:"data"`
		Err string `json:"error"`
	}

	payload := map[string]any{
		"name":        name,
		"size":        info.Size(),
		"contentType": contentType,
	}
	if err := c.post(ctx, "/api/attachments.create", payload, &createResp); err != nil {
		return "", err
	}
	if !createResp.OK {
		return "", fmt.Errorf("attachments.create: %s", createResp.Err)
	}

	// Step 2: Upload the file via multipart form to the upload URL.
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add form fields from the create response.
	for key, val := range createResp.Data.Form {
		if key == "contentType" || key == "maxUploadSize" {
			continue
		}
		_ = writer.WriteField(key, val)
	}

	// Add the file.
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", err
	}
	writer.Close()

	uploadURL := createResp.Data.UploadURL
	if !strings.HasPrefix(uploadURL, "http") {
		uploadURL = c.baseURL + uploadURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.auth.HeaderValue)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("uploading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("file upload failed (%d): %s", resp.StatusCode, string(raw))
	}

	// Return the attachment URL (relative).
	return createResp.Data.Attachment.URL, nil
}

func (c *Client) post(ctx context.Context, path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.auth.HeaderValue)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("outline API status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
