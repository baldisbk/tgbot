package config

import (
	"flag"
	"io/ioutil"

	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"github.com/baldisbk/tgbot_sample/pkg/poller"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

var defaultPath = "/etc/tgbot/config.yaml"
var develPath = "config.yaml"

var configPath = flag.String("config", "", "path to config")
var develMode = flag.Bool("devel", false, "development mode")

type Config struct {
	Path  string `yaml:"-"`
	Devel bool   `yaml:"-"`

	CacheConfig   usercache.Config `yaml:"user_cache"`
	FactoryConfig impl.Config      `yaml:"user_factory"`
	PollerConfig  poller.Config    `yaml:"poller"`
	TimerConfig   timer.Config     `yaml:"timer"`
	ApiConfig     tgapi.Config     `yaml:"tgapi"`
}

func ParseConfig() (*Config, error) {
	flag.Parse()

	config := Config{
		Devel: *develMode,
		Path:  *configPath,
	}
	if config.Path == "" {
		if config.Devel {
			config.Path = develPath
		} else {
			config.Path = defaultPath
		}
	}
	contents, err := ioutil.ReadFile(config.Path)
	if err != nil {
		return nil, xerrors.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, xerrors.Errorf("parse config: %w", err)
	}

	return &config, nil
}
