package metering

import (
	"database/sql"
	"encoding/json"
	meteringModel "hypercloud-api-server/metering/model"
	"hypercloud-api-server/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"k8s.io/klog"
)

const (
	DB_DRIVER = "mysql"
	//DB_URI = "root:tmax@tcp(mysql-service.hypercloud4-system.svc:3306)/metering?parseTime=true"
	DB_URI                = "tmax:tmax@tcp(192.168.6.116:3306)/metering?parseTime=true"
	METERING_INSERT_QUERY = "insert into metering.metering (id,namespace,cpu,memory,storage,gpu,public_ip,private_ip,metering_time,status) " +
		"values (?,?,truncate(?,2),?,?,truncate(?,2),?,?,?,?)"
	METERING_DELETE_QUERY = "truncate metering.metering"

	METERING_HOUR_INSERT_QUERY = "insert into metering.metering_hour values (?,?,?,?,?,?,?,?,?,?)"
	METERING_HOUR_SELECT_QUERY = "select id, namespace, truncate(sum(cpu)/count(*),2) as cpu, truncate(sum(memory)/count(*),0) as memory," +
		"truncate(sum(storage)/count(*),0) as storage, truncate(sum(gpu)/count(*),2) as gpu," +
		"truncate(sum(public_ip)/count(*),0) as public_ip, truncate(sum(private_ip)/count(*),0) as private_ip, " +
		"metering_time, status from metering.metering group by hour(metering_time), namespace"
	METERING_HOUR_UPDATE_QUERY = "update metering.metering_hour set status = 'Merged' where status = 'Success'"
	METERING_HOUR_DELETE_QUERY = "delete from metering.metering_hour where status = 'Merged'"

	METERING_DAY_INSERT_QUERY = "insert into metering.metering_day values (?,?,?,?,?,?,?,?,?,?)"
	METERING_DAY_SELECT_QUERY = "select id, namespace, truncate(sum(cpu)/count(*),2) as cpu, truncate(sum(memory)/count(*),0) as memory, " +
		"truncate(sum(storage)/count(*),0) as storage, truncate(sum(gpu)/count(*),2) as gpu, " +
		"truncate(sum(public_ip)/count(*),0) as public_ip, truncate(sum(private_ip)/count(*),0) as private_ip," +
		"metering_time, status from metering.metering_hour where status = 'Success' " +
		"group by day(metering_time), namespace"
	METERING_DAY_UPDATE_QUERY = "update metering.metering_day set status = 'Merged' where status = 'Success'"
	METERING_DAY_DELETE_QUERY = "delete from metering.metering_day where status = 'Merged'"

	METERING_MONTH_INSERT_QUERY = "insert into metering.metering_month values (?,?,?,?,?,?,?,?,?,?)"
	METERING_MONTH_SELECT_QUERY = "select id, namespace, truncate(sum(cpu)/count(*),2) as cpu, truncate(sum(memory)/count(*),0) as memory, " +
		"truncate(sum(storage)/count(*),0) as storage, truncate(sum(gpu)/count(*),2) as gpu, " +
		"truncate(sum(public_ip)/count(*),0) as public_ip, truncate(sum(private_ip)/count(*),0) as private_ip, " +
		"metering_time, status from metering.metering_day where status = 'Success' " +
		"group by month(metering_time), namespace"
	METERING_MONTH_UPDATE_QUERY = "update metering.metering_month set status = 'Merged' where status = 'Success'"
	METERING_MONTH_DELETE_QUERY = "delete from metering.metering_month where status = 'Merged'"

	METERING_YEAR_INSERT_QUERY = "insert into metering.metering_year values (?,?,?,?,?,?,?,?,?,?)"
	METERING_YEAR_SELECT_QUERY = "select id, namespace, truncate(sum(cpu)/count(*),2) as cpu, truncate(sum(memory)/count(*),0) as memory, " +
		"truncate(sum(storage)/count(*),0) as storage, truncate(sum(gpu)/count(*),2) as gpu, " +
		"truncate(sum(public_ip)/count(*),0) as public_ip, truncate(sum(private_ip)/count(*),0) as private_ip, " +
		"date_format(metering_time,'%Y-01-01 %H:00:00') as metering_time, status from metering.metering_month where status = 'Success' " +
		"group by year(metering_time), namespace"

	PROMETHEUS_URI                 = "http://10.101.168.154:9090/api/v1/query" //FIXME
	PROMETHEUS_GET_CPU_QUERY       = "sum(kube_pod_container_resource_requests{resource=\"cpu\"})by(namespace)"
	PROMETHEUS_GET_MEMORY_QUERY    = "sum(kube_pod_container_resource_requests{resource=\"memory\"})by(namespace)"
	PROMETHEUS_GET_STORAGE_QUERY   = "sum(kube_persistentvolumeclaim_resource_requests_storage_bytes)by(namespace)"
	PROMETHEUS_GET_PUBLIC_IP_QUERY = "count(kube_service_spec_type{type=\"LoadBalancer\"})by(namespace)"
)

