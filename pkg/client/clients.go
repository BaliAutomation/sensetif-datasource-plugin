package client

type Clients struct {
	Cassandra *CassandraClient
	Pulsar    *PulsarClient
	Stripe    *StripeClient
}
