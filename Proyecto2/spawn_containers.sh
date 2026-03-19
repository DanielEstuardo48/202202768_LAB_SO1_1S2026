#!/bin/bash

# Imagenes
IMAGES=(
    "roldyoran/go-client"   # Alto consumo RAM
    "alpine_cpu"            # Alto consumo CPU (imagen local)
    "alpine sleep 240"      # Bajo consumo
)

# Lanzar un contenedor aleatorio
launch_random() {
    CHOICE=$((RANDOM % 3))
    case $CHOICE in
        0)
            docker run -d --name "so1_$(date +%s%N)" roldyoran/go-client
            ;;
        1)
            docker run -d --name "so1_$(date +%s%N)" alpine sh -c \
                "while true; do echo '2^20' | bc > /dev/null; sleep 2; done"
            ;;
        2)
            docker run -d --name "so1_$(date +%s%N)" alpine sleep 240
            ;;
    esac
}

echo "[$(date)] Lanzando 5 contenedor..."
for i in $(seq 1 5); do
    launch_random
    sleep 0.2   # Pequeña pausa para evitar sobrecargar el sistema de inmediato
done
echo "[$(date)] Contenedores lanzados."