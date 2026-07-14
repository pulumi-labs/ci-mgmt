package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePackage_publishUsesConfiguredEnvironmentAliases(t *testing.T) {
	// Given
	configPath := filepath.Join("..", "..", "test-providers", "eks", ".ci-mgmt.yaml")
	config, err := LoadLocalConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	config.Env["NPM_TOKEN"] = "${{ secrets.CUSTOM_NPM_TOKEN }}"
	config.Env["NUGET_PUBLISH_KEY"] = "${{ secrets.CUSTOM_NUGET_KEY }}"
	config.Env["PYPI_API_TOKEN"] = "${{ secrets.CUSTOM_PYPI_TOKEN }}"

	outDir := t.TempDir()
	if err := GeneratePackage(GenerateOpts{
		RepositoryName: "example/pulumi-test",
		OutDir:         outDir,
		TemplateName:   "external-bridged-provider",
		Config:         config,
		SkipMigrations: true,
	}); err != nil {
		t.Fatalf("generate package: %v", err)
	}

	publishWorkflow, err := os.ReadFile(filepath.Join(outDir, ".github", "workflows", "publish.yml"))
	if err != nil {
		t.Fatalf("read publish workflow: %v", err)
	}

	// When
	generated := string(publishWorkflow)

	// Then
	for _, expected := range []string{
		"PYPI_PASSWORD: ${{ env.PYPI_API_TOKEN }}",
		"NODE_AUTH_TOKEN: ${{ env.NPM_TOKEN }}",
		"NUGET_PUBLISH_KEY: ${{ env.NUGET_PUBLISH_KEY }}",
	} {
		if !strings.Contains(generated, expected) {
			t.Errorf("generated publish workflow does not use configured environment alias %q", expected)
		}
	}
}
