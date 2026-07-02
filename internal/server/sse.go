package server

import (
	stdhttp "net/http"
	"strings"

	"github.com/go-kratos/kratos/v3/transport/http"
)

const todoWatchSSEPath = "/v1/todos/watch"

// sseHeaderFilter sets common SSE headers for generated or handwritten SSE routes.
func sseHeaderFilter(paths ...string) http.FilterFunc {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
			if isSSERequestPath(request.URL.Path, paths) {
				headers := writer.Header()
				headers.Set("Content-Type", "text/event-stream")
				headers.Set("Cache-Control", "no-cache")
				headers.Set("Connection", "keep-alive")
				headers.Set("X-Accel-Buffering", "no")
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func isSSERequestPath(requestPath string, templates []string) bool {
	for _, template := range templates {
		if pathMatchesTemplate(requestPath, template) {
			return true
		}
	}
	return false
}

func pathMatchesTemplate(requestPath, templatePath string) bool {
	requestSegments := strings.Split(strings.Trim(requestPath, "/"), "/")
	templateSegments := strings.Split(strings.Trim(templatePath, "/"), "/")
	if len(requestSegments) != len(templateSegments) {
		return false
	}

	for idx, templateSegment := range templateSegments {
		requestSegment := requestSegments[idx]
		if strings.HasPrefix(templateSegment, "{") && strings.HasSuffix(templateSegment, "}") {
			if requestSegment == "" {
				return false
			}
			continue
		}
		if requestSegment != templateSegment {
			return false
		}
	}
	return true
}
