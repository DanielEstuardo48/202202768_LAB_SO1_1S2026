#!/bin/bash
MODULE_PATH="/ruta/al/modulo/continfo.ko"
MODULE_NAME="continfo"

# Si ya está cargado, no hacer nada
if lsmod | grep -q "$MODULE_NAME"; then
    echo "Módulo ya cargado."
    exit 0
fi

insmod "$MODULE_PATH"
echo "Módulo cargado exitosamente."