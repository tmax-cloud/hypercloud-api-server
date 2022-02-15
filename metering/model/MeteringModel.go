package model

import "time"

type Metering struct {
	Id           string    `json:"id"`
	Namespace    string    `json:"namespace"`
	Cpu          float64   `json:"cpu"`
	Memory       uint64    `json:"memory"`
	Storage      uint64    `json:"storage"`
	Gpu          float64   `json:"gpu"`
	PublicIp     uint64    `json:"publicIp"`
	PrivateIp    uint64    `json:"privateIp"`
	TrafficIn    uint64    `json:"trafficIn"`
	TrafficOut   uint64    `json:"trafficOut"`
	MeteringTime time.Time `json:"meteringTime"`
}

type Metric struct {
	Metric map[string]string `json:"metric"`
	Value  []string          `json:"value"`
}

type MetricDataList struct {
	ResultType string   `json:"resultType"`
	Result     []Metric `json:"result"`
}

type MetricResponse struct {
	Status string         `json:"status"`
	Data   MetricDataList `json:"data"`
}
