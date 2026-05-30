package drivers

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"lfiber/internal/support/appctx"
	mailviews "lfiber/resources/views/mail"
)

// RenderTemplate renders an email view template using html/template and the embedded FS.
func RenderTemplate(templateName string, data map[string]any) (string, error) {
	if data == nil {
		data = make(map[string]any)
	}

	// 1. Get fallback values from config
	appName := "Go Fiber App"
	appURL := "http://localhost:3300"

	if app := appctx.App(); app != nil {
		if cfg := app.AppConfig(); cfg != nil {
			if cfg.App.Name != "" {
				appName = cfg.App.Name
			}
			if cfg.App.URL != "" {
				appURL = cfg.App.URL
			}
		}
	}

	// 2. Set default fields in data if not already set
	if _, ok := data["AppName"]; !ok {
		data["AppName"] = appName
	}
	if _, ok := data["AppUrl"]; !ok {
		data["AppUrl"] = appURL
	}
	if _, ok := data["Year"]; !ok {
		data["Year"] = time.Now().Year()
	}

	// 3. Create a new template and parse layout, components, and target template
	tmpl := template.New("mail")

	// Read layouts/default.tmpl
	layoutBytes, err := mailviews.FS.ReadFile("layouts/default.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read default layout: %w", err)
	}
	if _, err := tmpl.Parse(string(layoutBytes)); err != nil {
		return "", fmt.Errorf("failed to parse default layout: %w", err)
	}

	// Read and parse components
	components := []string{"header.tmpl", "footer.tmpl", "button.tmpl", "panel.tmpl"}
	for _, comp := range components {
		path := "components/" + comp
		compBytes, err := mailviews.FS.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read component %s: %w", comp, err)
		}
		if _, err := tmpl.Parse(string(compBytes)); err != nil {
			return "", fmt.Errorf("failed to parse component %s: %w", comp, err)
		}
	}

	// Read target template
	targetPath := templateName
	if !strings.HasSuffix(targetPath, ".tmpl") {
		targetPath += ".tmpl"
	}
	targetBytes, err := mailviews.FS.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read target template %s: %w", targetPath, err)
	}
	if _, err := tmpl.Parse(string(targetBytes)); err != nil {
		return "", fmt.Errorf("failed to parse target template %s: %w", targetPath, err)
	}

	// Render layout template (which calls header, footer, content)
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return "", fmt.Errorf("failed to execute mail template %s: %w", templateName, err)
	}

	return buf.String(), nil
}
