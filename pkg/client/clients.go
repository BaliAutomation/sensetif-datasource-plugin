package client

type Clients struct {
	Cassandra *CassandraClient
	Kafka     *KafkaClient
	Stripe    *StripeClient
}
