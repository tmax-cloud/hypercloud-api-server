package cluster

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pq "github.com/lib/pq"
	"k8s.io/klog"
)

const (
	DB_USER      = "postgres"
	DB_PASSWORD  = "tmax"
	DB_NAME      = "postgres"
	HOSTNAME     = "postgres-service.hypercloud5-system.svc"
	PORT         = 5432
	INSERT_QUERY = "INSERT INTO invitation (cluster, member, attribute, invitedTime) VALUES ($1, $2, $3, $4)"
	DELETE_QUERY = "DELETE FROM invitation WHERE cluster = $1, member = $2, attribute = $3"
)

var pg_con_info string

func init() {
	pg_con_info = fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		PORT, HOSTNAME, DB_USER, DB_PASSWORD, DB_NAME)
}

func waitForNotification(l *pq.Listener) {
	for {
		select {
		case n := <-l.Notify:
			fmt.Println("Received data from channel [", n.Channel, "] :")
			// Prepare notification payload for pretty print
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				fmt.Println("Error processing JSON: ", err)
				return
			}
			fmt.Println(string(prettyJSON.Bytes()))
			return
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func test() {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(pg_con_info, 10*time.Second, time.Minute, reportProblem)
	err := listener.Listen("events")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")
	for {
		waitForNotification(listener)
	}
}

func insert(item invitationInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec(INSERT_QUERY, item.cluster, item.member, item.attribute, time.Now())
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func get(cluster string, member string, attribute string) (*invitationInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	invitation := &invitationInfo{}
	var b strings.Builder
	b.WriteString("select * from invitation where 1=1 ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and member = '")
	b.WriteString(member)
	b.WriteString("' ")

	b.WriteString("and attribute = '")
	b.WriteString(attribute)
	b.WriteString("' ")
	query := b.String()
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(
			&invitation.id,
			&invitation.cluster,
			&invitation.member,
			&invitation.attribute,
		)
	}
	return invitation, nil

}

func delete(item invitationInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec(DELETE_QUERY, item.cluster, item.member, item.attribute)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}
