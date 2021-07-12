package caller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

var grafanaId string
var grafanaPw string

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func GetGrafanaUser(email string) int {
	grafanaId, grafanaPw = "admin", "admin"
	httpgeturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/users/lookup?loginOrEmail=" + email
	request, _ := http.NewRequest("GET", httpgeturl, nil)
	client := &http.Client{}
	resp, err := client.Do(request)
	var GrafanaUserGet util.Grafana_User_Get
	if err != nil {
		klog.Errorln(err)
		return 0
	} else {
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal([]byte(body), &GrafanaUserGet)
		klog.Infof(string(body))
		klog.Infof(strconv.Itoa(GrafanaUserGet.Id))
	}
	return GrafanaUserGet.Id
}

func CreateGrafanaUser(email string) {
	grafanaId, grafanaPw = "admin", "admin"
	httpposturl_user := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/admin/users"
	// get grafana api key
	klog.Infof("start to create grafana apikey")

	httpposturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/auth/keys"
	var GrafanaKeyBody util.GrafanaKeyBody

	GrafanaKeyBody.Name = RandomString(8)
	GrafanaKeyBody.Role = "Admin"
	GrafanaKeyBody.SecondsToLive = 300
	json_body, _ := json.Marshal(GrafanaKeyBody)
	request, _ := http.NewRequest("POST", httpposturl, bytes.NewBuffer(json_body))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		klog.Errorln(err)
		return
	} else {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)

		klog.Infof(string(body))
		var grafana_resp util.Grafana_key
		json.Unmarshal([]byte(body), &grafana_resp)
		util.GrafanaKey = "Bearer " + grafana_resp.Key
		klog.Infof(util.GrafanaKey)
		/////create grafana user
		klog.Infof("start to create grafana user")
	}
	var grafana_user_body util.Grafana_user
	grafana_user_body.Email = email
	grafana_user_body.Name = RandomString(8)
	grafana_user_body.Login = RandomString(8)
	grafana_user_body.Password = "1234"

	json_body, _ = json.Marshal(grafana_user_body)

	request, _ = http.NewRequest("POST", httpposturl_user, bytes.NewBuffer(json_body))
	klog.Infof(string(json_body))
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
	//	request.Header.Add("Authorization", util.GrafanaKey)
	client = &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		klog.Errorln(err)
		return
	} else {
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)

		str := string(respBody)
		klog.Info(str)
		klog.Info(" Create Grafana User " + email + " Success ")
	}

}

func CreateGrafanaPermission(email string, userId int, dashboardId int) {

	// get grafana api key
	klog.Infof("start to create grafana apikey")
	grafanaId, grafanaPw = "admin", "admin"
	httpposturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/auth/keys"
	var GrafanaKeyBody util.GrafanaKeyBody

	GrafanaKeyBody.Name = RandomString(8)
	GrafanaKeyBody.Role = "Admin"
	GrafanaKeyBody.SecondsToLive = 300
	json_body, _ := json.Marshal(GrafanaKeyBody)
	request, _ := http.NewRequest("POST", httpposturl, bytes.NewBuffer(json_body))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		klog.Errorln(err)
		return
	} else {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)

		klog.Infof(string(body))
		var grafana_resp util.Grafana_key
		json.Unmarshal([]byte(body), &grafana_resp)
		util.GrafanaKey = "Bearer " + grafana_resp.Key
		klog.Infof(util.GrafanaKey)
	}
	klog.Infof("start to handle dashboard permission")
	httpposturl_per := "http://" + util.GRAFANA_URI + "api/dashboards/id/" + strconv.Itoa(dashboardId) + "/permissions"
	permBody := `{
		"items": [

			{
			"userId": ` + strconv.Itoa(userId) + `,
			"permission": 1
			}
		]
	}`
	json_body, _ = json.Marshal(permBody)
	request, _ = http.NewRequest("POST", httpposturl_per, bytes.NewBuffer([]byte(permBody)))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", util.GrafanaKey)
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		klog.Errorln(err)
		return
	} else {
		defer response.Body.Close()
		resbody, _ := ioutil.ReadAll(response.Body)
		klog.Infof(string(resbody))
	}
}

