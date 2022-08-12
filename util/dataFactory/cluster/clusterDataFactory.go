package cluster

import (
	"context"
	"strings"
	"time"

	util "github.com/tmax-cloud/hypercloud-api-server/util"
	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"
	"k8s.io/apimachinery/pkg/types"

	// "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	// DB_USER             = "postgres"
	// DB_PASSWORD         = "tmax"
	// DB_NAME             = "postgres"
	// HOSTNAME            = "postgres-service.hypercloud5-system.svc"
	// PORT                = 5432
	INSERT_QUERY        = "INSERT INTO CLUSTER_MEMBER (namespace, cluster, member_id, member_name, attribute, role, status, createdTime, updatedTime) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	DELETE_QUERY        = "DELETE FROM CLUSTER_MEMBER WHERE namespace = $1 and cluster = $2 and member_id = $3 and attribute = $4"
	DELETE_ALL_QUERY    = "DELETE FROM CLUSTER_MEMBER WHERE namespace = $1 and cluster = $2"
	UPDATE_STATUS_QUERY = "UPDATE CLUSTER_MEMBER SET STATUS = 'invited', updatedTime = $1 WHERE namespace = $2 and cluster = $3 and member_id = $4 and attribute = $5 "
	UPDATE_ROLE_QUERY   = "UPDATE CLUSTER_MEMBER SET ROLE = '@@ROLE@@', updatedTime = $1  WHERE namespace = $2 and cluster = $3 and member_id = $4 and attribute = $5 "
)

var pg_con_info string

func Insert(item util.ClusterMemberInfo) error {
	_, err := db.Dbpool.Exec(context.TODO(), INSERT_QUERY, item.Namespace, item.Cluster, item.MemberId, item.MemberName, item.Attribute, item.Role, item.Status, time.Now(), time.Now())
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	return nil
}

func ListClusterMemberWithOutPending(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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

func ListClusterMember(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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

func ListClusterOwnerAndGroupMember(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and (attribute = 'group'")
	b.WriteString("or status = 'owner')")

	query := b.String()
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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

func ListClusterInvitedMember(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and (attribute = 'user'")
	b.WriteString("and status = 'invited')")

	query := b.String()
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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

func ListClusterGroup(cluster string, namespace string) ([]util.ClusterMemberInfo, error) {
	clusterMemberList := []util.ClusterMemberInfo{}
	var b strings.Builder

	b.WriteString("select * from CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and cluster = '")
	b.WriteString(cluster)
	b.WriteString("' ")

	b.WriteString("and (attribute = 'group')")

	query := b.String()
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
		return nil, err
	}
	defer rows.Close()

	//klog.V(3).Infoln(unsafe.Sizeof(rows))
	for rows.Next() {
		var clusterManagerNamespacedName types.NamespacedName
		rows.Scan(
			&clusterManagerNamespacedName.Namespace,
			&clusterManagerNamespacedName.Name,
		)
		clusterManagerNamespacedNameList = append(clusterManagerNamespacedNameList, clusterManagerNamespacedName)
	}
	//klog.V(3).Infoln("NS: " + clusterManagerNamespacedNameList[0].Namespace + " / Name: " + clusterManagerNamespacedNameList[0].Name)
	return clusterManagerNamespacedNameList, nil
}

func GetPendingUser(clusterMember util.ClusterMemberInfo) (*util.ClusterMemberInfo, error) {
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.V(1).Info(err)
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

	klog.V(3).Infoln("Query: " + UPDATE_STATUS_QUERY)
	klog.V(3).Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err := db.Dbpool.Exec(context.TODO(), UPDATE_STATUS_QUERY, time.Now(), item.Namespace, item.Cluster, item.MemberId, item.Attribute)
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	return nil
}

func UpdateRole(item util.ClusterMemberInfo) error {

	query := strings.Replace(UPDATE_ROLE_QUERY, "@@ROLE@@", item.Role, -1)
	klog.V(3).Infoln("Query: " + query)
	klog.V(3).Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err := db.Dbpool.Exec(context.TODO(), query, time.Now(), item.Namespace, item.Cluster, item.MemberId, item.Attribute)
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	return nil
}

func Delete(item util.ClusterMemberInfo) error {

	klog.V(3).Infoln("Query: " + DELETE_QUERY)
	klog.V(3).Infoln("Paremeters: " + item.Namespace + ", " + item.Cluster + ", " + item.MemberId + ", " + item.Attribute)

	_, err := db.Dbpool.Exec(context.TODO(), DELETE_QUERY, item.Namespace, item.Cluster, item.MemberId, item.Attribute)
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	return nil
}
func DeleteALL(namespace, cluster string) error {

	klog.V(3).Infoln("Query: " + DELETE_ALL_QUERY)
	klog.V(3).Infoln("Paremeters: " + namespace + ", " + cluster)

	_, err := db.Dbpool.Exec(context.TODO(), DELETE_ALL_QUERY, namespace, cluster)
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	return nil
}

func GetRemainClusterForSubject(namespace, subject, attribute string) (int, error) {
	var b strings.Builder
	var result int

	b.WriteString("select count(*) from  CLUSTER_MEMBER where 1=1 ")

	b.WriteString("and namespace = '")
	b.WriteString(namespace)
	b.WriteString("' ")

	b.WriteString("and member_id = '")
	b.WriteString(subject)
	b.WriteString("' ")

	b.WriteString("and attribute = '")
	b.WriteString(attribute)
	b.WriteString("' ")

	b.WriteString("and status not in ('pending') ")

	query := b.String()
	klog.V(3).Infoln("Query: " + query)
	rows, err := db.Dbpool.Query(context.TODO(), query)

	if err != nil {
		klog.V(1).Info(err)
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
