package audit

import (
	"database/sql"
	"fmt"

	pq "github.com/lib/pq"
	//hypercloudAudit "github.com/tmax-cloud/hypercloud-api-server/audit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
	"strings"
)

type ClaimListResponse struct {
	Claims    []Claim `json:"claimList"`
	RowsCount int64   `json:"rowsCount"`
}

type Claim struct {
	Id        string `json:"id"`
	Namespace string `json:"namespace"`
	Body      string `json:"body"`
}

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "tmax"
	DB_NAME     = "postgres"
	HOSTNAME    = "postgres-service.hypercloud5-system.svc"
	PORT        = 5432
)

var pg_con_info string

func init() {
	pg_con_info = fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		PORT, HOSTNAME, DB_USER, DB_PASSWORD, DB_NAME)
}

func NewNullString(s string) sql.NullString {
	if s == "null" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func Insert(items []audit.Event) {
	defer func() {
		if v := recover(); v != nil {
			klog.Errorln("capture a panic:", v)
		}
	}()

	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		klog.Error(err)
	}

	// Insert Metadata
	stmt, err := txn.Prepare(pq.CopyIn("audit", "id", "username", "useragent", "namespace", "apigroup", "apiversion", "resource", "name",
		"stage", "stagetimestamp", "verb", "code", "status", "reason", "message"))
	if err != nil {
		klog.Error(err)
	}

	for _, event := range items {
		_, err = stmt.Exec(event.AuditID,
			event.User.Username,
			event.UserAgent,
			NewNullString(event.ObjectRef.Namespace),
			NewNullString(event.ObjectRef.APIGroup),
			NewNullString(event.ObjectRef.APIVersion),
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.Stage,
			event.StageTimestamp.Time,
			event.Verb,
			event.ResponseStatus.Code,
			event.ResponseStatus.Status,
			event.ResponseStatus.Reason,
			event.ResponseStatus.Message)

		if err != nil {
			klog.Error(err)
		}
	}
	res, err := stmt.Exec()
	if err != nil {
		klog.Error(err)
	}

	err = stmt.Close()
	if err != nil {
		klog.Error(err)
	}

	// Insert Request Body
	stmt_body, err := txn.Prepare(pq.CopyIn("audit_body", "id", "namespace", "body"))
	if err != nil {
		klog.Error(err)
	}

	for _, event := range items {
		_, err = stmt_body.Exec(event.AuditID,
			NewNullString(event.ObjectRef.Namespace),
			event.RequestObject.Raw)

		if err != nil {
			klog.Error(err)
		}
	}
	res_body, err := stmt_body.Exec()
	if err != nil {
		klog.Error(err)
	}

	err = stmt_body.Close()
	if err != nil {
		klog.Error(err)
	}

	// Commit
	err = txn.Commit()
	if err != nil {
		klog.Error(err)
	}

	if count, err := res.RowsAffected(); err != nil {
		klog.Error(err)
	} else if _, err := res_body.RowsAffected(); err != nil {
		klog.Error(err)
	} else {
		klog.Info("Affected rows: ", count)
	}
}

func Get(query string) (audit.EventList, int64) {
	defer func() {
		if v := recover(); v != nil {
			klog.Errorln("capture a panic:", v)
		}
	}()

	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
	}
	defer rows.Close()

	eventList := audit.EventList{}
	var row_count int64
	for rows.Next() {
		var temp_namespace, temp_apigroup, temp_apiversion sql.NullString
		event := audit.Event{
			ObjectRef:      &audit.ObjectReference{},
			ResponseStatus: &metav1.Status{},
		}
		err := rows.Scan(
			&event.AuditID,
			&event.User.Username,
			&event.UserAgent,
			&temp_namespace,  //&event.ObjectRef.Namespace,
			&temp_apigroup,   //&event.ObjectRef.APIGroup,
			&temp_apiversion, //&event.ObjectRef.APIVersion,
			&event.ObjectRef.Resource,
			&event.ObjectRef.Name,
			&event.Stage,
			&event.StageTimestamp.Time,
			&event.Verb,
			&event.ResponseStatus.Code,
			&event.ResponseStatus.Status,
			&event.ResponseStatus.Reason,
			&event.ResponseStatus.Message,
			&row_count)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}
		if temp_namespace.Valid {
			event.ObjectRef.Namespace = temp_namespace.String
		} else {
			event.ObjectRef.Namespace = ""
		}

		if temp_apigroup.Valid {
			event.ObjectRef.APIGroup = temp_apigroup.String
		} else {
			event.ObjectRef.APIGroup = ""
		}

		if temp_apiversion.Valid {
			event.ObjectRef.APIVersion = temp_apiversion.String
		} else {
			event.ObjectRef.APIVersion = ""
		}
		eventList.Items = append(eventList.Items, event)
	}
	eventList.Kind = "EventList"
	eventList.APIVersion = "audit.k8s.io/v1"

	return eventList, row_count
}

func GetByJson(jquery string) ClaimListResponse {
	defer func() {
		if v := recover(); v != nil {
			klog.Errorln("capture a panic:", v)
		}
	}()

	klog.Infoln("query =", jquery)

	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	rows, err := db.Query(jquery)
	if err != nil {
		klog.Error(err)
	}
	defer rows.Close()

	var claimList ClaimListResponse
	for rows.Next() {
		var claim Claim
		var namespace sql.NullString

		err := rows.Scan(
			&claim.Id,
			&namespace,
			&claim.Body)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}
		if namespace.Valid {
			claim.Namespace = namespace.String
		} else {
			claim.Namespace = ""
		}
		claim.Body = strings.Replace(claim.Body, "\\", "", -1)

		claimList.Claims = append(claimList.Claims, claim)
	}
	return claimList
}

func GetMemberList(query string) ([]string, int64) {
	defer func() {
		if v := recover(); v != nil {
			klog.Errorln("capture a panic:", v)
		}
	}()

	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
	}
	defer rows.Close()

	// var memberList []string
	memberList := []string{}
	var row_count int64

	for rows.Next() {
		var member string
		err := rows.Scan(
			&member,
			&row_count)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}
		memberList = append(memberList, member)
	}

	return memberList, row_count
}
