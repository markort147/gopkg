package ymlcfg

import (
	"io"
	"os"
	"strings"
	"testing"
)

type mockReader struct {
	data string
	read bool
}

type mockCfg struct {
	Server struct {
		Host string `yaml:"Host"`
		Port int    `yaml:"Port"`
	} `yaml:"Server"`
	Log struct {
		Level  string `yaml:"Level"`
		Output string `yaml:"Output"`
	} `yaml:"Log"`
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.read {
		return 0, io.EOF
	}
	n = copy(p, m.data)
	m.read = true
	return n, nil
}

func TestLoadConfig(t *testing.T) {

	t.Run("testing private func loadConfig", func(t *testing.T) {
		reader := &mockReader{
			data: `Server:
  Host: localhost
  Port: 8080

Log:
  Level: debug
  Output: stdout`,
		}

		want := []struct {
			value     any
			retriever func(*mockCfg) any
		}{
			{"localhost", func(c *mockCfg) any { return c.Server.Host }},
			{8080, func(c *mockCfg) any { return c.Server.Port }},
			{"debug", func(c *mockCfg) any { return c.Log.Level }},
			{"stdout", func(c *mockCfg) any { return c.Log.Output }},
		}

		cfg, err := parseConfig[mockCfg](reader)
		if err != nil {
			t.Fatal(err)
		}

		for _, w := range want {
			if got := w.retriever(cfg); got != w.value {
				t.Errorf("Expected %v, got %v", w.value, got)
			}
		}
	})

	t.Run("testing private func loadConfig with invalid yaml", func(t *testing.T) {
		reader := &mockReader{
			data: `	Server:
  Host: localhost
  Port: 8080`,
		}

		if _, err := parseConfig[mockCfg](reader); !strings.Contains(err.Error(), "error loading config") {
			t.Errorf("Expected error containing %q, got '%v'", "error loading config", err)
		}
	})
}

func TestFromFile(t *testing.T) {

	t.Run("testing loading config from file", func(t *testing.T) {
		fileContent := `Server:
  Host: localhost
  Port: 8080

Log:
  Level: debug
  Output: stdout`

		file := os.TempDir() + "/config_test.yaml"
		defer os.Remove(file)
		if err := os.WriteFile(file, []byte(fileContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := ParseFile[mockCfg](file)
		if err != nil {
			t.Fatal(err)
		}

		want := []struct {
			value     any
			retriever func(*mockCfg) any
		}{
			{"localhost", func(c *mockCfg) any { return c.Server.Host }},
			{8080, func(c *mockCfg) any { return c.Server.Port }},
			{"debug", func(c *mockCfg) any { return c.Log.Level }},
			{"stdout", func(c *mockCfg) any { return c.Log.Output }},
		}

		for _, w := range want {
			if got := w.retriever(cfg); got != w.value {
				t.Errorf("Expected %v, got %v", w.value, got)
			}
		}
	})

	t.Run("testing loading config from invalid file", func(t *testing.T) {

		file := "infalid_file"

		if _, err := ParseFile[mockCfg](file); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