func CreateDashBoard(res http.ResponseWriter, req *http.Request) {

	body, err := ioutil.ReadAll(req.Body)
	klog.Infof(string(body))
	if err != nil {
		klog.Errorln(err)
	}

	var v util.Grafana_Namespace
	json.Unmarshal([]byte(body), &v)
	email := v.Email
	namespace := v.Namespace
	klog.Infof("Namespace Name is " + namespace)
	grafanaId, grafanaPw = "admin", "admin"
	klog.Infof("start to get api key")
	httpposturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/auth/keys"
	var GrafanaKeyBody util.GrafanaKeyBody
	GrafanaKeyBody.Name = RandomString(8)
	GrafanaKeyBody.Role = "Admin"
	GrafanaKeyBody.SecondsToLive = 300
	json_body, _ := json.Marshal(GrafanaKeyBody)
	request, _ := http.NewRequest("POST", httpposturl, bytes.NewBuffer(json_body))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, _ := client.Do(request)
	body, _ = ioutil.ReadAll(response.Body)

	var grafana_resp util.Grafana_key
	json.Unmarshal([]byte(body), &grafana_resp)
	util.GrafanaKey = "Bearer " + grafana_resp.Key

	klog.Infof("start to create Dashboard")
	httpposturl_dashboard := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/dashboards/db"
	DashBoardBody := `{
		"dashboard":
			{
				"annotations": {
				  "list": [
					{
					  "builtIn": 1,
					  "datasource": "-- Grafana --",
					  "enable": true,
					  "hide": true,
					  "iconColor": "rgba(0, 211, 255, 1)",
					  "name": "Annotations & Alerts",
					  "type": "dashboard"
					}
				  ]
				},
				"description": "prometheus operator ",
				"editable": true,
				"graphTooltip": 0,
				"id": null,
				
				"links": [],
				"panels": [
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 0
					},
					"id": 16,
					"panels": [],
					"repeat": null,
					"title": "Headlines",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"cacheTimeout": null,
					"colorBackground": false,
					"colorValue": false,
					"colors": [
					  "#299c46",
					  "rgba(237, 129, 40, 0.89)",
					  "#d44a3a"
					],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"format": "percentunit",
					"gauge": {
					  "maxValue": 100,
					  "minValue": 0,
					  "show": false,
					  "thresholdLabels": false,
					  "thresholdMarkers": true
					},
					"gridPos": {
					  "h": 3,
					  "w": 6,
					  "x": 0,
					  "y": 1
					},
					"id": 1,
					"interval": null,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"mappingType": 1,
					"mappingTypes": [
					  {
						"name": "value to text",
						"value": 1
					  },
					  {
						"name": "range to text",
						"value": 2
					  }
					],
					"maxDataPoints": 100,
					"nullPointMode": "null as zero",
					"nullText": null,
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"postfix": "",
					"postfixFontSize": "50%",
					"prefix": "",
					"prefixFontSize": "50%",
					"rangeMaps": [
					  {
						"from": "null",
						"text": "N/A",
						"to": "null"
					  }
					],
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"sparkline": {
					  "fillColor": "rgba(31, 118, 189, 0.18)",
					  "full": false,
					  "lineColor": "rgb(31, 120, 193)",
					  "show": false,
					  "ymax": null,
					  "ymin": null
					},
					"stack": false,
					"steppedLine": false,
					"tableColumn": "",
					"targets": [
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) / sum(kube_pod_container_resource_requests_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"})",
						"format": "time_series",
						"instant": true,
						"intervalFactor": 2,
						"refId": "A"
					  }
					],
					"thresholds": "70,80",
					"timeFrom": null,
					"timeShift": null,
					"title": "CPU Utilisation (from requests)",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "singlestat",
					"valueFontSize": "80%",
					"valueMaps": [
					  {
						"op": "=",
						"text": "N/A",
						"value": "null"
					  }
					],
					"valueName": "avg",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"cacheTimeout": null,
					"colorBackground": false,
					"colorValue": false,
					"colors": [
					  "#299c46",
					  "rgba(237, 129, 40, 0.89)",
					  "#d44a3a"
					],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"format": "percentunit",
					"gauge": {
					  "maxValue": 100,
					  "minValue": 0,
					  "show": false,
					  "thresholdLabels": false,
					  "thresholdMarkers": true
					},
					"gridPos": {
					  "h": 3,
					  "w": 6,
					  "x": 6,
					  "y": 1
					},
					"id": 2,
					"interval": null,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"mappingType": 1,
					"mappingTypes": [
					  {
						"name": "value to text",
						"value": 1
					  },
					  {
						"name": "range to text",
						"value": 2
					  }
					],
					"maxDataPoints": 100,
					"nullPointMode": "null as zero",
					"nullText": null,
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"postfix": "",
					"postfixFontSize": "50%",
					"prefix": "",
					"prefixFontSize": "50%",
					"rangeMaps": [
					  {
						"from": "null",
						"text": "N/A",
						"to": "null"
					  }
					],
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"sparkline": {
					  "fillColor": "rgba(31, 118, 189, 0.18)",
					  "full": false,
					  "lineColor": "rgb(31, 120, 193)",
					  "show": false,
					  "ymax": null,
					  "ymin": null
					},
					"stack": false,
					"steppedLine": false,
					"tableColumn": "",
					"targets": [
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) / sum(kube_pod_container_resource_limits_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"})",
						"format": "time_series",
						"instant": true,
						"intervalFactor": 2,
						"refId": "A"
					  }
					],
					"thresholds": "70,80",
					"timeFrom": null,
					"timeShift": null,
					"title": "CPU Utilisation (from limits)",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "singlestat",
					"valueFontSize": "80%",
					"valueMaps": [
					  {
						"op": "=",
						"text": "N/A",
						"value": "null"
					  }
					],
					"valueName": "avg",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"cacheTimeout": null,
					"colorBackground": false,
					"colorValue": false,
					"colors": [
					  "#299c46",
					  "rgba(237, 129, 40, 0.89)",
					  "#d44a3a"
					],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"format": "percentunit",
					"gauge": {
					  "maxValue": 100,
					  "minValue": 0,
					  "show": false,
					  "thresholdLabels": false,
					  "thresholdMarkers": true
					},
					"gridPos": {
					  "h": 3,
					  "w": 6,
					  "x": 12,
					  "y": 1
					},
					"id": 3,
					"interval": null,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"mappingType": 1,
					"mappingTypes": [
					  {
						"name": "value to text",
						"value": 1
					  },
					  {
						"name": "range to text",
						"value": 2
					  }
					],
					"maxDataPoints": 100,
					"nullPointMode": "null as zero",
					"nullText": null,
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"postfix": "",
					"postfixFontSize": "50%",
					"prefix": "",
					"prefixFontSize": "50%",
					"rangeMaps": [
					  {
						"from": "null",
						"text": "N/A",
						"to": "null"
					  }
					],
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"sparkline": {
					  "fillColor": "rgba(31, 118, 189, 0.18)",
					  "full": false,
					  "lineColor": "rgb(31, 120, 193)",
					  "show": false,
					  "ymax": null,
					  "ymin": null
					},
					"stack": false,
					"steppedLine": false,
					"tableColumn": "",
					"targets": [
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) / sum(kube_pod_container_resource_requests_memory_bytes{namespace=\"` + namespace + `\"})",
						"format": "time_series",
						"instant": true,
						"intervalFactor": 2,
						"refId": "A"
					  }
					],
					"thresholds": "70,80",
					"timeFrom": null,
					"timeShift": null,
					"title": "Memory Utilization (from requests)",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "singlestat",
					"valueFontSize": "80%",
					"valueMaps": [
					  {
						"op": "=",
						"text": "N/A",
						"value": "null"
					  }
					],
					"valueName": "avg",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"cacheTimeout": null,
					"colorBackground": false,
					"colorValue": false,
					"colors": [
					  "#299c46",
					  "rgba(237, 129, 40, 0.89)",
					  "#d44a3a"
					],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"format": "percentunit",
					"gauge": {
					  "maxValue": 100,
					  "minValue": 0,
					  "show": false,
					  "thresholdLabels": false,
					  "thresholdMarkers": true
					},
					"gridPos": {
					  "h": 3,
					  "w": 6,
					  "x": 18,
					  "y": 1
					},
					"id": 4,
					"interval": null,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"mappingType": 1,
					"mappingTypes": [
					  {
						"name": "value to text",
						"value": 1
					  },
					  {
						"name": "range to text",
						"value": 2
					  }
					],
					"maxDataPoints": 100,
					"nullPointMode": "null as zero",
					"nullText": null,
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"postfix": "",
					"postfixFontSize": "50%",
					"prefix": "",
					"prefixFontSize": "50%",
					"rangeMaps": [
					  {
						"from": "null",
						"text": "N/A",
						"to": "null"
					  }
					],
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"sparkline": {
					  "fillColor": "rgba(31, 118, 189, 0.18)",
					  "full": false,
					  "lineColor": "rgb(31, 120, 193)",
					  "show": false,
					  "ymax": null,
					  "ymin": null
					},
					"stack": false,
					"steppedLine": false,
					"tableColumn": "",
					"targets": [
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) / sum(kube_pod_container_resource_limits_memory_bytes{namespace=\"` + namespace + `\"})",
						"format": "time_series",
						"instant": true,
						"intervalFactor": 2,
						"refId": "A"
					  }
					],
					"thresholds": "70,80",
					"timeFrom": null,
					"timeShift": null,
					"title": "Memory Utilisation (from limits)",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "singlestat",
					"valueFontSize": "80%",
					"valueMaps": [
					  {
						"op": "=",
						"text": "N/A",
						"value": "null"
					  }
					],
					"valueName": "avg",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 4
					},
					"id": 17,
					"panels": [],
					"repeat": null,
					"title": "CPU Usage",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 5
					},
					"hiddenSeries": false,
					"id": 5,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [
					  {
						"alias": "quota - requests",
						"color": "#F2495C",
						"dashes": true,
						"fill": 0,
						"hideTooltip": true,
						"legend": false,
						"linewidth": 2,
						"stack": false
					  },
					  {
						"alias": "quota - limits",
						"color": "#FF9830",
						"dashes": true,
						"fill": 0,
						"hideTooltip": true,
						"legend": false,
						"linewidth": 2,
						"stack": false
					  }
					],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  },
					  {
						"expr": "scalar(kube_resourcequota{cluster=\"$cluster\", namespace=\"` + namespace + `\", type=\"hard\",resource=\"requests.cpu\"})",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "quota - requests",
						"legendLink": null,
						"refId": "B",
						"step": 10
					  },
					  {
						"expr": "scalar(kube_resourcequota{cluster=\"$cluster\", namespace=\"` + namespace + `\", type=\"hard\",resource=\"limits.cpu\"})",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "quota - limits",
						"legendLink": null,
						"refId": "C",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "CPU Usage",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 12
					},
					"id": 18,
					"panels": [],
					"repeat": null,
					"title": "CPU Quota",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"columns": [],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"fontSize": "100%",
					"gridPos": {
					  "h": 13,
					  "w": 24,
					  "x": 0,
					  "y": 13
					},
					"id": 6,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {},
					"pageSize": null,
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"showHeader": true,
					"sort": {
					  "col": 0,
					  "desc": true
					},
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"styles": [
					  {
						"alias": "Time",
						"align": "auto",
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"pattern": "Time",
						"type": "hidden"
					  },
					  {
						"alias": "CPU Usage",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #A",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "CPU Requests",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #B",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "CPU Requests %",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #C",
						"thresholds": [],
						"type": "number",
						"unit": "percentunit"
					  },
					  {
						"alias": "CPU Limits",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #D",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "CPU Limits %",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #E",
						"thresholds": [],
						"type": "number",
						"unit": "percentunit"
					  },
					  {
						"alias": "Pod",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": true,
						"linkTooltip": "Drill down",
						"linkUrl": "./d/6581e46e4e5c7ba40a07646395ef7b23/k8s-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-pod=$__cell",
						"pattern": "pod",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"pattern": "/.*/",
						"thresholds": [],
						"type": "string",
						"unit": "short"
					  }
					],
					"targets": [
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "A",
						"step": 10
					  },
					  {
						"expr": "sum(kube_pod_container_resource_requests_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "B",
						"step": 10
					  },
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod) / sum(kube_pod_container_resource_requests_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "C",
						"step": 10
					  },
					  {
						"expr": "sum(kube_pod_container_resource_limits_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "D",
						"step": 10
					  },
					  {
						"expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod) / sum(kube_pod_container_resource_limits_cpu_cores{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "E",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeShift": null,
					"title": "CPU Quota",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"transform": "table",
					"type": "table",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 26
					},
					"id": 19,
					"panels": [],
					"repeat": null,
					"title": "Memory Usage",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 27
					},
					"hiddenSeries": false,
					"id": 7,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [
					  {
						"alias": "quota - requests",
						"color": "#F2495C",
						"dashes": true,
						"fill": 0,
						"hideTooltip": true,
						"legend": false,
						"linewidth": 2,
						"stack": false
					  },
					  {
						"alias": "quota - limits",
						"color": "#FF9830",
						"dashes": true,
						"fill": 0,
						"hideTooltip": true,
						"legend": false,
						"linewidth": 2,
						"stack": false
					  }
					],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\", container!=\"\"}) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  },
					  {
						"expr": "scalar(kube_resourcequota{cluster=\"$cluster\", namespace=\"` + namespace + `\", type=\"hard\",resource=\"requests.memory\"})",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "quota - requests",
						"legendLink": null,
						"refId": "B",
						"step": 10
					  },
					  {
						"expr": "scalar(kube_resourcequota{cluster=\"$cluster\", namespace=\"` + namespace + `\", type=\"hard\",resource=\"limits.memory\"})",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "quota - limits",
						"legendLink": null,
						"refId": "C",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Memory Usage (w/o cache)",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "bytes",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 34
					},
					"id": 20,
					"panels": [],
					"repeat": null,
					"title": "Memory Quota",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"columns": [],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"fontSize": "100%",
					"gridPos": {
					  "h": 14,
					  "w": 24,
					  "x": 0,
					  "y": 35
					},
					"id": 8,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {},
					"pageSize": null,
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"showHeader": true,
					"sort": {
					  "col": 0,
					  "desc": true
					},
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"styles": [
					  {
						"alias": "Time",
						"align": "auto",
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"pattern": "Time",
						"type": "hidden"
					  },
					  {
						"alias": "Memory Usage",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #A",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Memory Requests",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #B",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Memory Requests %",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #C",
						"thresholds": [],
						"type": "number",
						"unit": "percentunit"
					  },
					  {
						"alias": "Memory Limits",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #D",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Memory Limits %",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #E",
						"thresholds": [],
						"type": "number",
						"unit": "percentunit"
					  },
					  {
						"alias": "Memory Usage (RSS)",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #F",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Memory Usage (Cache)",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #G",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Memory Usage (Swap)",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #H",
						"thresholds": [],
						"type": "number",
						"unit": "bytes"
					  },
					  {
						"alias": "Pod",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": true,
						"linkTooltip": "Drill down",
						"linkUrl": "./d/6581e46e4e5c7ba40a07646395ef7b23/k8s-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-pod=$__cell",
						"pattern": "pod",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"pattern": "/.*/",
						"thresholds": [],
						"type": "string",
						"unit": "short"
					  }
					],
					"targets": [
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "A",
						"step": 10
					  },
					  {
						"expr": "sum(kube_pod_container_resource_requests_memory_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "B",
						"step": 10
					  },
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod) / sum(kube_pod_container_resource_requests_memory_bytes{namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "C",
						"step": 10
					  },
					  {
						"expr": "sum(kube_pod_container_resource_limits_memory_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "D",
						"step": 10
					  },
					  {
						"expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod) / sum(kube_pod_container_resource_limits_memory_bytes{namespace=\"` + namespace + `\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "E",
						"step": 10
					  },
					  {
						"expr": "sum(container_memory_rss{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "F",
						"step": 10
					  },
					  {
						"expr": "sum(container_memory_cache{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "G",
						"step": 10
					  },
					  {
						"expr": "sum(container_memory_swap{cluster=\"$cluster\", namespace=\"` + namespace + `\",container!=\"\"}) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "H",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeShift": null,
					"title": "Memory Quota",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"transform": "table",
					"type": "table",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 49
					},
					"id": 21,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"columns": [],
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 1,
					"fontSize": "100%",
					"gridPos": {
					  "h": 13,
					  "w": 24,
					  "x": 0,
					  "y": 50
					},
					"id": 9,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {},
					"pageSize": null,
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"showHeader": true,
					"sort": {
					  "col": 0,
					  "desc": true
					},
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"styles": [
					  {
						"alias": "Time",
						"align": "auto",
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"pattern": "Time",
						"type": "hidden"
					  },
					  {
						"alias": "Current Receive Bandwidth",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #A",
						"thresholds": [],
						"type": "number",
						"unit": "Bps"
					  },
					  {
						"alias": "Current Transmit Bandwidth",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #B",
						"thresholds": [],
						"type": "number",
						"unit": "Bps"
					  },
					  {
						"alias": "Rate of Received Packets",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #C",
						"thresholds": [],
						"type": "number",
						"unit": "pps"
					  },
					  {
						"alias": "Rate of Transmitted Packets",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #D",
						"thresholds": [],
						"type": "number",
						"unit": "pps"
					  },
					  {
						"alias": "Rate of Received Packets Dropped",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #E",
						"thresholds": [],
						"type": "number",
						"unit": "pps"
					  },
					  {
						"alias": "Rate of Transmitted Packets Dropped",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": false,
						"linkTooltip": "Drill down",
						"linkUrl": "",
						"pattern": "Value #F",
						"thresholds": [],
						"type": "number",
						"unit": "pps"
					  },
					  {
						"alias": "Pod",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"link": true,
						"linkTooltip": "Drill down to pods",
						"linkUrl": "./d/6581e46e4e5c7ba40a07646395ef7b23/k8s-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-pod=$__cell",
						"pattern": "pod",
						"thresholds": [],
						"type": "number",
						"unit": "short"
					  },
					  {
						"alias": "",
						"align": "auto",
						"colorMode": null,
						"colors": [],
						"dateFormat": "YYYY-MM-DD HH:mm:ss",
						"decimals": 2,
						"pattern": "/.*/",
						"thresholds": [],
						"type": "string",
						"unit": "short"
					  }
					],
					"targets": [
					  {
						"expr": "sum(irate(container_network_receive_bytes_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "A",
						"step": 10
					  },
					  {
						"expr": "sum(irate(container_network_transmit_bytes_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "B",
						"step": 10
					  },
					  {
						"expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "C",
						"step": 10
					  },
					  {
						"expr": "sum(irate(container_network_transmit_packets_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "D",
						"step": 10
					  },
					  {
						"expr": "sum(irate(container_network_receive_packets_dropped_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "E",
						"step": 10
					  },
					  {
						"expr": "sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[5m])) by (pod)",
						"format": "table",
						"instant": true,
						"intervalFactor": 2,
						"legendFormat": "",
						"refId": "F",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeShift": null,
					"title": "Current Network Usage",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"transform": "table",
					"type": "table",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					]
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 63
					},
					"id": 22,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 64
					},
					"hiddenSeries": false,
					"id": 10,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_receive_bytes_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Receive Bandwidth",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 71
					},
					"id": 23,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 72
					},
					"hiddenSeries": false,
					"id": 11,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_transmit_bytes_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Transmit Bandwidth",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 79
					},
					"id": 24,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 80
					},
					"hiddenSeries": false,
					"id": 12,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Rate of Received Packets",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 87
					},
					"id": 25,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 88
					},
					"hiddenSeries": false,
					"id": 13,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Rate of Transmitted Packets",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 95
					},
					"id": 26,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 96
					},
					"hiddenSeries": false,
					"id": 14,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_receive_packets_dropped_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Rate of Received Packets Dropped",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"datasource": "prometheus",
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 103
					},
					"id": 27,
					"panels": [],
					"repeat": null,
					"title": "Network",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "$datasource",
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 7,
					  "w": 24,
					  "x": 0,
					  "y": 104
					},
					"hiddenSeries": false,
					"id": 15,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 0,
					"links": [],
					"nullPointMode": "null as zero",
					"options": {
					  "dataLinks": []
					},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": true,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\"` + namespace + `\"}[$__interval])) by (pod)",
						"format": "time_series",
						"intervalFactor": 2,
						"legendFormat": "{{pod}}",
						"legendLink": null,
						"refId": "A",
						"step": 10
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Rate of Transmitted Packets Dropped",
					"tooltip": {
					  "shared": false,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": false
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  }
				],
				"refresh": "10s",
				"schemaVersion": 20,
				"style": "dark",
				"tags": [
				  "kubernetes-mixin"
				],
				"templating": {
				  "list": [
					{
					  "current": {
						"text": "prometheus",
						"value": "prometheus"
					  },
					  "hide": 0,
					  "includeAll": false,
					  "label": null,
					  "multi": false,
					  "name": "datasource",
					  "options": [],
					  "query": "prometheus",
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "type": "datasource"
					},
					{
					  "allValue": null,
					  "current": {
						"isNone": true,
						"text": "None",
						"value": ""
					  },
					  "datasource": "$datasource",
					  "definition": "",
					  "hide": 2,
					  "includeAll": false,
					  "label": null,
					  "multi": false,
					  "name": "cluster",
					  "options": [],
					  "query": "label_values(kube_pod_info, cluster)",
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 1,
					  "tagValuesQuery": "",
					  "tags": [],
					  "tagsQuery": "",
					  "type": "query",
					  "useTags": false
					}
				  ]
				},
				"time": {
				  "from": "now-1h",
				  "to": "now"
				},
				"timepicker": {
				  "refresh_intervals": [
					"5s",
					"10s",
					"30s",
					"1m",
					"5m",
					"15m",
					"30m",
					"1h",
					"2h",
					"1d"
				  ],
				  "time_options": [
					"5m",
					"15m",
					"1h",
					"6h",
					"12h",
					"24h",
					"2d",
					"7d",
					"30d"
				  ]
				},
				"timezone": "",
				"title": "Kubernetes / Compute Resources / Namespace (Pods)-` + namespace + `",
				"version": 2,
				"uid": "` + namespace + `"
			  }
		}
	}`
	json_body, _ = json.Marshal(DashBoardBody)

	request_db, _ := http.NewRequest("POST", httpposturl_dashboard, bytes.NewBuffer([]byte(DashBoardBody)))
	request_db.Header.Add("Content-Type", "application/json; charset=UTF-8")
	//request.Header.Add("Authorization", util.GrafanaKey)
	client = &http.Client{}
	resp, err := client.Do(request_db)
	if err != nil {
		klog.Errorln(err)
	} else {
		defer resp.Body.Close()
		var grafana_resp_dash util.Grafana_Dashboad_resp

		dashbody, _ := ioutil.ReadAll(resp.Body)

		klog.Infof(string(dashbody))
		json.Unmarshal([]byte(dashbody), &grafana_resp_dash)

		dashboardId := grafana_resp_dash.Id
		klog.Infof("start to get grafana user info")
		userId := GetGrafanaUser(email)
		klog.Infof(strconv.Itoa(userId))
		CreateGrafanaPermission(email, userId, dashboardId)
	}

}

