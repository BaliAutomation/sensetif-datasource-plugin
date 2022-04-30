FROM golang:1.18-alpine as plugin

RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./pkg ./pkg

ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o ./dist/gpx_sensetif-datasource_linux_amd64 ./pkg

FROM scratch AS export-stage
COPY --from=plugin /app/dist/gpx_sensetif-datasource_linux_amd64 .