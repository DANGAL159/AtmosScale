# Tecnologias
Para entender esta arquitectura a nivel de ingeniería, lo mejor es imaginarla como una gran fábrica o una empresa de logística. Cada tecnología tiene un rol específico, diseñado para resolver un problema que la tecnología anterior no podía manejar.

## 1. La Infraestructura Base (El Terreno y los Edificios)

* **Docker:** Es la tecnología de "empacado". Antes, instalar software era un problema porque dependía del sistema operativo de cada computadora. Docker agarra tu código (Rust, Go), sus librerías y un mini-sistema operativo, y lo encierra en una caja hermética llamada **Contenedor**. Ese contenedor funciona exactamente igual en tu laptop, en Google Cloud o en un servidor en China.
* **Kubernetes (K8s):** Es el **Orquestador** (el gerente de la fábrica). Si tienes 1 o 2 contenedores de Docker, los manejas a mano. Pero si tienes 5,000 contenedores y uno se apaga por error, no puedes estar ahí para reiniciarlo. Kubernetes vigila los contenedores de Docker; si uno muere, lo revive en milisegundos. Si hay mucho tráfico, clona los contenedores (escalabilidad).
* **Docker vs. Kubernetes:** No son competencia, son equipo. Docker construye las cajas (contenedores), y Kubernetes es el barco carguero y la grúa que decide dónde poner esas cajas, cómo conectarlas a la red y asegurarse de que no se hundan.


## 2. La Puerta de Entrada (Redes)

* **Ingress Controller (NGINX):** Es el **Recepcionista o Policía de Tránsito**. Todos tus microservicios adentro de Kubernetes viven en una red privada y no tienen internet. El Ingress es la única puerta abierta al mundo exterior. Lee la URL que el usuario solicita (ej. `/tweet`) y, usando reglas de enrutamiento de capa 7, decide a qué contenedor interno enviarle esa petición.

## 3. El Procesamiento (Los Trabajadores)

* **Rust (La API REST):** Es un lenguaje de programación moderno obsesionado con dos cosas: la velocidad extrema (al nivel de C++) y la seguridad de memoria (evita que el programa colapse por fugas de RAM). En tu arquitectura, Rust es el escudo frontal. Recibe el impacto masivo de miles de peticiones de Locust y las acepta sin que el procesador se sature.
* **Go (Golang):** Es un lenguaje creado por Google, diseñado específicamente para aprovechar procesadores de múltiples núcleos a través de la **concurrencia** (Goroutines). Toma el mensaje que le dio Rust y puede procesarlo, multiplicarlo o enviarlo a múltiples destinos al mismo tiempo con un costo de memoria casi nulo.
* **gRPC:** Es el **Contrato de Comunicación** interna. Normalmente, las APIs hablan en JSON (texto plano, pesado de leer y escribir). gRPC usa Protobuf, que comprime la información en código binario puro. Es como hablar por radio con códigos militares cortos en lugar de enviar cartas largas. Hace que Rust y Go se comuniquen a la velocidad de la luz.

## 4. Los Amortiguadores (Message Brokers)

Cuando Go procesa miles de tweets, no puede intentar guardarlos todos de golpe en la base de datos porque la colapsaría. Necesita "salas de espera".

* **RabbitMQ:** Es el **Buzón de Correos tradicional**. Es una cola de mensajes muy rápida. Go deja el mensaje ahí, y RabbitMQ se asegura de entregárselo al Consumidor en cuanto esté libre. Una vez entregado, el mensaje se marca como procesado y desaparece. Es ideal para tareas inmediatas (ej. enviar un correo de confirmación).
* **Apache Kafka:** Es un **Diario o Bitácora (Log) inmutable**. A diferencia de RabbitMQ, Kafka no borra el mensaje cuando lo leen. Lo guarda ordenado en el disco. Esto permite que múltiples departamentos (consumidores) lean el mismo historial en diferentes momentos, o que si el sistema falla, puedan "rebobinar" la cinta y volver a procesar todo desde ayer. Está diseñado para flujos de *Big Data*.


## 5. El Almacenamiento Rápido

* **Redis:** Es una **Base de Datos en Memoria**. A diferencia de MySQL o PostgreSQL que guardan los datos en el disco duro (que es un proceso lento mecánico o de estado sólido), Redis guarda todo directamente en la memoria RAM de la computadora. Esto permite leer y escribir millones de datos en fracciones de milisegundo.
* **Valkey:** **Valkey ES Redis**. Recientemente, la empresa dueña de Redis cambió su licencia para cobrar a las grandes empresas en la nube. Como respuesta, la Fundación Linux y gigantes como Google y AWS tomaron el último código libre de Redis, le cambiaron el nombre a Valkey, y lo mantienen gratuito y de código abierto. Hacen exactamente lo mismo y usan los mismos comandos (como el `XREVRANGE` que usaste en Grafana).

En resumen: Rust recibe la avalancha de la calle, se la pasa en binario (gRPC) a Go, Go la mete en las salas de espera (Kafka/RabbitMQ) para no saturar nada, otros programas en Go (Consumidores) sacan los datos de esas salas uno por uno con calma, y los guardan en la memoria RAM ultrarrápida (Valkey) para que Grafana pueda dibujar las gráficas en tiempo real sin trabarse.