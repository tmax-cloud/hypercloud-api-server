package alert

import "time"

type Alert struct {
	Id       string    `json:"id"`
	Resource string    `json:"resource"`
	Name     string    `json:"cpu"`
	Memory   float64   `json:"memory"`
	Cpu      float64   `json:"cpu"`
	Status   string    `json:"status"`
	Alert    string    `json:"alert"`
	Time     time.Time `json:"meteringTime"`
}
