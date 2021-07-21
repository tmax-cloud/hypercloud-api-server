package audit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmax-cloud/hypercloud-api-server/util"
	"github.com/tmax-cloud/hypercloud-api-server/util/caller"
	auditDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/audit"
	corev1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

type urlParam struct {
	Search        string               `json:"search"`
	UserId        string               `json:"userId"`
	Namespace     string               `json:"namespace"`
	NamespaceList corev1.NamespaceList `json:"namespaceList"`
	Resource      string               `json:"resource"`
	StartTime     string               `json:"startTime"`
	EndTime       string               `json:"endTime"`
	Limit         string               `json:"limit"`
	Offset        string               `json:"offset"`
	Code          string               `json:"code"`
	Verb          string               `json:"verb"`
	Status        string               `json:"status"`
	Sort          []string             `json:"sort"`
}

type response struct {
	EventList        audit.EventList `json:"eventList"`
	RowsCount        int64           `json:"rowsCount"`
	ClusterName      string          `json:"clusterName"`
	ClusterNamespace string          `json:"clusterNamespace"`
}

type MemberListResponse struct {
	MemberList []string `json:"memberList"`
}

func ListAuditVerb(w http.ResponseWriter, r *http.Request) {
	//fixed size array
	var verbList = [...]string{"create", "update", "patch", "delete", "deletecollection", "LOGIN", "LOGOUT", "LOGIN_ERROR"}
	util.SetResponse(w, "", verbList, http.StatusOK)
	return
}

func ListAuditResource(w http.ResponseWriter, r *http.Request) {
	util.SetResponse(w, "", caller.AuditResourceList, http.StatusOK)
	return
}

