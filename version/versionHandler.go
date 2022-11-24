package version

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	versionModel "github.com/tmax-cloud/hypercloud-api-server/version/model"

	"net/http"
	"strings"

	"k8s.io/klog"
)

var (
	result     []versionModel.Module
	conf       versionModel.Config
	configSize int
)

func init() {
	// 1. READ CONFIG FILE
	// File path should be same with what declared on volume mount in yaml file.
	yamlFile, err := ioutil.ReadFile("/go/src/version/version.config")
	if err != nil {
		klog.V(1).Infoln(err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		klog.V(1).Infoln(err)
	}
	configSize = len(conf.Modules)
	result = make([]versionModel.Module, configSize)

	for _, mod := range conf.Modules {
		if mod.Name == "HyperAuth" {
			idx := strings.Index(mod.ReadinessProbe.HTTPGet.Path, "/auth/realms")
			if idx < 0 {
				klog.V(1).Infoln("Failed to get HyperAuth URL. Please check Configmap [version-config].")
				break
			}
			k8sApiCaller.HYPERAUTH_URL = mod.ReadinessProbe.HTTPGet.Path[:idx]
			k8sApiCaller.HYPERAUTH_REALM_PREFIX = mod.ReadinessProbe.HTTPGet.Path[idx:]
			break
		}
	}
}

// Get handles ~/version get method
func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /version")
	var wg sync.WaitGroup
	wg.Add(configSize)

	for idx, mod := range conf.Modules {
		go func(idx int, mod versionModel.ModuleInfo) { // GoRoutine
			defer wg.Done()
			// klog.V(3).Infoln("Module Name = ", mod.Name)
			result[idx].Name = mod.Name

			// If the moudle is HyperAuth,
			// Ask to hyperauth using given URL
			if mod.Name == "HyperAuth" {
				hyperauth_status, hyperauth_version := AskToHyperAuth(mod)
				result[idx].Status = hyperauth_status
				result[idx].Version = hyperauth_version
				return
			}

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

			ps := versionModel.NewPodStatus()

			if exist {
				if mod.ReadinessProbe.Exec.Command != nil {
					// by exec command
					for j := range podList.Items {
						stdout, stderr, err := k8sApiCaller.ExecCommand(podList.Items[j], mod.ReadinessProbe.Exec.Command, mod.ReadinessProbe.Exec.Container)
						output := stderr + stdout

						if err != nil {
							klog.V(1).Infoln(mod.Name, " exec command error : ", err)
							break
						} else {
							ps.Data[output]++
						}
					}
				} else if mod.ReadinessProbe.HTTPGet.Path != "" {
					// by HTTP
					for j := range podList.Items {
						var url string
						if mod.ReadinessProbe.HTTPGet.Scheme == "" || strings.EqualFold(mod.ReadinessProbe.HTTPGet.Scheme, "http") {
							url = "http://"
						} else if strings.EqualFold(mod.ReadinessProbe.HTTPGet.Scheme, "https") {
							url = "https://"
						}
						url += podList.Items[j].Status.PodIP + ":" + mod.ReadinessProbe.HTTPGet.Port + mod.ReadinessProbe.HTTPGet.Path
						http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // ignore certificate

						client := http.Client{
							Timeout: 15 * time.Second,
						}
						response, err := client.Get(url)
						if err != nil {
							klog.V(1).Infoln(mod.Name, " HTTP Error : ", err)
							break
						} else if response.StatusCode >= 200 && response.StatusCode < 400 {
							ps.Data["Running"]++
						}
						defer response.Body.Close()
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
							klog.V(1).Infoln(mod.Name, " TCP Error : ", err)
							break
						} else {
							ps.Data["Running"]++
						}
					}
				} else {
					// by Status.Phase
					for j := range podList.Items {
						if string(podList.Items[j].Status.Phase) == "Running" {
							ps.Data["Running"]++
						}
					}
				}
			}

			if !exist {
				//klog.V(1).Infoln(mod.Name, " cannot found pods using given label : ", labels)
				result[idx].Status = "Not Installed"
			} else if ps.Data["Running"] == len(podList.Items) {
				// if every pod is 'Running', the module is normal
				result[idx].Status = "Normal"
			} else {
				result[idx].Status = "Abnormal"
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
				//klog.V(1).Infoln(mod.Name, " cannot found pods using given label : ", labels)
				result[idx].Version = "Not Installed"
			} else if mod.VersionProbe.Exec.Command != nil {
				// by exec command
				stdout, stderr, err := k8sApiCaller.ExecCommand(podList.Items[0], mod.VersionProbe.Exec.Command, mod.VersionProbe.Container)
				output := stderr + stdout
				if err != nil {
					klog.V(1).Infoln(mod.Name, " exec command error : ", err)
				} else {
					result[idx].Version = ParsingVersion(output)
				}
			} else if podList.Items[0].Labels["version"] != "" {
				// by version label
				result[idx].Version = podList.Items[0].Labels["version"]
			} else {
				// by image tag
				if mod.VersionProbe.Container == "" {
					result[idx].Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
				} else {
					for j := range podList.Items[0].Spec.Containers {
						if podList.Items[0].Spec.Containers[j].Name == mod.VersionProbe.Container {
							result[idx].Version = ParsingVersion(podList.Items[0].Spec.Containers[j].Image)
							break
						}
					}
				}
			}
			// klog.V(3).Infoln(mod.Name + " status = " + result[idx].Status)
			// klog.V(3).Infoln(mod.Name + " version = " + result[idx].Version)
		}(idx, mod)
	}
	wg.Wait()

	// encode to JSON format and response
	util.SetResponse(res, "", result, http.StatusOK)
	return
}

// AppendStatusResult connects status of each pod to one string.
func AppendStatusResult(p versionModel.PodStatus) string {
	temp := ""
	for s, num := range p.Data {
		temp += s + "(" + strconv.Itoa(num) + "),"
	}
	return strings.TrimRight(temp, ",")
}

// ParsingVersion parses version using regular expression
func ParsingVersion(str string) string {
	isLatest, err := regexp.MatchString("latest", str)
	if err != nil {
		klog.V(1).Infoln(err)
	} else if isLatest {
		return "latest"
	}

	r, err := regexp.Compile(":[a-z]*[A-Z]*[0-9]*(\\.[0-9]+)*")
	if err != nil {
		klog.V(1).Infoln(err)
	}

	matches := r.FindAllString(str, -1)
	version := matches[len(matches)-1] // ignore private docker repo, which is ":5000"
	return strings.TrimLeft(version, ":")
}

func AskToHyperAuth(mod versionModel.ModuleInfo) (string, string) {
	var hyperauth_version string
	var hyperauth_status string

	url := mod.ReadinessProbe.HTTPGet.Path + "/version"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // ignore certificate

	client := http.Client{
		Timeout: 15 * time.Second,
	}
	response, err := client.Get(url)
	if err != nil {
		hyperauth_status = "Abnormal"
		klog.V(1).Infoln(mod.Name, " HTTPS Error : ", err)
		return "", ""
	} else if response.StatusCode >= 200 && response.StatusCode < 300 {
		hyperauth_status = "Normal"
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			klog.V(1).Infoln(err)
		} else {
			bodyString := string(bodyBytes)
			hyperauth_version = bodyString
		}
	} else {
		hyperauth_status = "Abnormal"
	}
	defer response.Body.Close()

	return hyperauth_status, hyperauth_version
}
