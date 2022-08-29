package audit

import (
	"context"
	"database/sql"

	//hypercloudAudit "github.com/tmax-cloud/hypercloud-api-server/audit"

	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
	//"strings"
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

var pg_con_info string

const (
	AUDIT_INSERT_QUERY = "INSERT INTO audit (ID, USERNAME, USERAGENT , NAMESPACE , APIGROUP , APIVERSION , RESOURCE , NAME , STAGE , STAGETIMESTAMP , VERB, CODE , STATUS , REASON , MESSAGE ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)"
	//AUDIT_BODY_INSERT_QUERY = "INSERT INTO audit_body (ID, NAMESPACE, BODY ) VALUES ($1, $2, $3)"
)

func Insert(items []audit.Event) {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	// //create batch
	// batch := &pgx.Batch{}
	// numInserts := len(items)

	// //load insert statements into batch queue
	// for _, event := range items {
	// 	batch.Queue(AUDIT_INSERT_QUERY,
	// 		event.AuditID,
	// 		event.User.Username,
	// 		event.UserAgent,
	// 		NewNullString(event.ObjectRef.Namespace),
	// 		NewNullString(event.ObjectRef.APIGroup),
	// 		NewNullString(event.ObjectRef.APIVersion),
	// 		event.ObjectRef.Resource,
	// 		event.ObjectRef.Name,
	// 		event.Stage,
	// 		event.StageTimestamp.Time,
	// 		event.Verb,
	// 		event.ResponseStatus.Code,
	// 		event.ResponseStatus.Status,
	// 		event.ResponseStatus.Reason,
	// 		event.ResponseStatus.Message)
	// }

	for _, event := range items {
		_, err := db.Dbpool.Exec(context.TODO(), AUDIT_INSERT_QUERY,
			event.AuditID,
			event.User.Username,
			event.UserAgent,
			db.NewNullString(event.ObjectRef.Namespace),
			db.NewNullString(event.ObjectRef.APIGroup),
			db.NewNullString(event.ObjectRef.APIVersion),
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
			klog.V(1).Info(err)
		}
	}

	// Insert Request Body
	/*
		for _, event := range items {
			_, err = db.Dbpool.Exec(context.TODO(), AUDIT_BODY_INSERT_QUERY,
				event.AuditID,
				NewNullString(event.ObjectRef.Namespace),
				event.RequestObject.Raw)

			if err != nil {
				klog.V(1).Info(err)
			}
		}
	*/

	// //send batch to connection pool
	// br := db.Dbpool.SendBatch(context.TODO(), batch)
	// //execute statements in batch queue
	// for i := 0; i < numInserts; i++ {
	// 	_, err := br.Exec()
	// 	if err != nil {
	// 		klog.V(1).Infoln(err)
	// 		// os.Exit(1)
	// 	}
	// }
	// // klog.V(3).Infoln("Successfully batch inserted data n")

	// //Compare length of results slice to size of table
	// klog.V(3).Infof("size of results: %d\n", numInserts)
	// //check size of table for number of rows inserted
	// // result of last SELECT statement
	// var rowsInserted int
	// err := br.QueryRow().Scan(&rowsInserted)
	// klog.V(3).Infof("size of table: %d\n", rowsInserted)

	// err = br.Close()
	// if err != nil {
	// 	klog.V(1).Infof("Unable to closer batch %v\n", err)
	// }
	klog.V(3).Info("Affected rows: ", len(items))
}

func Get(query string) (audit.EventList, int64) {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
			klog.V(1).Info(err)
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
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	klog.V(3).Infoln("query =", jquery)

	rows, err := db.Dbpool.Query(context.TODO(), jquery)
	if err != nil {
		klog.V(1).Info(err)
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
			klog.V(1).Info(err)
		}
		if namespace.Valid {
			claim.Namespace = namespace.String
		} else {
			claim.Namespace = ""
		}
		//claim.Body = strings.Replace(claim.Body, "\\", "", -1)

		claimList.Claims = append(claimList.Claims, claim)
	}
	return claimList
}

func GetMemberList(query string) ([]string, int64) {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.V(1).Info(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		klog.V(1).Info(err)
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
			klog.V(1).Info(err)
		}
		memberList = append(memberList, member)
	}

	return memberList, row_count
}
