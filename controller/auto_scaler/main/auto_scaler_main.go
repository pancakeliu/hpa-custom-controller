package main

import (
	"os"

	"github.com/lpc-win32/hpa-custom-controller/controller/auto_scaler"
	"github.com/lpc-win32/hpa-custom-controller/controller/options"
	k8s_type "github.com/lpc-win32/hpa-custom-controller/k8sv1beta1"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

func main() {
	config := options.NewHpaCustiomControllerOptions()
	config.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if err := config.ValidateFlags(); err != nil {
		glog.Infof("%v\n", err)
		os.Exit(-1)
	}

	as_obj, err := auto_scaler.NewAutoScaler(config)
	if err != nil {
		glog.Infof("%v\n", err)
		os.Exit(-1)
	}

	k8sScalerMetric := make(map[string]k8s_type.ScalerMetricSt)
	k8sScalerMetric["container_memory_usage_bytes"] = k8s_type.ScalerMetricSt{
		MinScaler: 30,
		MaxScaler: 30,
	}

	k8sScaleSpec_obj := &k8s_type.K8sScaleSpec{
		MinReplicas:   1,
		MaxReplicas:   3,
		Name:          "nginx-deployment-test-2",
		Namespace:     "default",
		ScalerMetrics: k8sScalerMetric,
	}

	k8sScalerMetricArr := []*k8s_type.K8sScaleSpec{k8sScaleSpec_obj}

	as_obj.ExecAutoScaler(k8sScalerMetricArr)
}
