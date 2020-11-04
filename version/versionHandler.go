package version

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"reflect"
	"time"

	yaml "gopkg.in/yaml.v2"

	"hypercloud-api-server/util"
	k8sApiCaller "hypercloud-api-server/util/Caller"

	"net/http"
	"strings"

	"k8s.io/klog"
)

// Get handles ~/version get method
func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET /version")
	var conf Config

	// 1. READ CONFIG FILE
	// If it runs on POD, path should be same with
	// what declared on volume in yaml file.
	yamlFile, err := ioutil.ReadFile("/config/module.config")
	if err != nil {
		klog.Errorln(err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		klog.Errorln(err)
	}
	configSize := len(conf.Modules)
	result := make([]Module, configSize)

	// Main algorithm
	for idx, mod := range conf.Modules {
		klog.Infoln("Module Name = ", mod.Name)
		result[idx].Name = mod.Name
		// 2. GET STATUS
		var labels string
		for i, label := range mod.Selector.MatchLabels.StatusLabel {
			if i == 0 {
				labels = label
			} else {
				labels += ", " + label
			}
		}
		podList, exist := k8sApiCaller.GetPodListByLabel(labels, mod.Namespace)

		ps := NewPodStatus()

		if !exist {
			klog.Errorln("cannot found pods using given label")
			result[idx].Status = "Not found"
		} else if mod.ReadinessProbe.Exec.Command != nil {
			// by exec command
			for j := range podList.Items {
				stdout, stderr, err := k8sApiCaller.ExecCommand(podList.Items[j], mod.ReadinessProbe.Exec.Command, mod.ReadinessProbe.Exec.Container)
				output := stderr + stdout

				if err != nil {
					klog.Errorln(mod.Name, " exec command error : ", err)
					ps.Data["Command error"]++
				} else {
					ps.Data[output]++
				}
			}
			result[idx].Status = AppendStatusResult(*ps)
		} else if mod.ReadinessProbe.HTTPGet.Path != "" {
			// by HTTP
			// This code can only work on POD
			var url string
			if mod.ReadinessProbe.HTTPGet.Scheme == "" || strings.EqualFold(mod.ReadinessProbe.HTTPGet.Scheme, "http") {
				url = "http://"
			} else if strings.EqualFold(mod.ReadinessProbe.HTTPGet.Scheme, "https") {
				url = "https://"
			}
			url += mod.ReadinessProbe.HTTPGet.ServiceName + "." + mod.Namespace + ":" + mod.ReadinessProbe.HTTPGet.Port + mod.ReadinessProbe.HTTPGet.Path
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // ignore certificate
			response, err := http.Get(url)
			if err != nil {
				klog.Errorln(err)
				result[idx].Status = "HTTP error"
			} else {
				if response.StatusCode >= 200 && response.StatusCode < 300 {
					result[idx].Status = "Ready"
				} else {
					result[idx].Status = "Not ready"
				}
			}
		} else if mod.ReadinessProbe.TCPSocket.Port != "" {
			// by Port
			port := mod.ReadinessProbe.TCPSocket.Port
			timeout := time.Second
			for j := range podList.Items {
				host := podList.Items[j].Status.PodIP

				conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
				defer conn.Close()
				if err != nil {
					klog.Errorln(err)
					ps.Data["Not ready"]++
				} else {
					ps.Data["Ready"]++
				}
			}
			result[idx].Status = AppendStatusResult(*ps)
		} else {
			// by Status.Phase
			for j := range podList.Items {
				ps.Data[string(podList.Items[j].Status.Phase)]++
			}
			result[idx].Status = AppendStatusResult(*ps)
		}

		// 3. GET VERSION
		if !(reflect.DeepEqual(mod.Selector.MatchLabels.StatusLabel, mod.Selector.MatchLabels.VersionLabel)) {
			labels = ""
			for i, label := range mod.Selector.MatchLabels.VersionLabel {
				if i == 0 {
					labels = label
				} else {
					labels += ", " + label
				}
			}
			podList, exist = k8sApiCaller.GetPodListByLabel(labels, mod.Namespace)
		}

		if !exist {
			klog.Errorln("cannot found pods using given label")
			result[idx].Version = "not found"
		} else if mod.VersionProbe.Exec.Command != nil {
			// by exec command
			stdout, stderr, err := k8sApiCaller.ExecCommand(podList.Items[0], mod.VersionProbe.Exec.Command, mod.VersionProbe.Exec.Container)
			output := stderr + stdout
			if err != nil {
				klog.Errorln(mod.Name, " exec command error : ", err)
				result[idx].Version = "command error"
			} else {
				result[idx].Version = ParsingVersion(output)
			}
		} else if podList.Items[0].Labels["version"] != "" {
			// by version label
			result[idx].Version = ParsingVersion(podList.Items[0].Labels["version"])
		} else {
			// by image tag
			result[idx].Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		}
	}

	// encode to JSON format and response
	res = util.SetResponse(res, "", result, http.StatusOK)
}
