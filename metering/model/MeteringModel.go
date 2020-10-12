package model

import "time"

type Metering struct {
	Id string				`json:"id"`
	Namespace string		`json:"namespace"`
	Cpu float64				`json:"cpu"`
	Memory float64			`json:"memory"`
	Storage float64			`json:"storage"`
	Gpu float64				`json:"gpu"`
	PublicIp int64			`json:"publicIp"`
	PrivateIp int64			`json:"privateIp`
	MeteringTime time.Time	`json:"meteringTime"`
}

type Metric struct {
	Metric map[string]string	`json:"metric"`
	Value []string				`json:"value"`
}

type MetricDataList struct {
	ResultType string		`json:"resultType"`
	Result []Metric			`json:"result"`
}

type MetricResponse struct {
	Status string 			`json:"status"`
	Data MetricDataList		`json:"data"`
}
