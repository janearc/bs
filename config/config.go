package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Version struct {
		BuildDate string `yaml:"build_date"`
		Build     string `yaml:"build"`
		Branch    string `yaml:"branch"`
	}
}

func LoadConfig(configFileName string, versionFileName string, secretsFileName string) (*Config, error) {
	if configFileName == "" {
		configFileName = "/app/config/bs/config.yml"
	}

	file, err := os.Open(configFileName)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to open config file %s", configFileName)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logrus.WithError(err).Fatalf("Failed to close config file %s", configFileName)
		}
		logrus.Infof("Parsed config file %s", configFileName)
	}(file)

	var config Config

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		logrus.WithError(err).Fatalf("Failed to decode config file %s", configFileName)
		return nil, err
	}

	// same as above, but for the version file
	if versionFileName == "" {
		versionFileName = "/app/config/version.yml"
	}
	vf, err := os.Open(versionFileName)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to open version file %s", versionFileName)
		return nil, err
	}
	defer func(vf *os.File) {
		err := vf.Close()
		if err != nil {
			logrus.WithError(err).Fatalf("Failed to close version file %s", versionFileName)
		}
		logrus.Infof("Parsed version file %s", versionFileName)
	}(vf)

	// Decode the version file
	vfDecoder := yaml.NewDecoder(vf)
	if err := vfDecoder.Decode(&config); err != nil {
		logrus.WithError(err).Fatalf("Failed to decode version file %s", versionFileName)
		return nil, err
	}

	if secretsFileName == "" {
		secretsFileName = "/app/config/bs/secrets.yml"
	}

	sf, err := os.Open(secretsFileName)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to open secrets file %s", secretsFileName)
		return nil, err
	}
	defer func(sf *os.File) {
		err := sf.Close()
		if err != nil {
			logrus.WithError(err).Fatalf("Failed to close secrets file %s", secretsFileName)
		}
		logrus.Infof("Parsed secrets file %s", secretsFileName)
	}(sf)

	secDecoder := yaml.NewDecoder(sf)
	if err := secDecoder.Decode(&config); err != nil {
		logrus.WithError(err).Fatalf("Failed to decode secrets file %s", secretsFileName)
		return nil, err
	}

	logrus.Info("returning from loadconfig")
	return &config, nil
}
