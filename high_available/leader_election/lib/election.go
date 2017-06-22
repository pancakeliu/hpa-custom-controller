package lib

import (
	"encoding/json"
	"os"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/leaderelection"
	"k8s.io/kubernetes/pkg/client/leaderelection/resourcelock"
	"k8s.io/kubernetes/pkg/client/record"
	// client "k8s.io/kubernetes/pkg/client/unversioned"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	startBackoff = time.Second
	maxBackoff   = time.Minute
)

func getCurrentLeader(electionId, namespace string, c *clientset.Clientset) (string, *api.Endpoints, error) {
	endpoints, err := c.Endpoints(namespace).Get(electionId)
	if err != nil {
		return "", nil, err
	}
	val, found := endpoints.Annotations[resourcelock.LeaderElectionRecordAnnotationKey]
	if !found {
		// 没有leader的时候，返回当前的endpoints竞选leader
		return "", endpoints, nil
	}
	electionRecord := resourcelock.LeaderElectionRecord{}
	if err := json.Unmarshal([]byte(val), &electionRecord); err != nil {
		return "", nil, err
	}
	// 返回当前的leader
	return electionRecord.HolderIdentity, endpoints, err
}

// NewElection creates an election.  'namespace'/'election' should be an existing Kubernetes Service
// 'id' is the id if this leader, should be unique.
func NewElection(electionId, id, namespace string, ttl time.Duration, callback func(leader string), c *clientset.Clientset, custom_func func(<-chan int), custom_func_chan chan int) (*leaderelection.LeaderElector, error) {
	_, err := c.Endpoints(namespace).Get(electionId)
	if err != nil {
		if errors.IsNotFound(err) {
			// 不存在electionID则创建新的
			_, err = c.Endpoints(namespace).Create(&api.Endpoints{
				ObjectMeta: api.ObjectMeta{
					Name: electionId,
				},
			})
			if err != nil && !errors.IsConflict(err) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 创建一个广播，用来宣告自己的leader状态
	broadcaster := record.NewBroadcaster()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	recorder := broadcaster.NewRecorder(api.EventSource{
		Component: "leader-elector",
		Host:      hostname,
	})

	leader, endpoints, err := getCurrentLeader(electionId, namespace, c)
	if err != nil {
		return nil, err
	}
	callback(leader)

	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(stop <-chan struct{}) {
			callback(id)
			go custom_func(custom_func_chan)
		},
		OnStoppedLeading: func() {
			leader, _, err := getCurrentLeader(electionId, namespace, c)
			if err != nil {
				glog.Errorf("failed to get leader: %v", err)
				// empty string means leader is unknown
				callback("")
				return
			}
			callback(leader)
			// 通知当前进程已不再是leader节点，工作进行停止执行
			custom_func_chan <- 1
		},
	}

	rl := resourcelock.EndpointsLock{
		EndpointsMeta: endpoints.ObjectMeta,
		Client:        c,
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	}

	config := leaderelection.LeaderElectionConfig{
		Lock:          &rl,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callbacks,
	}

	return leaderelection.NewLeaderElector(config)
}

// RunElection runs an election given an leader elector.  Doesn't return.
func RunElection(e *leaderelection.LeaderElector) {
	wait.Forever(e.Run, 0)
}
