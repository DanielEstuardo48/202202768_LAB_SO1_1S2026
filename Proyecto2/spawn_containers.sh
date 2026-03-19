#!/bin/bash

echo "[$(date)] Lanzando contenedores..."

# Siempre lanza al menos 1 go-client (alto RAM)
docker run -d --name "so1_$(date +%s%N)" roldyoran/go-client

# Siempre lanza al menos 1 alpine CPU (alto CPU)
docker run -d --name "so1_$(date +%s%N)" alpine sh -c \
    "while true; do echo '2^20' | bc > /dev/null; sleep 2; done"

# Siempre lanza al menos 1 alpine sleep (bajo consumo)
docker run -d --name "so1_$(date +%s%N)" alpine sleep 240

# Los últimos 2 son aleatorios
for i in 1 2; do
    CHOICE=$((RANDOM % 3))
    case $CHOICE in
        0) docker run -d --name "so1_$(date +%s%N)" roldyoran/go-client ;;
        1) docker run -d --name "so1_$(date +%s%N)" alpine sh -c \
               "while true; do echo '2^20' | bc > /dev/null; sleep 2; done" ;;
        2) docker run -d --name "so1_$(date +%s%N)" alpine sleep 240 ;;
    esac
    sleep 0.2
done

echo "[$(date)] Listo."