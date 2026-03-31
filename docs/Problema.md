# Caso de Estudio: Sistema de Telemetría Meteorológica Nacional

## Contexto del Negocio:
El Instituto Nacional de Meteorología está desplegando una nueva red de dispositivos IoT (Internet de las Cosas) en todos los municipios del país. Estos sensores están programados para emitir lecturas de temperatura, humedad y estado del clima en tiempo real. 

## El Problema Actual:
Actualmente, el instituto cuenta con un sistema monolítico tradicional. Durante condiciones climáticas normales, el sistema funciona adecuadamente. Sin embargo, durante eventos climáticos extremos (como tormentas o frentes fríos), los sensores aumentan su frecuencia de transmisión para emitir alertas. Esto genera picos masivos de tráfico concurrente que el servidor central no puede soportar, provocando:

1. **Pérdida de datos críticos:** El servidor rechaza conexiones por saturación, perdiendo lecturas vitales en el momento que más se necesitan.
2. **Caída del sistema (Downtime):** Si la base de datos se satura escribiendo la información, todo el sistema de recepción colapsa.
3. **Latencia en la toma de decisiones:** Los tableros de visualización tardan minutos en actualizarse porque consultan directamente bases de datos transaccionales lentas, impidiendo que las autoridades emitan alertas tempranas a la población.

## Requerimientos del Nuevo Sistema:
La junta directiva solicita el diseño de una nueva arquitectura distribuida que cumpla con los siguientes objetivos:

1. **Alta Concurrencia en la Entrada:** El punto de entrada del sistema debe ser capaz de recibir miles de peticiones HTTP por segundo sin degradar su rendimiento, actuando como un embudo de alta velocidad.
2. **Tolerancia a Fallos y Cero Pérdida de Datos:** El sistema debe implementar un mecanismo de amortiguación (buffers/colas). Si la base de datos se cae o se reinicia, los datos de los sensores no deben perderse, sino retenerse temporalmente hasta que el sistema de almacenamiento vuelva a estar en línea.
3. **Procesamiento Asíncrono:** La recepción del dato debe estar totalmente separada de su escritura. El sensor debe recibir un mensaje de "Éxito" en milisegundos, sin tener que esperar a que el dato se guarde en disco.
4. **Visualización en Tiempo Real:** Se requiere un tablero de mando dinámico que lea la información con latencias de sub-milisegundos, permitiendo ver las fluctuaciones de temperatura y humedad al instante.
5. **Resiliencia y Auto-recuperación:** La infraestructura debe ser capaz de detectar si un componente interno falla, aislar el error y reiniciar dicho componente automáticamente sin intervención humana.

---

## Traducción de los requerimientos a la arquitectura

* *"Alta concurrencia en la entrada"* = Necesito un lenguaje de bajo nivel muy rápido (Rust) y un Ingress Controller.
* *"Tolerancia a fallos y amortiguación"* = Necesito Message Brokers (Kafka / RabbitMQ).
* *"Procesamiento asíncrono"* = Necesito separar en microservicios y usar RPC rápido (Go + gRPC).
* *"Visualización en sub-milisegundos"* = Necesito una base de datos en memoria (Redis/Valkey) y Grafana.
* *"Auto-recuperación"* = Necesito orquestar con Kubernetes.

