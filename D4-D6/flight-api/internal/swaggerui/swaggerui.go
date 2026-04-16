package swaggerui

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
)

// Mount регистрирует GET /docs (Swagger UI) и GET /openapi.yaml.
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /docs", handleDocs)
	mux.HandleFunc("GET /openapi.yaml", handleOpenAPIYAML)
}

// DocsHandler — публичный http.HandlerFunc для встраивания в другие роутеры (Gin/Echo и т.п.).
func DocsHandler(w http.ResponseWriter, r *http.Request) { handleDocs(w, r) }

// OpenAPIYAMLHandler — публичный http.HandlerFunc для /openapi.yaml.
func OpenAPIYAMLHandler(w http.ResponseWriter, r *http.Request) { handleOpenAPIYAML(w, r) }

func handleDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerPage))
}

func handleOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := readOpenAPISpec()
	if err != nil {
		http.Error(w, "openapi spec not found: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(data)
}

func readOpenAPISpec() ([]byte, error) {
	if p := os.Getenv("OPENAPI_PATH"); p != "" {
		return os.ReadFile(p)
	}
	wd, _ := os.Getwd()
	candidates := []string{
		"openapi.yaml",
		filepath.Join(wd, "openapi.yaml"),
		filepath.Join(wd, "..", "openapi.yaml"),
		filepath.Join(wd, "flight-api", "openapi.yaml"),
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, errors.New("openapi.yaml (задайте OPENAPI_PATH)")
}

const swaggerPage = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8" />
  <title>Flight API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" crossorigin />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
<script>
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "/openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis],
    layout: "BaseLayout",
    tryItOutEnabled: true,
  });
};
</script>
</body>
</html>
`
