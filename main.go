package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"routerConfigPrototype/config"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.NewRouterConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := cfg.Watch(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// Pass in the self-updating config values
	go printConfigValues(ctx, config.StaticValue[int]{Value: 12}, cfg.ETCDConfigs())

	go testUsingSignalToReload(cfg)

	for {
		time.Sleep(99 * time.Second)
	}
}

func printConfigValues(ctx context.Context, connCount config.Value[int], etcdCfgs config.Value[[]*config.ETCDConfig]) {
	t := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-t.C:
			for _, e := range etcdCfgs.Get() {
				fmt.Printf("ETCDConfig: %v\n", e)
			}

			fmt.Printf("ConnCount: %v\n", connCount.Get())
		case <-ctx.Done():
			return
		}
	}
}

func testUsingSignalToReload(cfg *config.RouterConfig) {
	// Also listen for SIGHUP to reload the config
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)

	for {
		<-ch
		fmt.Println("Received SIGHUP")
		err := cfg.Reload()
		if err != nil {
			fmt.Println(err)
		}
	}
}
