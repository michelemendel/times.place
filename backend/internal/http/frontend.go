package http

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// getFrontendFS returns a filesystem for serving frontend assets and the absolute path used.
// Uses the build directory on disk (frontend/build/).
// Checks FRONTEND_BUILD_DIR first, then paths relative to current working directory.
func getFrontendFS() (fs.FS, string, error) {
	cwd, _ := os.Getwd()

	// 1. Explicit env (absolute or relative to CWD)
	if dir := os.Getenv("FRONTEND_BUILD_DIR"); dir != "" {
		absPath, err := filepath.Abs(dir)
		if err == nil {
			if info, err := os.Stat(absPath); err == nil && info.IsDir() {
				return os.DirFS(absPath), absPath, nil
			}
		}
	}

	// 2. Paths relative to current working directory (covers workspace root, backend/, etc.)
	candidates := []string{
		filepath.Join(cwd, "frontend", "build"),     // CWD is workspace root
		filepath.Join(cwd, "..", "frontend", "build"), // CWD is backend/
		filepath.Join(cwd, "..", "..", "frontend", "build"),
	}
	// Legacy relative paths (resolved against CWD)
	candidates = append(candidates,
		filepath.Join("..", "..", "frontend", "build"),
		"frontend/build",
	)

	for _, absPath := range candidates {
		// Resolve to absolute if not already
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(cwd, absPath)
		}
		absPath = filepath.Clean(absPath)
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return os.DirFS(absPath), absPath, nil
		}
	}

	return nil, "", os.ErrNotExist
}

// setupFrontendRoutes sets up routes for serving frontend static files and SPA fallback
// This should be called AFTER API routes are registered
func setupFrontendRoutes(e *echo.Echo) error {
	frontendFS, usedPath, err := getFrontendFS()
	if err != nil {
		return err
	}
	e.Logger.Infof("Serving frontend from %s", usedPath)

	// Read index.html for SPA fallback
	var indexFile []byte
	indexFile, err = fs.ReadFile(frontendFS, "index.html")
	if err != nil {
		// If index.html doesn't exist, the frontend hasn't been built yet
		// This is okay - the route will just return 404
		indexFile = nil
	}

	// Create a static file server
	fileSystem := http.FS(frontendFS)
	fileServer := http.FileServer(fileSystem)

	// Register catch-all route for frontend
	// This must be registered AFTER all API routes
	e.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path

		// Don't serve frontend for API routes
		if strings.HasPrefix(path, "/api") {
			return c.NoContent(http.StatusNotFound)
		}

		// Try to serve the requested file
		// Remove leading slash for filesystem lookup
		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Check if file exists
		if _, err := fs.Stat(frontendFS, filePath); err == nil {
			// File exists, serve it
			return echo.WrapHandler(fileServer)(c)
		}

		// File doesn't exist, serve index.html for SPA routing (if available)
		if indexFile != nil {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			return c.Blob(http.StatusOK, echo.MIMETextHTMLCharsetUTF8, indexFile)
		}

		// No index.html available, return 404
		return c.NoContent(http.StatusNotFound)
	})

	return nil
}
