package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TestClient demonstrates all HTTP methods and binding types
type TestClient struct {
	baseURL string
	client  *http.Client
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Helper function to make HTTP requests
func (tc *TestClient) makeRequest(method, endpoint string, body interface{}, headers map[string]string, queryParams map[string]string) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Build URL with query parameters
	u, err := url.Parse(tc.baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if queryParams != nil {
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("🔄 %s %s\n", method, u.String())
	if len(headers) > 0 {
		fmt.Printf("   Headers: %v\n", headers)
	}
	if body != nil {
		bodyJson, _ := json.MarshalIndent(body, "   ", "  ")
		fmt.Printf("   Body: %s\n", string(bodyJson))
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Helper function to print response
func (tc *TestClient) printResponse(resp *http.Response) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Response: %d %s\n", resp.StatusCode, resp.Status)

	// Pretty print JSON response if possible
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		prettyJson, _ := json.MarshalIndent(jsonData, "   ", "  ")
		fmt.Printf("   %s\n", string(prettyJson))
	} else {
		fmt.Printf("   %s\n", string(body))
	}
	fmt.Println()
}

// Test all GET requests
func (tc *TestClient) testGETRequests() {
	fmt.Println("🟢 Testing GET Requests")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test 1: List Users with Query Parameters
	fmt.Println("Test 1: List Users with pagination and filtering")
	resp, err := tc.makeRequest("GET", "/api/v1/users", nil, nil, map[string]string{
		"page":            "1",
		"page_size":       "5",
		"sort_by":         "name",
		"sort_order":      "asc",
		"status":          "active",
		"include_deleted": "false",
	})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 2: Get User with Path Parameters
	fmt.Println("Test 2: Get specific user with path parameter")
	resp, err = tc.makeRequest("GET", "/api/v1/users/user-123", nil, nil, map[string]string{
		"fields":          "id,username,email",
		"include_profile": "true",
		"include_posts":   "true",
	})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 3: Search Users with Headers and Complex Query Parameters
	fmt.Println("Test 3: Search users with headers and advanced filtering")
	resp, err = tc.makeRequest("GET", "/api/v1/users/search", nil,
		map[string]string{
			"X-Client-ID":  "demo-client-001",
			"X-Request-ID": "req-" + fmt.Sprintf("%d", time.Now().Unix()),
			"User-Agent":   "TestClient/1.0",
			"X-API-Key":    "demo-api-key-12345",
		},
		map[string]string{
			"q":       "alice",
			"fields":  "id,username,email",
			"limit":   "10",
			"lat":     "37.7749",
			"lng":     "-122.4194",
			"radius":  "50",
			"min_age": "18",
			"max_age": "65",
			"country": "US",
			"city":    "San Francisco",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test all POST requests
func (tc *TestClient) testPOSTRequests() {
	fmt.Println("🟡 Testing POST Requests")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test 1: Create User with JSON Body
	fmt.Println("Test 1: Create user with JSON body")
	createUserBody := map[string]interface{}{
		"username":  "testuser123",
		"email":     "testuser@example.com",
		"password":  "securepassword123",
		"full_name": "Test User",
		"phone":     "12345678901",
		"age":       28,
		"gender":    "other",
		"bio":       "This is a test user created via API",
		"address": map[string]interface{}{
			"street":      "123 Test Street",
			"street2":     "Apt 4B",
			"city":        "Test City",
			"state":       "CA",
			"country":     "US",
			"postal_code": "12345",
			"latitude":    37.7749,
			"longitude":   -122.4194,
			"is_primary":  true,
			"type":        "home",
		},
		"hobbies":   []string{"coding", "reading", "hiking"},
		"languages": []string{"English", "Spanish", "French"},
		"social_links": map[string]string{
			"twitter":  "https://twitter.com/testuser",
			"linkedin": "https://linkedin.com/in/testuser",
		},
		"preferences": map[string]string{
			"theme":    "dark",
			"language": "en",
		},
		"settings": map[string]interface{}{
			"email_notifications": true,
			"push_notifications":  false,
			"theme":               "dark",
			"language":            "en",
			"timezone":            "America/Los_Angeles",
			"two_factor_enabled":  false,
			"preferences": map[string]string{
				"notification_sound": "chime",
				"auto_save":          "true",
			},
		},
		"agree_terms":          true,
		"subscribe_newsletter": true,
		"referral_code":        "REF123456",
		"tags":                 []string{"beta-user", "early-adopter"},
		"metadata": map[string]string{
			"source":   "api-test",
			"version":  "1.0",
			"test_run": "true",
		},
	}

	resp, err := tc.makeRequest("POST", "/api/v1/users", createUserBody,
		map[string]string{
			"Authorization": "Bearer demo-secret-key",
		}, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 2: Create Post with Mixed Parameters (Path + Query + Headers + JSON Body)
	fmt.Println("Test 2: Create user post with mixed parameters")
	createPostBody := map[string]interface{}{
		"title":            "My Comprehensive Test Post",
		"content":          "This is a comprehensive test post created via API to demonstrate all the various binding types and validation rules that our system supports.",
		"excerpt":          "A comprehensive test post demonstrating API capabilities",
		"category":         "technology",
		"tags":             []string{"api", "testing", "demo", "comprehensive"},
		"visibility":       "public",
		"allow_comments":   true,
		"publish_at":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"meta_title":       "Comprehensive Test Post - API Demo",
		"meta_description": "This post demonstrates comprehensive API testing capabilities with various parameter types and validation rules.",
		"seo_keywords":     []string{"api", "testing", "demo", "rest", "gin"},
		"images": []string{
			"https://example.com/images/post1.jpg",
			"https://example.com/images/post2.png",
		},
		"attachments": []string{
			"https://example.com/files/document.pdf",
		},
		"custom_fields": map[string]string{
			"template": "blog-post",
			"featured": "true",
			"priority": "high",
		},
		"external_id": "ext-post-12345",
	}

	resp, err = tc.makeRequest("POST", "/api/v1/users/user-123/posts", createPostBody,
		map[string]string{
			"Authorization":    "Bearer demo-secret-key",
			"Content-Type":     "application/json",
			"User-Agent":       "TestClient/1.0",
			"X-Client-Version": "1.2.3",
		},
		map[string]string{
			"draft":            "false",
			"source":           "api",
			"notify_followers": "true",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 3: Form Data Registration (simulated with JSON for demo)
	fmt.Println("Test 3: User registration with form-style data")
	registerBody := map[string]interface{}{
		"username":         "formuser123",
		"email":            "formuser@example.com",
		"password":         "password123",
		"confirm_password": "password123",
		"first_name":       "Form",
		"last_name":        "User",
		"birth_date":       "1990-05-15",
		"phone":            "09876543210",
		"gender":           "prefer_not_to_say",
		"country":          "US",
		"timezone":         "America/New_York",
		"interests":        []string{"technology", "science", "music"},
		"skills":           []string{"programming", "design"},
		"newsletter":       "weekly",
		"marketing_emails": true,
		"captcha":          "123456",
		"invite_code":      "INVITE2024",
		"utm_source":       "google",
		"utm_medium":       "cpc",
		"utm_campaign":     "spring2024",
		"referrer":         "https://google.com",
	}

	resp, err = tc.makeRequest("POST", "/api/v1/users/register", registerBody, nil, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test PUT requests
func (tc *TestClient) testPUTRequests() {
	fmt.Println("🔵 Testing PUT Requests")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test 1: Complete User Update
	fmt.Println("Test 1: Complete user update")
	updateUserBody := map[string]interface{}{
		"username":  "updateduser123",
		"email":     "updated@example.com",
		"full_name": "Updated Test User",
		"phone":     "19876543210",
		"age":       30,
		"bio":       "This user profile has been completely updated via PUT request",
		"status":    "active",
		"roles":     []string{"user", "premium"},
		"address": map[string]interface{}{
			"street":      "456 Updated Ave",
			"city":        "Updated City",
			"state":       "NY",
			"country":     "US",
			"postal_code": "54321",
		},
		"settings": map[string]interface{}{
			"email_notifications": false,
			"push_notifications":  true,
			"theme":               "light",
			"language":            "en",
			"timezone":            "America/New_York",
		},
		"updated_at": time.Now().Format(time.RFC3339),
		"version":    2,
	}

	resp, err := tc.makeRequest("PUT", "/api/v1/users/user-123", updateUserBody,
		map[string]string{
			"Authorization": "Bearer demo-secret-key",
			"If-Match":      "version-1",
		},
		map[string]string{
			"send_notification": "true",
			"reason":            "User requested profile update",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 2: Profile Update with Partial Body
	fmt.Println("Test 2: User profile update with partial body")
	profileUpdateBody := map[string]interface{}{
		"profile": map[string]interface{}{
			"bio":        "Updated bio via profile endpoint",
			"avatar_url": "https://example.com/new-avatar.jpg",
			"website":    "https://updated-website.com",
			"location":   "New York, NY",
			"birth_date": "1992-03-15",
			"hobbies":    []string{"photography", "travel", "cooking"},
			"social_links": map[string]string{
				"instagram": "https://instagram.com/user",
				"github":    "https://github.com/user",
			},
			"is_public": true,
			"verified":  false,
		},
	}

	resp, err = tc.makeRequest("PUT", "/api/v1/users/user-123/profile", profileUpdateBody, nil, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test PATCH requests
func (tc *TestClient) testPATCHRequests() {
	fmt.Println("🟠 Testing PATCH Requests")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test: Partial User Update
	fmt.Println("Test: Partial user update with PATCH")
	patchUserBody := map[string]interface{}{
		"username":  "patcheduser",
		"email":     "patched@example.com",
		"full_name": "Patched User Name",
		"status":    "active",
		"profile_patches": map[string]string{
			"bio":      "Updated bio via PATCH",
			"location": "Patched Location",
		},
		"settings_patches": map[string]string{
			"theme":    "dark",
			"language": "es",
		},
		"address_patches": map[string]string{
			"city":  "Patched City",
			"state": "CA",
		},
		"add_roles":    []string{"moderator"},
		"remove_roles": []string{},
		"add_tags":     []string{"patch-test", "updated"},
		"remove_tags":  []string{},
		"patch_reason": "User requested partial profile update",
		"patch_metadata": map[string]string{
			"source":    "api-patch-test",
			"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	resp, err := tc.makeRequest("PATCH", "/api/v1/users/user-123", patchUserBody,
		map[string]string{
			"Authorization":       "Bearer demo-secret-key",
			"If-Match":            "version-2",
			"If-Unmodified-Since": time.Now().Add(-1 * time.Hour).Format(http.TimeFormat),
			"X-Patch-Source":      "client-test",
		}, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test DELETE requests
func (tc *TestClient) testDELETERequests() {
	fmt.Println("🔴 Testing DELETE Requests")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test 1: Single User Deletion
	fmt.Println("Test 1: Single user deletion")
	resp, err := tc.makeRequest("DELETE", "/api/v1/users/user-to-delete", nil,
		map[string]string{
			"Authorization":    "Bearer admin-secret-key",
			"X-Confirm-Delete": "DELETE",
			"X-Admin-Token":    "admin-token-12345",
		},
		map[string]string{
			"hard_delete":   "false",
			"reason":        "User requested account deletion",
			"transfer_data": "false",
			"transfer_to":   "",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 2: Batch User Deletion
	fmt.Println("Test 2: Batch user deletion")
	resp, err = tc.makeRequest("DELETE", "/api/v1/users", nil,
		map[string]string{
			"Authorization":   "Bearer admin-secret-key",
			"X-Batch-Confirm": "BATCH_DELETE_CONFIRMED",
			"X-Operation-ID":  "batch-op-" + fmt.Sprintf("%d", time.Now().Unix()),
		},
		map[string]string{
			"user_ids":    "user-001,user-002,user-003",
			"hard_delete": "false",
			"reason":      "Bulk cleanup operation",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test complex scenarios
func (tc *TestClient) testComplexScenarios() {
	fmt.Println("🟣 Testing Complex Scenarios")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test 1: Nested Path Parameters
	fmt.Println("Test 1: Nested path parameters - Get post comments")
	resp, err := tc.makeRequest("GET", "/api/v1/users/user-123/posts/post-456/comments", nil,
		map[string]string{
			"X-User-Context":    "authenticated",
			"X-Client-Timezone": "America/Los_Angeles",
		},
		map[string]string{
			"page":            "1",
			"per_page":        "20",
			"sort":            "created_at",
			"order":           "desc",
			"status":          "published",
			"include_replies": "true",
			"include_hidden":  "false",
			"since":           time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"until":           time.Now().Format(time.RFC3339),
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Test 2: Multiple Route Bindings - User Profile
	fmt.Println("Test 2: Multiple route bindings - User profile")
	resp, err = tc.makeRequest("GET", "/api/v1/users/user-789/profile", nil,
		map[string]string{
			"X-Viewer-ID":    "viewer-123",
			"X-Access-Token": "access-token-xyz",
		},
		map[string]string{
			"sections":          "basic,stats,posts",
			"include_stats":     "true",
			"include_posts":     "true",
			"include_followers": "false",
			"context":           "friend",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// Alternative route for same endpoint
	fmt.Println("Test 3: Alternative route binding - Profile via /profiles")
	resp, err = tc.makeRequest("GET", "/api/v1/profiles/user-789/data", nil, nil,
		map[string]string{
			"data_types": "profile,posts,stats",
			"format":     "json",
			"start_date": "2024-01-01",
			"end_date":   "2024-12-31",
			"aggregate":  "true",
			"group_by":   "month",
		})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test different content types (simulated)
func (tc *TestClient) testContentTypes() {
	fmt.Println("🎨 Testing Different Content Types")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test: Multi-format data processing
	fmt.Println("Test: Multi-format data processing")

	// JSON format
	jsonData := map[string]interface{}{
		"data_type": "json",
		"format":    "structured",
		"options": map[string]string{
			"validate": "true",
			"prettify": "false",
		},
		"content":         `{"message": "Hello, World!", "timestamp": "2024-01-01T00:00:00Z"}`,
		"validate_schema": true,
		"encoding":        "utf-8",
	}

	resp, err := tc.makeRequest("POST", "/api/v1/data/process", jsonData,
		map[string]string{
			"Content-Type": "application/json",
		}, nil)
	if err != nil {
		fmt.Printf("❌ JSON Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

// Test server health and info
func (tc *TestClient) testServerInfo() {
	fmt.Println("ℹ️  Testing Server Information")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Health check
	fmt.Println("Test 1: Health check")
	resp, err := tc.makeRequest("GET", "/health", nil, nil, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}

	// API documentation
	fmt.Println("Test 2: API documentation")
	resp, err = tc.makeRequest("GET", "/", nil, nil, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		tc.printResponse(resp)
	}
}

func main() {
	fmt.Println("🚀 Starting Complete API Test Client")
	fmt.Println("Testing comprehensive HTTP methods, parameter bindings, and validation")
	fmt.Println("Server should be running on http://localhost:8080")
	fmt.Println()

	client := NewTestClient("http://localhost:8080")

	// Wait a moment to ensure server is ready
	time.Sleep(1 * time.Second)

	// Test server info first
	client.testServerInfo()

	// Test all HTTP methods
	client.testGETRequests()
	client.testPOSTRequests()
	client.testPUTRequests()
	client.testPATCHRequests()
	client.testDELETERequests()

	// Test complex scenarios
	client.testComplexScenarios()

	// Test different content types
	client.testContentTypes()

	fmt.Println("✅ All tests completed!")
	fmt.Println("Review the output above to verify all HTTP methods, parameter bindings,")
	fmt.Println("validation rules, and middleware functionality are working correctly.")
}
