package consumer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	"k8s.io/klog"
)

type TopicEvent struct {
	Type      string            `json:"type"`
	UserName  string            `json:"userName"`
	UserId    string            `json:"userId"`
	Time      float32           `json:"time"`
	RealmId   string            `json:"realmId"`
	ClientId  string            `json:"clientId"`
	SessionId string            `json:"sessionId"`
	IpAddress string            `json:"ipAddress"`
	Error     string            `json:"error"`
	Details   map[string]string `json:"details"`
}

func HyperauthConsumer() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

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
	version, err := sarama.ParseKafkaVersion("2.0.1")

	consumerConfig := sarama.NewConfig()
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.ClientID = "hypercloud-api-server" //FIXME
	consumerConfig.Version = version
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
		klog.Errorf("Error closing client: %v", err)
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

	for message := range claim.Messages() {
		klog.Infof("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")

		var topicEvent TopicEvent
		if err := json.Unmarshal(message.Value, &topicEvent); err != nil {
			klog.Error("make topicEvent Struct failed : ", err)
		}

		//LOGIC HERE!!
		switch topicEvent.Type {
		case "USER_DELETE":
			klog.Info("User [ " + topicEvent.UserName + " ] Deleted !")
			// Delete NamespaceClaim with Creator Annotation
			k8sApiCaller.DeleteNSCWithUser(topicEvent.UserName)

			// Delete ResourceQuotaClaim with Creator Annotation
			k8sApiCaller.DeleteRQCWithUser(topicEvent.UserName)

			// Delete RoleBindingClaim with Creator Annotation
			k8sApiCaller.DeleteRBCWithUser(topicEvent.UserName)
			break
		default:
			klog.Info("Unknown Event Published from Hyperauth, Do nothing!")

		}

	}

	return nil
}
