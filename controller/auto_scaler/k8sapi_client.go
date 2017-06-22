package auto_scaler

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

/*var (
	kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
)*/

type K8sapiClient struct {
	K8sapi *kubernetes.Clientset
}

func NewK8sapiClient(kubeconfig string) *K8sapiClient {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	k8sapiClient := K8sapiClient{K8sapi: clientset}

	return &k8sapiClient
}

func (k8s *K8sapiClient) K8sGetReplicas(namespace, name string) (int32, error) {
	deployments, err := k8s.K8sapi.Extensions().Deployments(namespace).Get(name)
	if err != nil {
		return -1, err
	}

	spec_num := deployments.Spec.Replicas
	if spec_num == nil {
		err = errors.New("Get Spec Replicas error")
		return -1, err
	}

	return *spec_num, err
}

func (k8s *K8sapiClient) K8sScaler(namespace, name string, scalereplicas int32) error {
	scale, err := k8s.K8sapi.Extensions().Scales(namespace).Get("deployment", name)

	check_err(err)
	scale.Spec.Replicas = scalereplicas
	_, err = k8s.K8sapi.Extensions().Scales(namespace).Update("deployment", scale)
	check_err(err)
	return err
}

func check_err(err error) {
	if err != nil {
		fmt.Printf("go err from apiserver: %s\n", err)
	}
}
