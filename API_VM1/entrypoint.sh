#!/bin/sh
# Iniciar API1 en segundo plano
./api1 &
# Iniciar API2 en segundo plano
./api2 &

# Esperar para que el contenedor no se cierre
wait -n