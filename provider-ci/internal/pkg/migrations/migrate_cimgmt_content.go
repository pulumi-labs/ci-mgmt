package migrations

import (
	"fmt"
	"path/filepath"
	"strings"
)

// migrate any ci-mgmt.yml content
type migrateCimgmtContent struct{}

func (migrateCimgmtContent) Name() string {
	return "Migrate entries from .ci-mgmt.yml to the top level mise.toml override file"
}
func (migrateCimgmtContent) ShouldRun(templateName string) bool {
	return true
}

// This currently:
// - migrates the tool overrides from .ci-mgmt.yaml to a root level mise.toml
// - removes the `esc` section
func (migrateCimgmtContent) Migrate(templateName, outDir string) error {
	ciMgmtPath := filepath.Join(outDir, ".ci-mgmt.yaml")
	cimgmt, err := newCimgmtYaml(ciMgmtPath)
	if err != nil {
		return err
	}

	misePath := filepath.Join(outDir, "mise.toml")
	mise, err := newTomlFile(misePath)
	if err != nil {
		return err
	}

	toolVersions := cimgmt.getFieldNode("toolVersions")
	// if we don't override any toolVersions then we don't need to do anything
	if toolVersions != nil {
		if len(mise.content) == 0 {
			mise.content = []byte("# Overwrites mise configuration at .config/mise.toml\n[tools]\n")
		}

		miseTools := []sectionEntry{}

		// convert any toolVersions overrides to mise tool entries
		toolVersionsMap := nodeToMap(toolVersions)
		for tool, version := range toolVersionsMap {
			if tool == "go" {
				// don't use go overrides anymore
				continue
			}
			version = strings.TrimSuffix(version, ".x")
			if tool == "java" {
				version = fmt.Sprintf("corretto-%s", version)
			}
			miseTools = append(miseTools, sectionEntry{
				key:   tool,
				value: version,
			})
		}

		updated, err := mise.ensureSectionEntries("tools", miseTools)
		if err != nil {
			return fmt.Errorf("error writing toolVersions to mise.toml: %w", err)
		}
		if updated {
			err := mise.writeFile()
			if err != nil {
				return err
			}
		}

		// Finally remove the toolVersions from .ci-mgmt.yaml
		cimgmt.deleteKey("toolVersions")
	}

	esc := cimgmt.getFieldNode("esc")
	// if we didn't configure `esc`, we don't need to do anything.
	if esc != nil {
		cimgmt.deleteKey("esc")
	}

	return cimgmt.writeFile()
}
