package dataFactory

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"k8s.io/klog/v2"
)

var connStr string
var ctx context.Context
var Dbpool *pgxpool.Pool
var DBPassWordPath string

// var DBRootPW stringe

const (
	DB_DRIVER   = "postgres"
	DB_USER     = "postgres"
	DB_PASSWORD = "tmax"
	DB_NAME     = "postgres"
	HOSTNAME    = "timescaledb-service.hypercloud5-system.svc"
	PORT        = "5432"
)

func CreateConnection() {
	var err error
	// content, err := ioutil.ReadFile(DBPassWordPath)
	// if err != nil {
	// 	klog.Errorln(err)
	// 	return
	// }
	// dbRootPW := string(content)

	connStr = DB_DRIVER + "://" + DB_USER + ":" + DB_PASSWORD + "@" + HOSTNAME + ":" + PORT + "/" + DB_NAME
	// 치환
	//connStr = strings.Replace(connStr, "{DB_ROOT_PW}", dbRootPW, -1)
	ctx = context.Background()

	Dbpool, err = pgxpool.Connect(ctx, connStr)
	if err != nil {
		klog.Errorf("Unable to connect to database: %v\n", err)
		panic(err)
	}
}
