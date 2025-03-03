package ymlcfg

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func parseConfig[T any](reader io.Reader) (*T, error) {
	decoder := yaml.NewDecoder(reader)

	cfg := new(T)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}
	return cfg, nil
}

func ParseFile[T any](file string) (*T, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	return parseConfig[T](f)
}
