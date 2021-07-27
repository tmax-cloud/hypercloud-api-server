package grafana

import (
	"net/http"

	//_ "github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/tmax-cloud/hypercloud-api-server/util"
)

func Get(res http.ResponseWriter, req *http.Request) {
	util.SetResponse(res, "Test Success", nil, http.StatusOK)
}

func Search(res http.ResponseWriter, req *http.Request) {
	metrics := []string{
		"billing_by_account",
		"billing_by_region",
		"billing_by_instance",
		"billing_by_metrics",
	}

	util.SetResponse(res, "", metrics, http.StatusOK)
}

func Query(res http.ResponseWriter, req *http.Request) {

}

func Annotations(res http.ResponseWriter, req *http.Request) {

}
