package cluster

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pq "github.com/lib/pq"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/apimachinery/pkg/types"

	// "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	DB_USER             = "postgres"
	DB_PASSWORD         = "tmax"
	DB_NAME             = "postgres"
	HOSTNAME            = "postgres-service.hypercloud5-system.svc"
	PORT                = 5432
	INSERT_QUERY        = "INSERT INTO CLUSTER_MEMBER (namespace, cluster, member_id, member_name, attribute, role, status, createdTime, updatedTime) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	DELETE_QUERY        = "DELETE FROM CLUSTER_MEMBER WHERE namespace = $1 and cluster = $2 and member_id = $3 and attribute = $4"
	DELETE_ALL_QUERY    = "DELETE FROM CLUSTER_MEMBER WHERE namespace = $1 and cluster = $2"
	UPDATE_STATUS_QUERY = "UPDATE CLUSTER_MEMBER SET STATUS = 'invited' WHERE namespace = $1 and cluster = $2 and member_id = $3 and attribute = $4 and updatedTime = $5"
	UPDATE_ROLE_QUERY   = "UPDATE CLUSTER_MEMBER SET ROLE = '@@ROLE@@' WHERE namespace = $1 and cluster = $2 and member_id = $3 and attribute = $4 and updatedTime = $5"
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

func Insert(item util.ClusterMemberInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec(INSERT_QUERY, item.Namespace, item.Cluster, item.MemberId, item.MemberName, item.Attribute, item.Role, item.Status, time.Now(), time.Now())
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func ListClusterMemberWithOutPending(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and status not in ('pending') ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clusterMember := util.ClusterMemberInfo{}
		rows.Scan(
			&clusterMember.Id,
			&clusterMember.Namespace,
			&clusterMember.Cluster,
			&clusterMember.MemberId,
			&clusterMember.MemberName,
			&clusterMember.Attribute,
			&clusterMember.Role,
			&clusterMember.Status,
			&clusterMember.CreatedTime,
			&clusterMember.UpdatedTime,
		)
		clusterMemberList = append(clusterMemberList, clusterMember)
	}
	return clusterMemberList, nil
}

func ListAllClusterMember(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clusterMember := util.ClusterMemberInfo{}
		rows.Scan(
			&clusterMember.Id,
			&clusterMember.Namespace,
			&clusterMember.Cluster,
			&clusterMember.MemberId,
			&clusterMember.MemberName,
			&clusterMember.Attribute,
			&clusterMember.Role,
			&clusterMember.Status,
			&clusterMember.CreatedTime,
			&clusterMember.UpdatedTime,
		)
		clusterMemberList = append(clusterMemberList, clusterMember)
	}
	return clusterMemberList, nil
}

func ListAllClusterUser(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and attribute = 'user'")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clusterMember := util.ClusterMemberInfo{}
		rows.Scan(
			&clusterMember.Id,
			&clusterMember.Namespace,
			&clusterMember.Cluster,
			&clusterMember.MemberId,
			&clusterMember.MemberName,
			&clusterMember.Attribute,
			&clusterMember.Role,
			&clusterMember.Status,
			&clusterMember.CreatedTime,
			&clusterMember.UpdatedTime,
		)
		clusterMemberList = append(clusterMemberList, clusterMember)
	}
	return clusterMemberList, nil
}

func ListAllClusterGroup(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and attribute = 'group'")
	b.WriteString("or status = 'owner' ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clusterMember := util.ClusterMemberInfo{}
		rows.Scan(
			&clusterMember.Id,
			&clusterMember.Namespace,
			&clusterMember.Cluster,
			&clusterMember.MemberId,
			&clusterMember.MemberName,
			&clusterMember.Attribute,
			&clusterMember.Role,
			&clusterMember.Status,
			&clusterMember.CreatedTime,
			&clusterMember.UpdatedTime,
		)
		clusterMemberList = append(clusterMemberList, clusterMember)
	}
	return clusterMemberList, nil
}

func ListClusterInNamespace(userId string, userGroups []string, namespace string) ([]string, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterNameList := []string{}
	var b strings.Builder

	b.WriteString("select cluster from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and (namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and member_id = '")
	b.WriteString(userId)
	b.WriteString("' ")

	for _, userGroup := range userGroups {
		b.WriteString("or member_id = '")
		b.WriteString(userGroup)
		b.WriteString("' ")
	}
	b.WriteString(") ")
	b.WriteString("and status not in ('pending') ")

	b.WriteString("group by cluster")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var clusterNmae string
		rows.Scan(
			&clusterNmae,
		)
		clusterNameList = append(clusterNameList, clusterNmae)
	}
	return clusterNameList, nil
}

