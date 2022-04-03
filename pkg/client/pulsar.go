package client

import (
	"context"
	"fmt"
	"time"

	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type PulsarClient struct {
	client    pulsar.Client
	producers map[string]pulsar.Producer
}

func (p *PulsarClient) Send(topic string, key string, value []byte) string {
	topic = model.Namespace + "/" + topic
	parts, e := p.client.TopicPartitions(topic)
	if e != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to create a producer for topic %s - Error=%+v", topic, e))
		return ""
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Partitions of %s : %+v", topic, parts))
	}
	producer := p.getOrCreateProducer(topic)
	if producer == nil {
		return ""
	}
	message := &pulsar.ProducerMessage{
		Payload: value,
		Key:     key,
	}
	msgId, err := producer.Send(context.Background(), message)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, message.Key, message.Value))
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Sent message on topic %s with key %s. Id: %s. Data: %+v\n", producer.Topic(), message.Key, msgId, string(message.Payload)))
	}
	return string(msgId.Serialize())
}

func (p *PulsarClient) getOrCreateProducer(topic string) pulsar.Producer {
	producer := p.producers[topic]
	if producer == nil {
		var err error
		options := pulsar.ProducerOptions{
			Topic:           topic,
			DisableBatching: true,
		}
		producer, err = p.client.CreateProducer(options)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Failed to create a producer for topic %s - Error=%+v", topic, err))
			return nil
		} else {
			log.DefaultLogger.Info(fmt.Sprintf("Created a new producer for topic %s", topic))
		}
		p.producers[topic] = producer
	}
	return producer
}

func (p *PulsarClient) Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error) {
	return p.client.Subscribe(options)
}

func (p *PulsarClient) InitializePulsar(pulsarHosts string, clientId string) {
	p.producers = make(map[string]pulsar.Producer)
	var err error
	p.client, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:               pulsarHosts,
		ConnectionTimeout: 30 * time.Second,
		OperationTimeout:  30 * time.Second,
	})
	if err != nil {
		log.DefaultLogger.Error("Failed to initialize Pulsar: " + err.Error())
		return
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Connecting %s to Pulsar cluster %s.", clientId, pulsarHosts))
	}
	defer p.client.Close()
}
