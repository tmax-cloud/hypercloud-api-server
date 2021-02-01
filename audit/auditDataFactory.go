package audit

import (
	"database/sql"
	"fmt"

	pq "github.com/lib/pq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

const (
	DB_USER     = "audit"
	DB_PASSWORD = "tmax"
	DB_NAME     = "audit"
	HOSTNAME    = "postgres-service.hypercloud5-system.svc"
	PORT        = 5432
)

var pg_con_info string

func init() {
	pg_con_info = fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		PORT, HOSTNAME, DB_USER, DB_PASSWORD, DB_NAME)
}

func insert(items []audit.Event) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		klog.Error(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("audit", "id", "username", "useragent", "namespace", "apigroup", "apiversion", "resource", "name",
		"stage", "stagetimestamp", "verb", "code", "status", "reason", "message"))
	if err != nil {
		klog.Error(err)
	}

	for _, event := range items {
		_, err = stmt.Exec(event.AuditID,
			event.User.Username,
			event.UserAgent,
			event.ObjectRef.Namespace,
			event.ObjectRef.APIGroup,
			event.ObjectRef.APIVersion,
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

	err = txn.Commit()
	if err != nil {
		klog.Error(err)
	}

	if count, err := res.RowsAffected(); err != nil {
		klog.Error(err)
	} else {
		klog.Info("Affected rows: ", count)
	}
}

func get(query string) (audit.EventList, int64) {
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
		event := audit.Event{
			ObjectRef:      &audit.ObjectReference{},
			ResponseStatus: &metav1.Status{},
		}
		err := rows.Scan(
			&event.AuditID,
			&event.User.Username,
			&event.UserAgent,
			&event.ObjectRef.Namespace,
			&event.ObjectRef.APIGroup,
			&event.ObjectRef.APIVersion,
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
		eventList.Items = append(eventList.Items, event)
	}
	eventList.Kind = "EventList"
	eventList.APIVersion = "audit.k8s.io/v1"

	return eventList, row_count
}
