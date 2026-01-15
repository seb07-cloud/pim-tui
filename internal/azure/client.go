package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
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
	pimCred    azcore.TokenCredential // Cached credential for PIM API
	httpClient *http.Client
	userID     string
	tenant     *Tenant // Cached tenant info
	useAzRest  bool    // Use 'az rest' command instead of HTTP client
}

// NewClient creates a new Azure client
// It tries 'az rest' first (which handles Graph API auth properly), then falls back to SDK
func NewClient() (*Client, error) {
	// Test if 'az rest' works for Graph API calls
	// 'az rest' automatically handles token acquisition with proper scopes
	cmd := exec.Command("az", "rest", "--method", "GET", "--url", "https://graph.microsoft.com/v1.0/me?$select=id", "--query", "id", "-o", "tsv")
	if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
		return &Client{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			useAzRest:  true,
		}, nil
	}

	// Fall back to DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return &Client{
		cred:       cred,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		useAzRest:  false,
	}, nil
}

func (c *Client) graphRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	if c.useAzRest {
		return c.azRestRequest(method, url, body)
	}

	// Use SDK credential
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
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// azRestRequest uses 'az rest' command which handles Graph API auth properly
func (c *Client) azRestRequest(method, url string, body interface{}) ([]byte, error) {
	args := []string{"rest", "--method", method, "--url", url}

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		args = append(args, "--body", string(jsonBody))
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("az rest failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("az rest failed: %w", err)
	}

	return output, nil
}

// pimRequest makes requests to the PIM Governance API (api.azrbac.mspim.azure.com)
// This API uses the same token as ARM and works with Azure CLI credentials
func (c *Client) pimRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	// Lazily initialize PIM credential (cached for all subsequent calls)
	if c.pimCred == nil {
		cred, err := azidentity.NewAzureCLICredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create CLI credential: %w", err)
		}
		c.pimCred = cred
	}

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
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("PIM API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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
