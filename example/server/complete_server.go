package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kenka/ginpb/example/api"
	"github.com/go-kenka/ginpb/middleware"
)

// CompleteExampleServiceImpl implements the CompleteExampleService interface
type CompleteExampleServiceImpl struct {
	// In-memory storage for demonstration
	users map[string]*api.User
	posts map[string]*api.Post
}

// NewCompleteExampleService creates a new service instance
func NewCompleteExampleService() *CompleteExampleServiceImpl {
	return &CompleteExampleServiceImpl{
		users: make(map[string]*api.User),
		posts: make(map[string]*api.Post),
	}
}

// ========== GET Request Implementations ==========

func (s *CompleteExampleServiceImpl) ListUsers(ctx context.Context, req *api.ListUsersRequest) (*api.ListUsersResponse, error) {
	fmt.Printf("ListUsers called with: page=%d, page_size=%d, sort_by=%s\n",
		req.Page, req.PageSize, req.SortBy)

	// Mock response
	users := []*api.User{
		{
			Id:       "user-001",
			Username: "alice",
			Email:    "alice@example.com",
			FullName: "Alice Smith",
			Age:      28,
			Status:   "active",
		},
		{
			Id:       "user-002",
			Username: "bob",
			Email:    "bob@example.com",
			FullName: "Bob Johnson",
			Age:      32,
			Status:   "active",
		},
	}

	return &api.ListUsersResponse{
		Users:      users,
		TotalCount: int32(len(users)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		HasNext:    false,
	}, nil
}

func (s *CompleteExampleServiceImpl) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	fmt.Printf("GetUser called with user_id=%s, fields=%v\n", req.UserId, req.Fields)

	user, exists := s.users[req.UserId]
	if !exists {
		// Create a mock user if it doesn't exist
		user = &api.User{
			Id:       req.UserId,
			Username: "demo-user",
			Email:    "demo@example.com",
			FullName: "Demo User",
			Age:      25,
			Status:   "active",
		}
	}

	response := &api.GetUserResponse{
		User: user,
	}

	if req.IncludeProfile {
		response.Profile = &api.UserProfile{
			Bio:       "This is a demo user profile",
			AvatarUrl: "https://example.com/avatar.jpg",
			IsPublic:  true,
		}
	}

	if req.IncludePosts {
		response.Posts = []*api.Post{
			{
				Id:      "post-001",
				UserId:  req.UserId,
				Title:   "My First Post",
				Content: "This is the content of my first post",
				Status:  "published",
			},
		}
	}

	return response, nil
}

func (s *CompleteExampleServiceImpl) SearchUsers(ctx context.Context, req *api.SearchUsersRequest) (*api.SearchUsersResponse, error) {
	fmt.Printf("SearchUsers called with query=%s, client_id=%s, api_key=%s\n",
		req.Query, req.ClientId, req.ApiKey)

	// Mock search results
	users := []*api.User{
		{
			Id:       "search-result-001",
			Username: req.Query + "-user",
			Email:    req.Query + "@search.com",
			FullName: "Search Result User",
			Status:   "active",
		},
	}

	return &api.SearchUsersResponse{
		Users:       users,
		TotalCount:  1,
		Query:       req.Query,
		SearchTime:  0.025,
		Suggestions: []string{"suggestion1", "suggestion2"},
	}, nil
}

// ========== POST Request Implementations ==========

func (s *CompleteExampleServiceImpl) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	fmt.Printf("CreateUser called with username=%s, email=%s\n", req.Username, req.Email)

	// Create new user
	userID := "user-" + strconv.FormatInt(time.Now().Unix(), 10)
	user := &api.User{
		Id:          userID,
		Username:    req.Username,
		Email:       req.Email,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Age:         req.Age,
		Gender:      req.Gender,
		Bio:         req.Bio,
		Status:      "active",
		Roles:       []string{"user"},
		Address:     req.Address,
		Settings:    req.Settings,
		Hobbies:     req.Hobbies,
		Languages:   req.Languages,
		SocialLinks: req.SocialLinks,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Store user
	s.users[userID] = user

	return &api.CreateUserResponse{
		User:            user,
		Message:         "User created successfully",
		ActivationToken: "activation-token-12345",
		Warnings:        []string{},
	}, nil
}

