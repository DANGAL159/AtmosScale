// main.go: Consumidor de Kafka en Go que guarda los datos en Valkey (Redis)
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	// 1. Variables de entorno para Kafka
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "weather-tweets"
	}
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "weather-consumer-group"
	}

	// 2. Variables de entorno para Valkey (Redis)
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6379"
	}

	// 3. Conexión a Valkey usando el cliente de Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})
	
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Error conectando a Valkey: %v", err)
	}
	log.Println("Conexión exitosa a Valkey!")

	// 4. Configurar el "Reader" (Lector) de Kafka
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		GroupID:  kafkaGroupID,
		Topic:    kafkaTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	log.Println("Consumidor de Kafka iniciado. Esperando mensajes...")

	// 5. Ciclo infinito: Leer de Kafka y guardar en Valkey
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error leyendo mensaje de Kafka: %v", err)
			continue
		}

		// Decodificamos el JSON entrante solo para validar (y por si quieres hacer algún filtro)
		var tweet map[string]interface{}
		if err := json.Unmarshal(m.Value, &tweet); err != nil {
			log.Printf("Error decodificando JSON: %v", err)
			continue
		}

		// 6. Guardamos en Valkey usando Redis Streams (XADD)
		// Guardaremos todo bajo la llave "tweets_kafka"
		_, err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "tweets_kafka",
			Values: tweet, 
		}).Result()

		if err != nil {
			log.Printf("Error insertando en Valkey: %v", err)
		} else {
			log.Printf("Tweet de %v guardado en Valkey", tweet["municipality"])
		}
		
		// Opcional (pero recomendado): Agregar una expiración (TTL) a la llave principal 
		// para evitar saturar la memoria RAM de Valkey a largo plazo
		rdb.Expire(ctx, "tweets_kafka", 2*time.Hour)
	}
}