package config

import (
	"fmt"

	"github.com/toolkits/pkg/file"
)

type Config struct {
	Logger   loggerSection   `yaml:"logger"`
	Voice    voiceSection    `yaml:"voice"`
	Consumer consumerSection `yaml:"consumer"`
	Redis    redisSection    `yaml:"redis"`
}

type loggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

type redisSection struct {
	Addr    string         `yaml:"addr"`
	Pass    string         `yaml:"pass"`
	DB      int            `yaml:"db"`
	Idle    int            `yaml:"idle"`
	Timeout timeoutSection `yaml:"timeout"`
}

type timeoutSection struct {
	Conn  int `yaml:"conn"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}

type consumerSection struct {
	Queue  string `yaml:"queue"`
	Worker int    `yaml:"worker"`
}

type voiceSection struct {
    Message   	     string   `yaml:"message"`
	Mobiles          []string `yaml:"mobiles"`
	TtsCode          string   `yaml:"ttscode"`
	CalledShowNumber string   `yaml:"calledshownumber"`
	TtsParam         ttsParam `yaml:"ttsparam"`

}

type ttsParam struct {
	Sname string `yaml:"sname"`

}
var yaml Config

func Get() Config {
	return yaml
}

func ParseConfig(yf string) error {
	err := file.ReadYaml(yf, &yaml)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", yf, err)
	}
	return nil
}
