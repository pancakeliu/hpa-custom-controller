package main

import (
    "fmt"
    "flag"
    "os"
    "io/ioutil"

    "gopkg.in/yaml.v2"
    "github.com/golang/glog"
    "github.com/lpc-win32/hpa-custom-controller/hpactl/cmd"
)

type EtcdConfig struct {
    EtcdString []string
}

func initConfig() (*EtcdConfig, error) {
    // 读取hpactl-config配置文件
    config_file_path := "/root/hpactl-config.yaml"
    data, err := ioutil.ReadFile(config_file_path)
    if err != nil {
        glog.Errorf("read file %s error", config_file_path)
        return nil, err
    }
    etcd_config := &EtcdConfig{}

    // 解析yaml配置文件
    err = yaml.Unmarshal([]byte(data), etcd_config)
    if err != nil {
        glog.Error("yaml config file unmarshal error")
    }

    return etcd_config, nil
}

func exec_hpactl(c *cmd.Command, cmds []string) {
    switch cmds[0] {
        case "help":
            c.Help()
        case "create":
            if len(cmds) != 2 {
                fmt.Println("Usage error!")
                c.Help()
            }
            c.Create(cmds[1])
        case "update":
            if len(cmds) != 2 {
                fmt.Println("Usage error!")
                c.Help()
            }
            c.Update(cmds[1])
        case "get":
            if len(cmds) != 2 {
                fmt.Println("Usage error!")
                c.Help()
            }
            c.Get(cmds[1])
        case "delete":
            if len(cmds) != 3 {
                fmt.Println("Usage error!")
                c.Delete(cmds[1], cmds[2])
            }
    }
}

func main() {
    flag.Parse()

    config, err := initConfig()
    if err != nil {
        glog.Error("InitConfig error")
        os.Exit(-1)
    }

    fmt.Printf("etcd is %s\n", config.EtcdString)

    cmd_manage := cmd.NewCommand(config.EtcdString)
    cmds := flag.Args()
    if len(cmds) == 0 {
        cmd_manage.Help()
        os.Exit(1)
    }

    // 执行hpactl方法
    exec_hpactl(cmd_manage, cmds)
}