func ListClusterAllNamespace(userId string, userGroups []string) ([]types.NamespacedName, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterManagerNamespacedNameList := []types.NamespacedName{}
	var b strings.Builder

	b.WriteString("select namespace, cluster from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and (member_id = '")
	b.WriteString(userId)
	b.WriteString("' ")

	for _, userGroup := range userGroups {
		b.WriteString("or member_id = '")
		b.WriteString(userGroup)
		b.WriteString("' ")
	}
	b.WriteString(") ")

	b.WriteString("and status not in ('pending') ")

	b.WriteString("group by namespace, cluster")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var clusterManagerNamespacedName types.NamespacedName
		rows.Scan(
			&clusterManagerNamespacedName.Namespace,
			&clusterManagerNamespacedName.Name,
		)
		clusterManagerNamespacedNameList = append(clusterManagerNamespacedNameList, clusterManagerNamespacedName)
	}
	return clusterManagerNamespacedNameList, nil
}

func GetPendingUser(clusterMember util.ClusterMemberInfo) (*util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	// clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(clusterMember.Namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(clusterMember.Cluster)
	b.WriteString("' ")

	b.WriteString("and member_id = '")
	b.WriteString(clusterMember.MemberId)
	b.WriteString("' ")

	b.WriteString("and attribute = 'user' ")

	b.WriteString("and status = 'pending' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()
	ret := util.ClusterMemberInfo{}
	if rows.Next() {
		rows.Scan(
			&ret.Id,
			&ret.Namespace,
			&ret.Cluster,
			&ret.MemberId,
			&ret.MemberName,
			&ret.Attribute,
			&ret.Role,
			&ret.Status,
			&ret.CreatedTime,
			&ret.UpdatedTime,
		)
	}
	return &ret, nil
}

func ListPendingUser(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer db.Close()
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder
	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	if cluster != "" {
		b.WriteString("and cluster = '")
		b.WriteString(cluster)
		b.WriteString("' ")
	}

	b.WriteString("and status = 'pending' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clusterMember := util.ClusterMemberInfo{}
		rows.Scan(
			&clusterMember.Id,
			&clusterMember.Namespace,
			&clusterMember.Cluster,
			&clusterMember.MemberId,
			&clusterMember.MemberName,
			&clusterMember.Attribute,
			&clusterMember.Role,
			&clusterMember.Status,
			&clusterMember.CreatedTime,
			&clusterMember.UpdatedTime,
		)
		clusterMemberList = append(clusterMemberList, clusterMember)
	}
	return clusterMemberList, nil
}

func GetInvitedGroup(clusterMember util.ClusterMemberInfo) (int, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	defer db.Close()
	var result int
	var b strings.Builder

	b.WriteString("select count(*) from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and cluster = '")
	b.WriteString(clusterMember.Cluster)
	b.WriteString("' ")

	b.WriteString("and member_id = '")
	b.WriteString(clusterMember.MemberId)
	b.WriteString("' ")

	b.WriteString("and attribute = '")
	b.WriteString(clusterMember.Attribute)
	b.WriteString("' ")

	b.WriteString("and status = '")
	b.WriteString(clusterMember.Status)
	b.WriteString("' ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(
			&result,
		)
	}
	return result, nil
}

func UpdateStatus(item *util.ClusterMemberInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	klog.Infoln("Query: " + UPDATE_STATUS_QUERY)
	klog.Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err = db.Exec(UPDATE_STATUS_QUERY, item.Namespace, item.Cluster, item.MemberId, item.Attribute, time.Now())
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func UpdateRole(item util.ClusterMemberInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	query := strings.Replace(UPDATE_ROLE_QUERY, "@@ROLE@@", item.Role, -1)
	klog.Infoln("Query: " + query)
	klog.Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err = db.Exec(query, item.Namespace, item.Cluster, item.MemberId, item.Attribute, time.Now())
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func Delete(item util.ClusterMemberInfo) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	klog.Infoln("Query: " + DELETE_QUERY)
	klog.Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err = db.Exec(DELETE_QUERY, item.Namespace, item.Cluster, item.MemberId, item.Attribute)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}
func DeleteALL(namespace, cluster string) error {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return err
	}
	defer db.Close()

	klog.Infoln("Query: " + DELETE_ALL_QUERY)
	klog.Infoln("Paremeters: " + namespace + ", " + cluster)

	_, err = db.Exec(DELETE_ALL_QUERY, namespace, cluster)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func GetRemainClusterForSubject(namespace, cluster, subject, attribute string) (int, error) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	defer db.Close()
	var b strings.Builder
	var result int

	b.WriteString("select count(*) from  CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and member_id = '")
	b.WriteString(subject)
	b.WriteString("' ")

	b.WriteString("and attribute = '")
	b.WriteString(attribute)
	b.WriteString("' ")

	b.WriteString("and status not in ('pending') ")

	query := b.String()
	klog.Infoln("Query: " + query)
	rows, err := db.Query(query)

	if err != nil {
		klog.Error(err)
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(
			&result,
		)
	}
	return result, nil
}
