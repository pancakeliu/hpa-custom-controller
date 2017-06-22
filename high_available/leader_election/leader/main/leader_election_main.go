package main

import (
	"fmt"
	"github.com/lpc-win32/hpa-custom-controller/high_available/leader_election/leader"
	"github.com/lpc-win32/hpa-custom-controller/controller/options"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

func main() {
	config := options.NewHpaCustiomControllerOptions()
	config.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if err := config.ValidateFlags(); err != nil {
		glog.Errorf("%v\n", err)
		os.Exit(1)
	}

	leaderClient, err := leader.NewLeader(config)
	if err != nil {
		glog.Errorf("%v", err)
		os.Exit(1)
	}

	work_func := func(stop <-chan int) {
		for {
			select {
			case <-stop:
				fmt.Print("run over\n")
				return
			default:
			}
			fmt.Print("Run now...\n")
			time.Sleep(time.Second)
		}
	}

	work_func_chan := make(chan int)

	leaderClient.LeaderRun(work_func, work_func_chan)
}
