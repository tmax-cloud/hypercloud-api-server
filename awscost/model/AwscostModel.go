package model

type Awscost struct {
	Metrics map[string]*Metric `json: metrics`
}

type Metric struct {
	Amount float64 `json: amount`
	Unit   string  `json: unit`
}

func NewAwscost() *Awscost {
	p := Awscost{}
	p.Metrics = map[string]*Metric{}
	return &p
}
