package middleware

import (
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"

	"github.com/shsiddhant/mone-pore/internal/route"
	"github.com/shsiddhant/mone-pore/internal/session"
)

// UnlockedJournalRequired returns a middleware to ensure that the
// requested journal has been unlocked in the current session.
// Requests for locked journals are redirected to home.
func UnlockedJournalRequired(
	sessionManager *scs.SessionManager,
) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			key := session.JournalUnlockedKey(journalID)

			if !sessionManager.GetBool(r.Context(), key) {
				http.Redirect(w, r, route.HomeURL(), http.StatusSeeOther)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
