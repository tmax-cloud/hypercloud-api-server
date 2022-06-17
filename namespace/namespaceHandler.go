package namespace

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET/namespace")
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	labelSelector := queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// userGroups := queryParams.Get(util.QUERY_PARAMETER_USER_GROUP) //ex) hypercloud4,tmaxcloud,.....
	var status int

	klog.Infoln("userId : ", userId)
	if userId == "" {
		out := "userId is missing"
		status = http.StatusBadRequest
		util.SetResponse(res, out, nil, status)
		return
	}

	klog.Infoln("limit : ", limit)
	klog.Infoln("labelSelector : ", labelSelector)

	// var userGroups []string

	// if userGroup != "" {
	// 	userGroups = strings.Split(userGroup, ",")
	// }

	nsList := k8sApiCaller.GetAccessibleNS(userId, labelSelector, userGroups)

	//make OutDO
	if nsList.ResourceVersion != "" {
		status = http.StatusOK
		if len(nsList.Items) > 0 {
			if limit != "" {
				limitInt, _ := strconv.Atoi(limit)
				if len(nsList.Items) < limitInt {
					limitInt = len(nsList.Items)
				}
				nsList.Items = nsList.Items[:limitInt]
			}
		}
	} else {
		status = http.StatusForbidden
	}
	util.SetResponse(res, "", nsList, status)
}

func Put(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** PUT/namespace")
	klog.Infoln("Trial Namespace Period Extend Service Start")

	queryParams := req.URL.Query()
	nsName := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	addPeriod := queryParams.Get(util.QUERY_PARAMETER_PERIOD)
	klog.Infoln("Namespace Name : " + nsName)
	klog.Infoln("Add Period : " + addPeriod)

	namespace := k8sApiCaller.GetNamespace(nsName)

	if namespace == nil {
		klog.Infoln("333")
		status := http.StatusBadRequest
		out := "namespace is not exist"
		util.SetResponse(res, out, nil, status)
		return
	}

	if namespace.Labels != nil && namespace.Labels["fromClaim"] != "" && namespace.Labels["trial"] == "t" && namespace.Labels["period"] != "" && namespace.Annotations != nil && namespace.Annotations["owner"] != "" {
		oldPeriodInt, _ := strconv.Atoi(namespace.Labels["period"])
		addPeriodInt, _ := strconv.Atoi(addPeriod)
		newPeriod := strconv.Itoa(oldPeriodInt + addPeriodInt)
		namespace.Labels["period"] = newPeriod
		k8sApiCaller.UpdateNamespace(namespace)

		status := http.StatusOK
		out := "Trial Namespace Period Extend Service Success"
		util.SetResponse(res, out, nil, status)
	} else {
		status := http.StatusBadRequest
		out := "namespace is not trial version"
		util.SetResponse(res, out, nil, status)
	}
}

func Post(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** Post/namespace")
	hub.broadcast <- true
	klog.Infoln("broadcast namespace list to all websocket client")
	out := "broadcast namespace list to all websocket client"
	util.SetResponse(res, out, nil, http.StatusOK)
}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** OPTIONS/namespace")
	out := "**** OPTIONS/namespace"
	util.SetResponse(res, out, nil, http.StatusOK)
}

func Websocket(w http.ResponseWriter, r *http.Request) {
	conn, err := util.UpgradeWebsocket(w, r)
	if err != nil {
		klog.Errorln(err)
		return
	}

	queryParams := r.URL.Query()
	cond := urlParam{
		UserId:        queryParams.Get(util.QUERY_PARAMETER_USER_ID),
		Limit:         queryParams.Get(util.QUERY_PARAMETER_LIMIT),
		LabelSelector: queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR),
		UserGroup:     queryParams[util.QUERY_PARAMETER_USER_GROUP],
	}

	client := &Client{hub: hub, conn: conn, cond: cond, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func GetNSList(userId string, labelSelector string, userGroups []string, limit string) ([]byte, error) {
	nsList := k8sApiCaller.GetAccessibleNS(userId, labelSelector, userGroups)

	if nsList.ResourceVersion != "" {
		if len(nsList.Items) > 0 {
			if limit != "" {
				limitInt, _ := strconv.Atoi(limit)
				if len(nsList.Items) < limitInt {
					limitInt = len(nsList.Items)
				}
				nsList.Items = nsList.Items[:limitInt]
			}
		}
	}

	nsListBytes, err := json.Marshal(nsList)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return nsListBytes, nil
}
