package util

const (
	HEADER_PARAMETER_AUTHORIZATION = "authorization"
	QUERY_PARAMETER_OFFSET         = "offset"
	QUERY_PARAMETER_LIMIT          = "limit"
	QUERY_PARAMETER_NAMESPACE      = "namespace"
	QUERY_PARAMETER_USER_ID        = "userId"
	QUERY_PARAMETER_TIMEUNIT       = "timeUnit"
	QUERY_PARAMETER_STARTTIME      = "startTime"
	QUERY_PARAMETER_ENDTIME        = "endTime"
	QUERY_PARAMETER_SORT           = "sort"
	QUERY_PARAMETER_RESOURCE       = "resource"
	QUERY_PARAMETER_CODE           = "code"
	QUERY_PARAMETER_CONTINUE       = "continue"
	QUERY_PARAMETER_LABEL_SELECTOR = "labelSelector"
	QUERY_PARAMETER_PERIOD         = "period"
	QUERY_PARAMETER_NAME           = "name"
	QUERY_PARAMETER_USER_GROUP     = "userGroup"
	QUERY_PARAMETER_GRANULARITY    = "granularity"
	QUERY_PARAMETER_METRICS        = "metrics"
	QUERY_PARAMETER_DIMENSION      = "dimension"
	QUERY_PARAMETER_KEY            = "key"
	QUERY_PARAMETER_VALUE          = "value"
	QUERY_PARAMETER_API            = "api"
	QUERY_PARAMETER_ACCOUNT        = "account"
	QUERY_PARAMETER_KIND           = "kind"
	QUERY_PARAMETER_TYPE           = "type"
	QUERY_PARAMETER_HOST           = "host"

	//HyperAuth
	HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN = "/auth/realms/master/protocol/openid-connect/token"
	HYPERAUTH_SERVICE_NAME_USER_DETAIL    = "/auth/realms/tmax/user/"

	HYPERCLOUD_KUBECTL_NAMESPACE   = "hypercloud-kubectl"
	HYPERCLOUD_KUBECTL_PREFIX      = "hypercloud-kubectl-"
	HYPERCLOUD_KUBECTL_IMAGE       = "bitnami/kubectl:1.25.3"
	HYPERCLOUD_KUBECTL_LABEL_KEY   = "hypercloud"
	HYPERCLOUD_KUBECTL_LABEL_VALUE = "kubectl"

	HYPERCLOUD4_NAMESPACE       = "hypercloud4-system"
	HYPERCLOUD4_CLAIM_API_GROUP = "claim.tmax.io"

	CLAIM_API_GROUP             = "claim.tmax.io"
	CLAIM_API_Kind              = "clusterclaims"
	CLAIM_API_GROUP_VERSION     = "claim.tmax.io/v1alpha1"
	CLUSTER_API_GROUP           = "cluster.tmax.io"
	CLUSTER_API_Kind            = "clustermanagers"
	CLUSTER_API_GROUP_VERSION   = "cluster.tmax.io/v1alpha1"
	HYPERCLOUD_SYSTEM_NAMESPACE = "hypercloud5-system"

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
