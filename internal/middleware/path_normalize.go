package middleware

import (
	"net/http"
	"strings"
)

// NormalizePathMiddleware accepts redundant slashes in incoming paths before
// they reach Go's ServeMux. AquaDX accepts this form, and several ALL.Net
// clients build PowerOn as serverURL + "/sys/servlet/PowerOn" even when the
// configured server URL already has a trailing slash.
func NormalizePathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "//") {
			parts := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
			r.URL.Path = "/" + strings.Join(parts, "/")
			r.URL.RawPath = ""
		}
		next.ServeHTTP(w, r)
	})
}
