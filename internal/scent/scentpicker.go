package scent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const filename = "scent.yml"

type ValidatorConfig struct {
	Name        string   `yaml:"name"`
	Extensions  []string `yaml:"extensions"`
	Runner      string   `yaml:"runner"`
	TestPattern string   `yaml:"test_pattern"`
}

type RunnerConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Dir     string `yaml:"dir"`
}

type Scent struct {
	WatchPaths []string          `yaml:"watch_paths"`
	Validators []ValidatorConfig `yaml:"validators"`
	Runners    []RunnerConfig    `yaml:"runners"`
}

func (s *Scent) IsTestFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	base := filepath.Base(filename)
	for _, v := range s.Validators {
		for _, e := range v.Extensions {
			if strings.ToLower(e) != ext {
				continue
			}
			if v.TestPattern != "" {
				matched, err := filepath.Match(v.TestPattern, base)
				return err == nil && matched
			}
			return isDefaultTestFile(base, ext)
		}
	}
	return false
}

func isDefaultTestFile(base, ext string) bool {
	switch ext {
	case ".go":
		return strings.HasSuffix(base, "_test.go")
	case ".py":
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	case ".js", ".ts", ".jsx", ".tsx":
		return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.")
	case ".rb":
		return strings.HasSuffix(base, "_spec.rb") || strings.HasSuffix(base, "_test.rb")
	case ".rs":
		return strings.HasSuffix(base, "_test.rs")
	}
	return strings.Contains(base, "test")
}

func Load(dir string) (*Scent, error) {
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var s Scent
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &s, nil
}

func (s *Scent) RunnerForFile(changedFile string) *RunnerConfig {
	ext := strings.ToLower(filepath.Ext(changedFile))
	for _, v := range s.Validators {
		for _, e := range v.Extensions {
			if strings.ToLower(e) == ext {
				return s.findRunner(v.Runner)
			}
		}
	}
	return nil
}

func (s *Scent) findRunner(name string) *RunnerConfig {
	for i, r := range s.Runners {
		if r.Name == name {
			return &s.Runners[i]
		}
	}
	return nil
}

func (r *RunnerConfig) Execute() (string, bool) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", r.Command)
	} else {
		cmd = exec.Command("sh", "-c", r.Command)
	}
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}
