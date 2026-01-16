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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	graphBaseURL = "https://graph.microsoft.com/v1.0"
	graphBetaURL = "https://graph.microsoft.com/beta"

	// PIM Governance API for Entra ID roles and groups
	pimBaseURL = "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess"
)

type Client struct {
	cred       azcore.TokenCredential
	pimCred    azcore.TokenCredential // Credential for PIM API
	httpClient *http.Client
	userID     string
	tenant     *Tenant // Cached tenant info
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
		pimCred:    cred, // Same credential works for all scopes
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) graphRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Retry with exponential backoff for rate limiting
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			// Reset body reader for retry
			if body != nil {
				jsonBody, _ := json.Marshal(body)
				reqBody = bytes.NewReader(jsonBody)
			}
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

	return nil, fmt.Errorf("Graph API request failed after %d retries", maxRetries)
}

// pimRequest makes requests to the PIM Governance API (api.azrbac.mspim.azure.com)
// This API uses the same token as ARM and works with Azure CLI credentials
func (c *Client) pimRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	token, err := c.pimCred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://api.azrbac.mspim.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get PIM token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Retry with exponential backoff for rate limiting
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			// Reset body reader for retry
			if body != nil {
				jsonBody, _ := json.Marshal(body)
				reqBody = bytes.NewReader(jsonBody)
			}
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
			return nil, fmt.Errorf("PIM API error %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("PIM API request failed after %d retries", maxRetries)
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
