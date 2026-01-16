package azure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// mockCredential returns a static token for testing
type mockCredential struct{}

func (m *mockCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "mock-token",
		ExpiresOn: time.Now().Add(1 * time.Hour),
	}, nil
}

// newTestClient creates a client configured to use the test server
func newTestClient(serverURL string) *Client {
	return &Client{
		cred:       &mockCredential{},
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestGetCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectedUserID string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "success returns user ID",
			responseCode:   200,
			responseBody:   `{"id": "user-12345-abc"}`,
			expectedUserID: "user-12345-abc",
			expectError:    false,
		},
		{
			name:           "401 unauthorized returns auth error",
			responseCode:   401,
			responseBody:   `{"error": {"code": "Unauthorized", "message": "Token expired"}}`,
			expectedUserID: "",
			expectError:    true,
			errorContains:  "API error 401",
		},
		{
			name:           "500 server error returns error",
			responseCode:   500,
			responseBody:   `{"error": {"code": "InternalServerError", "message": "Something went wrong"}}`,
			expectedUserID: "",
			expectError:    true,
			errorContains:  "API error 500",
		},
		{
			name:           "malformed JSON returns parse error",
			responseCode:   200,
			responseBody:   `{"id": "not-closed`,
			expectedUserID: "",
			expectError:    true,
			errorContains:  "unexpected end of JSON",
		},
		{
			name:           "empty response returns empty ID",
			responseCode:   200,
			responseBody:   `{}`,
			expectedUserID: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path contains /me
				if !strings.Contains(r.URL.Path, "/me") {
					t.Errorf("expected path to contain /me, got %s", r.URL.Path)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer mock-token" {
					t.Errorf("expected Authorization header 'Bearer mock-token', got '%s'", authHeader)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			// Override the graphBaseURL by calling graphRequest with the test server URL
			// Since graphRequest is unexported and uses hardcoded URL, we test via a wrapper approach
			// For this test, we intercept by using a custom implementation

			// Create a client that will use our mock server
			// We need to test graphRequest indirectly through the actual method
			// This requires modifying how we call the API

			// Use a custom HTTP client that redirects requests
			client.httpClient = &http.Client{
				Transport: &testTransport{
					baseURL:    server.URL,
					realClient: http.DefaultTransport,
				},
				Timeout: 5 * time.Second,
			}

			ctx := context.Background()
			userID, err := client.GetCurrentUser(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if userID != tt.expectedUserID {
				t.Errorf("expected userID '%s', got '%s'", tt.expectedUserID, userID)
			}
		})
	}
}

