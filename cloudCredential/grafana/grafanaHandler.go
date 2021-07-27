package grafana

import (
	"net/http"

	//_ "github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/tmax-cloud/hypercloud-api-server/util"
)

func Get(res http.ResponseWriter, req *http.Request) {
	util.SetResponse(res, "Test Success", "", http.StatusOK)
}

func Search(res http.ResponseWriter, req *http.Request) {

}

func Query(res http.ResponseWriter, req *http.Request) {

}

func Annotations(res http.ResponseWriter, req *http.Request) {

}
