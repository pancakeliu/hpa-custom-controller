package auto_scaler

import (
	"fmt"
	"sync"
	"time"

	"github.com/lpc-win32/hpa-custom-controller/controller/options"
	k8s_type "github.com/lpc-win32/hpa-custom-controller/k8sv1beta1"

	"github.com/golang/glog"
)

type AutoScaler struct {
	prom_client             *PromethuesClient
	k8s_client              *K8sapiClient
	metric_reduce_timestamp map[string]int64
}

func NewAutoScaler(opt *options.Options) (*AutoScaler, error) {
	promethues_client, err := NewPromethuesClient(opt)
	if err != nil {
		glog.Error("PromethuesClient create error")
		return nil, err
	}

	k8s_client := NewK8sapiClient(opt.K8sApi_config_path)

	return &AutoScaler{
		prom_client:             promethues_client,
		k8s_client:              k8s_client,
		metric_reduce_timestamp: make(map[string]int64),
	}, nil
}

func (ac_obj *AutoScaler) ExecAutoScaler(k8sScaleSpecs []*k8s_type.K8sScaleSpec) int {
	var wg sync.WaitGroup

	if k8sScaleSpecs == nil {
		glog.Error("k8sSaleSpec array is nil")
		return -1
	}
	// Read all K8sScaleSpec
	for _, scale := range k8sScaleSpecs {
		wg.Add(1)

		go func(scale *k8s_type.K8sScaleSpec) {
			defer wg.Done()

			current_replicas, err := ac_obj.k8s_client.K8sGetReplicas(scale.Namespace, scale.Name)
			if err != nil {
				glog.Errorf("K8sGetReplicas failure.. err: %v", err)
				return
			}
			if current_replicas < scale.MinReplicas {
				ret_err := ac_obj.k8s_client.K8sScaler(scale.Namespace, scale.Name, scale.MinReplicas)
				if ret_err != nil {
					glog.Error("K8sScaler reduce replicas failure...")
					return
				}
				fmt.Printf("namespace: %s, name: %s replicas is %d, less than MinReplicas, scaler : %d\n",
					scale.Namespace, scale.Name, current_replicas, scale.MinReplicas)
			}
			if current_replicas > scale.MaxReplicas {
				ret_err := ac_obj.k8s_client.K8sScaler(scale.Namespace, scale.Name, scale.MaxReplicas)
				if ret_err != nil {
					glog.Error("K8sScaler reduce replicas failure...")
					return
				}
				fmt.Printf("namespace: %s, name: %s replicas is %d, greater than MaxReplicas, scaler : %d\n",
					scale.Namespace, scale.Name, current_replicas, scale.MinReplicas)
			}

			var max_add_replicas int32 // 最大增加的扩容副本数
			var max_del_replicas int32 // 最大缩容后的副本数
			var sum_metric_vals float64

			for scale_metric_key, scale_metric_val := range scale.ScalerMetrics {
				fmt.Printf("scale_metric_key is %s\n", scale_metric_key)
				values := ac_obj.prom_client.GetPromethues(scale_metric_key)

				// 计算出需要增加的副本数
				var add_replicas int32
				for _, metric_value := range values {
					// fmt.Println("%f\n", metric_value.Float())
					if metric_value.Float() > scale_metric_val.MaxScaler {
						add_replicas++
					}
					// sum_metric_vals用于计算缩容后的期望副本数，在不扩容的时候该值有效
					sum_metric_vals += metric_value.Float()
				}
				fmt.Printf("sum_metric_vals = %f\n", sum_metric_vals)
				if max_add_replicas < add_replicas {
					max_add_replicas = add_replicas
				}
				// 缩容期望副本数计算
                if scale_metric_val.MinScaler == 0 {
				    expect_del_replicas := int32(sum_metric_vals/scale_metric_val.MinScaler) + 1
				    if max_del_replicas < expect_del_replicas {
					    max_del_replicas = expect_del_replicas
				    }
                }
			}
			// 不需要扩容则考虑缩容的问题
			if max_add_replicas == 0 {
				// 判断时间间隔有无超过规定时间
				canScaler := false

                // 当计算的max_del_replicas为0不用考虑缩容问题，直接退出当前协程
                if max_del_replicas == 0 {
                    return
                }

				if timestamp, ok := ac_obj.metric_reduce_timestamp[scale.Namespace+scale.Name]; ok {
					current_timestamp := time.Now().Unix()
					if current_timestamp-timestamp > 5*60 {
						ac_obj.metric_reduce_timestamp[scale.Namespace+scale.Name] = time.Now().Unix()
						canScaler = true
					}
				} else {
					// ac_obj.metric_reduce_timestamp[scale.Namespace+scale.Name] = time.Now().Unix()
					canScaler = true
				}

				if canScaler {
					// 缩容前副本数值得检查
					if max_del_replicas > scale.MaxReplicas {
						max_del_replicas = scale.MaxReplicas
					}
					if max_del_replicas < scale.MinReplicas {
						max_del_replicas = scale.MinReplicas
					}

					fmt.Printf("namespace : %s, name : %s\n", scale.Namespace, scale.Name)
					current_replicas, err := ac_obj.k8s_client.K8sGetReplicas(scale.Namespace, scale.Name)
					if err != nil {
						glog.Errorf("K8sGetReplicas failure.. err: %v", err)
						return
					}
					fmt.Printf("Current_replicas is %d\n", current_replicas)

					// 计算出的副本数小于当前副本数方可执行缩容
					if current_replicas > max_del_replicas {
						// 通知k8s-api-server执行缩容
						fmt.Printf("开始缩容, %s %s 期望副本数为：%d\n", scale.Namespace, scale.Name, max_del_replicas)
						ret_err := ac_obj.k8s_client.K8sScaler(scale.Namespace, scale.Name, max_del_replicas)
						if ret_err != nil {
							glog.Error("K8sScaler reduce replicas failure...")
							return
						}
						// 成功缩容后记录当前时间
						ac_obj.metric_reduce_timestamp[scale.Namespace+scale.Name] = time.Now().Unix()
					}
				}
			} else {
				// 需要扩若：提取当前relicas
				current_replicas, err := ac_obj.k8s_client.K8sGetReplicas(scale.Namespace, scale.Name)
				if err != nil {
					glog.Errorf("K8sGetReplicas failure.. err: %v", err)
					return
				}
				expect_replicas := current_replicas + max_add_replicas

				// 限制副本数不可超出最大副本数
				if expect_replicas > scale.MaxReplicas {
					expect_replicas = scale.MaxReplicas
				}

				fmt.Printf("开始扩容, %s %s 期望副本数为：%d\n", scale.Namespace, scale.Name, expect_replicas)
				// 通知k8s-api-server执行扩容
				ret_err := ac_obj.k8s_client.K8sScaler(scale.Namespace, scale.Name, expect_replicas)
				if ret_err != nil {
					glog.Error("K8sScaler add replicas failure...")
					return
				}
			}
			return
		}(scale)
	}
	wg.Wait()

	return 0
}
