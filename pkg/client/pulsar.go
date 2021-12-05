package client

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"time"
)

type Pulsar interface {
	Send(topic string, key string, value []byte)
}

type PulsarClient struct {
	client    pulsar.Client
	producers map[string]pulsar.Producer
}

func (p *PulsarClient) Send(topic string, key string, value []byte) string {
	properties := make(map[string]string)
	schema := pulsar.NewBytesSchema(properties)
	producer := p.producers[topic]
	if producer == nil {
		var err error
		producer, err = p.client.CreateProducer(pulsar.ProducerOptions{
			Topic:  topic,
			Name:   "grafana-producer",
			Schema: schema,
		})
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Failed to create a producer for topic %s - Error=%s, %+v", topic, err.Error(), err), err)
			return ""
		}
	}

	message := &pulsar.ProducerMessage{
		Payload: value,
		Key:     key,
	}
	msgId, err := producer.Send(context.Background(), message)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, message.Key, message.Value))
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Sent message to key %s on topic %s. Id: %s. Data: %+v\n", message.Key, producer.Topic(), msgId, message.Value))
	}
	return string(msgId.Serialize())
}

func (p *PulsarClient) InitializePulsar(hosts string, clientId string) {
	var err error
	p.client, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:               hosts,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		log.DefaultLogger.Error("Failed to initialize Pulsar: " + err.Error())
		return
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Connecting %s to Pulsar cluster %s.", clientId, hosts))
	}
}
