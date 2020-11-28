package main

import (
	alert "hypercloud-api-server/alert"
	metering "hypercloud-api-server/metering"
	"hypercloud-api-server/namespace"
	"hypercloud-api-server/namespaceClaim"
	user "hypercloud-api-server/user"
	version "hypercloud-api-server/version"

	"k8s.io/klog"

	"net/http"

	"github.com/robfig/cron"
)

const (
	TEST = "<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"    <meta charset=\"UTF-8\">\r\n" +
		"    <title>HyperCloud 서비스 신청 승인 완료</title>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<div style=\"border: #c5c5c8 0.06rem solid; border-bottom: 0; width: 42.5rem; height: 53.82rem; padding: 0 1.25rem\">\r\n" +
		"    <header>\r\n" +
		"        <div style=\"margin: 0;\">\r\n" +
		"            <p style=\"font-size: 1rem; font-weight: bold; color: #333333; line-height: 3rem; letter-spacing: 0; border-bottom: #c5c5c8 0.06rem solid;\">\r\n" +
		"                HyperCloud 서비스 신청 승인 완료\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </header>\r\n" +
		"    <section>\r\n" +
		"        <figure style=\"text-align: center;\">\r\n" +
		"            <img style=\"margin: 0.94rem 0;\"\r\n" +
		"                 src=\"cid:trial-approval\">\r\n" +
		"        </figure>\r\n" +
		"        <div style=\"width: 35.70rem; margin: 0 2.75rem;\">\r\n" +
		"            <p style=\"font-size: 1.5rem; font-weight: bold; line-height: 3rem;\">\r\n" +
		"                축하합니다.\r\n" +
		"            </p>\r\n" +
		"            <p style=\"line-height: 1.38rem;\">\r\n" +
		"                고객님의 Trial 서비스 신청이 성공적으로 승인되었습니다. <br>\r\n" +
		"                지금 바로 티맥스의 소프트웨어와 검증을 거친 오픈소스 서비스를 결합한 클라우드 플랫폼, <br>\r\n" +
		"                HyperCloud를 이용해 보세요. <br>\r\n" +
		"                <br>\r\n" +
		"                네임스페이스 이름 : <span style=\"font-weight: 600;\">%%NAMESPACE_NAME%%</span> <br>\r\n" +
		"                Trial 기한 : %%TRIAL_START_TIME%% ~ %%TRIAL_END_TIME%% <br>\r\n" +
		"                <br>\r\n" +
		"                리소스 정보 <br>\r\n" +
		"                -CPU : 1 Core <br>\r\n" +
		"                -Memory : 4 GIB <br>\r\n" +
		"                -Storage : 4 GIB <br>\r\n" +
		"                <br>\r\n" +
		"<!--                <span style=\"font-weight: 600;\">승인사유</span> <br>-->\r\n" +
		"                <br>\r\n" +
		"\r\n" +
		"                감사합니다. <br>\r\n" +
		"                TmaxCloud 드림.\r\n" +
		"            </p>\r\n" +
		"            <p style=\"margin: 3rem 0;\">\r\n" +
		"                <a href=\"https://console.tmaxcloud.com\">Tmax Console 바로가기 ></a>\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </section>\r\n" +
		"</div>\r\n" +
		"<footer style=\"background-color: #3669B3; width: 45.12rem; height: 1.88rem; font-size: 0.75rem; color: #FFFFFF; display: flex;\r\n" +
		"    align-items: center; justify-content: center;\">\r\n" +
		"    <div>\r\n" +
		"        COPYRIGHT2020. TMAX A&C., LTD. ALL RIGHTS RESERVED\r\n" +
		"    </div>\r\n" +
		"</footer>\r\n" +
		"</body>\r\n" +
		"</html>"
)

func main() {
	// Metering Cron Job
	cronJob := cron.New()
	cronJob.AddFunc("0 */5 * ? * *", metering.MeteringJob)
	cronJob.Start()

	// Req multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/user", serveUser)
	mux.HandleFunc("/metering", serveMetering)
	mux.HandleFunc("/namespace", serveNamespace)
	mux.HandleFunc("/alert", serveAlert)
	mux.HandleFunc("/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/version", serveVersion)

	// HTTP Server Start
	klog.Info("Starting Hypercloud-Operator-API server...")
	if err := http.ListenAndServe(":80", mux); err != nil {
		klog.Errorf("Failed to listen and serve Hypercloud-Operator-API server: %s", err)
	}
	klog.Info("Started Hypercloud-Operator-API server")

}

func serveNamespace(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespace.Get(res, req)
	case http.MethodPut:
		namespace.Put(res, req)
	case http.MethodOptions:
		namespace.Options(res, req)
	default:
		//error
	}
}

func serveNamespaceClaim(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespaceClaim.Get(res, req)
	case http.MethodPut:
		namespaceClaim.Put(res, req)
	case http.MethodOptions:
		namespaceClaim.Options(res, req)
	default:
		//error
	}
}

func serveUser(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		user.Post(res, req)
	case http.MethodDelete:
		user.Delete(res, req)
	case http.MethodOptions:
		user.Options(res, req)
	default:
		//error
	}
}

func serveMetering(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		metering.Get(res, req)
	case http.MethodOptions:
		metering.Options(res, req)
	default:
		//error
	}
}

func serveAlert(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		alert.Post(res, req)
	default:
		//error
	}
}
func serveVersion(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		version.Get(res, req)

	default:
		//error
	}
}
