// main.go: Router en Go que recibe datos de la API en Rust y los envía a ambos gRPC Servers (Writers) de Kafka y RabbitMQ
package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Importamos el código generado por protoc. 
	// Asegúrate de que este path coincida con tu go.mod
	pb "wethertweet/proto" 
)

// 1. Estructura para parsear el JSON que envía la API de Rust
type TweetRequest struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
}

func main() {
	// 2. Leer variables de entorno para saber dónde están los gRPC Servers (Writers)
	// En K8s, estas variables apuntarán a los Services de los Deployments 2 y 3.
	kafkaWriterAddr := os.Getenv("KAFKA_WRITER_ADDR")
	if kafkaWriterAddr == "" {
		kafkaWriterAddr = "localhost:50051" // Puerto por defecto para pruebas locales
	}

	rabbitWriterAddr := os.Getenv("RABBIT_WRITER_ADDR")
	if rabbitWriterAddr == "" {
		rabbitWriterAddr = "localhost:50052"
	}

	// 3. Crear conexión persistente con el gRPC Server de Kafka
	connKafka, err := grpc.NewClient(kafkaWriterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error conectando al Writer de Kafka: %v", err)
	}
	defer connKafka.Close()
	kafkaClient := pb.NewWeatherTweetServiceClient(connKafka)

	// 4. Crear conexión persistente con el gRPC Server de RabbitMQ
	connRabbit, err := grpc.NewClient(rabbitWriterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error conectando al Writer de RabbitMQ: %v", err)
	}
	defer connRabbit.Close()
	rabbitClient := pb.NewWeatherTweetServiceClient(connRabbit)

	// 5. Inicializar el servidor web súper rápido con Fiber
	app := fiber.New()

	app.Post("/tweet", func(c *fiber.Ctx) error {
		// Leer el JSON entrante
		var req TweetRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
		}

		// Mapear los datos del JSON a la estructura de Protobuf
		grpcReq := &pb.WeatherTweetRequest{
			Municipality: pb.Municipalities(req.Municipality),
			Temperature:  req.Temperature,
			Humidity:     req.Humidity,
			Weather:      pb.Weathers(req.Weather),
		}

		// 6. Maximizar Concurrencia: Enviar a ambos brokers al mismo tiempo
		var wg sync.WaitGroup
		wg.Add(2) // Esperaremos a 2 procesos

		// Proceso 1: Enviar a Kafka
		go func() {
			defer wg.Done() // Avisar que terminó cuando salga de la función
			// Damos un timeout de 2 segundos para no bloquear el sistema si el writer falla
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			
			_, err := kafkaClient.SendTweet(ctx, grpcReq)
			if err != nil {
				log.Printf("Fallo envío a Kafka: %v", err)
			}
		}()

		// Proceso 2: Enviar a RabbitMQ
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			
			_, err := rabbitClient.SendTweet(ctx, grpcReq)
			if err != nil {
				log.Printf("Fallo envío a RabbitMQ: %v", err)
			}
		}()

		// Esperar a que ambas Goroutines terminen su trabajo
		wg.Wait()

		// Responder inmediatamente a la API de Rust
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "Tweet procesado y enviado a los writers",
		})
	})

	log.Println("Go Router (Deployment 1) escuchando en puerto 8080...")
	log.Fatal(app.Listen(":8080"))
}