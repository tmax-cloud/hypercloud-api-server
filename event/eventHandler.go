package event

import (
	"net/http"
	"strconv"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	eventDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/event"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /event")
	queryParams := req.URL.Query()
	offset := queryParams.Get(util.QUERY_PARAMETER_OFFSET)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	namespace := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	startTime := queryParams.Get(util.QUERY_PARAMETER_STARTTIME)
	endTime := queryParams.Get(util.QUERY_PARAMETER_ENDTIME)
	sorts := queryParams[util.QUERY_PARAMETER_SORT]
	kind := queryParams.Get(util.QUERY_PARAMETER_KIND)
	typ := queryParams.Get(util.QUERY_PARAMETER_TYPE)
	host := queryParams.Get(util.QUERY_PARAMETER_HOST)

	query := "select * from event"
	query = makeTimeRange(startTime, endTime, query)

	if namespace != "" {
		query += " and namespace='" + namespace + "'"
	}
	if kind != "" {
		query += " and kind='" + kind + "'"
	}
	if typ != "" {
		query += " and type='" + typ + "'"
	}
	if host != "" {
		query += " and host='" + host + "'"
	}

	if len(sorts) > 0 {
		query += " order by "
		for _, sort := range sorts {
			order := " asc, "
			if sort[0] == '-' {
				order = " desc, "
				sort = sort[1:]
			}
			query += sort
			query += order
		}
		query += "last_timestamp desc"
	} else {
		query += " order by last_timestamp desc"
	}

	if limit != "" {
		query += " limit " + limit
	} else {
		query += " limit 100"
	}

	if offset != "" {
		query += " offset " + offset
	} else {
		query += " offset 0"
	}

	if eventDataList, err := eventDataFactory.GetEventDataFromDB(query); err != nil {
		util.SetResponse(res, "", err, http.StatusInternalServerError)
	} else {
		util.SetResponse(res, "", convertEventv1toCorev1(eventDataList), http.StatusOK)
	}
}

func makeTimeRange(startTime string, endTime string, query string) string {
	var start int64
	start = 0
	end := time.Now().Unix()

	if startTime != "" {
		start, _ = strconv.ParseInt(startTime, 10, 64)
	}
	if endTime != "" {
		end, _ = strconv.ParseInt(endTime, 10, 64)
	}
	startTime = time.Unix(start, 0).Format("2006-01-02 15:04:05")
	endTime = time.Unix(end, 0).Format("2006-01-02 15:04:05")

	query += " where ('" + startTime + "' between first_timestamp and last_timestamp) or ('" + startTime + "' <= first_timestamp and '" + endTime + "' >= first_timestamp)"

	return query
}

func convertEventv1toCorev1(evs []eventv1.Event) []corev1.Event {
	var cvs []corev1.Event

	for _, ev := range evs {
		var cv corev1.Event
		cv.InvolvedObject.Namespace = ev.Regarding.Namespace
		cv.InvolvedObject.Kind = ev.Regarding.Kind
		cv.InvolvedObject.Name = ev.Regarding.Name
		cv.InvolvedObject.UID = ev.Regarding.UID
		cv.InvolvedObject.APIVersion = ev.Regarding.APIVersion
		cv.InvolvedObject.FieldPath = ev.Regarding.FieldPath
		cv.Action = ev.Action
		cv.Reason = ev.Reason
		cv.Message = ev.Note
		cv.ReportingController = ev.ReportingController
		cv.ReportingInstance = ev.ReportingInstance
		cv.Source.Host = ev.DeprecatedSource.Host
		cv.Count = ev.DeprecatedCount
		cv.Type = ev.Type
		cv.FirstTimestamp = ev.DeprecatedFirstTimestamp
		cv.LastTimestamp = ev.DeprecatedLastTimestamp
		cvs = append(cvs, cv)
	}

	return cvs
}
