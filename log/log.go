package log

import (
	"fmt"
	"io"
	"os"
	"reflect"

	glog "github.com/labstack/gommon/log"
)

/*
=== GLOBAL LOGGER CONFIGURATION ===
This file is used to configure the global logger for the application.
The global logger is used to log messages that are not specific to a particular package.
==================================
*/

type Config struct {
	Level  glog.Lvl
	Output io.Writer
}

var Logger = glog.New("global")

func Init(cfg *Config) error {
	if err := fixConfig(cfg); err != nil {
		return fmt.Errorf("log configuration error: %w", err)
	}

	Logger.SetLevel(cfg.Level)
	Logger.SetOutput(cfg.Output)
	return nil
}

func fixConfig(cfg *Config) error {
	if cfg.Level == 0 {
		cfg.Level = glog.INFO
	}
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return nil
}

func Close() {
	logOut := Logger.Output()
	if reflect.TypeOf(logOut) == reflect.TypeOf((*os.File)(nil)) {
		logOut.(*os.File).Close()
	}
}

func Test() {
	Logger.SetLevel(glog.DEBUG)
	Logger.SetOutput(os.Stdout)
}
