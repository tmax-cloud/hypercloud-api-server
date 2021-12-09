package consumer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	guuid "github.com/google/uuid"
	haudit "github.com/tmax-cloud/hypercloud-api-server/audit"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"

	// "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

var KafkaGroupId string

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
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				fmt.Println("panic 1 =  " + err.Error())
			} else {
				fmt.Printf("Panic happened with %v", r)
				fmt.Println()
			}
			go HyperauthConsumer()
		} else {
			fmt.Println("what???")
		}
	}()
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	/*
		kubectl create secret tls hypercloud-kafka-secret2 --cert=./hypercloud-api-server.crt,hypercloud-root-ca.crt --key=./hypercloud-api-server.key -n hypercloud5-system
		kubectl create secret generic hypercloud-kafka-secret2 --from-file=./hypercloud-api-server.crt --from-file=./hypercloud-root-ca.crt --from-file=./hypercloud-api-server.key -n hypercloud5-system
	*/
	tlsConfig, err := NewTLSConfig("./etc/ssl/tls.crt",
		"./etc/ssl/tls.key",
		"./etc/ssl/ca.crt")

	if err != nil {
		klog.Errorln(err)
		return
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
	consumerGroupId := KafkaGroupId
	//

	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())

	var kafka1_addr string
	// var kafka2_addr string
	// var kafka3_addr string
	// if os.Getenv("kafka1_addr") != "DNS" {
	// 	kafka1_addr = os.Getenv("kafka1_addr")
	// 	kafka2_addr = os.Getenv("kafka2_addr")
	// 	kafka3_addr = os.Getenv("kafka3_addr")
	// } else {
	kafka1_addr = "kafka-kafka-bootstrap.hyperauth.svc.cluster.local:9092"
	// kafka2_addr = "kafka-2.hyperauth.svc.cluster.local:9092"
	// kafka3_addr = "kafka-3.hyperauth.svc.cluster.local:9092"
	// }

	client, err := sarama.NewConsumerGroup([]string{kafka1_addr}, consumerGroupId, consumerConfig)
	if err != nil {
		klog.Errorln("Error creating consumer group client: %v", err)
		time.Sleep(time.Minute * 1)
		panic("Try Reconnection to Kafka...")
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
			// Delete NamespaceClaim
			k8sApiCaller.DeleteNSCWithUser(topicEvent.UserName)

			// Delete ResourceQuotaClaim
			k8sApiCaller.DeleteRQCWithUser(topicEvent.UserName)

			// Delete RoleBindingClaim
			k8sApiCaller.DeleteRBCWithUser(topicEvent.UserName)

			// Delete ClusterRoleBinding
			k8sApiCaller.DeleteCRBWithUser(topicEvent.UserName)

			// Delete RoleBinding
			k8sApiCaller.DeleteRBWithUser(topicEvent.UserName)

			k8sApiCaller.DeleteGrafanaUser(topicEvent.UserName)

			// 사용자에 대해서..
			// cluster의 주인인 경우.. 클러스터와 관련된 모든걸 지운다.
			// 1. master에서 clm 지우면 됨 (+db)
			// 2. claim 도 지움
			// 3. fedresource도 지움...
			// cluster의 멤버인 경우
			// 1. db에서 해당 사용자에 대한것 모두 삭제
			// 2. clm에 대한 rolebinding도 삭제
			// 3. remote에서 rolebinding도 삭제
			break
		case "REGISTER":
			k8sApiCaller.CreateGrafanaUser(topicEvent.UserName)
			klog.Info("Grafana User [ " + topicEvent.UserName + " ] Created !")
		case "LOGIN":
			klog.Info("login")
			event := audit.Event{
				AuditID: types.UID(guuid.New().String()),
				User: authv1.UserInfo{
					Username: topicEvent.UserName,
				},
				Verb: topicEvent.Type,
				ObjectRef: &audit.ObjectReference{
					Resource:   "users",
					Namespace:  "null",
					Name:       topicEvent.UserName,
					APIGroup:   "null",
					APIVersion: "null",
				},
				ResponseStatus: &metav1.Status{
					Code:   200,
					Status: "Success",
					// Message: string(message.Value),
				},
				StageTimestamp: metav1.MicroTime{
					Time: time.Unix(int64(topicEvent.Time/1000), 0),
				},
			}
			if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
				if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
					haudit.EventBuffer.Buffer <- event
				} else {
					klog.Error("###########   event is dropped.     ############")
				}
			}

		case "LOGOUT":
			klog.Info("LOGOUT")
			event := audit.Event{
				AuditID: types.UID(guuid.New().String()),
				User: authv1.UserInfo{
					Username: topicEvent.UserName,
				},
				Verb: topicEvent.Type,
				ObjectRef: &audit.ObjectReference{
					Resource:   "users",
					Namespace:  "null",
					Name:       topicEvent.UserName,
					APIGroup:   "null",
					APIVersion: "null",
				},
				ResponseStatus: &metav1.Status{
					Code:   200,
					Status: "Success",
					// Message: string(message.Value),
				},
				StageTimestamp: metav1.MicroTime{
					Time: time.Unix(int64(topicEvent.Time/1000), 0),
				},
			}
			if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
				if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
					haudit.EventBuffer.Buffer <- event
				} else {
					klog.Error("###########   event is dropped.     ############")
				}
			}

		case "LOGIN_ERROR":
			klog.Info("LOGIN_ERROR")
			// if topicEvent.
			event := audit.Event{
				AuditID: types.UID(guuid.New().String()),
				User: authv1.UserInfo{
					Username: topicEvent.UserName,
				},
				Verb: topicEvent.Type,
				ObjectRef: &audit.ObjectReference{
					Resource:   "users",
					Namespace:  "null",
					Name:       topicEvent.UserName,
					APIGroup:   "null",
					APIVersion: "null",
				},
				ResponseStatus: &metav1.Status{
					Code:   400,
					Status: "Failure",
					Reason: metav1.StatusReason(topicEvent.Error),
					// Message: string(message.Value),
				},
				StageTimestamp: metav1.MicroTime{
					Time: time.Unix(int64(topicEvent.Time/1000), 0),
				},
			}
			if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
				if len(haudit.EventBuffer.Buffer) < haudit.BufferSize {
					haudit.EventBuffer.Buffer <- event
				} else {
					klog.Error("###########   event is dropped.     ############")
				}
			}

		default:
			// klog.Info("Unknown Event Published from Hyperauth, Do nothing!")
		}

	}

	return nil
}
