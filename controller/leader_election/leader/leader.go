/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package leader

import (
	"fmt"
	"time"

	election "github.com/lpc-win32/hpa-custom-controller/controller/leader_election/lib"
	"github.com/lpc-win32/hpa-custom-controller/controller/options"

	"github.com/golang/glog"
	flag "github.com/spf13/pflag"
	// "k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	kubectl_util "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

/*var (
	flags = flag.NewFlagSet(
		`elector --election=<name>`,
		flag.ExitOnError)
	name      = flags.String("election", "", "The name of the election")
	id        = flags.String("id", "", "The id of this participant")
	namespace = flags.String("election-namespace", api.NamespaceDefault, "The Kubernetes namespace for this election")
	ttl       = flags.Duration("ttl", 10*time.Second, "The TTL for this election")
	inCluster = flags.Bool("use-cluster-credentials", false, "Should this request use cluster credentials?")
	addr      = flags.String("http", "", "If non-empty, stand up a simple webserver that reports the leader state")

	leader = &LeaderData{}
)*/

type Leader struct {
	flags     *flag.FlagSet
	name      string
	id        string
	namespace string
	ttl       time.Duration
	inCluster bool
}

func NewLeader(opt *options.Options) (*Leader, error) {
	return &Leader{
		flags:     flag.NewFlagSet(`elector --election=<name>`, flag.ExitOnError),
		name:      opt.Leader_election_name,
		id:        opt.Leader_election_id,
		namespace: opt.Leader_election_namespace,
		ttl:       opt.Leader_election_ttl,
		inCluster: false,
	}, nil
}

func (le *Leader) makeClient() (*clientset.Clientset, error) {
	var cfg *restclient.Config
	var err error

	if le.inCluster {
		if cfg, err = restclient.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		clientConfig := kubectl_util.DefaultClientConfig(le.flags)
		if cfg, err = clientConfig.ClientConfig(); err != nil {
			return nil, err
		}
	}
	// 创建一个新的client
	// return client.New(cfg)
	return clientset.NewForConfig(cfg)
}

// LeaderData represents information about the current leader
type LeaderData struct {
	Name string `json:"name"`
}

func (le *Leader) LeaderRun(custom_func_ptr func(<-chan int), custom_func_chan chan int) {
	kubeClient, err := le.makeClient()
	if err != nil {
		glog.Fatalf("error connecting to the client: %v", err)
	}

	// callback(leader_id)
	fn := func(str string) {
		// 记录leader
		le.name = str
		fmt.Printf("%s is the leader\n", le.name)
	}

	// e为leaderelection对象
	e, err := election.NewElection(le.name, le.id, le.namespace, le.ttl, fn, kubeClient, custom_func_ptr, custom_func_chan)
	if err != nil {
		glog.Fatalf("failed to create election: %v", err)
	}
	go election.RunElection(e)

	select {}
}
