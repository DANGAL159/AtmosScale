// main.go: Writer de Kafka en Go que recibe datos del go-router y los envía a Kafka para su posterior consumo por el consumidor de Kafka
package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	pb "wethertweet/proto"
)

// server es la estructura que implementa los métodos de gRPC
type server struct {
	pb.UnimplementedWeatherTweetServiceServer
	kafkaWriter *kafka.Writer // Mantenemos la conexión a Kafka abierta
}

// Estructura para enviar a Kafka en formato JSON
type KafkaMessage struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

// SendTweet es la función que se ejecuta CADA VEZ que el go-router nos envía un dato
func (s *server) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	
	// Convertimos los ENUM de Protobuf a Strings legibles para la Base de Datos
	msg := KafkaMessage{
		Municipality: req.GetMunicipality().String(),
		Temperature:  req.GetTemperature(),
		Humidity:     req.GetHumidity(),
		Weather:      req.GetWeather().String(),
	}

	// Convertimos a JSON (bytes)
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error al convertir a JSON: %v", err)
		return &pb.WeatherTweetResponse{Status: "Error interno"}, err
	}

	// Escribimos el mensaje en Kafka
	err = s.kafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Value: msgBytes,
		},
	)

	if err != nil {
		log.Printf("Error escribiendo en Kafka: %v", err)
		return &pb.WeatherTweetResponse{Status: "Error en Kafka"}, err
	}

	// Respondemos al go-router que todo salió bien
	return &pb.WeatherTweetResponse{Status: "Guardado en Kafka exitosamente"}, nil
}

func main() {
	// 1. Configurar variables de entorno
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051" // Puerto estándar de gRPC
	}

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092" // Puerto estándar de Kafka
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "weather-tweets"
	}

	// 2. Configurar el Writer de Kafka (Se crea una vez y se reutiliza)
	kw := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{}, // Balanceo de carga si hay múltiples particiones
	}
	defer kw.Close()

	// 3. Abrir el puerto TCP
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Fallo al escuchar en puerto %s: %v", port, err)
	}

	// 4. Crear y registrar el servidor gRPC
	grpcServer := grpc.NewServer()
	
	// Registramos nuestro servicio pasándole la conexión de Kafka
	pb.RegisterWeatherTweetServiceServer(grpcServer, &server{
		kafkaWriter: kw,
	})

	log.Printf("Writer de Kafka (Deployment 2) escuchando en puerto %s...", port)
	
	// 5. Iniciar el servidor (Esto bloquea el hilo principal para que no se cierre)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Fallo al iniciar el servidor gRPC: %v", err)
	}
}