package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type GCS struct {
	ProjectId           string `mapstructure:"project_id"`
	BucketName          string `mapstructure:"bucket_name"`
	Secret              string `mapstructure:"image_secret"`
	ServiceAccountKey   string `mapstructure:"service_account_key"`
	ServiceAccountEmail string `mapstructure:"service_account_email"`
}

type App struct {
	Port  int  `mapstructure:"port"`
	Debug bool `mapstructure:"debug"`
}

type Config struct {
	GCS GCS `mapstructure:"gcs"`
	App App `mapstructure:"app"`
}

func LoadConfig() (config *Config, err error) {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while reading the config")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while unmarshal the config")
	}

	return
}
