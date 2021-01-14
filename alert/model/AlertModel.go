package alert

type Alertaudit struct {
	Resource  string `json:"resource"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Alert     string `json:"alert"`
	Namespace string `json:"namespace"`
	Message   string `json:"message"`
	Instance  string `json:"instance"`
}
