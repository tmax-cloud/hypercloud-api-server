package caller

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tmax-cloud/hypercloud-api-server/util"

	"k8s.io/klog"
)

var (
	HYPERAUTH_URL          string
	HYPERAUTH_REALM_PREFIX string
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // ignore certificate
}

func setHyperAuthURL(serviceName string) string {
	return HYPERAUTH_URL + serviceName
}

func LoginAsAdmin() string {
	klog.V(3).Infoln(" [HyperAuth] Login as Admin Service")
	// Make Body for Content-Type (application/x-www-form-urlencoded)
	id, password, err := GetHyperAuthAdminAccount()
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", id)
	data.Set("password", password)
	data.Set("client_id", "admin-cli")

	// Make Request Object
	req, err := http.NewRequest("POST", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN), strings.NewReader(data.Encode()))
	if err != nil {
		klog.V(1).Infoln(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Request with Client Object
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(1).Infoln(err)
		panic(err)
	}
	defer resp.Body.Close()

	// Result
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes) // byte to string
	// klog.V(3).Infoln("Result string  : ", str)

	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		klog.V(1).Infoln(err)
	}
	accessToken := resultJson["access_token"].(string)
	return accessToken
}

func GetHyperAuthUserDetail(userId string) (map[string]interface{}, error) {
	adminToken := LoginAsAdmin()

	req, err := http.NewRequest("GET", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_USER_DETAIL)+userId, nil)
	q := req.URL.Query()
	q.Add("token", adminToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes) // byte to string
	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		klog.V(1).Infoln("Result string  : ", str)
		klog.V(1).Infoln(err)
		return nil, err
	}

	return resultJson, nil
}

func GetHyperAuthGroupByUser(userId string) ([]string, error) {
	userInfo, err := GetHyperAuthUserDetail(userId)
	if err != nil {
		klog.V(1).Infoln(err)
		return []string{}, err
	}

	groups := userInfo["groups"].([]interface{})

	var result []string
	for _, group := range groups {
		result = append(result, group.(string))
	}
	// klog.V(3).Infoln(result)

	return result, nil
}
