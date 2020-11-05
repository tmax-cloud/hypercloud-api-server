package alert

import (
	"encoding/json"
	"fmt"
	alertModel "hypercloud-api-server/alert/model"
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/apiserver/pkg/apis/audit"

	"k8s.io/klog"
)

//alert post
func Post(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** POST /alert")
	//queryParams := req.URL.Query()
	//var test1 alertModel.Alert
	req.ParseForm()
	log.Println(req.Form)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Get Alert Boday : %s\n", string(body))
	var v alertModel.Alert
	json.Unmarshal([]byte(string(body)), &v)

	for i := 0; i < len(v.Alert); i++ {
		if v.Alert[i] == '_' {
			fmt.Printf("%s\n", v.Alert)
			v.Resource = v.Alert[0:i]
			v.Status = v.Alert[i+1 : len(v.Alert)]
		}
	}
	fmt.Printf("status : %s\nresource : %s\nalert : %s\nnamespace : %s\nmessage : %s\nname : %s\n", v.Status, v.Resource, v.Alert, v.Namespace, v.Message, v.Name)

	var pop audit.EventList
	pop.Kind = "EventList"
	pop.APIVersion = "audit.k8s.io/v1"
	pop.Items[0].Stage = "ResponseComplete"
	pop.Items[0].Verb = "alert"
	pop.Items[0].ObjectRef.Resource = v.Resource
	pop.Items[0].ObjectRef.Namespace = v.Namespace
	pop.Items[0].ObjectRef.Name = v.Name
	pop.Items[0].ResponseStatus.Status = v.Status
	pop.Items[0].ResponseStatus.Message = v.Status
	//now := time.Now()
	//pop.Items[0].StageTimestamp = now.String

	resp, err := http.PostForm("http://example.com/form", &pop)

}

func bodyString(w http.ResponseWriter, r *http.Request) {
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	fmt.Fprintln(w, string(body))

}
