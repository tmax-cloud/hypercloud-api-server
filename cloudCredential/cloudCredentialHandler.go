package cloudcredential

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /cloudCredential")
	queryParams := req.URL.Query()
	api := queryParams.Get(util.QUERY_PARAMETER_API)
	resource := queryParams.Get(util.QUERY_PARAMETER_RESOURCE)
	namespace := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	account := queryParams.Get(util.QUERY_PARAMETER_ACCOUNT)

	klog.V(3).Infoln("API = ", api, " / RESOURCE = ", resource, " / NAMESPACE = ", namespace, " / ACCOUNT = ", account)

	reqURL := "http://" + resource + "-credential-server-service." + namespace + ".svc.cluster.local" // DEFAULT URL
	switch api {
	case "billing", "Billing", "BILLING":
		reqURL += "/billing"
		param := "?"
		param += AppendParam(req, util.QUERY_PARAMETER_STARTTIME)
		param += AppendParam(req, util.QUERY_PARAMETER_ENDTIME)
		param += AppendParam(req, util.QUERY_PARAMETER_GRANULARITY)
		param += AppendParamArray(req, util.QUERY_PARAMETER_METRICS)
		param += AppendParam(req, util.QUERY_PARAMETER_DIMENSION)
		param += AppendParam(req, util.QUERY_PARAMETER_SORT)
		param = strings.TrimRight(param, "&")
		reqURL += param
	default:
		klog.V(1).Infoln("NO API is given")
		util.SetResponse(res, "", "NO API is given", http.StatusBadRequest)
		return
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}
	klog.V(3).Infoln("Request URL = ", reqURL)
	response, err := client.Get(reqURL)
	if err != nil {
		klog.V(1).Infoln("HTTP Error : ", err)
		util.SetResponse(res, "", err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.V(1).Infoln(err.Error())
		util.SetResponse(res, "Failed to read response body", nil, http.StatusInternalServerError)
		return
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		klog.V(1).Infoln(err.Error())
	}
	klog.V(3).Infoln("Results: ", data)

	util.SetResponse(res, "", data, http.StatusOK)
}

func AppendParam(req *http.Request, param string) string {
	p := req.URL.Query().Get(param)
	if p == "" {
		return ""
	}
	return param + "=" + p + "&"
}

func AppendParamArray(req *http.Request, param string) string {
	p := req.URL.Query()[param]
	result := ""
	for _, content := range p {
		result += param + "=" + content + "&"
	}
	return result
}
