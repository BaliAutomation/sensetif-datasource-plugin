lint:
	~/go/bin/golangci-lint run
build:
	go build -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg

clean:
	rm -rf dist/*
