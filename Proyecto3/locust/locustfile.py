from locust import HttpUser, task, between
import random
import json
from datetime import datetime, timezone

COUNTRIES = ["USA", "RUS", "CHN", "ESP", "GTM"] # Lista de países para los reportes militares

class MilitaryReportUser(HttpUser):
    # Espera entre 0.5 y 2 segundos entre requests
    wait_time = between(0.5, 2)
    
    @task
    def send_report(self):
        # Genera un reporte militar con datos aleatorios
        payload = {
            "country": random.choice(COUNTRIES),
            "warplanes_in_air": random.randint(0, 50),
            "warships_in_water": random.randint(0, 30),
            "timestamp": datetime.now(timezone.utc).isoformat()
        }

        self.client.post(
            "/grpc-202202768",
            json=payload,
            headers={"Content-Type": "application/json"}
        )