var t time.Time

func MeteringJob() {
	t = time.Now()
	klog.Infoln("============= Metering Time =============")
	klog.Infoln("Current Time   : ", t.Format("2006-01-02 15:04:05"))
	klog.Infoln("minute of hour : ", t.Minute())
	klog.Infoln("hour of day    : ", t.Hour())
	klog.Infoln("day of month   : ", t.Day())
	klog.Infoln("day of year    : ", t.YearDay())

	if t.Minute() == 0 {
		// Insert into metering_hour
		insertMeteringHour()
	} else if t.Hour() == 0 {
		// Insert into metering_day
		insertMeteringDay()
	} else if t.Day() == 1 {
		// Insert into metering_month
		insertMeteringMonth()
	} else if t.YearDay() == 1 {
		// Insert into metering_year
		insertMeteringYear()
	}

	meteringData := makeMeteringMap()

	klog.Infoln("============= Metering Data =============")
	for key, value := range meteringData {
		klog.Infoln(key+"/cpu : ", value.Cpu)
		klog.Infoln(key+"/memory : ", value.Memory)
		klog.Infoln(key+"/storage : ", value.Storage)
		klog.Infoln(key+"/publicIp : ", value.PublicIp)
		klog.Infoln("-----------------------------------------")
	}
	//Insert into metering (new data)
	insertMeteringData(meteringData)

	deleteMeteringData()

}

func deleteMeteringData() {

}

func insertMeteringData(meteringData map[string]*meteringModel.Metering) {
	klog.Infoln("Insert into METERING Start!!")
	klog.Infoln("Current Time : " + t.Format("2006-01-02 15:04:00"))

	db, err := sql.Open(DB_DRIVER, DB_URI)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	for key, data := range meteringData {
		_, err := db.Exec(METERING_INSERT_QUERY,
			uuid.New(),
			key,
			data.Cpu,
			data.Memory,
			data.Storage,
			data.Gpu,
			data.PublicIp,
			data.PrivateIp,
			t.Format("2006-01-02 15:04:00"), "Success")

		if err != nil {
			klog.Error(err)
		}
	}

	klog.Infoln("Insert into METERING Success!!")
}

func makeMeteringMap() map[string]*meteringModel.Metering {
	var meteringData = make(map[string]*meteringModel.Metering)
	cpu := getMeteringData(PROMETHEUS_GET_CPU_QUERY)
	for _, metric := range cpu.Result {
		var keys []string
		for k := range meteringData {
			keys = append(keys, k)
		}
		if util.Contains(keys, metric.Metric["namespace"]) {
			meteringData[metric.Metric["namespace"]].Cpu, _ = strconv.ParseFloat(metric.Value[1], 64)
		} else {
			metering := new(meteringModel.Metering)
			metering.Namespace = metric.Metric["namespace"]
			metering.Cpu, _ = strconv.ParseFloat(metric.Value[1], 64)
			meteringData[metric.Metric["namespace"]] = metering
		}
	}

	memory := getMeteringData(PROMETHEUS_GET_MEMORY_QUERY)
	for _, metric := range memory.Result {
		var keys []string
		for k := range meteringData {
			keys = append(keys, k)
		}
		if util.Contains(keys, metric.Metric["namespace"]) {
			meteringData[metric.Metric["namespace"]].Memory, _ = strconv.ParseFloat(metric.Value[1], 64)
		} else {
			metering := new(meteringModel.Metering)
			metering.Namespace = metric.Metric["namespace"]
			metering.Memory, _ = strconv.ParseFloat(metric.Value[1], 64)
			meteringData[metric.Metric["namespace"]] = metering
		}
	}

	storage := getMeteringData(PROMETHEUS_GET_STORAGE_QUERY)
	for _, metric := range storage.Result {
		var keys []string
		for k := range meteringData {
			keys = append(keys, k)
		}
		if util.Contains(keys, metric.Metric["namespace"]) {
			meteringData[metric.Metric["namespace"]].Storage, _ = strconv.ParseFloat(metric.Value[1], 64)
		} else {
			metering := new(meteringModel.Metering)
			metering.Namespace = metric.Metric["namespace"]
			metering.Storage, _ = strconv.ParseFloat(metric.Value[1], 64)
			meteringData[metric.Metric["namespace"]] = metering
		}
	}

	publicIp := getMeteringData(PROMETHEUS_GET_PUBLIC_IP_QUERY)
	for _, metric := range publicIp.Result {
		var keys []string
		for k := range meteringData {
			keys = append(keys, k)
		}
		if util.Contains(keys, metric.Metric["namespace"]) {
			meteringData[metric.Metric["namespace"]].PublicIp, _ = strconv.ParseInt(metric.Value[1], 10, 64)
		} else {
			metering := new(meteringModel.Metering)
			metering.Namespace = metric.Metric["namespace"]
			metering.PublicIp, _ = strconv.ParseInt(metric.Value[1], 10, 64)
			meteringData[metric.Metric["namespace"]] = metering
		}
	}
	return meteringData
}

