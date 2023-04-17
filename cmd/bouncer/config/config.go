package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	failConfig int = 13
)

// Conf - struct for config file.
type Conf struct {
	Redis RedisConfig `yaml:"redis,omitempty"`
}

// RedisConfig - struct for redis config.
type RedisConfig struct {
	Addr     string `yaml:"address,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Parse - creates new config struct from file.
func (c *Conf) Parse(filename string) error {
	// Read config file
	config, err := c.readFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal config file to struct
	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	return nil
}

// readFile reads config file and returns its content
func (c *Conf) readFile(fileName string) ([]byte, error) {
	if len(fileName) == 0 {
		return nil, fmt.Errorf("no config file specified")
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("config file not found")
	}

	return data, nil
}
