package http

import (
	"net/http"
	"path/filepath"
)

// RegisterSwagger mounts Swagger UI and OpenAPI spec routes into your existing mux.
func RegisterSwagger(mux *http.ServeMux, swaggerDir string) {
	if swaggerDir == "" {
		return
	}
	fs := http.FileServer(http.Dir(filepath.Clean(swaggerDir)))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
}
