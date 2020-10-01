package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

const (
	failConfig int = 13
)

type Conf struct {
	Redis RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Addr     string `yaml:"address"`
	Password string `yaml:"password"`
}

func (c *Conf) GetConf(filename string) *Conf {
	config := readConfigFile(filename)

	err := yaml.Unmarshal(config, c)
	if err != nil {
		exitError("unable to parse config file")
	}

	return c
}

func readConfigFile(filename string) []byte {
	if len(filename) == 0 {
		exitError("no config file specified")
	}

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		exitError("config file not found")
	}

	return fileBytes
}

func exitError(reason string) {
	fmt.Printf("%s", reason)
	os.Exit(failConfig)
}
