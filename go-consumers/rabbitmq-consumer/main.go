// main.go: Consumidor de RabbitMQ en Go que guarda los datos en Valkey (Redis)
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Variables de entorno para RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	queueName := os.Getenv("RABBITMQ_QUEUE")
	if queueName == "" {
		queueName = "weather-tweets"
	}

	// 2. Variables de entorno para Valkey (Redis)
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6379"
	}

	// 3. Conexión a Valkey
	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Error conectando a Valkey: %v", err)
	}
	log.Println("Conexión exitosa a Valkey!")

	// 4. Conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}
	defer ch.Close()

	// Nos aseguramos de que la cola exista antes de intentar leer
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Error declarando la cola: %v", err)
	}

	// 5. Empezar a consumir los mensajes
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack (le avisa a RabbitMQ que ya procesamos el mensaje para que lo borre de la cola)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Error registrando el consumidor: %v", err)
	}

	log.Println("Consumidor de RabbitMQ iniciado. Esperando mensajes...")

	// 6. Bucle infinito leyendo el canal de Go
	for d := range msgs {
		var tweet map[string]interface{}
		
		if err := json.Unmarshal(d.Body, &tweet); err != nil {
			log.Printf("Error decodificando JSON: %v", err)
			continue
		}

		// Guardamos en Valkey bajo la llave "tweets_rabbit"
		_, err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "tweets_rabbit",
			Values: tweet,
		}).Result()

		if err != nil {
			log.Printf("Error insertando en Valkey: %v", err)
		} else {
			log.Printf("Tweet de %v guardado en Valkey (desde RabbitMQ)", tweet["municipality"])
		}

		// TTL de 2 horas para evitar saturación
		rdb.Expire(ctx, "tweets_rabbit", 2*time.Hour)
	}
}