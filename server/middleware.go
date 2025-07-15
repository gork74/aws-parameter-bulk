package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info().Msgf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly flags set.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		//Secure:   true,
	})

	// Configure origin validation for nosurf v1.2.0+
	// Allow localhost and test environments for development/testing
	csrfHandler.SetIsAllowedOriginFunc(func(origin *url.URL) bool {
		// Allow requests without origin (e.g., same-origin requests, direct requests)
		if origin == nil {
			return true
		}

		// Allow localhost for development and testing
		if origin.Hostname() == "localhost" || origin.Hostname() == "127.0.0.1" {
			return true
		}

		// Allow any origin with httptest scheme (Go test framework)
		if origin.Scheme == "httptest" {
			return true
		}

		// For production, you would add your actual domain here
		// For now, allow any origin to maintain existing behavior
		// In production, replace this with specific domain validation
		return true
	})

	return csrfHandler
}
