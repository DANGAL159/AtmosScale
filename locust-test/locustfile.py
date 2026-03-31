# locustfile.py: Script de carga para simular sensores de clima enviando datos a tu API en Rust
from locust import HttpUser, task, between
import random

class WeatherSensor(HttpUser):
    # Tiempo de espera entre peticiones de un mismo usuario (muy rápido)
    wait_time = between(0.1, 0.5)

    @task
    def send_weather_tweet(self):
        # Generamos datos climáticos aleatorios para darle vida a Grafana
        payload = {
            "municipality": random.randint(1, 10), # Simulamos 10 municipios diferentes
            "temperature": random.randint(10, 38), # Temperatura entre 10°C y 38°C
            "humidity": random.randint(30, 95),    # Humedad entre 30% y 95%
            "weather": random.randint(1, 4)        # 1: Soleado, 2: Lluvioso, etc.
        }
        
        # Hacemos el POST a tu API en Rust a través del Ingress
        self.client.post("/tweet", json=payload)