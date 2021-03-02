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

	//HyperAuth
	//HYPERAUTH_URL = "http://hyperauth.hyperauth"
	HYPERAUTH_URL                                    = "http://192.168.6.151"
	HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN            = "auth/realms/master/protocol/openid-connect/token"
	HYPERAUTH_SERVICE_NAME_USER_DETAIL_WITHOUT_TOKEN = "auth/realms/tmax/user/"

	HYPERCLOUD4_NAMESPACE       = "hypercloud4-system"
	HYPERCLOUD4_CLAIM_API_GROUP = "claim.tmax.io"

	CLAIM_API_GROUP             = "claim.tmax.io"
	CLAIM_API_Kind              = "clusterclaims"
	CLAIM_API_GROUP_VERSION     = "claim.tmax.io/v1alpha1"
	CLUSTER_API_GROUP           = "cluster.tmax.io"
	CLUSTER_API_Kind            = "clustermanagers"
	CLUSTER_API_GROUP_VERSION   = "cluster.tmax.io/v1alpha1"
	HYPERCLOUD_SYSTEM_NAMESPACE = "hypercloud5-system"
)
