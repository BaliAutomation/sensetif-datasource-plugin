module github.com/BaliAutomation/sensetif-datasource

go 1.15

require (
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/gocql/gocql v0.0.0-20210425135552-909f2a77f46e // newest version doesn't resolve inside Goland, but does from Make. Weird!
	github.com/grafana/grafana-plugin-sdk-go v0.105.0
)
