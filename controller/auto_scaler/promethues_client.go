package auto_scaler

import (
	"fmt"
	"github.com/lpc-win32/hpa-custom-controller/controller/options"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/tidwall/gjson"
)

type PromethuesClient struct {
	promethues_ip_port string
}

func NewPromethuesClient(opt *options.Options) (*PromethuesClient, error) {
	return &PromethuesClient{
		promethues_ip_port: opt.Promethues_ip_port,
	}, nil
}

func (pc *PromethuesClient) GetPromethues(promethues_query_expression string) []gjson.Result {
	// Get one minute date from promethues
	serverPath := "/api/v1/query?query=%v"
	tmp_url := fmt.Sprintf(serverPath, promethues_query_expression)
	url := fmt.Sprintf("http://%v%v", pc.promethues_ip_port, tmp_url)

	res, err := http.Get(url)
	if err != nil {
		glog.Info(err)
		return nil
	}
	jsonString, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		glog.Info(err)
		return nil
	}

	pre_metric := gjson.Get(string(jsonString), "data.result.#.value")
	if len(pre_metric.Array()) == 0 {
		glog.Info("Promethues return value is null")
		return nil
	}
	metric := gjson.Get(string(jsonString), "data.result.#.value.1")

	// 取value值
	return metric.Array()
}
