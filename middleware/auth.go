package middleware

import (
	"context"
	"net/http"
	"strings"

	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/utils"
)

// UserContextKey is the key for storing user in context
type contextKey string

const UserContextKey contextKey = "user"

// AuthenticatedUser returns middleware that ensures user is logged in via session or API token
func AuthenticatedUser(sessionService *models.SessionService, apiTokenService *models.APITokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *models.User
			var err error
			
			// Try API token authentication first
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				user, err = apiTokenService.ValidateToken(token)
			} else {
				// Try session-based authentication
				user, err = utils.IsUserLoggedIn(r, sessionService)
			}
			
			if err != nil || user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that checks if user has required role
func RequireRole(minRole int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// For role-based access, we check specific role permissions
			permissions := models.GetPermissions(user.Role)
			
			switch minRole {
			case models.RoleAdministrator:
				if !permissions.CanViewAdmin {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			case models.RoleEditor:
				if !permissions.CanEditPosts {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			case models.RoleViewer:
				if !permissions.CanViewUnpublished {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission returns middleware that checks for specific permission
func RequirePermission(checkPermission func(models.UserPermissions) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			permissions := models.GetPermissions(user.Role)
			if !checkPermission(permissions) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves user from request context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// Enhanced middleware for API endpoints that supports both API tokens and the old token system
func APIAuthMiddleware(legacyToken string, apiTokenService *models.APITokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			token := strings.TrimPrefix(authHeader, "Bearer ")
			
			// First try the legacy token for backwards compatibility
			if token == legacyToken {
				// For legacy token, we don't have user context, so we continue without user in context
				next.ServeHTTP(w, r)
				return
			}
			
			// Try API token authentication
			user, err := apiTokenService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Add user to context for API endpoints
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}