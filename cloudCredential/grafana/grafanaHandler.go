package grafana

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	//cc "github.com/tmax-cloud/hypercloud-api-server/cloudCredential"
	"github.com/tmax-cloud/hypercloud-api-server/cloudCredential/model"
	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

type QueryResponse struct {
	Target     string    `json:"target"`
	Datapoints [][]int64 `json:"datapoints"` // [ ["value", "timestamp in milliseconds"], ["value", ...], ...]
}

var query_response []QueryResponse

const (
	BILLING_BY_ACCOUNT  = "billing_by_account"
	BILLING_BY_REGION   = "billing_by_region"
	BILLING_BY_INSTANCE = "billing_by_instance"
	BILLING_BY_METRICS  = "billing_by_metrics"
)

func Get(res http.ResponseWriter, req *http.Request) {
	util.SetResponse(res, "Test Success", nil, http.StatusOK)
}

func Search(res http.ResponseWriter, req *http.Request) {
	available_query := []string{
		BILLING_BY_ACCOUNT,
		BILLING_BY_REGION,
		BILLING_BY_INSTANCE,
		BILLING_BY_METRICS,
	}

	util.SetResponse(res, "", available_query, http.StatusOK)
}

func Query(res http.ResponseWriter, req *http.Request) {

	var resultJson map[string]interface{}
	bytes, _ := ioutil.ReadAll(req.Body)
	str := string(bytes)
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		klog.V(1).Infoln(err.Error())
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	targets := resultJson["targets"].([]interface{})
	klog.V(3).Infoln("targets =", targets)

	/* SAMPLE TARGETS STRUCT
	[
		{
			refId: "A"
			target: "billing_by_instance"
			type: "timeserie"
			data: {
				additional-key : additional-value
			}
		},
		...
	]
	*/
	time := time.Now().Unix()
	//time *= 1000
	for i := range targets {
		query := targets[i].(map[string]interface{})
		target := query["target"]

		// TODO : additional data로 cc의 이름 받아서 특정 cloudcredential api server에 보내도록 짜야함
		// cc_name := query["data"].(map[string]interface{})["cloudcredential"]
		// klog.V(3).Infoln("cloudcredential resource name =", cc_name)

		switch target {
		case BILLING_BY_ACCOUNT:
			klog.V(3).Infoln("Handling billing_by_account...")
			requestURL := "https://localhost:443/cloudCredential?api=billing&resource=swlee-aws&namespace=hypercloud5-system&startTime=1625717822&endTime=" + strconv.FormatInt(time, 10)
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // skip TLS
			resp, err := http.Get(requestURL)
			if err != nil {
				klog.V(1).Infoln(err.Error())
				util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.V(1).Infoln(err.Error())
				util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
				return
			}

			var data []model.Awscost
			err = json.Unmarshal(body, &data)
			if err != nil {
				klog.V(1).Infoln(err.Error())
			}

			addData(BILLING_BY_ACCOUNT, data)

		case BILLING_BY_REGION:
			klog.V(3).Infoln("Handling billing_by_region...")
		case BILLING_BY_INSTANCE:
			klog.V(3).Infoln("Handling billing_by_instance...")
		case BILLING_BY_METRICS:
			klog.V(3).Infoln("Handling billing_by_metrics...")
		default:
			klog.V(1).Infoln("Invalid target")
			util.SetResponse(res, "", nil, http.StatusBadRequest)
			return
		}
	}

	util.SetResponse(res, "", query_response, http.StatusOK)
}

func Annotations(res http.ResponseWriter, req *http.Request) {

}

func addData(target string, data []model.Awscost) {
	// if target already exists,
	// append data to it
	for i := range query_response {
		if query_response[i].Target == target {
			time := time.Now().Unix()
			time *= 1000
			query_response[i].Datapoints = append(query_response[i].Datapoints, []int64{int64(data[0].Metrics["BlendedCost"].Amount), (time)})
			return
		}
	}

	// if not,
	// generate new object for the target
	temp := QueryResponse{}
	temp.Target = target
	time := time.Now().Unix()
	time *= 1000
	temp.Datapoints = append(temp.Datapoints, []int64{int64(data[0].Metrics["BlendedCost"].Amount), (time)})
	query_response = append(query_response, temp)
}
