package config

import (
	"github.com/go-bumbu/config"
)

type AppCfg struct {
	Env  Env
	Msgs []Msg
}

type Env struct {
	LogLevel   string
	Production bool
}

// Default represents the basic set of sensible defaults
var defaultCfg = AppCfg{
	Env: Env{
		LogLevel:   "info",
		Production: true,
	},
}

type Msg struct {
	Level string
	Msg   string
}

func Get(file string) (AppCfg, error) {
	configMsg := []Msg{}
	cfg := AppCfg{}
	var err error
	_, err = config.Load(
		config.Defaults{Item: defaultCfg},
		config.CfgFile{Path: file, Mandatory: false},
		config.EnvVar{Prefix: "BUMBU"},
		config.Unmarshal{Item: &cfg},
		config.Writer{Fn: func(level, msg string) {
			if level == config.InfoLevel {
				configMsg = append(configMsg, Msg{Level: "info", Msg: msg})
			}
			if level == config.DebugLevel {
				configMsg = append(configMsg, Msg{Level: "debug", Msg: msg})
			}
		}},
	)
	cfg.Msgs = configMsg
	return cfg, err
}