func getMeteringData(query string) meteringModel.MetricDataList {
	var metricResponse meteringModel.MetricResponse
	// Make Request Object
	req, err := http.NewRequest("GET", PROMETHEUS_URI, nil)
	if err != nil {
		panic(err)
	}

	// Add QueryParameter
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()
	//klog.Infoln("request URL  : ", req.URL.String())

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
	//klog.Infoln("Result string  : ", str)

	if err := json.Unmarshal([]byte(str), &metricResponse); err != nil {
	}
	//klog.Infoln("Result MetricResponse  : ", metricResponse)
	return metricResponse.Data
}

func insertMeteringYear() {
	klog.Infoln("Insert into METERING_YEAR Start!!")
	klog.Infoln("Current Time : " + t.Format("2006-01-02 15:04:00"))

	db, err := sql.Open(DB_DRIVER, DB_URI)
	if err != nil {
		klog.Error(err)
		return
	}
	defer db.Close()

	rows, err := db.Query(METERING_MONTH_SELECT_QUERY)
	defer rows.Close()

	if err != nil {
		klog.Error(err)
		return
	}

	var meteringData meteringModel.Metering
	var status string
	for rows.Next() {
		err := rows.Scan(
			&meteringData.Id,
			&meteringData.Namespace,
			&meteringData.Cpu,
			&meteringData.Memory,
			&meteringData.Storage,
			&meteringData.Gpu,
			&meteringData.PublicIp,
			&meteringData.PrivateIp,
			&meteringData.MeteringTime,
			&status)
		if err != nil {
			klog.Error(err)
			return
		}

		_, err = db.Exec(METERING_YEAR_INSERT_QUERY,
			uuid.New(),
			meteringData.Namespace,
			meteringData.Cpu,
			meteringData.Memory,
			meteringData.Storage,
			meteringData.Gpu,
			meteringData.PublicIp,
			meteringData.PrivateIp,
			meteringData.MeteringTime.AddDate(0, -(util.MonthToInt(meteringData.MeteringTime.Month())-1),
				-(meteringData.MeteringTime.Day()-1)).Format("2006-01-02 00:00:00"),
			//date_format(metering_time,'%Y-01-01 00:00:00'),
			status)
		if err != nil {
			klog.Error(err)
			return
		}
	}
	klog.Infoln("Insert into METERING_YEAR Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Update METERING_MONTH Past data to 'Merged' Start!!")
	_, err = db.Exec(METERING_MONTH_UPDATE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Update METERING_MONTH Past data to 'Merged' Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Delete METERING for past 1 year Start!!")
	_, err = db.Exec(METERING_MONTH_DELETE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Delete METERING for past 1 year Success!!")
}

func insertMeteringMonth() {
	klog.Infoln("Insert into METERING_MONTH Start!!")
	klog.Infoln("Current Time : " + t.Format("2006-01-02 15:04:00"))

	db, err := sql.Open(DB_DRIVER, DB_URI)
	if err != nil {
		klog.Error(err)
		return
	}
	defer db.Close()

	rows, err := db.Query(METERING_MONTH_SELECT_QUERY)
	defer rows.Close()

	if err != nil {
		klog.Error(err)
		return
	}

	var meteringData meteringModel.Metering
	var status string
	for rows.Next() {
		err := rows.Scan(
			&meteringData.Id,
			&meteringData.Namespace,
			&meteringData.Cpu,
			&meteringData.Memory,
			&meteringData.Storage,
			&meteringData.Gpu,
			&meteringData.PublicIp,
			&meteringData.PrivateIp,
			&meteringData.MeteringTime,
			&status)
		if err != nil {
			klog.Error(err)
			return
		}

		_, err = db.Exec(METERING_MONTH_INSERT_QUERY,
			uuid.New(),
			meteringData.Namespace,
			meteringData.Cpu,
			meteringData.Memory,
			meteringData.Storage,
			meteringData.Gpu,
			meteringData.PublicIp,
			meteringData.PrivateIp,
			meteringData.MeteringTime.AddDate(0, 0,
				-(meteringData.MeteringTime.Day()-1)).Format("2006-01-02 00:00:00"),
			//date_format(metering_time,'%Y-%m-01 00:00:00'),
			status)
		if err != nil {
			klog.Error(err)
			return
		}
	}
	klog.Infoln("Insert into METERING_MONTH Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Update METERING_DAY Past data to 'Merged' Start!!")
	_, err = db.Exec(METERING_DAY_UPDATE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Update METERING_DAY Past data to 'Merged' Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Delete METERING for past 1 month Start!!")
	_, err = db.Exec(METERING_DAY_DELETE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Delete METERING for past 1 month Success!!")
}

func insertMeteringDay() {
	klog.Infoln("Insert into METERING_DAY Start!!")
	klog.Infoln("Current Time : " + t.Format("2006-01-02 15:04:00"))

	db, err := sql.Open(DB_DRIVER, DB_URI)
	if err != nil {
		klog.Error(err)
		return
	}
	defer db.Close()

	rows, err := db.Query(METERING_DAY_SELECT_QUERY)
	defer rows.Close()

	if err != nil {
		klog.Error(err)
		return
	}

	var meteringData meteringModel.Metering
	var status string
	for rows.Next() {
		err := rows.Scan(
			&meteringData.Id,
			&meteringData.Namespace,
			&meteringData.Cpu,
			&meteringData.Memory,
			&meteringData.Storage,
			&meteringData.Gpu,
			&meteringData.PublicIp,
			&meteringData.PrivateIp,
			&meteringData.MeteringTime,
			&status)
		if err != nil {
			klog.Error(err)
			return
		}

		_, err = db.Exec(METERING_DAY_INSERT_QUERY,
			uuid.New(),
			meteringData.Namespace,
			meteringData.Cpu,
			meteringData.Memory,
			meteringData.Storage,
			meteringData.Gpu,
			meteringData.PublicIp,
			meteringData.PrivateIp,
			meteringData.MeteringTime.Format("2006-01-02 00:00:00"), //date_format(metering_time,'%Y-%m-%d 00:00:00')
			status)
		if err != nil {
			klog.Error(err)
			return
		}
	}
	klog.Infoln("Insert into METERING_DAY Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Update METERING_HOUR Past data to 'Merged' Start!!")
	_, err = db.Exec(METERING_HOUR_UPDATE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Update METERING_HOUR Past data to 'Merged' Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Delete METERING for past 1 day Start!!")
	_, err = db.Exec(METERING_HOUR_DELETE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Delete METERING for past 1 day Success!!")
}

func insertMeteringHour() {
	klog.Infoln("Insert into METERING_HOUR Start!!")
	klog.Infoln("Current Time : " + t.Format("2006-01-02 15:04:00"))

	db, err := sql.Open(DB_DRIVER, DB_URI)
	if err != nil {
		klog.Error(err)
		return
	}
	defer db.Close()

	rows, err := db.Query(METERING_HOUR_SELECT_QUERY)
	defer rows.Close()

	if err != nil {
		klog.Error(err)
		return
	}

	var meteringData meteringModel.Metering
	var status string
	for rows.Next() {
		err := rows.Scan(
			&meteringData.Id,
			&meteringData.Namespace,
			&meteringData.Cpu,
			&meteringData.Memory,
			&meteringData.Storage,
			&meteringData.Gpu,
			&meteringData.PublicIp,
			&meteringData.PrivateIp,
			&meteringData.MeteringTime,
			&status)
		if err != nil {
			klog.Error(err)
			return
		}

		_, err = db.Exec(METERING_HOUR_INSERT_QUERY,
			uuid.New(),
			meteringData.Namespace,
			meteringData.Cpu,
			meteringData.Memory,
			meteringData.Storage,
			meteringData.Gpu,
			meteringData.PublicIp,
			meteringData.PrivateIp,
			meteringData.MeteringTime.Format("2006-01-02 15:00:00"),
			status)
		if err != nil {
			klog.Error(err)
			return
		}
	}
	klog.Infoln("Insert into METERING_HOUR Success!!")
	klog.Infoln("--------------------------------------")
	klog.Infoln("Delete METERING for past 1 hour Start!!")
	_, err = db.Exec(METERING_DELETE_QUERY)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Infoln("Delete METERING for past 1 hour Success!!")
}
