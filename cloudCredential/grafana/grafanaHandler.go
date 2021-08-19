package grafana

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
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
		klog.Errorln("Error occured during unmarshaling request body")
		util.SetResponse(res, "", nil, http.StatusInternalServerError)
		return
	}
	targets := resultJson["targets"].([]interface{})
	klog.Infoln("targets =", targets)

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
	var qr []QueryResponse
	var temp QueryResponse
	for i := range targets {
		query := targets[i].(map[string]interface{})
		target := query["target"]

		// TODO : additional data로 cc의 이름 받아서 특정 cloudcredential api server에 보내도록 짜야함
		// cc_name := query["data"].(map[string]interface{})["cloudcredential"]
		// klog.Infoln("cloudcredential resource name =", cc_name)

		switch target {
		case BILLING_BY_ACCOUNT:
			klog.Infoln("Handling billing_by_account...")
			requestURL := "https://localhost:443/cloudCredential?api=billing&resource=swlee-aws&namespace=hypercloud5-system&endTime=1625717822&startTime=1625065200"
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // skip TLS
			resp, err := http.Get(requestURL)
			if err != nil {
				klog.Errorln("Error while calling /cloudCredential API")
				util.SetResponse(res, "Error while calling /cloudCredential API", nil, http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Errorln(err.Error())
				util.SetResponse(res, "Failed to read response body", nil, http.StatusInternalServerError)
				return
			}

			var data []model.Awscost //interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				klog.Errorln(err.Error())
			}

			klog.Infoln("cloudcredential API response.data ==", data)

			temp = QueryResponse{}
			temp.Target = "billing_by_account"
			time := time.Now().Unix()
			time *= 1000
			temp.Datapoints = append(temp.Datapoints, []int64{int64(data[0].Metrics["BlendedCost"].Amount), (time)})
			qr = append(qr, temp)

		case BILLING_BY_REGION:
			klog.Infoln("Handling billing_by_region...")
		case BILLING_BY_INSTANCE:
			klog.Infoln("Handling billing_by_instance...")
		case BILLING_BY_METRICS:
			klog.Infoln("Handling billing_by_metrics...")
		default:
			klog.Errorln("Invalid target")
			util.SetResponse(res, "", nil, http.StatusBadRequest)
			return
		}
	}

	// var qr []QueryResponse
	// var temp QueryResponse
	// temp = QueryResponse{}
	// temp.Target = "billing_by_account"
	// time := time.Now().Unix()
	// time *= 1000
	// temp.Datapoints = append(temp.Datapoints, []int64{3, (time - 10000*1000)})
	// temp.Datapoints = append(temp.Datapoints, []int64{1, (time)})
	// qr = append(qr, temp)

	// temp = QueryResponse{}
	// temp.Target = "billing_by_region"
	// temp.Datapoints = append(temp.Datapoints, []int64{8, (time - 1000*1000)})
	// temp.Datapoints = append(temp.Datapoints, []int64{3, (time)})
	// qr = append(qr, temp)

	//klog.Infoln("qr =", qr)

	util.SetResponse(res, "", qr, http.StatusOK)
}

func Annotations(res http.ResponseWriter, req *http.Request) {

}
