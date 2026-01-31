// Package fileprefix provides a golangci-lint module plugin for filename checks.
package fileprefix

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("fileprefix", New)
}

// Settings controls the behavior of the fileprefix linter.
type Settings struct {
	Message     string   `json:"message"`
	ExcludeDirs []string `json:"excludeDirs"`
}

// Plugin implements the golangci-lint module plugin interface.
type Plugin struct {
	settings Settings
}

// New builds a new plugin instance with the provided settings.
func New(settings any) (register.LinterPlugin, error) {
	cfg, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}
	if cfg.Message == "" {
		cfg.Message = "filename contains package name; consider extracting the logic into a separate package"
	}
	return &Plugin{settings: cfg}, nil
}

// BuildAnalyzers returns analyzers for the custom linter.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "fileprefix",
			Doc:  "disallow filenames that start with the package name",
			Run:  p.run,
		},
	}, nil
}

// GetLoadMode reports the analysis load mode.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func (p *Plugin) run(pass *analysis.Pass) (any, error) {
	if len(pass.Files) == 0 {
		return nil, nil
	}

	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".go") {
		return nil, nil
	}

	if strings.HasSuffix(base, "_test.go") {
		return nil, nil
	}

	dir := filepath.Base(filepath.Dir(filename))
	if dir == "cmd" || dir == "main" {
		return nil, nil
	}

	for _, excluded := range p.settings.ExcludeDirs {
		if dir == excluded {
			return nil, nil
		}
	}

	name := strings.TrimSuffix(base, ".go")
	pattern := regexp.MustCompile("^" + regexp.QuoteMeta(dir) + "(_|[a-z])")
	if pattern.MatchString(name) {
		pass.Report(analysis.Diagnostic{
			Pos:     pass.Files[0].Pos(),
			Message: p.settings.Message,
		})
	}

	return nil, nil
}
