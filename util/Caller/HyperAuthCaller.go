package Caller

import (
	"encoding/json"
	"hypercloud-api-server/util"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func setHyperAuthURL (serviceName string) string {
	hyperauthHttpPort := "8080"
	if os.Getenv("HYPERAUTH_HTTP_PORT") != "" {
		hyperauthHttpPort =  os.Getenv("HYPERAUTH_HTTP_PORT")
	}
	return util.HYPERAUTH_URL + ":" + hyperauthHttpPort + "/" + serviceName
}

func LoginAsAdmin() string {
	klog.Infoln(" [HyperAuth] Login as Admin Service")
	// Make Body for Content-Type (application/x-www-form-urlencoded)
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", "admin")
	data.Set("password", "admin")
	data.Set("client_id", "admin-cli")

	// Make Request Object
	req, err := http.NewRequest("POST", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN), strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Request with Client Object
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Result
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes) // byte to string
	klog.Infoln("Result string  : ", str)

	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
	}
	accessToken := resultJson["access_token"].(string)
	return accessToken
}

func getUserDetailWithoutToken ( userId string ) map[string]interface{} {
	klog.Infoln(" [HyperAuth] HyperAuth Get User Detail Without Token Service")

	// Make Request Object
	req, err := http.NewRequest("GET", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_USER_DETAIL_WITHOUT_TOKEN) + userId,nil)
	if err != nil {
		panic(err)
	}

	// Request with Client Object
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Result
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes) // byte to string
	klog.Infoln("Result string  : ", str)

	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
	}
	return resultJson
}