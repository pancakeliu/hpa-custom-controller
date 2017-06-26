package k8sv1beta1

type ScalerMetricSt struct {
	MinScaler float64
	MaxScaler float64
}
type K8sScaleSpec struct {
	MinReplicas   int32  `json:"minReplicas,omitempty"`
	MaxReplicas   int32  `json:"maxReplicas,omitempty"`
	Name          string `json:"name,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ScalerMetrics map[string]ScalerMetricSt
}
