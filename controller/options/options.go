package options

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	flag "github.com/spf13/pflag"
)

type Options struct {
	Leader_election_name       string
	Leader_election_id         string
	Leader_election_namespace  string
	Leader_election_ttl        time.Duration
	Leader_election_in_cluster bool

	Promethues_ip_port string

	K8sApi_config_path string
}

func NewHpaCustiomControllerOptions() *Options {
	return &Options{
		Leader_election_name:      "",
		Leader_election_id:        "",
		Leader_election_namespace: "",
		Leader_election_ttl:       10 * time.Second,

		Promethues_ip_port: "",
		K8sApi_config_path: "",
	}
}

func (opt *Options) ValidateFlags() error {
	var errorsFound bool

	if opt.Leader_election_id == "" {
		errorsFound = true
		glog.Errorf("--leader-election-id parameter can't be empty")
	}
	if opt.Promethues_ip_port == "" {
		errorsFound = true
		glog.Errorf("promethues_ip_port parameter can't be empty")
	}
	if opt.Leader_election_name == "" {
		errorsFound = true
		glog.Errorf("leader_election_name parameter can't be empty")
	}
	if opt.K8sApi_config_path == "" {
		errorsFound = true
		glog.Errorf("k8sapi_config_path can't be empty")
	}

	// Judge errorsFound's value
	if errorsFound {
		return fmt.Errorf("failed to validate all input parameters")
	}

	return nil
}

func (opt *Options) AddFlags(fs *flag.FlagSet) {
	// leader-election-id用于ledaer的竞选，需动态获取
	fs.StringVar(&opt.Leader_election_id, "leader-election-id", opt.Leader_election_id,
		"set opt.leader_election_id")
}
