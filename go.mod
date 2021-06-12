module github.com/BaliAutomation/sensetif-datasource-plugin

go 1.15

require (
	github.com/gocql/gocql v0.0.0-20210425135552-909f2a77f46e	// newest version doesn't resolve inside Goland, but does from Make. Weird!
	github.com/grafana/grafana-plugin-sdk-go v0.92.0
	google.golang.org/genproto v0.0.0-20200911024640-645f7a48b24f // indirect
)
