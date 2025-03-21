package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/thomascriley/aptutil/log"
	"github.com/thomascriley/aptutil/mirror"
)

const (
	defaultConfigPath = "/etc/apt/mirror.toml"
)

var (
	configPath = flag.String("f", defaultConfigPath, "configuration file name")
)

func main() {
	flag.Parse()

	config := mirror.NewConfig()
	md, err := toml.DecodeFile(*configPath, config)
	if err != nil {
		log.ErrorExit(err)
	}
	if len(md.Undecoded()) > 0 {
		log.Error("invalid config keys", map[string]interface{}{
			"keys": fmt.Sprintf("%#v", md.Undecoded()),
		})
		os.Exit(1)
	}

	err = config.Log.Apply()
	if err != nil {
		log.ErrorExit(err)
	}

	err = mirror.Run(config, flag.Args())
	if err != nil {
		log.ErrorExit(err)
	}
}
