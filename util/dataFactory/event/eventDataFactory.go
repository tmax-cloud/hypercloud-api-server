package event

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"
	eventv1 "k8s.io/api/events/v1"
	"k8s.io/klog"
)

const (
	EVENT_SELECT_BEFORE_INSERT_QUERY = "SELECT name FROM event WHERE uid = $1 and reason = $2 and first_timestamp = $3"
	EVENT_INSERT_QUERY               = "INSERT INTO event (namespace, kind, name, uid, apiversion, fieldpath, action, reason, note, reporting_controller, reporting_instance, host, count, type, first_timestamp, last_timestamp)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"
		//" ON CONFLICT (uid, reason, first_timestamp) DO UPDATE SET last_timestamp = $16, count = $13"
	EVENT_UPDATE_QUERY = "UPDATE event SET last_timestamp = $1, count = $2 WHERE uid = $3 and first_timestamp = $4 and reason = $5"
)

func Get() {
	// TODO : implement get function for client
}

func Insert(e *eventv1.Event) {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	// Fisrt, check if there is already same event in DB.
	// If not, INSERT
	// else, UPDATE
	var err error
	err = SELECT_BEFORE_INSERT(string(e.Regarding.UID), e.Reason, e.DeprecatedFirstTimestamp.Time)

	if err == pgx.ErrNoRows {
		_, err = db.Dbpool.Exec(context.TODO(), EVENT_INSERT_QUERY,
			db.NewNullString(e.Regarding.Namespace),
			db.NewNullString(e.Regarding.Kind),
			db.NewNullString(e.Regarding.Name),
			string(e.Regarding.UID), // some event uses node name for uid
			db.NewNullString(e.Regarding.APIVersion),
			db.NewNullString(e.Regarding.FieldPath),
			db.NewNullString(e.Action),
			db.NewNullString(e.Reason),
			db.NewNullString(e.Note),
			db.NewNullString(e.ReportingController),
			db.NewNullString(e.ReportingInstance),
			db.NewNullString(e.DeprecatedSource.Host),
			e.DeprecatedCount,
			db.NewNullString(e.Type),
			e.DeprecatedFirstTimestamp.Time,
			e.DeprecatedLastTimestamp.Time)
	} else if err != nil {
		klog.V(1).Info("Error occurs during check whether the event is already existed")
		return
	} else {
		if err := UpdateEventRow(e.DeprecatedLastTimestamp.Time, e.DeprecatedCount, string(e.Regarding.UID), e.DeprecatedFirstTimestamp.Time, e.Reason); err != nil {
			klog.V(1).Info("Failed to Update Event for [", e.Regarding.Kind, " ", e.Regarding.Name, "].")
			return
		}
	}

	klog.V(5).Info("Event for [", e.Regarding.Kind, " ", e.Regarding.Name, "] is successfuly inserted.")
}

func SELECT_BEFORE_INSERT(uid string, reason string, firstTime time.Time) error {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	var name string
	var err error

	err = db.Dbpool.QueryRow(context.TODO(), EVENT_SELECT_BEFORE_INSERT_QUERY, uid, reason, firstTime).Scan(&name)
	if err == pgx.ErrNoRows {
		klog.V(5).Info("No existing event, Do INSERT")
	} else if err != nil {
		klog.V(1).Info(err)
	}
	return err
}

func UpdateEventRow(lastTime time.Time, count int32, uid string, firstTime time.Time, reason string) error {
	_, err := db.Dbpool.Exec(context.TODO(), EVENT_UPDATE_QUERY,
		lastTime, count, uid, firstTime, reason)
	if err != nil {
		klog.V(1).Info(err)
	}
	return err
}
