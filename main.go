package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	sarama "github.com/Shopify/sarama"
	"github.com/tmax-cloud/hypercloud-api-server/alert"
	cluster "github.com/tmax-cloud/hypercloud-api-server/cluster"
	claim "github.com/tmax-cloud/hypercloud-api-server/clusterClaim"
	metering "github.com/tmax-cloud/hypercloud-api-server/metering"
	"github.com/tmax-cloud/hypercloud-api-server/namespace"
	"github.com/tmax-cloud/hypercloud-api-server/namespaceClaim"
	user "github.com/tmax-cloud/hypercloud-api-server/user"
	version "github.com/tmax-cloud/hypercloud-api-server/version"
	"k8s.io/klog"

	"net/http"

	"github.com/robfig/cron"
)

func main() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	// Get Hypercloud Operating Mode!!!
	hcMode := os.Getenv("HC_MODE")

	// For Log file
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	file, err := os.OpenFile(
		"./logs/api-server.log",
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		os.FileMode(0644),
	)
	if err != nil {
		klog.Error(err, "Error Open", "./logs/api-server")
		return
	}
	defer file.Close()
	w := io.MultiWriter(file, os.Stdout)
	klog.SetOutput(w)

	// Logging Cron Job
	cronJob := cron.New()
	cronJob.AddFunc("1 0 0 * * ?", func() {
		input, err := ioutil.ReadFile("./logs/api-server.log")
		if err != nil {
			klog.Error(err)
			return
		}
		err = ioutil.WriteFile("./logs/api-server"+time.Now().AddDate(0, 0, -1).Format("2006-01-02")+".log", input, 0644)
		if err != nil {
			klog.Error(err, "Error creating", "./logs/api-server")
			return
		}
		klog.Info("Log BackUp Success")
		os.Truncate("./logs/api-server.log", 0)
		file.Seek(0, os.SEEK_SET)
	})

	// Metering Cron Job
	cronJob.AddFunc("0 */1 * ? * *", metering.MeteringJob)
	cronJob.Start()

	// Hyperauth Event Consumer
	go hyperauthConsumer()

	// Req multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/user", serveUser)
	mux.HandleFunc("/metering", serveMetering)
	mux.HandleFunc("/namespace", serveNamespace)
	mux.HandleFunc("/alert", serveAlert)
	mux.HandleFunc("/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/version", serveVersion)

	if hcMode != "single" {
		// for multi mode only
		mux.HandleFunc("/clusterclaim", serveClusterClaim)
		mux.HandleFunc("/cluster", serveCluster)
		mux.HandleFunc("/cluster/owner", serveClusterOwner)
		mux.HandleFunc("/cluster/member", serveClusterMember)
		mux.HandleFunc("/test/", serveTest)
	}

	// HTTP Server Start
	klog.Info("Starting Hypercloud5-API server...")
	klog.Flush()

	if err := http.ListenAndServe(":80", mux); err != nil {
		klog.Errorf("Failed to listen and serve Hypercloud5-API server: %s", err)
	}
	klog.Info("Started Hypercloud5-API server")

}

func hyperauthConsumer() {
	/*
		kubectl create secret tls hypercloud-kafka-secret2 --cert=./hypercloud-api-server.crt,hypercloud-root-ca.crt --key=./hypercloud-api-server.key -n hypercloud5-system
		kubectl create secret generic hypercloud-kafka-secret2 --from-file=./hypercloud-api-server.crt --from-file=./hypercloud-root-ca.crt --from-file=./hypercloud-api-server.key -n hypercloud5-system
	*/
	tlsConfig, err := NewTLSConfig("./etc/ssl/hypercloud-api-server.crt",
		"./etc/ssl/hypercloud-api-server.key",
		"./etc/ssl/hypercloud-root-ca.crt")
	if err != nil {
		klog.Fatal(err)
	}
	// This can be used on test server if domain does not match cert:
	// tlsConfig.InsecureSkipVerify = true

	// Consumer Config!!!
	consumerConfig := sarama.NewConfig()
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.ClientID = "hypercloud-api-server" //FIXME
	consumerGroupId := "hypercloud-api-server"
	//

	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup([]string{"kafka-1.hyperauth:9092", "kafka-2.hyperauth:9092", "kafka-3.hyperauth:9092"}, consumerGroupId, consumerConfig)
	if err != nil {
		klog.Error("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		topic := "tmax"
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, []string{topic}, &consumer); err != nil {
				klog.Error("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	klog.Info("hypercloud-api-server consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		klog.Info("terminating: context cancelled")
	case <-sigterm:
		klog.Info("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		klog.Error("Error closing client: %v", err)
	}
}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		klog.Info("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}

func serveNamespace(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespace.Get(res, req)
	case http.MethodPut:
		namespace.Put(res, req)
	case http.MethodOptions:
		namespace.Options(res, req)
	default:
		//error
	}
}

func serveNamespaceClaim(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespaceClaim.Get(res, req)
	case http.MethodPut:
		namespaceClaim.Put(res, req)
	case http.MethodOptions:
		namespaceClaim.Options(res, req)
	default:
		//error
	}
}

func serveUser(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		user.Post(res, req)
	case http.MethodDelete:
		user.Delete(res, req)
	case http.MethodOptions:
		user.Options(res, req)
	default:
		//error
	}
}

func serveMetering(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		metering.Get(res, req)
	case http.MethodOptions:
		metering.Options(res, req)
	default:
		//error
	}
}

func serveAlert(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		alert.Post(res, req)
	case http.MethodGet:
		alert.Get(res, req)
	default:
		//error
	}
}
func serveVersion(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		version.Get(res, req)
	default:
		//error
	}
}

func serveTest(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	klog.Info("Request body: \n", string(body))
}

func serveClusterClaim(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		//curl -XPUT 172.22.6.2:32319/api/master/clusterclaim?userId=sangwon_cho@tmax.co.kr
		claim.List(res, req)
	case http.MethodPost:
	case http.MethodPut:
		//curl -XPUT 172.22.6.2:32319/api/master/clusterclaim?userId=sangwon_cho@tmax.co.kr\&clusterClaim=test-d5n92\&admit=true
		claim.Put(res, req)
	case http.MethodDelete:
	default:
	}
}

func serveCluster(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.List(res, req)
	case http.MethodPost:
	case http.MethodPut:
		// invite multiple users
		cluster.Put(res, req)
	case http.MethodDelete:
	default:
	}
}

func serveClusterOwner(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.ListOwner(res, req)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	default:
	}
}

func serveClusterMember(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.ListMember(res, req)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	default:
	}
}
