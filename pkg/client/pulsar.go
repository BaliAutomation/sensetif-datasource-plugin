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
	client pulsar.Client
}

func (p *PulsarClient) Send(topic string, key string, value []byte) string {
	properties := make(map[string]string)
	schema := pulsar.NewBytesSchema(properties)
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic:  topic,
		Name:   "grafana-producer",
		Schema: schema,
	})
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, key, value))
	}

	//producer.SendAsync(context.Background(), &pulsar.ProducerMessage{
	//	Payload: value,
	//	Key:     key,
	//}, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
	//	if err != nil {
	//		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, message.Key, message.Value))
	//	} else {
	//		log.DefaultLogger.Info(fmt.Sprintf("Sent message to key %s on topic %s. Data: %+v\n", message.Key, producer.Topic(), message.Value))
	//	}
	//	defer producer.Close()
	//})
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
	producer.Close()
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
	}
}