func (s *CompleteExampleServiceImpl) RegisterUser(ctx context.Context, req *api.RegisterUserRequest) (*api.RegisterUserResponse, error) {
	fmt.Printf("RegisterUser called with username=%s, email=%s\n", req.Username, req.Email)

	// Validate passwords match
	if req.Password != req.ConfirmPassword {
		return &api.RegisterUserResponse{
			Success:          false,
			Message:          "Passwords do not match",
			ValidationErrors: []string{"Password confirmation does not match"},
		}, nil
	}

	userID := "reg-user-" + strconv.FormatInt(time.Now().Unix(), 10)

	return &api.RegisterUserResponse{
		Success:       true,
		UserId:        userID,
		ActivationUrl: "https://example.com/activate?token=abc123",
		Message:       "Registration successful. Please check your email for activation.",
		Warnings:      []string{},
	}, nil
}

func (s *CompleteExampleServiceImpl) CreatePost(ctx context.Context, req *api.CreatePostRequest) (*api.CreatePostResponse, error) {
	fmt.Printf("CreatePost called with user_id=%s, title=%s\n", req.UserId, req.Title)

	// Validate authorization
	if req.Authorization == "" || req.Authorization[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid authorization header")
	}

	postID := "post-" + strconv.FormatInt(time.Now().Unix(), 10)
	post := &api.Post{
		Id:              postID,
		UserId:          req.UserId,
		Title:           req.Title,
		Content:         req.Content,
		Excerpt:         req.Excerpt,
		Category:        req.Category,
		Tags:            req.Tags,
		Status:          "draft",
		Visibility:      req.Visibility,
		AllowComments:   req.AllowComments,
		ImageUrls:       req.ImageUrls,
		AttachmentUrls:  req.AttachmentUrls,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		SeoKeywords:     req.SeoKeywords,
		CustomFields:    req.CustomFields,
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}

	if req.PublishAt != "" {
		post.PublishedAt = req.PublishAt
		post.Status = "published"
	}

	// Store post
	s.posts[postID] = post

	return &api.CreatePostResponse{
		Post:             post,
		Message:          "Post created successfully",
		EditUrl:          fmt.Sprintf("https://example.com/posts/%s/edit", postID),
		PreviewUrl:       fmt.Sprintf("https://example.com/posts/%s/preview", postID),
		RequiresApproval: false,
	}, nil
}

// ========== PUT Request Implementations ==========

func (s *CompleteExampleServiceImpl) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserResponse, error) {
	fmt.Printf("UpdateUser called with user_id=%s, username=%s\n", req.UserId, req.Username)

	// Check if user exists
	user, exists := s.users[req.UserId]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserId)
	}

	// Update user fields
	user.Username = req.Username
	user.Email = req.Email
	user.FullName = req.FullName
	user.Phone = req.Phone
	user.Age = req.Age
	user.Bio = req.Bio
	user.Status = req.Status
	user.Roles = req.Roles
	user.Address = req.Address
	user.Settings = req.Settings
	user.UpdatedAt = req.UpdatedAt
	user.Version = req.Version

	// Store updated user
	s.users[req.UserId] = user

	return &api.UpdateUserResponse{
		User:                      user,
		Message:                   "User updated successfully",
		EmailVerificationRequired: req.Email != user.Email,
		VerificationUrl:           "https://example.com/verify-email",
		UpdatedFields:             []string{"username", "email", "full_name"},
	}, nil
}

func (s *CompleteExampleServiceImpl) UpdateProfile(ctx context.Context, req *api.UpdateProfileRequest) (*api.UpdateProfileResponse, error) {
	fmt.Printf("UpdateProfile called with user_id=%s\n", req.UserId)

	return &api.UpdateProfileResponse{
		Profile:       req.Profile,
		Message:       "Profile updated successfully",
		UpdatedFields: []string{"bio", "avatar_url"},
	}, nil
}

// ========== PATCH Request Implementations ==========

