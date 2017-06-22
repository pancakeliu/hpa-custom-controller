package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lpc-win32/hpa-custom-controller/controller/auto_scaler"
	"github.com/lpc-win32/hpa-custom-controller/controller/leader_election/leader"
	controller_options "github.com/lpc-win32/hpa-custom-controller/controller/options"
	"github.com/lpc-win32/hpa-custom-controller/global_options"
	"github.com/lpc-win32/hpa-custom-controller/collector/etcd_converge"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

func main() {
	hpaconfig := flag.String("hpa-config", "./config", "input hpaconfig path.")
	c, _ := global_options.InitAutoScaleConfig(*hpaconfig)
	autoscaler := etcd_converge.NewAutoScaler(c)
	config := &controller_options.Options{
		Leader_election_name:       c.Leader_election_name,
		Leader_election_id:         "",
		Leader_election_namespace:  c.Leader_election_namespace,
		Leader_election_ttl:        c.Leader_election_ttl,
		Leader_election_in_cluster: c.Leader_election_in_cluster,
		Promethues_ip_port:         c.Prometheus_ip_port,
		K8sApi_config_path:         c.K8sapi_config_path,
	}

	// 单独从命令行中提取 leader-election-id
	config.AddFlags(pflag.CommandLine)
	pflag.Parse()

	// 验证参数的合法性
	if err := config.ValidateFlags(); err != nil {
		glog.Errorf("%v\n", err)
		os.Exit(1)
	}

	as_obj, err := auto_scaler.NewAutoScaler(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	leaderClient, err := leader.NewLeader(config)
	if err != nil {
		glog.Errorf("%v", err)
		os.Exit(1)
	}

	// work_func & work_func_stop_chan用于leader选举，只有leader才能执行work_func
	// work_func_stop_chan用于leader接收退出信号
	// TODO
	work_func := func(stop <-chan int) {
		// 传递一个stop channel，用于控制工作循环的退出
		autoscaler.Run(as_obj, stop)
	}
	work_func_stop_chan := make(chan int)

	leaderClient.LeaderRun(work_func, work_func_stop_chan)
}
