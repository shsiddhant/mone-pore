package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/shsiddhant/mone-pore/internal/route"
)

// AdminSetupRequired returns a middleware that checks if admin is setup.
// If admin is not setup then redirects every non admin setup url.
// If admin is setup, then redirect admin setup url to home.
// Otherwise, serve the handler.
func AdminSetupRequired(
	checkAdminSetup func(ctx context.Context) (bool, error),
) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			adminSetup, err := checkAdminSetup(r.Context())
			if err != nil {
				log.Fatalf("Fatal: Failed to verify admin setup status: %v", err)
			}

			// Allow access to /static
			if strings.HasPrefix(r.URL.Path, "/static/") {
				handler.ServeHTTP(w, r)
				return
			}

			// If admin is not setup then redirect every non admin setup url.z
			if !adminSetup && route.SetupAdminURL() != r.URL.Path {
				http.Redirect(w, r, route.SetupAdminURL(), http.StatusSeeOther)
				return
			}

			// If admin is setup, then redirect admin setup url to home.
			if adminSetup && route.SetupAdminURL() == r.URL.Path {
				http.Redirect(w, r, route.HomeURL(), http.StatusSeeOther)
				return
			}

			handler.ServeHTTP(w, r)
		})

	}
}
