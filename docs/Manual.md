# Guía de Desarrollo y Construcción

A continuación, se detalla el proceso paso a paso para la creación de los microservicios, compilación de binarios y construcción de las imágenes Docker, los comandos de inicialización siempre se deben de realizar en la carpeta raíz y los de compilación o creacion de imagenes en la carpeta actual trabajada, a menos que se indique lo contrario.

## 0. Prerrequisitos e Instalaciones Necesarias
> **Entorno de ejecución**:  
> Esta guía utiliza `k3d` por defecto para garantizar compatibilidad multiplataforma (Windows/macOS/Linux).  
> Si estás en **Linux nativo** y prefieres usar `k3s` directamente para mayor rendimiento debes de modificar los comandos de la guía.
> Se realizó en un sistema Windows 11

Para poder compilar los microservicios, generar los contratos gRPC y levantar la infraestructura local, es estrictamente necesario contar con las siguientes herramientas instaladas en el entorno de desarrollo (Windows 10/11). 

Se recomienda utilizar `winget` (el administrador de paquetes de Windows) desde una terminal de PowerShell o CMD ejecutada como Administrador.

### 0.1 Motor de Contenedores y Orquestación
El proyecto utiliza Docker como motor de contenedores y K3d para virtualizar el clúster de Kubernetes.

* **Docker Desktop:** Necesario para compilar las imágenes. (Requiere tener WSL 2 habilitado en Windows).
  * *Descarga manual desde:* [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)
* **Kubectl:** Herramienta de línea de comandos para interactuar con Kubernetes.
```cmd
  winget install Kubernetes.kubectl
```
* **K3d:** Wrapper ligero para levantar clústeres de K3s (Kubernetes) dentro de Docker.
```cmd
  winget install Rancher.k3d
```

### 0.2 Lenguajes de Programación
El ecosistema está construido en Rust (API REST), Go (Router y Workers) y Python (Scripts de pruebas y documentación).

* **Go (Golang):** Para compilar el router, los escritores y los consumidores.
```cmd
  winget install GoLang.Go
```
* **Rust y Cargo:** Para compilar la API de entrada.
```cmd
  winget install Rustlang.Rustup
```
* **Python y Pip:** Requeridos para ejecutar Locust (Pruebas de carga) y MkDocs (Generación de este sitio estático).
```cmd
  winget install Python.Python.3.11
```

### 0.3 Compilador de Contratos (Protocol Buffers)
Para que Go y Rust puedan comunicarse mediante gRPC, es necesario el compilador de Protobuf para transformar el archivo `.proto` en código fuente.

* **Protoc:**
  ```cmd
  winget install ProtocolBuffers.protoc
  ```

### 0.4 Validar Instalaciones
Una vez instaladas todas las herramientas, es recomendable cerrar la terminal, abrir una nueva y verificar que las variables de entorno se hayan configurado correctamente ejecutando:

```cmd
docker --version
kubectl version --client
k3d --version
go version
cargo --version
python --version
protoc --version

```

*Si todos los comandos devuelven una versión, el entorno está listo para continuar con el Paso 1.*


## 1. Zot Registry (Repositorio de Imágenes)
Para cumplir con los estándares de arquitectura, todas las imágenes generadas se almacenan en un registro OCI privado (Zot) antes de ser consumidas por Kubernetes.

**Comando para levantar el servidor Zot localmente:**
```bash
k3d registry create zot-registry --port 5000 --image ghcr.io/project-zot/zot-linux-amd64:latest

```

## 2. Contrato de Comunicación (gRPC Protobuf)
El primer paso fue definir el contrato estricto de comunicación entre los servicios usando `.proto` y compilarlo para generar el código de Go.

**Comandos de inicialización:**
```bash
mkdir proto
cd proto
go mod init wethertweet/proto

```

<br>[COLOCAR CODIGO DEL ARCHIVO: proto/tweet.proto]


**Comandos de compilación:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc --go_out=. --go-grpc_out=. proto/tweet.proto
go mod tidy

