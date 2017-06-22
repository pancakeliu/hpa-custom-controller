package etcd_client

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func NewClient(kubeconfig []string) (client.Client, error) {
	cfg := client.Config{
		Endpoints: kubeconfig,
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)

	return c, err
}

type EtcdClient struct {
	Client client.Client
}

func (c *EtcdClient) EtcdSet(key, value string) error {
	kapi := client.NewKeysAPI(c.Client)
	_, err := kapi.Set(context.Background(), key, value, nil)
	return err
}
func (c *EtcdClient) EtcdGet(key string) (string, error) {
	kapi := client.NewKeysAPI(c.Client)
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return "", err
	}
	return resp.Node.Value, err
}

func (c *EtcdClient) EtcdDelete(key string) error {
	kapi := client.NewKeysAPI(c.Client)
	_, err := kapi.Delete(context.Background(), key, nil)
	return err
}
func (c *EtcdClient) EtcdList(dir string) ([]string, error) {
	kapi := client.NewKeysAPI(c.Client)
	resp, err := kapi.Get(context.Background(), dir, nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Get error!")
		return nil, err
	}
	list := make([]string, 5)
	//	var list []string
	i := 0
	for _, value := range resp.Node.Nodes {
		list[i] = value.Key
		i++
	}
	return list, err
}

func (c *EtcdClient) EtcdUpdate(key, value string) error {
	kapi := client.NewKeysAPI(c.Client)
	_, err := kapi.Update(context.Background(), key, value)
	return err
}