func TestGetTenant(t *testing.T) {
	tests := []struct {
		name             string
		responseCode     int
		responseBody     string
		expectedTenantID string
		expectedName     string
		expectError      bool
		errorContains    string
	}{
		{
			name:         "success returns tenant info",
			responseCode: 200,
			responseBody: `{"value": [{"id": "tenant-abc-123", "displayName": "Contoso"}]}`,
			expectedTenantID: "tenant-abc-123",
			expectedName:     "Contoso",
			expectError:      false,
		},
		{
			name:          "404 not found returns error",
			responseCode:  404,
			responseBody:  `{"error": {"code": "NotFound", "message": "Resource not found"}}`,
			expectError:   true,
			errorContains: "API error 404",
		},
		{
			name:          "empty value array returns error",
			responseCode:  200,
			responseBody:  `{"value": []}`,
			expectError:   true,
			errorContains: "no organization found",
		},
		{
			name:          "malformed JSON returns parse error",
			responseCode:  200,
			responseBody:  `{"value": [{"id": "broken`,
			expectError:   true,
			errorContains: "unexpected end of JSON",
		},
		{
			name:         "403 forbidden returns auth error",
			responseCode: 403,
			responseBody: `{"error": {"code": "Forbidden", "message": "Insufficient privileges"}}`,
			expectError:  true,
			errorContains: "API error 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path contains /organization
				if !strings.Contains(r.URL.Path, "/organization") {
					t.Errorf("expected path to contain /organization, got %s", r.URL.Path)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer mock-token" {
					t.Errorf("expected Authorization header 'Bearer mock-token', got '%s'", authHeader)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			client.httpClient = &http.Client{
				Transport: &testTransport{
					baseURL:    server.URL,
					realClient: http.DefaultTransport,
				},
				Timeout: 5 * time.Second,
			}

			ctx := context.Background()
			tenant, err := client.GetTenant(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tenant.ID != tt.expectedTenantID {
				t.Errorf("expected tenant ID '%s', got '%s'", tt.expectedTenantID, tenant.ID)
			}
			if tenant.DisplayName != tt.expectedName {
				t.Errorf("expected tenant name '%s', got '%s'", tt.expectedName, tenant.DisplayName)
			}
		})
	}
}

func TestGraphRequestRetryBehavior(t *testing.T) {
	tests := []struct {
		name          string
		responses     []mockResponse // Series of responses to return
		expectError   bool
		errorContains string
		expectRetries int
	}{
		{
			name: "429 then success - retries and succeeds",
			responses: []mockResponse{
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 200, body: `{"id": "user-after-retry"}`},
			},
			expectError:   false,
			expectRetries: 1,
		},
		{
			name: "multiple 429 then success - retries and succeeds",
			responses: []mockResponse{
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 200, body: `{"id": "user-after-retries"}`},
			},
			expectError:   false,
			expectRetries: 2,
		},
		{
			name: "all 429 - exceeds retries and fails",
			responses: []mockResponse{
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 429, body: `{"error": "rate limited"}`},
				{code: 429, body: `{"error": "rate limited"}`},
			},
			expectError:   true,
			errorContains: "API error 429",
			expectRetries: 3, // Max retries
		},
		{
			name: "500 does not retry - fails immediately",
			responses: []mockResponse{
				{code: 500, body: `{"error": "server error"}`},
			},
			expectError:   true,
			errorContains: "API error 500",
			expectRetries: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idx := requestCount
				if idx >= len(tt.responses) {
					idx = len(tt.responses) - 1
				}
				resp := tt.responses[idx]
				requestCount++

				w.WriteHeader(resp.code)
				w.Write([]byte(resp.body))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			client.httpClient = &http.Client{
				Transport: &testTransport{
					baseURL:    server.URL,
					realClient: http.DefaultTransport,
				},
				Timeout: 30 * time.Second, // Longer timeout for retries
			}

			ctx := context.Background()
			// Clear any cached userID to force API call
			client.userID = ""
			_, err := client.GetCurrentUser(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Verify retry count: initial request + retries
			expectedRequests := tt.expectRetries + 1
			if requestCount != expectedRequests {
				t.Errorf("expected %d requests (1 initial + %d retries), got %d", expectedRequests, tt.expectRetries, requestCount)
			}
		})
	}
}

func TestGetCurrentUserCaching(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"id": "cached-user-id"})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.httpClient = &http.Client{
		Transport: &testTransport{
			baseURL:    server.URL,
			realClient: http.DefaultTransport,
		},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()

	// First call should make a request
	userID1, err := client.GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if userID1 != "cached-user-id" {
		t.Errorf("expected 'cached-user-id', got '%s'", userID1)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after first call, got %d", requestCount)
	}

	// Second call should use cached value
	userID2, err := client.GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if userID2 != "cached-user-id" {
		t.Errorf("expected 'cached-user-id', got '%s'", userID2)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after second call (cached), got %d", requestCount)
	}
}

func TestGetTenantCaching(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []map[string]string{
				{"id": "cached-tenant-id", "displayName": "Cached Tenant"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.httpClient = &http.Client{
		Transport: &testTransport{
			baseURL:    server.URL,
			realClient: http.DefaultTransport,
		},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()

	// First call should make a request
	tenant1, err := client.GetTenant(ctx)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if tenant1.ID != "cached-tenant-id" {
		t.Errorf("expected 'cached-tenant-id', got '%s'", tenant1.ID)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after first call, got %d", requestCount)
	}

	// Second call should use cached value
	tenant2, err := client.GetTenant(ctx)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if tenant2.ID != "cached-tenant-id" {
		t.Errorf("expected 'cached-tenant-id', got '%s'", tenant2.ID)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after second call (cached), got %d", requestCount)
	}
}

// mockResponse represents a single HTTP response for retry testing
type mockResponse struct {
	code int
	body string
}

// testTransport redirects requests to the test server while preserving the path
type testTransport struct {
	baseURL    string
	realClient http.RoundTripper
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect the request to our test server while keeping the path
	testURL := t.baseURL + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}

	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	return t.realClient.RoundTrip(newReq)
}
