package transport

import (
	"net/http"
	"strings"
)

// Middleware used to authenticate and authorize users.
// Uses the url to check which resource is being accessed
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if we are trying to login or creating a new client
		if r.URL.Path == "/login" || r.URL.Path == "/newClient" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header["Authorization"]
		if len(authHeader) == 0 {
			ErrorResponse(w, http.StatusUnauthorized, "Not allowed")
			return
		}

		token := authHeader[0]
		user := s.AuthN.Verify(token)

		if user == nil {
			ErrorResponse(w, http.StatusUnauthorized, "Not allowed")
			return
		}

		// Try to get the requested resource from the url.
		// If not users or apartments, then it should be login/createUser
		// and not authorization is needed.
		requestedResource := getResource(r.URL.Path)
		if requestedResource != "" {
			op := getOp(r.Method)

			if !s.AuthZ.Allowed(user.Role, requestedResource, op) {
				ErrorResponse(w, http.StatusForbidden, "Not allowed")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) ContentTypeJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getOp(method string) Permission {
	meth2Perm := make(map[string]Permission)
	meth2Perm["POST"] = Create
	meth2Perm["GET"] = Read
	meth2Perm["PATCH"] = Update
	meth2Perm["DELETE"] = Delete

	return meth2Perm[strings.ToUpper(method)]
}

// Returns the desired resource by inferring it from the url
func getResource(urlPath string) string {
	if strings.HasPrefix(urlPath, "/users") {
		return "users"
	} else if strings.HasPrefix(urlPath, "/apartments") {
		return "apartments"
	}
	return ""
}
