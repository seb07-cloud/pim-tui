// Package azure provides API clients for Azure PIM, Graph, and ARM services.
// Uses azidentity.AzureCLICredential for authentication - requires `az login` before use.
// All API calls are direct HTTP requests with SDK-managed tokens (no subprocess execution).
package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

const (
	graphBaseURL = "https://graph.microsoft.com/v1.0"
	graphBetaURL = "https://graph.microsoft.com/beta"

	// PIM Governance API for Entra ID roles and groups
	pimBaseURL = "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess"
)

type Client struct {
	cred        azcore.TokenCredential
	httpClient  *http.Client
	userID      string
	tenant      *Tenant    // Cached tenant info
	tokenExpiry time.Time  // When current access token expires
}

// apiScope represents the different Azure API endpoints
type apiScope string

const (
	scopeGraph = apiScope("https://graph.microsoft.com/.default")
	scopePIM   = apiScope("https://api.azrbac.mspim.azure.com/.default")
	scopeARM   = apiScope("https://management.azure.com/.default")
)

// authRecordPath returns the path to the cached authentication record file.
func authRecordPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	pimDir := filepath.Join(configDir, "pim-tui")
	if err := os.MkdirAll(pimDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(pimDir, "auth_record.json"), nil
}

// loadAuthRecord loads a previously stored AuthenticationRecord from disk.
func loadAuthRecord() (azidentity.AuthenticationRecord, error) {
	record := azidentity.AuthenticationRecord{}
	path, err := authRecordPath()
	if err != nil {
		return record, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return record, err
	}
	err = json.Unmarshal(b, &record)
	return record, err
}

// storeAuthRecord persists an AuthenticationRecord to disk.
func storeAuthRecord(record azidentity.AuthenticationRecord) error {
	path, err := authRecordPath()
	if err != nil {
		return err
	}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

// ClearAuthRecord removes the cached authentication record.
func ClearAuthRecord() error {
	path, err := authRecordPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// GetAuthCachePath returns the path where auth credentials are cached.
// Useful for debugging authentication issues.
func GetAuthCachePath() string {
	path, err := authRecordPath()
	if err != nil {
		return fmt.Sprintf("(error: %v)", err)
	}
	return path
}

// HasCachedAuth returns true if there's a cached authentication record.
func HasCachedAuth() bool {
	path, err := authRecordPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// NewClient creates a new Azure client using Azure CLI credentials.
// Requires the user to have run `az login` before calling this function.
func NewClient() (*Client, error) {
	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure CLI credential: %w", err)
	}

	return &Client{
		cred:       cred,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// NewClientWithCachedAuth tries to create a client using cached authentication.
// Returns nil if no cached auth exists or if it's invalid, without error.
// This enables silent authentication without browser prompts on app restart.
func NewClientWithCachedAuth(ctx context.Context) (*Client, error) {
	// Try to load stored authentication record
	record, err := loadAuthRecord()
	if err != nil {
		// No cached auth available
		return nil, nil
	}

	// Create persistent cache
	c, err := cache.New(&cache.Options{Name: "pim-tui"})
	if err != nil {
		// Persistent caching not available in this environment
		return nil, nil
	}

	// Create credential with cached record
	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		AuthenticationRecord: record,
		Cache:                c,
	})
	if err != nil {
		return nil, nil
	}

	// Test if the cached credential works by getting a token silently
	_, err = cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		// Cached auth is invalid or expired, clear it
		_ = ClearAuthRecord()
		return nil, nil
	}

	return &Client{
		cred:       cred,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// AuthenticateWithBrowser performs interactive browser authentication flow.
// Opens the default browser for the user to authenticate with Azure.
// Caches the authentication for future use (no re-auth on app restart).
// Returns a new Client on success, or error on failure/timeout.
func AuthenticateWithBrowser(ctx context.Context) (*Client, error) {
	// Create persistent cache
	c, err := cache.New(&cache.Options{Name: "pim-tui"})

	var cred *azidentity.InteractiveBrowserCredential
	if err != nil {
		// Fall back to non-persistent auth if cache unavailable
		cred, err = azidentity.NewInteractiveBrowserCredential(nil)
	} else {
		cred, err = azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
			Cache: c,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create browser credential: %w", err)
	}

	// Trigger the auth flow and get authentication record
	record, err := cred.Authenticate(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("browser authentication failed: %w", err)
	}

	// Store the authentication record for future silent auth
	if storeErr := storeAuthRecord(record); storeErr != nil {
		// Log but don't fail - auth still works, just won't be cached
		fmt.Fprintf(os.Stderr, "Warning: could not cache auth: %v\n", storeErr)
	}

	return &Client{
		cred:       cred,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// TokenExpiry returns when the current access token expires.
// Returns zero time if no token has been acquired yet.
func (c *Client) TokenExpiry() time.Time {
	return c.tokenExpiry
}

// doRequest makes an authenticated HTTP request to an Azure API with retry logic.
// It handles token acquisition, request body marshaling, and exponential backoff on 429 errors.
func (c *Client) doRequest(ctx context.Context, scope apiScope, method, url string, body interface{}) ([]byte, error) {
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{string(scope)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Track token expiry for informational purposes
	c.tokenExpiry = token.ExpiresOn

	// Marshal body once outside the retry loop
	var jsonBody []byte
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Retry with exponential backoff for rate limiting
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}

		var reqBody io.Reader
		if jsonBody != nil {
			reqBody = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Retry on 429 Too Many Requests
		if resp.StatusCode == 429 && attempt < maxRetries {
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("API request failed after %d retries", maxRetries)
}

// graphRequest makes requests to the Microsoft Graph API
func (c *Client) graphRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, scopeGraph, method, url, body)
}

// pimRequest makes requests to the PIM Governance API (api.azrbac.mspim.azure.com)
func (c *Client) pimRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, scopePIM, method, url, body)
}

func (c *Client) GetCurrentUser(ctx context.Context) (string, error) {
	if c.userID != "" {
		return c.userID, nil
	}

	data, err := c.graphRequest(ctx, "GET", graphBaseURL+"/me?$select=id", nil)
	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	c.userID = result.ID
	return c.userID, nil
}

// GetCurrentUserInfo returns the user's display name and email
func (c *Client) GetCurrentUserInfo(ctx context.Context) (displayName, email string, err error) {
	data, err := c.graphRequest(ctx, "GET", graphBaseURL+"/me?$select=displayName,userPrincipalName", nil)
	if err != nil {
		return "", "", err
	}

	var result struct {
		DisplayName       string `json:"displayName"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", "", err
	}

	return result.DisplayName, result.UserPrincipalName, nil
}

func (c *Client) GetTenant(ctx context.Context) (*Tenant, error) {
	if c.tenant != nil {
		return c.tenant, nil
	}

	data, err := c.graphRequest(ctx, "GET", graphBaseURL+"/organization?$select=id,displayName", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Value) == 0 {
		return nil, fmt.Errorf("no organization found")
	}

	c.tenant = &Tenant{
		ID:          result.Value[0].ID,
		DisplayName: result.Value[0].DisplayName,
	}
	return c.tenant, nil
}
