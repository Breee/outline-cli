package oidc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// BrowserLoginResult holds the outcome of a browser-based OAuth2 login.
type BrowserLoginResult struct {
	// AccessToken is the Outline OAuth2 access token.
	AccessToken string
}

// BrowserLoginSession holds the state of an in-progress browser OAuth2 login.
type BrowserLoginSession struct {
	// BrowserURL is the URL the user should open in their browser.
	BrowserURL string
	resultCh   chan callbackResult
	srv        *http.Server
	listener   net.Listener
}

// oauthServerMeta holds Outline's OAuth2 server metadata.
type oauthServerMeta struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	RegistrationEndpoint  string `json:"registration_endpoint"`
}

// dcrResponse is the response from dynamic client registration.
type dcrResponse struct {
	ClientID string `json:"client_id"`
}

// StartBrowserLogin initiates Outline's OAuth2 authorization code flow with PKCE.
// It registers a public OAuth client via DCR, starts a local callback server,
// and returns a session with the browser URL. Call Wait() to block until login completes.
func StartBrowserLogin(ctx context.Context, outlineURL string, listenPort int) (*BrowserLoginSession, error) {
	outlineURL = strings.TrimRight(outlineURL, "/")

	// Step 1: Discover Outline's OAuth2 endpoints.
	meta, err := discoverOAuth(ctx, outlineURL)
	if err != nil {
		return nil, err
	}

	// Step 2: Register a public client via DCR.
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", listenPort)
	clientID, err := registerClient(ctx, meta.RegistrationEndpoint, redirectURI)
	if err != nil {
		return nil, err
	}

	// Step 3: Generate PKCE code verifier + challenge.
	codeVerifier, err := randomString(43)
	if err != nil {
		return nil, fmt.Errorf("generating code verifier: %w", err)
	}
	challengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(challengeHash[:])

	// Generate state nonce.
	stateNonce, err := randomString(24)
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	// Step 4: Build the authorization URL.
	authURL, err := buildAuthURL(meta.AuthorizationEndpoint, clientID, redirectURI, stateNonce, codeChallenge)
	if err != nil {
		return nil, err
	}

	// Step 5: Start local HTTP server.
	resultCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			resultCh <- callbackResult{err: fmt.Errorf("OAuth error: %s: %s", errParam, desc)}
			http.Error(w, errParam+": "+desc, http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code == "" {
			resultCh <- callbackResult{err: errors.New("no code in callback")}
			http.Error(w, "no code", http.StatusBadRequest)
			return
		}
		if state != stateNonce {
			resultCh <- callbackResult{err: errors.New("state mismatch")}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		// Exchange code for token.
		token, exchangeErr := exchangeCode(r.Context(), meta.TokenEndpoint, clientID, code, redirectURI, codeVerifier)
		if exchangeErr != nil {
			resultCh <- callbackResult{err: exchangeErr}
			http.Error(w, "Token exchange failed: "+exchangeErr.Error(), http.StatusInternalServerError)
			return
		}

		resultCh <- callbackResult{token: token}
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:60px">
<h2>Authentication successful!</h2><p>You can close this tab and return to the CLI.</p>
<script>setTimeout(function(){window.close()},3000)</script></body></html>`)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", listenPort))
	if err != nil {
		return nil, fmt.Errorf("starting local server: %w", err)
	}
	srv := &http.Server{Handler: mux}

	go func() {
		if srvErr := srv.Serve(listener); srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
			resultCh <- callbackResult{err: srvErr}
		}
	}()

	return &BrowserLoginSession{
		BrowserURL: authURL,
		resultCh:   resultCh,
		srv:        srv,
		listener:   listener,
	}, nil
}

// Wait blocks until the user completes authentication or the context is cancelled.
func (s *BrowserLoginSession) Wait(ctx context.Context) (*BrowserLoginResult, error) {
	defer func() { _ = s.srv.Shutdown(context.Background()) }()

	var result callbackResult
	select {
	case result = <-s.resultCh:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if result.err != nil {
		return nil, result.err
	}
	return &BrowserLoginResult{AccessToken: result.token}, nil
}

type callbackResult struct {
	token string
	err   error
}

// discoverOAuth fetches Outline's OAuth2 server metadata.
func discoverOAuth(ctx context.Context, outlineURL string) (*oauthServerMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, outlineURL+"/.well-known/oauth-authorization-server", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching OAuth metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth metadata returned status %d", resp.StatusCode)
	}
	var meta oauthServerMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decoding OAuth metadata: %w", err)
	}
	if meta.AuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return nil, errors.New("incomplete OAuth metadata from server")
	}
	return &meta, nil
}

// registerClient registers a public OAuth client via dynamic client registration.
func registerClient(ctx context.Context, endpoint, redirectURI string) (string, error) {
	if endpoint == "" {
		return "", errors.New("server does not support dynamic client registration")
	}
	body, _ := json.Marshal(map[string]any{
		"client_name":                "outline-cli",
		"redirect_uris":              []string{redirectURI},
		"token_endpoint_auth_method": "none",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("registering OAuth client: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("client registration failed (%d): %s", resp.StatusCode, string(respBody))
	}
	var dcr dcrResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcr); err != nil {
		return "", fmt.Errorf("decoding DCR response: %w", err)
	}
	if dcr.ClientID == "" {
		return "", errors.New("no client_id in DCR response")
	}
	return dcr.ClientID, nil
}

// buildAuthURL constructs the authorization URL with PKCE parameters.
func buildAuthURL(endpoint, clientID, redirectURI, state, codeChallenge string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("scope", "read write")
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// exchangeCode exchanges an authorization code for an access token.
func exchangeCode(ctx context.Context, tokenEndpoint, clientID, code, redirectURI, codeVerifier string) (string, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error: %s", tokenResp.Error)
	}
	if tokenResp.AccessToken == "" {
		return "", errors.New("no access_token in response")
	}
	return tokenResp.AccessToken, nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
