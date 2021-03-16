package model

// ModuleInfo is a strcut for storing configMap file.
// It must support all possible cases described in configMap.
type ModuleInfo struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Selector  struct {
		MatchLabels struct {
			StatusLabel  []string `yaml:"statusLabel"`
			VersionLabel []string `yaml:"versionLabel"`
		} `yaml:"matchLabels"`
	} `yaml:"selector"`
	ReadinessProbe struct {
		Exec struct {
			Command   []string `yaml:"command"`
			Container string   `yaml:"container"`
		} `yaml:"exec"`
		HTTPGet struct {
			Path   string `yaml:"path"`
			Port   string `yaml:"port"`
			Scheme string `yaml:"scheme"`
		} `yaml:"httpGet"`
		TCPSocket struct {
			Port string `yaml:"port"`
		} `yaml:"tcpSocket"`
	} `yaml:"readinessProbe"`
	VersionProbe struct {
		Container string `yaml:"container"`
		Exec      struct {
			Command []string `yaml:"command"`
		} `yaml:"exec"`
	} `yaml:"versionProbe"`
}

// Config struct is array of ModuleInfo.
type Config struct {
	Modules []ModuleInfo `yaml:"modules"`
}

// Module struct is for storing result and returning to client.
type Module struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

// PodStatus struct is for temporarily storing status of each pod.
type PodStatus struct {
	Data map[string]int
}

// NewPodStatus initializes PodStatus struct.
func NewPodStatus() *PodStatus {
	p := PodStatus{}
	p.Data = map[string]int{}
	return &p
}
