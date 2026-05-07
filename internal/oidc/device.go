package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type discoveryDoc struct {
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

type DeviceAuthorization struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func DeviceLogin(ctx context.Context, issuer, clientID string, scopes []string) (*oauth2.Token, DeviceAuthorization, error) {
	doc, err := discover(ctx, issuer)
	if err != nil {
		return nil, DeviceAuthorization{}, err
	}
	if doc.TokenEndpoint == "" || doc.DeviceAuthorizationEndpoint == "" {
		return nil, DeviceAuthorization{}, errors.New("issuer does not expose device authorization flow")
	}

	device, err := requestDeviceCode(ctx, doc.DeviceAuthorizationEndpoint, clientID, scopes)
	if err != nil {
		return nil, DeviceAuthorization{}, err
	}

	interval := time.Duration(device.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	token, err := pollToken(ctx, doc.TokenEndpoint, clientID, device.DeviceCode, interval)
	if err != nil {
		return nil, device, err
	}
	return token, device, nil
}

func discover(ctx context.Context, issuer string) (discoveryDoc, error) {
	url := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return discoveryDoc{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return discoveryDoc{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return discoveryDoc{}, fmt.Errorf("oidc discovery failed with status %d", resp.StatusCode)
	}
	var doc discoveryDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return discoveryDoc{}, err
	}
	return doc, nil
}

func requestDeviceCode(ctx context.Context, endpoint, clientID string, scopes []string) (DeviceAuthorization, error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	if len(scopes) > 0 {
		values.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return DeviceAuthorization{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return DeviceAuthorization{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return DeviceAuthorization{}, fmt.Errorf("device authorization failed with status %d", resp.StatusCode)
	}

	var device DeviceAuthorization
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return DeviceAuthorization{}, err
	}
	if device.DeviceCode == "" {
		return DeviceAuthorization{}, errors.New("empty device code from oidc provider")
	}
	return device, nil
}

func pollToken(ctx context.Context, tokenEndpoint, clientID, deviceCode string, interval time.Duration) (*oauth2.Token, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			values := url.Values{}
			values.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
			values.Set("client_id", clientID)
			values.Set("device_code", deviceCode)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(values.Encode()))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}

			var payload struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
				ExpiresIn   int    `json:"expires_in"`
				Error       string `json:"error"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
			resp.Body.Close()
			if decodeErr != nil {
				return nil, decodeErr
			}

			switch payload.Error {
			case "authorization_pending":
				continue
			case "slow_down":
				interval += 2 * time.Second
				ticker.Reset(interval)
				continue
			case "":
				if payload.AccessToken == "" {
					return nil, errors.New("oidc token response missing access_token")
				}
				return &oauth2.Token{AccessToken: payload.AccessToken, TokenType: payload.TokenType}, nil
			default:
				return nil, fmt.Errorf("oidc token flow failed: %s", payload.Error)
			}
		}
	}
}
