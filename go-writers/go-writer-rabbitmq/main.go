// main.go: Writer de RabbitMQ en Go que recibe datos del go-router y los envía a RabbitMQ para su posterior consumo por el consumidor de RabbitMQ
package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"

	pb "wethertweet/proto"
)

// server implementa el servicio gRPC
type server struct {
	pb.UnimplementedWeatherTweetServiceServer
	rabbitConn *amqp.Connection // Mantenemos la conexión principal abierta
	queueName  string
}

// Estructura para enviar a RabbitMQ
type RabbitMessage struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

// SendTweet se ejecuta cada vez que el Router (Deployment 1) nos manda datos
func (s *server) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	
	// Abrimos un canal (channel) temporal y ligero para esta petición
	ch, err := s.rabbitConn.Channel()
	if err != nil {
		log.Printf("Error abriendo canal en RabbitMQ: %v", err)
		return &pb.WeatherTweetResponse{Status: "Error interno"}, err
	}
	defer ch.Close() // Muy importante cerrarlo al terminar

	msg := RabbitMessage{
		Municipality: req.GetMunicipality().String(),
		Temperature:  req.GetTemperature(),
		Humidity:     req.GetHumidity(),
		Weather:      req.GetWeather().String(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return &pb.WeatherTweetResponse{Status: "Error convirtiendo a JSON"}, err
	}

	// Publicamos el mensaje en la cola
	err = ch.PublishWithContext(ctx,
		"",          // exchange (usamos el default)
		s.queueName, // routing key (es el nombre de la cola)
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)

	if err != nil {
		log.Printf("Error publicando en RabbitMQ: %v", err)
		return &pb.WeatherTweetResponse{Status: "Error en RabbitMQ"}, err
	}

	return &pb.WeatherTweetResponse{Status: "Guardado en RabbitMQ exitosamente"}, nil
}

func main() {
	// 1. Configurar variables de entorno
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052" // Usamos el 50052 para diferenciarlo de Kafka
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/" // URL por defecto de RabbitMQ
	}

	queueName := os.Getenv("RABBITMQ_QUEUE")
	if queueName == "" {
		queueName = "weather-tweets"
	}

	// 2. Conectar a RabbitMQ (esta conexión pesada se hace solo una vez)
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Fallo al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 3. Declarar la cola al iniciar para asegurarnos de que exista
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Fallo al abrir canal inicial: %v", err)
	}
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable: los mensajes sobreviven si RabbitMQ se reinicia
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Fallo al declarar la cola: %v", err)
	}
	ch.Close()

	// 4. Iniciar servidor TCP
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Fallo al escuchar en puerto %s: %v", port, err)
	}

	// 5. Registrar gRPC Server
	grpcServer := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(grpcServer, &server{
		rabbitConn: conn,
		queueName:  queueName,
	})

	log.Printf("Writer de RabbitMQ (Deployment 3) escuchando en puerto %s...", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Fallo al iniciar el servidor gRPC: %v", err)
	}
}