func DeleteGrafanaDashboard(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	namespace := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	grafanaId, grafanaPw = "admin", "admin"
	// get grafana api key
	klog.Infof("start to create grafana apikey")

	httpposturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/auth/keys"
	var GrafanaKeyBody util.GrafanaKeyBody

	GrafanaKeyBody.Name = RandomString(8)
	GrafanaKeyBody.Role = "Admin"
	GrafanaKeyBody.SecondsToLive = 300
	json_body, _ := json.Marshal(GrafanaKeyBody)
	request, _ := http.NewRequest("POST", httpposturl, bytes.NewBuffer(json_body))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)

		klog.Infof(string(body))
		var grafana_resp util.Grafana_key
		json.Unmarshal([]byte(body), &grafana_resp)
		util.GrafanaKey = "Bearer " + grafana_resp.Key
		klog.Infof(util.GrafanaKey)
	}

	httpposturl_dashboard := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/dashboards/uid/" + namespace
	request, _ = http.NewRequest("DELETE", httpposturl_dashboard, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", util.GrafanaKey)
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		if err != nil {
			klog.Errorln(err)

		}
		klog.Infof(string(body))
	}
}

func DeleteGrafanaUser(email string) {
	id := GetGrafanaUser(email)
	grafanaId, grafanaPw = "admin", "admin"
	httpurl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/admin/users/" + strconv.Itoa(id)

	request_user_delete, _ := http.NewRequest("DELETE", httpurl, nil)

	client := &http.Client{}
	resp, err := client.Do(request_user_delete)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer resp.Body.Close()
		respBody, _ := ioutil.ReadAll(resp.Body)
		klog.Infof(string(respBody))
	}

}
