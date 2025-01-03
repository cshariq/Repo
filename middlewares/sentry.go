package middlewares

import (
	"context"
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/hackclub/hackatime/utils"
)

// SentryMiddleware is a wrapper around sentryhttp to include user information to traces
type SentryMiddleware struct {
	handler http.Handler
}

func NewSentryMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		}).Handle(&SentryMiddleware{handler: h})
	}
}

func (h *SentryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "-", "-")
	h.handler.ServeHTTP(w, r.WithContext(ctx))
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if user := GetPrincipal(r); user != nil {
			hub.Scope().SetUser(sentry.User{ID: user.ID})
		}

		// Parse user agent
		userAgent := r.Header.Get("User-Agent")
		_, editor, err := utils.ParseUserAgent(userAgent)
		if err == nil && editor != "" {
			hub.Scope().SetTag("editor", editor)
		}

		// Attach request body if available
		if r.Body != nil {
			if body, err := io.ReadAll(r.Body); err == nil {
				hub.Scope().SetExtra("request_body", string(body))
			}
		}
	}
}
