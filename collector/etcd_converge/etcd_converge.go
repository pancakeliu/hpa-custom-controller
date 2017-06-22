package etcd_converge

import (
	"encoding/json"
	//"flag"
	"fmt"
	"github.com/lpc-win32/hpa-custom-controller/controller/auto_scaler"
	//	controller_options "hpa-custom-controller/controller/options"
	"github.com/lpc-win32/hpa-custom-controller/collector/etcd_client"
	"github.com/lpc-win32/hpa-custom-controller/k8sv1beta1"
	"github.com/lpc-win32/hpa-custom-controller/global_options"

	"k8s.io/client-go/pkg/util/clock"

	//	"math"
	//	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
)

type Autoscaler struct {
	EtcdClient *etcd_client.EtcdClient
}

var ArrSpec []*k8sv1beta1.K8sScaleSpec

// 防止etcd没有获取到，维护一个map
// var MapSpec map[string]*k8sv1beta1.K8sScaleSpec

func NewAutoScaler(config *global_options.AutoScalerConfig) *Autoscaler {
	//	str1 := make([]string, 3)
	//	str1[0] = "http://127.0.0.1:2379"
	etcd, _ := etcd_client.NewClient(config.Etcdstring)
	c := etcd_client.EtcdClient{}
	c.Client = etcd
	autoscaler := &Autoscaler{EtcdClient: &c}
	ArrSpec = []*k8sv1beta1.K8sScaleSpec{}
	// MapSpec = make(map[string]*k8sv1beta1.K8sScaleSpec)

	return autoscaler
}

func (h *Autoscaler) Run(autoscaler *auto_scaler.AutoScaler, stop <-chan int) {
	kclock := clock.RealClock{}
	ticker := kclock.Tick(30 * time.Second)
	h.Informer(autoscaler)

	for {
		select {
		case <-stop:
			glog.Info("This pod is not leader now...")
			return
		case <-ticker:
			h.Informer(autoscaler)
		}
	}
}

func (h *Autoscaler) Informer(autoscaler *auto_scaler.AutoScaler) {
	// TODO
	ArrSpec = nil
	Spec := k8sv1beta1.K8sScaleSpec{}
	namespaces, err := h.EtcdClient.EtcdList("/api/v1/autoscalers")

	if err == nil {
		for _, value := range namespaces {
			if len(value) != 0 {
				namespace := strings.Split(value, "/")
				str := "/api/v1/autoscalers/" + namespace[4]
				names, _ := h.EtcdClient.EtcdList(string(str))
				for _, value1 := range names {
					if len(value1) != 0 {
						name := strings.Split(value1, "/")
						str1 := str + "/" + name[5]
						key, etcd_err := h.EtcdClient.EtcdGet(string(str1))
						if etcd_err != nil {
							// 当获取失败时取上次结果
							// spec_ptr, ok := MapSpec[value+value1]
							// if !ok {
							//	break
							// }
							// Spec = *spec_ptr
                            key, etcd_err = h.EtcdClient.EtcdGet(string(str1))
                            if etcd_err != nil {
                                break
                            }
						}
						err := json.Unmarshal([]byte(key), &Spec)
						if err != nil {
							fmt.Println(err)
						}
						// 函数调用
						// fmt.Println(Spec)
						ArrSpec = append(ArrSpec, &Spec)

						// 将上次的结果存储至Map集合中
						// MapSpec[value+value1] = &Spec
						//	h.QueueWatch.Push(hr)
					}
				}
			}
		}
	}
	autoscaler.ExecAutoScaler(ArrSpec)
}
