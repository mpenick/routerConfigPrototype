package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

type routerConfig struct {
	ConnCount  int           `yaml:"connCount"`
	ETCConfigs []*ETCDConfig `yaml:"etcdConfigs"`
	Config     *os.File      `yaml:"-"`
}

type RouterConfig struct {
	configFileName string
	wrapped        atomic.Pointer[routerConfig]
}

func NewRouterConfig(args []string) (*RouterConfig, error) {
	var cfg routerConfig
	parser, err := kong.New(&cfg)

	if _, err = parser.Parse(args); err != nil {
		return nil, err
	}

	var configFileName string
	if cfg.Config != nil {
		defer cfg.Config.Close()

		configFileName = cfg.Config.Name()

		bytes, err := io.ReadAll(cfg.Config)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(bytes, &cfg)
		if err != nil {
			return nil, err
		}
	}

	r := &RouterConfig{configFileName: configFileName}
	r.wrapped.Store(&cfg)
	return r, nil
}

func (r *RouterConfig) Reload() error {
	var cfg routerConfig

	f, err := os.Open(r.configFileName)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return err
	}

	r.wrapped.Store(&cfg)

	return nil
}

var (
	ErrNoConfigFile = errors.New("no config file specified")
)

func (r *RouterConfig) Watch(ctx context.Context) error {
	if r.configFileName == "" {
		return ErrNoConfigFile
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	dir, _ := filepath.Split(r.configFileName)
	err = w.Add(dir)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			break
		case event, ok := <-w.Events:
			if !ok {
				break
			}

			if event.Name == r.configFileName {
				err = r.Reload()
				// Do something with failure
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func (r *RouterConfig) ConnCount() Value[int] {
	return valueFunc[int](func() int {
		return r.wrapped.Load().ConnCount
	})
}

func (r *RouterConfig) ETCDConfigs() Value[[]*ETCDConfig] {
	return valueFunc[[]*ETCDConfig](func() []*ETCDConfig {
		return r.wrapped.Load().ETCConfigs
	})
}
