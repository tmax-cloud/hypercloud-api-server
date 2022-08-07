package event

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	EVENT_SELECT_BEFORE_INSERT_QUERY = "SELECT name FROM event WHERE uid = $1 and reason = $2 and first_timestamp = $3"
	EVENT_INSERT_QUERY               = "INSERT INTO event (namespace, kind, name, uid, apiversion, fieldpath, action, reason, note, reporting_controller, reporting_instance, host, count, type, first_timestamp, last_timestamp)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"
		//" ON CONFLICT (uid, reason, first_timestamp) DO UPDATE SET last_timestamp = $16, count = $13"
	EVENT_UPDATE_QUERY = "UPDATE event SET last_timestamp = $1, count = $2 WHERE uid = $3 and first_timestamp = $4 and reason = $5"
)

func GetEventDataFromDB(query string) ([]eventv1.Event, error) {
	klog.V(3).Infoln("=== query ===")
	klog.V(3).Infoln(query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
		return nil, err
	}
	defer rows.Close()

	var eventList []eventv1.Event
	for rows.Next() {
		var eventData eventv1.Event
		var first_timestamp, last_timestamp time.Time
		err := rows.Scan(
			&eventData.Regarding.Namespace,
			&eventData.Regarding.Kind,
			&eventData.Regarding.Name,
			&eventData.Regarding.UID,
			&eventData.Regarding.APIVersion,
			&eventData.Regarding.FieldPath,
			&eventData.Action,
			&eventData.Reason,
			&eventData.Note,
			&eventData.ReportingController,
			&eventData.ReportingInstance,
			&eventData.DeprecatedSource.Host,
			&eventData.DeprecatedCount,
			&eventData.Type,
			&first_timestamp,
			&last_timestamp)
		//&eventData.DeprecatedFirstTimestamp,
		//&eventData.DeprecatedLastTimestamp)
		if err != nil {
			klog.V(1).Info(err)
			return nil, err
		}
		eventData.DeprecatedFirstTimestamp = metav1.Time{Time: first_timestamp}
		eventData.DeprecatedLastTimestamp = metav1.Time{Time: last_timestamp}
		eventList = append(eventList, eventData)
	}
	return eventList, nil
}

func Insert(e *eventv1.Event) {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	// Error handling when time information comes into e.EventTime not e.DeprecatedFirstTimestamp,
	// i.e., e.DeprecatedFirstTimestamp comes into 0001-01-01 00:00:00, which is before 1969-01-01 01:01:01
	if e.DeprecatedFirstTimestamp.Time.Before(time.Date(1969, time.Month(1), 1, 1, 1, 1, 1, time.UTC)) {
		e.DeprecatedFirstTimestamp.Time = e.EventTime.Time
		e.DeprecatedLastTimestamp.Time = time.Now()
	}

	// Fisrt, check if there is already same event in DB.
	// If not, INSERT
	// else, UPDATE
	var err error
	err = SelectBeforeInsert(string(e.Regarding.UID), e.Reason, e.DeprecatedFirstTimestamp.Time)

	if err == pgx.ErrNoRows {
		if err := InsertNewEvent(e); err != nil {
			klog.V(1).Info("Failed to Insert new Event for [", e.Regarding.Kind, " ", e.Regarding.Name, "]")
			return
		}
	} else if err != nil {
		klog.V(1).Info("Error occurs during check whether the event is already existed")
		return
	} else {
		if err := UpdateEventRow(e.DeprecatedLastTimestamp.Time, e.DeprecatedCount, string(e.Regarding.UID), e.DeprecatedFirstTimestamp.Time, e.Reason); err != nil {
			klog.V(1).Info("Failed to Update Event for [", e.Regarding.Kind, " ", e.Regarding.Name, "]")
			return
		}
	}

	klog.V(5).Info("Event for [", e.Regarding.Kind, " ", e.Regarding.Name, "] is successfuly inserted")
}

func SelectBeforeInsert(uid string, reason string, firstTime time.Time) error {
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

func InsertNewEvent(e *eventv1.Event) error {
	defer func() {
		if v := recover(); v != nil {
			klog.V(1).Infoln("capture a panic:", v)
		}
	}()

	_, err := db.Dbpool.Exec(context.TODO(), EVENT_INSERT_QUERY,
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