```

## 3. API REST en Rust (Punto de Entrada)
Se desarrolló una API web utilizando Actix-Web para soportar la alta carga inicial.

**Comandos de inicialización:**
```bash
cargo new api-rust
cd api-rust

```

<br>[COLOCAR CODIGO DEL ARCHIVO: api-rust/Cargo.toml]
<br>[COLOCAR CODIGO DEL ARCHIVO: api-rust/src/main.rs]
<br>[COLOCAR CODIGO DEL ARCHIVO: api-rust/Dockerfile]


**Construcción de la imagen Docker:**
```bash
docker build -t api-rust:local .
docker tag api-rust:local localhost:5000/api-rust:local
docker push localhost:5000/api-rust:local

```

## 4. Router Go (Deployment 1 - REST a gRPC)
Este servicio recibe el JSON de Rust por REST y dispara las peticiones gRPC concurrentemente.

**Comandos de inicialización:**
```bash
mkdir go-router
cd go-router
go mod init go-router

go get google.golang.org/grpc
go get [github.com/gofiber/fiber/v2](https://github.com/gofiber/fiber/v2)
go mod edit -replace wethertweet/proto=../proto
go mod tidy

```

<br>[COLOCAR CODIGO DEL ARCHIVO: go-router/main.go]
<br>[COLOCAR CODIGO DEL ARCHIVO: go-router/Dockerfile]

**Construcción de la imagen Docker (Ejecutar desde la raíz del proyecto):**
```bash
cd ..
docker build -f go-router/Dockerfile -t go-router:local .
docker tag go-router:local localhost:5000/go-router:local
docker push localhost:5000/go-router:local

```

## 5. Writers gRPC (Deployments 2 y 3)
Servidores encargados de recibir gRPC y publicar en Kafka y RabbitMQ respectivamente.

### Writer Kafka
**Comandos de inicialización:**
```bash
mkdir -p go-writers/go-writer-kafka
cd go-writers/go-writer-kafka
go mod init go-writer-kafka

go get google.golang.org/grpc
go get [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go)
go mod edit -replace wethertweet/proto=../../proto
go mod tidy

```

<br>[COLOCAR CODIGO DEL ARCHIVO: go-writers/go-writer-kafka/main.go]
<br>[COLOCAR CODIGO DEL ARCHIVO: go-writers/go-writer-kafka/Dockerfile]

**Construcción de la imagen Docker (Ejecutar desde la raíz del proyecto):**
```bash
cd ../..
docker build -f go-writers/go-writer-kafka/Dockerfile -t go-writer-kafka:local .
docker tag go-writer-kafka:local localhost:5000/go-writer-kafka:local
docker push localhost:5000/go-writer-kafka:local

```

### Writer RabbitMQ
**Comandos de inicialización:**
```bash
mkdir -p go-writers/go-writer-rabbitmq
cd go-writers/go-writer-rabbitmq
go mod init go-writer-rabbitmq

go get google.golang.org/grpc
go get [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
go mod edit -replace wethertweet/proto=../../proto
go mod tidy

```

<br>[COLOCAR CODIGO DEL ARCHIVO: go-writers/rabbitmq/main.go]
<br>[COLOCAR CODIGO DEL ARCHIVO: go-writers/rabbitmq/Dockerfile]

**Construcción de la imagen Docker (Ejecutar desde la raíz del proyecto):**
```bash
cd ../..
docker build -f go-writers/go-writer-rabbitmq/Dockerfile -t go-writer-rabbitmq:local .
docker tag go-writer-rabbitmq:local localhost:5000/go-writer-rabbitmq:local
docker push localhost:5000/go-writer-rabbitmq:local

```

## 6. Consumidores de Eventos (Deployments)
Procesos independientes que leen de las colas y persisten los datos en Valkey mediante Streams.

### Consumer Kafka
**Comandos de inicialización:**
```bash
mkdir -p go-consumers/kafka-consumer
cd go-consumers/kafka-consumer
go mod init kafka-consumer

go get [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go)
go get [github.com/redis/go-redis/v9](https://github.com/redis/go-redis/v9)

```

<br>[COLOCAR CODIGO DEL ARCHIVO: go-consumers/kafka-consumer/main.go]
<br>[COLOCAR CODIGO DEL ARCHIVO: go-consumers/kafka-consumer/Dockerfile]

**Construcción de la imagen Docker:**
```bash
docker build -t kafka-consumer:local .
docker tag kafka-consumer:local localhost:5000/kafka-consumer:local
docker push localhost:5000/kafka-consumer:local

```

### Consumer RabbitMQ
**Comandos de inicialización:**
```bash
mkdir -p go-consumers/rabbitmq-consumer
cd go-consumers/rabbitmq-consumer
go mod init rabbitmq-consumer

go get [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
go get [github.com/redis/go-redis/v9](https://github.com/redis/go-redis/v9)

```

<br>[COLOCAR CODIGO DEL ARCHIVO: go-consumers/rabbitmq-consumer/main.go]
<br>[COLOCAR CODIGO DEL ARCHIVO: go-consumers/rabbitmq-consumer/Dockerfile]

**Construcción de la imagen Docker:**
```bash
docker build -t rabbitmq-consumer:local .
docker tag rabbitmq-consumer:local localhost:5000/rabbitmq-consumer:local
docker push localhost:5000/rabbitmq-consumer:local

```

## 7. Orquestación con Kubernetes (K3d)
Para simular un entorno de producción, utilizamos K3d para levantar un clúster local de K3s. Se deshabilitó el Ingress por defecto (Traefik) para cumplir con el requisito de utilizar NGINX.

### 7.1 Creación del clúster e importación de imágenes:

Dependiendo del entorno, existen dos formas de levantar la infraestructura base:

#### Opción A: Entorno de Desarrollo Rápido (Imágenes locales)**
Ideal para pruebas rápidas sin necesidad de un registro externo. Utilizar la carpeta k8s-docker-images, las rutas ya estan adecuadas

```yml
containers:
      - name: 
        image:
```

```bash
# Crear clúster con 1 server y 2 agents, mapeando el puerto 80 al 8080 local
k3d cluster create clima-cluster --servers 1 --agents 2 --port "8080:80@loadbalancer" --k3s-arg "--disable=traefik@server:0"

# Inyectar las imágenes compiladas localmente directo a los nodos de K3d 
k3d image import api-rust:local go-router:local go-writer-kafka:local go-writer-rabbitmq:local kafka-consumer:local rabbitmq-consumer:local -c clima-cluster

```

#### Opción B: Entorno de Producción (Usando Zot Registry)
Cumple con el estándar de arquitectura utilizando un registro OCI para el pull de imágenes. Utilizar la carpeta k8s-docker-images, las rutas ya estan adecuadas

```yml
containers:
      - name: 
        image:
```

```bash
# 1. Crear el registro Zot local en el puerto 5000
Se realizo en el paso 1.
k3d registry create zot-registry --port 5000 --image ghcr.io/project-zot/zot-linux-amd64:latest

# 1. Crear el clúster enlazado directamente a ese registro
k3d cluster create clima-cluster --servers 1 --agents 2 --port "8080:80@loadbalancer" --k3s-arg "--disable=traefik@server:0" --registry-use k3d-zot-registry:5000

```

> Importante: Para que esto funcione, en todos tus archivos .yaml de los Deployments y APIs creadas, el nombre de la imagen debe llevar el prefijo del registro. Ejemplo: en lugar de image: api-rust:local, debe decir image: localhost:5000/api-rust:local


### 7.2 Infraestructura Base y Message Brokers
Se desplegaron las bases de datos y los brokers dentro de un namespace dedicado llamado `weather-system`.

<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/00-namespace.yaml]
<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/01-valkey.yaml]
<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/02-rabbitmq.yaml]

**Aplicar manifiestos base:**
```bash
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-valkey.yaml
kubectl apply -f k8s/02-rabbitmq.yaml

```

**Despliegue de Kafka (Modo KRaft con Strimzi):**
Para Kafka, utilizamos el operador oficial de Strimzi en su versión más reciente, implementando la arquitectura KRaft (sin Zookeeper) para mayor eficiencia.

```bash
# Instalación del operador Strimzi
kubectl create -f 'https://strimzi.io/install/latest?namespace=weather-system' -n weather-system

```

<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/03-kafka.yaml]

**Aplicar manifiesto base:**
```bash
kubectl apply -f k8s/03-kafka.yaml

```

### 7.3 Despliegue de Microservicios
Se aplicaron los manifiestos para levantar los contenedores de Go y Rust, asignando variables de entorno para su comunicación interna y límites de recursos (CPU/Memoria).

<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/04-go-router.yaml]
<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/05-go-writers.yaml]
<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/06-go-consumers.yaml]
<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/07-api-rust.yaml]

**Aplicar manifiestos base:**
```bash
kubectl apply -f k8s/04-go-router.yaml
kubectl apply -f k8s/05-go-writers.yaml
kubectl apply -f k8s/06-go-consumers.yaml
kubectl apply -f k8s/07-api-rust.yaml

```

### 7.4 Ingress Controller (NGINX)
Para exponer la API REST al exterior, se instaló NGINX Ingress Controller y se configuró las reglas de enrutamiento.

```bash
# Instalación de NGINX Ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

```

**Aplicar manifiesto base(esperar 2 min sino marca error: no endpoints available for service "ingress-nginx-controller-admission"):**
```bash
kubectl apply -f k8s/08-ingress.yaml

```

### 7.5 Dashboard de Monitoreo (Grafana)
Para visualizar los datos almacenados en la base de datos en memoria, se desplegó Grafana. El acceso a Grafana se configuró a través del Ingress Controller en la ruta raíz /, del puerto 8080, por lo que puedes acceder a ella en localhost:8080/

<br>[COLOCAR CODIGO DEL ARCHIVO: k8s/09-grafana.yaml]

**Aplicar manifiesto:**
```bash
kubectl apply -f k8s/09-grafana.yaml

```

#### Configuración de Paneles en Grafana:
* Se configuró un Data Source apuntando al servicio interno en url  `valkey-service:6379`.
* Se crearon paneles tipo "Time series" consultando los Streams de Valkey (`tweets_rabbit` / `tweets_kafka`).
* Para optimizar la carga bajo tráfico intenso, se utilizó el comando `XREVRANGE` limitando la consulta a los últimos 500 registros.

### 7.6 Prueba End-to-End
Una vez aplicados todos los manifiestos, se verificó el flujo completo enviando un payload JSON al Ingress expuesto en el puerto 8080 local, se debe esperar 5-10 min para que todo este creado, pueden ocurrir errores con kafka que se resuelven con `kubectl delete pod -l app=kafka-consumer -n weather-system`:

```cmd
curl -X POST http://localhost:8080/tweet -H "Content-Type: application/json" -d "{\"municipality\": 1, \"temperature\": 25, \"humidity\": 60, \"weather\": 2}"

```

```pwsh
curl.exe -X POST http://localhost:8080/tweet -H "Content-Type: application/json" -d '{"municipality": 1, "temperature": 25, "humidity": 60, "weather": 2}'

```


## 8. Pruebas de Carga con Locust
Para someter la arquitectura a pruebas de estrés y validar la escalabilidad, se utilizó Locust para simular múltiples sensores enviando datos concurrentes.

**Comandos de inicialización:**
```bash
mkdir -p locust-test
cd locust-test

```

<br>[COLOCAR CODIGO DEL ARCHIVO: locust-test/locustfile.py]

**Comando de ejecución:**
```bash
pip install locust
locust -f locustfile.py

```

## 9. Comandos Útiles y Resolución de Problemas (Troubleshooting) 

En arquitecturas distribuidas y basadas en microservicios, es fundamental saber rastrear el flujo de los datos y encontrar cuellos de botella. A continuación, se presentan los comandos más útiles de `kubectl` para la administración del clúster (todos asumen el uso del namespace `weather-system`).

### 9.1 Monitoreo de Estado
Para verificar que todos los contenedores están levantados y corriendo sin errores (como `ImagePullBackOff` o `CrashLoopBackOff`):
```bash
# Ver el estado resumido de todos los Pods
kubectl get pods -n weather-system

# Ver absolutamente todos los recursos (Servicios, Deployments, Pods)
kubectl get all -n weather-system

```

### 9.2 Lectura de Logs (Rastrear el camino del dato)
Si los datos no llegan a la base de datos, se debe revisar el log de cada microservicio en orden. Se utilizan las "etiquetas" (`-l app=...`) para no tener que escribir el nombre exacto del pod generado aleatoriamente.

```bash
# 1. Ver si la API Rust recibió la petición
kubectl logs -l app=api-rust -n weather-system

# 2. Ver si el Router de Go la distribuyó por gRPC
kubectl logs -l app=go-router -n weather-system

# 3. Ver si los consumidores leyeron de los brokers y guardaron en Valkey
kubectl logs -l app=kafka-consumer -n weather-system
kubectl logs -l app=rabbitmq-consumer -n weather-system

```

### 9.3 Reinicio Forzado de Pods
Si un servicio se queda atascado (por ejemplo, si intentó conectarse a Kafka antes de que este terminara de iniciar), la mejor práctica en Kubernetes no es "apagar y prender" el clúster, sino matar el Pod defectuoso. El `Deployment` o `StatefulSet` creará uno nuevo y limpio automáticamente en segundos.

```bash
# Reiniciar todos los pods de un microservicio específico
kubectl delete pod -l app=kafka-consumer -n weather-system

# Eliminar un pod específico por su nombre (útil para bases de datos atascadas)
kubectl delete pod valkey-0 -n weather-system

```

### 9.4 Inspección Profunda de Errores
Si un Pod nunca pasa a estado `Running`, los logs normales no funcionarán. Se debe "describir" el pod para ver los eventos internos de Kubernetes (errores de red, falta de memoria, imagen no encontrada):

```bash
kubectl describe pod <nombre-exacto-del-pod> -n weather-system

```

### 9.5 Túneles Directos (Port-Forwarding)
Si el Ingress Controller falla o si se necesita acceder temporalmente a un servicio interno que no está expuesto al público (como una base de datos o un dashboard de administración), se puede abrir un túnel seguro directo desde la máquina local hasta el clúster:

```bash
# Sintaxis: kubectl port-forward svc/<nombre-servicio> <puerto-local>:<puerto-contenedor> -n <namespace>

# Ejemplo: Acceder a Grafana de forma privada (bypass del Ingress)
kubectl port-forward svc/grafana-service 3000:3000 -n weather-system

```

>*(El túnel se mantiene abierto mientras la terminal siga corriendo. Se cierra con `Ctrl+C`).*

### 9.6 Destrucción Total del Entorno (Tear Down)
Durante la fase de desarrollo, es muy común que la base de datos se corrompa por pruebas mal hechas o que se necesite empezar desde un lienzo en blanco. En lugar de borrar pod por pod, la mejor práctica es destruir el clúster local por completo.

#### Eliminar el clúster de Kubernetes:
Esto apagará y borrará todos los contenedores, redes y volúmenes asociados a K3s.
```bash
k3d cluster delete clima-cluster

# Detener
k3d cluster stop clima-cluster

# Iniciar
k3d cluster start clima-cluster

```
#### Eliminar el registro de imágenes (Zot):
Si utilizaste la Opción B para levantar el servidor Zot administrado por K3d, debes eliminarlo por separado:

```Bash
k3d registry delete k3d-zot-registry
```

> (Nota: Después de ejecutar estos comandos, tu máquina quedará completamente limpia de los recursos de este proyecto. Para volver a empezar, simplemente debes regresar a los comandos del paso 7.1).