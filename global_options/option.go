package global_options

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type AutoScalerConfig struct {
	Etcdstring                 []string
	Leader_election_name       string
	Leader_election_namespace  string
	Leader_election_ttl        time.Duration
	Leader_election_in_cluster bool
	Prometheus_ip_port         string
	K8sapi_config_path         string
}

/*var data = `
etcdstring: ['127.0.0.1:2379','127.0.0.1:4001']
`*/
func InitAutoScaleConfig(hpaconfig string) (*AutoScalerConfig, error) {
	data, err := ioutil.ReadFile(hpaconfig)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println(string(data))
	config := &AutoScalerConfig{}
	err = yaml.Unmarshal([]byte(data), config)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println(config)
	return config, nil
}
