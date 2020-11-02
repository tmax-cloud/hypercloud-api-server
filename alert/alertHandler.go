package alert

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	log.Println(string(body))
	fmt.Printf("%s\n", string(body))
}

func bodyString(w http.ResponseWriter, r *http.Request) {
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	fmt.Fprintln(w, string(body))

}
