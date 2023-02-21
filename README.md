# Jaeger tracing for golang

## Links

1. Jaeger: https://www.jaegertracing.io/docs/1.42/
2. Jaeger client libraries: https://www.jaegertracing.io/docs/1.42/client-libraries/
3. OpenTelemetry GO Getting Started: https://opentelemetry.io/docs/instrumentation/go/getting-started/
4. OpenTelemetry Demo: https://opentelemetry.io/ecosystem/demo/
5. OpenTelemetry-Go Contrib: https://github.com/open-telemetry/opentelemetry-go-contrib
6. opentelemetry-go-contrib/instrumentation/google.golang.org/grpc/otelgrpc/example: https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc/otelgrpc/example
7. opentelemetry-go-contrib/instrumentation/github.com/gin-gonic/gin/otelgin/example: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/example/server.go
8. Implementing OpenTelemetry in a Gin application: https://signoz.io/blog/opentelemetry-gin/
9. opentelemetry-go/example/jaeger: https://github.com/open-telemetry/opentelemetry-go/blob/main/example/jaeger/main.go
10. Take OpenTracing for a HotROD ride: https://medium.com/opentracing/take-opentracing-for-a-hotrod-ride-f6e3141f7941

## Start Jaeger

`$ docker run -d --name jaeger -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 -e COLLECTOR_OTLP_ENABLED=true -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 4317:4317 -p 4318:4318 -p 14250:14250 -p 14268:14268 -p 14269:14269 -p 9411:9411 jaegertracing/all-in-one:1.42`

UI: http://localhost:16686/

## Gin example

[./gin/main.go](./gin/main.go)

### Run

`$ go run ./gin`

### CURLs

```curl
curl --request GET \
  --url http://localhost:8080/albums \
  --header 'Content-Type: application/json'
```

```curl
curl --request GET \
  --url http://localhost:8080/albums/8
```

```curl
curl --request POST \
  --url http://localhost:8080/albums \
  --header 'Content-Type: application/json' \
  --data '{
	"id": "4",
	"title": "The Modern Sound of Betty Carter",
	"artist": "Betty Carter",
	"price": 49.99
}'
```

## GRPC example

[./grpcserver/main.go](./grpcserver/main.go)
[./grpcclient/main.go](./grpcclient/main.go)

### Run

`$ go run ./grpcserver`
`$ go run ./grpcclient`