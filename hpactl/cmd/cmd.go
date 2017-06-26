package cmd

import (
    "encoding/json"
    "io/ioutil"
    "fmt"

    "gopkg.in/yaml.v2"

    "github.com/golang/glog"
    "github.com/lpc-win32/hpa-custom-controller/collector/etcd_client"
    "github.com/lpc-win32/hpa-custom-controller/k8sv1beta1"
)

type Command struct {
    EtcdClient *etcd_client.EtcdClient
}

func NewCommand(command_strs []string) *Command {
    client, _ := etcd_client.NewClient(command_strs)
    c_obj := etcd_client.EtcdClient{}
    c_obj.Client = client

    return &Command{ EtcdClient: &c_obj }
}

func (c *Command) Help() {
    fmt.Println("hpactl usage:")
    fmt.Println("    create xxx.yaml (insert k-v into etcd)")
    fmt.Println("    delete namespace name")
    fmt.Println("    get    namespace")
    fmt.Println("    update xxx.yaml (update k-v)")
}

func (c *Command) Create(config_file_path string) error {
    data, err := ioutil.ReadFile(config_file_path)
    if err != nil {
        glog.Errorf("config file %s read error", config_file_path)
        return err
    }
    spec := &k8sv1beta1.K8sScaleSpec{}
    err = yaml.Unmarshal([]byte(data), spec)
    if err != nil {
        glog.Error("yaml unmarshal error")
        return err
    }
    key := "/api/v1/autoscalers/" + spec.Namespace + "/" + spec.Name
    value, err := json.Marshal(spec)
    if err != nil {
        glog.Error("json marshal error")
        return err
    }

    err = c.EtcdClient.EtcdSet(string(key), string(value))
    if err != nil {
        glog.Errorf("key: %s, value: %s etcd set error", string(key), string(value))
        return err
    }

    return nil
}

func (c *Command) Delete(namespace string, name string) error {
    key := "/api/v1/autoscalers/" + namespace + "/" + name
    err := c.EtcdClient.EtcdDelete(key)
    if err != nil {
        glog.Error("etcd delete error")
        return err
    }
    return nil
}

func (c *Command) Get(namespace string) error {
    key_prefix := "api/v1/autoscalers/" + namespace
    key_list, err := c.EtcdClient.EtcdList(key_prefix)
    if err != nil {
        glog.Error("etcd get error")
        return err
    }
    for _, key := range key_list {
        val, _ := c.EtcdClient.EtcdGet(string(key))
        spec := &k8sv1beta1.K8sScaleSpec{}
        err = json.Unmarshal([]byte(val), spec)
        if err != nil {
            glog.Error("key:%s json unmarshal error")
            return err
        }
    }
    return nil
}

func (c *Command) Update(path string) error {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        glog.Errorf("read file %s error", path)
        return err
    }
    spec := &k8sv1beta1.K8sScaleSpec{}
    err = yaml.Unmarshal([]byte(data), spec)
    if err != nil {
        glog.Error("yaml unmarshal error")
        return err
    }

    key := "/api/v1/autoscalers/" + spec.Namespace + "/" + spec.Name
    value, err := json.Marshal(spec)
    if err != nil {
        glog.Error("json marshal error")
        return err
    }
    err = c.EtcdClient.EtcdUpdate(key, string(value))
    if err != nil {
        glog.Error("ercd update error")
        return err
    }
    return nil
}
