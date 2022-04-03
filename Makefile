lint:
	~/go/bin/golangci-lint run

build_frontend:
	export NODE_OPTIONS=--openssl-legacy-provider && npm run dev
build_backend:
	docker build --pull --rm -f "backend-build.Dockerfile" -t sensetifdatasourceplugin:latest --output dist "."

build:
	make build_frontend && make build_backend

clean:
	rm -rf dist/*
