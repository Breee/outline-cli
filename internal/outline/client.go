package outline

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type Document struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
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

func (c *Client) CreateDocument(ctx context.Context, collectionID, title, text string, publish bool) (Document, error) {
	payload := map[string]any{
		"collectionId": collectionID,
		"title":        title,
		"text":         text,
		"publish":      publish,
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
