package middleware

import (
	// "log"
	"net/http"
)

// CORSMiddleware wraps an existing http.Handler to inject universal CORS headers
func CORSMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Dynamic Origin Mirroring
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// log.Printf("[CORSMiddleware] origin %s\n", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*") // Fallback for non-browser tools like Postman
			// log.Printf("[CORSMiddleware] origin *\n")
		}

		// Universal Origin Allow (Change "*" to "http://localhost:5173" in strict production)
		// w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allowed Methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allowed Headers
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle the Preflight "OPTIONS" request from the browser
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request down to your actual API handler
		next.ServeHTTP(w, r)
	})
}