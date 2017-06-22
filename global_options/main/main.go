package main

import (
	"flag"
	"fmt"
	"github.com/lpc-win32/hpa-custom-controller/global_options"
)

func main() {
	hpaconfig := flag.String("hpaconfig", "./config", "Input your hpaconfigure.")
	config, _ := global_options.InitAutoScaleConfig(*hpaconfig)
	fmt.Println((*config).Etcdstring)
}
