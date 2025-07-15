package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"check-backup-dicom-exams/model"
)

func LoadConfig(path string) (model.Config, error) {
	var cfg model.Config
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, err
}
