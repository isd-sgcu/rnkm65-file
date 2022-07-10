package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
)

type GCS struct {
	ProjectId           string `mapstructure:"project_id"`
	BucketName          string `mapstructure:"bucket_name"`
	Secret              string `mapstructure:"image_secret"`
	ServiceAccountEmail string `mapstructure:"service_account_email"`
	ServiceAccountKey   []byte
	ServiceAccountJSON  []byte
}

type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSL      string `mapstructure:"ssl"`
}

type App struct {
	Port     int  `mapstructure:"port"`
	Debug    bool `mapstructure:"debug"`
	CacheTTL int  `mapstructure:"cache_ttl"`
}

type Config struct {
	GCS      GCS      `mapstructure:"gcs"`
	App      App      `mapstructure:"app"`
	Database Database `mapstructure:"database"`
	Redis    Redis    `mapstructure:"redis"`
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

	config.GCS.ServiceAccountJSON, err = loadFile("./config/gcs-service-account.json")
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while unmarshal the config")
	}

	config.GCS.ServiceAccountKey, err = loadFile("./config/gcs-private-key.pem")
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while unmarshal the config")
	}

	return
}

func loadFile(path string) ([]byte, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}