func (s *CompleteExampleServiceImpl) PatchUser(ctx context.Context, req *api.PatchUserRequest) (*api.PatchUserResponse, error) {
	fmt.Printf("PatchUser called with user_id=%s\n", req.UserId)

	// Check if user exists
	user, exists := s.users[req.UserId]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserId)
	}

	var updatedFields []string
	var appliedOperations []string

	// Apply partial updates
	if req.Username != "" {
		user.Username = req.Username
		updatedFields = append(updatedFields, "username")
	}
	if req.Email != "" {
		user.Email = req.Email
		updatedFields = append(updatedFields, "email")
	}
	if req.FullName != "" {
		user.FullName = req.FullName
		updatedFields = append(updatedFields, "full_name")
	}
	if req.Status != "" {
		user.Status = req.Status
		updatedFields = append(updatedFields, "status")
	}

	// Apply array operations
	if len(req.AddRoles) > 0 {
		user.Roles = append(user.Roles, req.AddRoles...)
		appliedOperations = append(appliedOperations, "add_roles")
	}
	if len(req.RemoveRoles) > 0 {
		// Remove roles logic
		appliedOperations = append(appliedOperations, "remove_roles")
	}

	user.UpdatedAt = time.Now().Format(time.RFC3339)
	s.users[req.UserId] = user

	return &api.PatchUserResponse{
		User:              user,
		PatchedFields:     updatedFields,
		AppliedOperations: appliedOperations,
		Message:           "User patched successfully",
	}, nil
}

// ========== DELETE Request Implementations ==========

func (s *CompleteExampleServiceImpl) DeleteUser(ctx context.Context, req *api.DeleteUserRequest) (*api.DeleteUserResponse, error) {
	fmt.Printf("DeleteUser called with user_id=%s, hard_delete=%v\n", req.UserId, req.HardDelete)

	// Validate confirmation
	if req.Confirmation != "DELETE" {
		return nil, fmt.Errorf("invalid confirmation: expected 'DELETE', got '%s'", req.Confirmation)
	}

	// Check if user exists
	_, exists := s.users[req.UserId]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserId)
	}

	// Delete user
	delete(s.users, req.UserId)

	return &api.DeleteUserResponse{
		Success:          true,
		Message:          "User deleted successfully",
		DeletedAt:        time.Now().Format(time.RFC3339),
		IsRecoverable:    !req.HardDelete,
		RecoveryDeadline: time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
		BackupLocation:   "backup-storage://users/" + req.UserId,
	}, nil
}

func (s *CompleteExampleServiceImpl) BatchDeleteUsers(ctx context.Context, req *api.BatchDeleteUsersRequest) (*api.BatchDeleteUsersResponse, error) {
	fmt.Printf("BatchDeleteUsers called with %d user_ids\n", len(req.UserIds))

	var deletedIds []string
	var errors []*api.BatchError

	for _, userID := range req.UserIds {
		if _, exists := s.users[userID]; exists {
			delete(s.users, userID)
			deletedIds = append(deletedIds, userID)
		} else {
			errors = append(errors, &api.BatchError{
				Id:           userID,
				ErrorCode:    "USER_NOT_FOUND",
				ErrorMessage: "User not found",
				Details:      map[string]string{"user_id": userID},
			})
		}
	}

	return &api.BatchDeleteUsersResponse{
		TotalRequested:      int32(len(req.UserIds)),
		SuccessfullyDeleted: int32(len(deletedIds)),
		FailedDeletions:     int32(len(errors)),
		DeletedUserIds:      deletedIds,
		Errors:              errors,
		OperationId:         "batch-op-" + strconv.FormatInt(time.Now().Unix(), 10),
		Message:             fmt.Sprintf("Batch delete completed: %d successful, %d failed", len(deletedIds), len(errors)),
	}, nil
}

// ========== Complex Scenario Implementations ==========

func (s *CompleteExampleServiceImpl) GetPostComments(ctx context.Context, req *api.GetPostCommentsRequest) (*api.GetPostCommentsResponse, error) {
	fmt.Printf("GetPostComments called with post_id=%s, comment_id=%s\n", req.PostId, req.CommentId)

	// Mock comments
	comments := []*api.Comment{
		{
			Id:         "comment-001",
			PostId:     req.PostId,
			UserId:     "user-001",
			Content:    "This is a great post!",
			Status:     "published",
			LikeCount:  5,
			ReplyCount: 2,
			CreatedAt:  time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		},
	}

	return &api.GetPostCommentsResponse{
		Comments:   comments,
		TotalCount: int32(len(comments)),
		Page:       req.Page,
		PerPage:    req.PerPage,
		HasMore:    false,
		Stats: &api.CommentStats{
			TotalComments:     int32(len(comments)),
			PublishedComments: int32(len(comments)),
			HiddenComments:    0,
			TotalReplies:      2,
			AverageRating:     4.5,
			FlaggedCount:      0,
		},
	}, nil
}

