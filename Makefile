lint:
	~/go/bin/golangci-lint run

	
build:
	export NODE_OPTIONS=--openssl-legacy-provider && npm run dev
	docker build --pull --rm -f "backend-build.Dockerfile" -t sensetifdatasourceplugin:latest --output dist "."

clean:
	rm -rf dist/*
