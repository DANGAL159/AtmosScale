module go-writer-rabbitmq

go 1.24.0

replace wethertweet/proto => ../../proto

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	google.golang.org/grpc v1.79.3
	wethertweet/proto v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