func (s *CompleteExampleServiceImpl) GetUserProfile(ctx context.Context, req *api.GetUserProfileRequest) (*api.GetUserProfileResponse, error) {
	fmt.Printf("GetUserProfile called with user_id=%s\n", req.UserId)

	user, exists := s.users[req.UserId]
	if !exists {
		user = &api.User{
			Id:       req.UserId,
			Username: "demo-profile-user",
			Email:    "profile@example.com",
			FullName: "Profile Demo User",
			Status:   "active",
		}
	}

	profile := &api.UserProfile{
		Bio:       "This is a comprehensive user profile",
		AvatarUrl: "https://example.com/avatars/user.jpg",
		Website:   "https://userwebsite.com",
		Location:  "San Francisco, CA",
		BirthDate: "1990-01-01",
		Hobbies:   []string{"coding", "reading", "hiking"},
		SocialLinks: map[string]string{
			"twitter":  "https://twitter.com/user",
			"linkedin": "https://linkedin.com/in/user",
		},
		IsPublic: true,
		Verified: true,
	}

	stats := &api.UserStats{
		PostCount:      42,
		FollowerCount:  1234,
		FollowingCount: 567,
		LikeCount:      890,
		CommentCount:   123,
		EngagementRate: 8.5,
		LastActivity:   time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		ProfileViews:   5678,
	}

	return &api.GetUserProfileResponse{
		User:              user,
		Profile:           profile,
		Stats:             stats,
		RecentPosts:       []*api.Post{},
		Followers:         []*api.User{},
		IsFollowing:       false,
		CanMessage:        true,
		ProfileVisibility: "public",
	}, nil
}

// ========== Server Setup ==========

func main() {
	// Create service instance
	service := NewCompleteExampleService()

	// Create Gin router with middleware
	r := gin.New()

	// Add global middleware
	r.Use(middleware.LoggingWithConfig(middleware.LoggingConfig{
		Format: middleware.LogFormatJSON,
	}))
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Create operation-specific middleware map
	operationMiddleware := map[string][]gin.HandlerFunc{
		"CreateUser": {
			middleware.BearerAuth("demo-secret-key"),
		},
		"UpdateUser": {
			middleware.BearerAuth("demo-secret-key"),
		},
		"DeleteUser": {
			middleware.BearerAuth("admin-secret-key"),
		},
		"BatchDeleteUsers": {
			middleware.BearerAuth("admin-secret-key"),
		},
	}

	// Register service with operation-specific middleware using function options
	api.RegisterCompleteExampleServiceHTTPServer(r, service,
		api.WithCompleteExampleServiceOperationMiddlewares(operationMiddleware),
	)

	// Add health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// Add API documentation endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Complete Example Service",
			"version": "1.0.0",
			"endpoints": gin.H{
				"users": gin.H{
					"GET /api/v1/users":             "List users with pagination and filtering",
					"GET /api/v1/users/:user_id":    "Get specific user details",
					"POST /api/v1/users":            "Create new user (requires auth)",
					"PUT /api/v1/users/:user_id":    "Update user (requires auth)",
					"PATCH /api/v1/users/:user_id":  "Partial update user (requires auth)",
					"DELETE /api/v1/users/:user_id": "Delete user (requires admin auth)",
				},
				"search": gin.H{
					"GET /api/v1/users/search": "Search users with advanced filters",
				},
				"posts": gin.H{
					"POST /api/v1/users/:user_id/posts":                  "Create user post",
					"GET /api/v1/users/:user_id/posts/:post_id/comments": "Get post comments",
				},
				"profiles": gin.H{
					"GET /api/v1/users/:user_id/profile": "Get user profile",
					"PUT /api/v1/users/:user_id/profile": "Update user profile",
				},
				"utility": gin.H{
					"GET /health": "Health check endpoint",
					"GET /":       "API documentation",
				},
			},
			"middleware": gin.H{
				"global": []string{"logging", "recovery", "cors"},
				"auth_required": []string{
					"CreateUser", "UpdateUser", "DeleteUser", "BatchDeleteUsers",
				},
			},
		})
	})

	// Start server
	fmt.Println("🚀 Complete Example Server starting on :8080")
	fmt.Println("📖 API Documentation: http://localhost:8080")
	fmt.Println("💚 Health Check: http://localhost:8080/health")
	fmt.Println("🔍 Example requests:")
	fmt.Println("   GET http://localhost:8080/api/v1/users?page=1&page_size=10")
	fmt.Println("   GET http://localhost:8080/api/v1/users/search?q=test&client_id=demo&X-API-Key=demo-key")
	fmt.Println("   POST http://localhost:8080/api/v1/users (with JSON body)")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
