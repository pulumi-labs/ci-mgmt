package pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubAppCredentials_renderCustomExpressions_whenConfigured(t *testing.T) {
	// Given
	configPath := filepath.Join(t.TempDir(), ".ci-mgmt.yaml")
	configContent := `github-app:
  enabled: true
  id: ${{ secrets.RELEASE_APP_ID }}
  private-key: ${{ secrets.RELEASE_APP_PEM }}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	config, err := LoadLocalConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	tmpl, err := parseTemplate(templateFS, "templates/base/.github/workflows/test.yml")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	// When
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateContext{Config: config}); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	content := rendered.String()

	// Then
	for _, want := range []string{
		"app-id: ${{ secrets.RELEASE_APP_ID }}",
		"private-key: ${{ secrets.RELEASE_APP_PEM }}",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected rendered workflow to contain %q, got:\n%s", want, content)
		}
	}
	for _, unwanted := range []string{
		"PULUMI_PROVIDER_AUTOMATION_APP_ID",
		"PULUMI_PROVIDER_AUTOMATION_PRIVATE_KEY",
	} {
		if strings.Contains(content, unwanted) {
			t.Fatalf("expected rendered workflow not to contain %q, got:\n%s", unwanted, content)
		}
	}
}

func TestDefaultGitHubAppCredentials_matchExistingWorkflowSecretNames(t *testing.T) {
	// Given
	config, err := loadDefaultConfig()
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}
	tmpl, err := parseTemplate(templateFS, "templates/base/.github/workflows/test.yml")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	// When
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateContext{Config: config}); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	content := rendered.String()

	// Then
	for _, want := range []string{
		"app-id: ${{ secrets.PULUMI_PROVIDER_AUTOMATION_APP_ID }}",
		"private-key: ${{ secrets.PULUMI_PROVIDER_AUTOMATION_PRIVATE_KEY }}",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected rendered workflow to contain %q, got:\n%s", want, content)
		}
	}
}
