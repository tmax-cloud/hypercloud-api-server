package util

const(
	HEADER_PARAMETER_AUTHORIZATION = "authorization";
	QUERY_PARAMETER_OFFSET = "offset";
	QUERY_PARAMETER_LIMIT = "limit";
	QUERY_PARAMETER_NAMESPACE = "namespace";
	QUERY_PARAMETER_USER_ID = "userId";
	QUERY_PARAMETER_TIMEUNIT = "timeUnit";
	QUERY_PARAMETER_STARTTIME = "startTime";
	QUERY_PARAMETER_ENDTIME = "endTime";
	QUERY_PARAMETER_SORT = "sort";
	QUERY_PARAMETER_RESOURCE = "resource";
	QUERY_PARAMETER_CODE = "code";
	QUERY_PARAMETER_CONTINUE = "continue";
	QUERY_PARAMETER_LABEL_SELECTOR = "labelSelector";
	QUERY_PARAMETER_PERIOD = "period";

	//HyperAuth
	//HYPERAUTH_URL = "http://hyperauth.hyperauth"
	HYPERAUTH_URL = "http://192.168.6.151"
	HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN = "auth/realms/master/protocol/openid-connect/token";
	HYPERAUTH_SERVICE_NAME_USER_DETAIL_WITHOUT_TOKEN = "auth/realms/tmax/user/";
)
