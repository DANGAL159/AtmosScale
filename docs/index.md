# Sistema de Tweets del Clima ⛅

Bienvenido a la documentación técnica del proyecto "Tweets del Clima".

Este proyecto consiste en la implementación de una arquitectura distribuida y escalable para simular la recepción y procesamiento masivo de datos climáticos en tiempo real. 

## Arquitectura del Sistema
El sistema está diseñado para soportar una alta concurrencia generada por Locust[cite: 16], pasando por diferentes capas de procesamiento:

1. **API REST (Rust):** Punto de entrada ultrarrápido que recibe el tráfico.
2. **Router gRPC (Go):** Servicio puente que recibe peticiones REST y las distribuye concurrentemente vía gRPC.
3. **Writers gRPC (Go):** Servidores encargados de publicar los mensajes en los Message Brokers (Kafka y RabbitMQ).
4. **Message Brokers:** Tecnologías de encolamiento (Kafka y RabbitMQ) para la comunicación asíncrona.
5. **Consumers (Go):** Servicios que leen continuamente las colas y almacenan los datos.
6. **Almacenamiento (Valkey):** Base de datos en memoria para almacenamiento de ultra alta velocidad.

Navega a la sección del **Manual** para ver el paso a paso del desarrollo y los comandos de construcción.