func UpdateAuditResource() {
	klog.Infoln("Update Audit resource list")
	caller.UpdateAuditResourceList()
	return
}
func AddAudit(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var eventList audit.EventList
	clusterName := r.URL.Query().Get(util.QUERY_PARAMETER_CLUSTER_NAME)
	clusterNamespace := r.URL.Query().Get(util.QUERY_PARAMETER_CLUSTER_NAMESPACE)
	if clusterName == "" {
		clusterName = "master"
	}
	if clusterNamespace == "" {
		clusterNamespace = ""
	}

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if err := json.Unmarshal(body, &eventList); err != nil {
		util.SetResponse(w, "", nil, http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	for _, event := range eventList.Items {
		if event.ResponseStatus.Status == "" && (event.ResponseStatus.Code/100) == 2 {
			event.ResponseStatus.Status = "Success"
		}
	}

	auditDataFactory.Insert(eventList.Items, clusterName, clusterNamespace)
	if len(hub.clients) > 0 {
		hub.broadcast <- eventList
	}
	util.SetResponse(w, "", nil, http.StatusOK)
}

func AddAuditBatch(w http.ResponseWriter, r *http.Request) {
	klog.Info("AddAuditBatch")
	var body []byte
	var event audit.Event
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if err := json.Unmarshal(body, &event); err != nil {
		klog.Error(err)
		util.SetResponse(w, "", nil, http.StatusInternalServerError)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	event.AuditID = types.UID(uuid.New().String())
	if event.StageTimestamp.Time.IsZero() {
		event.StageTimestamp.Time = time.Now()
	}

	clusterName := r.URL.Query().Get(util.QUERY_PARAMETER_CLUSTER_NAME)
	clusterNamespace := r.URL.Query().Get(util.QUERY_PARAMETER_CLUSTER_NAMESPACE)
	if clusterName == "" {
		clusterName = "master"
	}
	if clusterNamespace == "" {
		clusterNamespace = ""
	}

	if len(EventBuffer.Buffer) < BufferSize {
		EventBuffer.Buffer <- event
		EventBuffer.clusterName <- clusterName
		EventBuffer.clusterNamespace <- clusterNamespace
	} else {
		klog.Error("event is dropped.")
	}
	util.SetResponse(w, "", nil, http.StatusOK)
}

func MemberSuggestions(res http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
	search := r.URL.Query().Get("search")

	var b strings.Builder
	b.WriteString("select username, count(*) from audit where 1=1 ")

	if sarResult, err := caller.CreateSubjectAccessReview(userId, nil, "", "namespaces", "", "", "list"); err != nil {
		klog.Errorln(err)
	} else if sarResult.Status.Allowed == false {
		b.WriteString("and username = '")
		b.WriteString(userId)
		b.WriteString("' ")
	}
	b.WriteString("and username like '")
	b.WriteString(search)
	if search != "" {
		b.WriteString("%' group by username order by count desc limit 5")
	} else {
		b.WriteString("' group by username order by count desc limit 5")
	}
	query := b.String()
	klog.Info("query: ", query)
	memberList, _ := auditDataFactory.GetMemberList(query)

	memberListResponse := MemberListResponse{
		MemberList: memberList,
		// RowsCount:  count,
	}

	util.SetResponse(res, "", memberListResponse, http.StatusOK)
}

func GetAudit(res http.ResponseWriter, req *http.Request) {
	var nsList corev1.NamespaceList
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// ns get줘도 감사기록은 안보이게..
	nsListSAR, err := caller.CreateSubjectAccessReview(userId, userGroups, "", "namespaces", "", "", "list")
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, "", nil, http.StatusInternalServerError)
		return
	}

	if !nsListSAR.Status.Allowed {
		if queryParams.Get("namespace") == "" {
			util.SetResponse(res, "Non-admin users should select namespace.", nil, http.StatusBadRequest)
			return
		}
		tmp := []string{}
		// list ns w/ labelselector
		if nsList = caller.GetAccessibleNS(userId, "", userGroups); len(nsList.Items) == 0 {
			util.SetResponse(res, "no ns", nil, http.StatusOK)
			return
		}
		for _, item := range nsList.Items {
			if item.Annotations["owner"] == userId {
				tmp = append(tmp, item.Name)
			}
		}
		if !util.Contains(tmp, queryParams.Get("namespace")) {
			util.SetResponse(res, "Not authorized", nil, http.StatusForbidden)
			return
		}
	}

	// search := r.URL.Query().Get("search")

	urlParam := urlParam{}
	urlParam.Search = queryParams.Get("search")
	urlParam.UserId = userId
	urlParam.Namespace = queryParams.Get("namespace")
	urlParam.Resource = queryParams.Get("resource")
	urlParam.Limit = queryParams.Get("limit")
	urlParam.Offset = queryParams.Get("offset")
	urlParam.Code = queryParams.Get("code")
	urlParam.Verb = queryParams.Get("verb")
	urlParam.Sort = queryParams["sort"]
	urlParam.StartTime = queryParams.Get("startTime")
	urlParam.EndTime = queryParams.Get("endTime")
	urlParam.Status = queryParams.Get("status")
	urlParam.NamespaceList = nsList

	query := queryBuilder(urlParam)
	eventList, count := auditDataFactory.Get(query) // 반환 값 추가해야함

	response := response{
		EventList: eventList,
		RowsCount: count,
	}

	util.SetResponse(res, "", response, http.StatusOK)
}

func queryBuilder(param urlParam) string {
	// search := param.Search
	// userId := param.UserId
	namespace := param.Namespace
	resource := param.Resource
	startTime := param.StartTime
	endTime := param.EndTime
	limit := param.Limit
	offset := param.Offset
	code := param.Code
	verb := param.Verb
	sort := param.Sort
	status := param.Status
	// nsList := param.NamespaceList

	var b strings.Builder
	b.WriteString("select *, count(*) over() as full_count from audit where 1=1 ")

	// if startTime != "" && endTime != "" {
	// 	b.WriteString("and stagetimestamp between to_timestamp(")
	// 	b.WriteString(startTime)
	// 	b.WriteString(") and to_timestamp(")
	// 	b.WriteString(endTime)
	// 	b.WriteString(")")
	// }

	////////////////////////////////////////////////////////////////////////////////////////
	// b.WriteString(") as sub where 1=1 ")

	// if search != "" {
	// 	parsedSearch := strings.Replace(search, "_", "\\_", -1)
	// 	b.WriteString("and username like '")
	// 	b.WriteString(parsedSearch)
	// 	b.WriteString("%' ")
	// }

	if startTime != "" && endTime != "" {
		b.WriteString("and stagetimestamp between to_timestamp(")
		b.WriteString(startTime)
		b.WriteString(") and to_timestamp(")
		b.WriteString(endTime)
		b.WriteString(") ")
	}

	if namespace != "" {
		b.WriteString("and namespace = '")
		b.WriteString(namespace)
		b.WriteString("' ")
	}

	if resource != "" {
		b.WriteString("and resource = '")
		b.WriteString(resource)
		b.WriteString("' ")
	}

	if status != "" {
		b.WriteString("and status = '")
		b.WriteString(status)
		b.WriteString("' ")
	}

	if verb != "" {
		b.WriteString("and verb = '")
		b.WriteString(verb)
		b.WriteString("' ")
	}

	if code != "" {
		codeNum, _ := strconv.ParseInt(code, 10, 32)
		lowerBound := (codeNum / 100) * 100
		upperBound := lowerBound + 99
		b.WriteString("and code between '")
		b.WriteString(fmt.Sprintf("%v", lowerBound))
		b.WriteString("' and '")
		b.WriteString(fmt.Sprintf("%v '", upperBound))
	}

	if sort != nil && len(sort) > 0 {
		b.WriteString("order by ")
		for _, s := range sort {
			order := " asc, "
			if string(s[0]) == "-" {
				order = " desc, "
				s = s[1:]
			}
			b.WriteString(s)
			b.WriteString(order)
		}
		b.WriteString("stagetimestamp desc ")
	} else {
		b.WriteString("order by stagetimestamp desc ")
	}

	if limit != "" {
		b.WriteString("limit ")
		b.WriteString(limit)
		b.WriteString(" ")
	} else {
		b.WriteString("limit 100 ")
	}

	if offset != "" {
		b.WriteString("offset ")
		b.WriteString(offset)
		b.WriteString(" ")
	} else {
		b.WriteString("offset 0 ")
	}
	query := b.String()

	klog.Info("query: ", query)
	return query
}
