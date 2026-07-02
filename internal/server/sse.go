package server

import (
	stdhttp "net/http"
	"strings"

	"github.com/go-kratos/kratos/v3/transport/http"
)

const todoWatchSSEPath = "/v1/todos/watch"

// sseHeaderFilter 为 SSE 路由设置通用响应头。
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

// isSSERequestPath 判断请求路径是否命中 SSE 路由模板。
func isSSERequestPath(requestPath string, templates []string) bool {
	for _, template := range templates {
		if pathMatchesTemplate(requestPath, template) {
			return true
		}
	}
	return false
}

// pathMatchesTemplate 判断请求路径是否匹配 Kratos 风格路径模板。
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
