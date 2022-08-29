package metering

import (
	"context"
	"net/http"
	"strconv"
	"time"

	meteringModel "github.com/tmax-cloud/hypercloud-api-server/metering/model"
	"github.com/tmax-cloud/hypercloud-api-server/util"
	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"

	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /metering")
	queryParams := req.URL.Query()
	offset := queryParams.Get(util.QUERY_PARAMETER_OFFSET)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	namespace := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	timeUnit := queryParams.Get(util.QUERY_PARAMETER_TIMEUNIT)
	startTime := queryParams.Get(util.QUERY_PARAMETER_STARTTIME)
	endTime := queryParams.Get(util.QUERY_PARAMETER_ENDTIME)
	sorts := queryParams[util.QUERY_PARAMETER_SORT]

	if timeUnit == "" || !(timeUnit == "hour" || timeUnit == "day" || timeUnit == "month" || timeUnit == "year") {
		timeUnit = "day" // default time unit
	}
	var query string
	query = makeTimeRange(timeUnit, startTime, endTime, query)

	if namespace != "" {
		query += " and namespace='" + namespace + "'"
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
		query += "metering_time desc"
	} else {
		query += " order by metering_time desc"
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

	if meteringDataList, err := getMeteringDataFromDB(query); err != nil {
		util.SetResponse(res, "", err, http.StatusInternalServerError)
	} else {
		util.SetResponse(res, "", meteringDataList, http.StatusOK)
	}
}

func getMeteringDataFromDB(query string) ([]meteringModel.Metering, error) {
	klog.V(3).Infoln("=== query ===")
	klog.V(3).Infoln(query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
		return nil, err
	}
	defer rows.Close()

	var meteringList []meteringModel.Metering
	var status string
	for rows.Next() {
		var meteringData meteringModel.Metering
		err := rows.Scan(
			&meteringData.Id,
			&meteringData.Namespace,
			&meteringData.Cpu,
			&meteringData.Memory,
			&meteringData.Storage,
			&meteringData.Gpu,
			&meteringData.PublicIp,
			&meteringData.PrivateIp,
			&meteringData.TrafficIn,
			&meteringData.TrafficOut,
			&meteringData.MeteringTime,
			&status)
		if err != nil {
			klog.V(1).Info(err)
			return nil, err
		}
		meteringList = append(meteringList, meteringData)
	}
	return meteringList, nil
}

func makeTimeRange(timeUnit string, startTime string, endTime string, query string) string {
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

	switch timeUnit {
	case "hour":
		query += "select * from metering_hour"
	case "day":
		query += "select * from metering_day"
	case "month":
		query += "select * from metering_month"
	case "year":
		query += "select * from metering_year"
	}
	query += " where metering_time between '" + startTime + "' and '" + endTime + "'"
	return query
}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** OPTIONS/metering")
	out := "**** OPTIONS/metering"
	util.SetResponse(res, out, nil, http.StatusOK)
}
