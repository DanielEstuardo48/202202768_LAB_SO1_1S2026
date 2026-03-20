# Manual de Instalación — Proyecto 2
## Sonda de Kernel en C y Daemon en Go para la Telemetría de Contenedores
**Sistemas Operativos 1 — Universidad San Carlos de Guatemala**  
**Facultad de Ingeniería — Ingeniería en Ciencias y Sistemas**  
**Nombre: Daniel Estuardo Salvatierra Macajola**  
**Carnet: 202202768**

---

## Índice

1. [Requisitos Previos](#1-requisitos-previos)
2. [Preparación de la Máquina Virtual](#2-preparación-de-la-máquina-virtual)
3. [Instalación de Dependencias](#3-instalación-de-dependencias)
4. [Clonar el Repositorio](#4-clonar-el-repositorio)
5. [Compilación del Módulo de Kernel](#5-compilación-del-módulo-de-kernel)
6. [Configuración de Scripts](#6-configuración-de-scripts)
7. [Compilación del Daemon en Go](#7-compilación-del-daemon-en-go)
8. [Configuración de Docker Compose](#8-configuración-de-docker-compose)
9. [Instalación y Configuración de Grafana](#9-instalación-y-configuración-de-grafana)
10. [Configuración del Dashboard](#10-configuración-del-dashboard)
11. [Verificación de la Instalación](#11-verificación-de-la-instalación)
12. [Ejecución del Sistema Completo](#12-ejecución-del-sistema-completo)
13. [Detención del Sistema](#13-detención-del-sistema)

---

## 1. Requisitos Previos

Antes de iniciar la instalación, verificar que se cuenta con lo siguiente:

### Hardware mínimo requerido
| Recurso | Mínimo | Recomendado |
|---|---|---|
| CPU | 2 núcleos | 4 núcleos |
| RAM | 4 GB | 6 GB |
| Almacenamiento | 25 GB | 30 GB |
| Conexión a internet | Requerida | Requerida |

### Sistema Operativo
- **Ubuntu Server 22.04 LTS** (probado y recomendado)
- Kernel Linux 5.15 o superior

Verificar la versión del kernel:
```bash
uname -r
```

### Acceso requerido
- Usuario con privilegios `sudo`
- Acceso a internet para descargar imágenes Docker y paquetes

---

## 2. Preparación de la Máquina Virtual

Si se está usando una máquina virtual (VirtualBox, VMware, etc.), configurar lo siguiente antes de instalar:

### Configuración de red
Usar **Adaptador Puente (Bridged)** o **NAT** para tener acceso a internet desde la VM.

### Espacio en disco
Asegurarse de que el disco virtual tenga al menos **25 GB** de espacio libre:
```bash
df -h
```

### Actualizar el sistema operativo
```bash
sudo apt update && sudo apt upgrade -y
sudo reboot
```

Después del reinicio, continuar con el paso 3.

---

## 3. Instalación de Dependencias

Instalar todos los paquetes necesarios en el siguiente orden:

### 3.1 Herramientas de compilación del kernel
```bash
sudo apt install -y linux-headers-$(uname -r) build-essential
```

Verificar que los headers se instalaron correctamente:
```bash
ls /lib/modules/$(uname -r)/build
```
Debe mostrar el directorio de compilación del kernel. Si no existe, la compilación del módulo fallará.

### 3.2 Docker
```bash
# Instalar Docker
sudo apt install -y docker.io

# Habilitar e iniciar el servicio
sudo systemctl enable docker
sudo systemctl start docker

# Verificar instalación
docker --version
```

Salida esperada: `Docker version 24.x.x`

### 3.3 Docker Compose
```bash
# Instalar Docker Compose plugin
sudo apt install -y docker-compose-plugin

# Verificar instalación
docker compose version
```

Salida esperada: `Docker Compose version v2.x.x`

### 3.4 Go (Golang)

Descargar e instalar Go 1.21 o superior:
```bash
# Descargar Go (ajustar versión si es necesario)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# Extraer en /usr/local
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Configurar PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verificar instalación
go version
```

Salida esperada: `go version go1.21.0 linux/amd64`

### 3.5 Git
```bash
sudo apt install -y git

# Verificar
git --version
```

### 3.6 bc (calculadora para contenedores de prueba)
```bash
sudo apt install -y bc
```

---

## 4. Clonar el Repositorio

```bash
# Clonar el repositorio del proyecto
git clone https://github.com/DanielEstuardo48/202202768_LAB_SO1_1S2026.git

# Ingresar a la carpeta del proyecto 2
cd 202202768_LAB_SO1_1S2026/proyecto2

# Verificar los archivos presentes
ls -la
```

Debe mostrar los siguientes archivos:
```
continfo.c
Makefile
docker-compose.yml
load_module.sh
spawn_containers.sh
daemon/
```

> **Nota:** Si los archivos están directamente en el home (`~/`) en lugar de en una carpeta del repositorio, ajustar las rutas de los pasos siguientes reemplazando la ruta del repositorio por `~/`.

---

## 5. Compilación del Módulo de Kernel

### 5.1 Compilar
Desde la carpeta donde está `continfo.c` y el `Makefile`:
```bash
make
```

Salida esperada al finalizar:
```
make -C /lib/modules/x.x.x-generic/build M=/home/danii modules
  CC [M]  /home/danii/continfo.o
  MODPOST /home/danii/Module.symvers
  CC [M]  /home/danii/continfo.mod.o
  LD [M]  /home/danii/continfo.ko
```

Verificar que se generó el archivo `.ko`:
```bash
ls -la continfo.ko
```

### 5.2 Cargar el módulo por primera vez
```bash
sudo insmod continfo.ko
```

### 5.3 Verificar que el módulo está cargado
```bash
lsmod | grep continfo
```

Salida esperada: `continfo   xxxxx  0`

### 5.4 Verificar el archivo en /proc
```bash
cat /proc/continfo_pr2_so1_202202768
```

Debe mostrar un JSON con información del sistema y los procesos activos.

### 5.5 Verificar el log del kernel
```bash
sudo dmesg | tail -5
```

Debe aparecer: `continfo: Módulo cargado. Leyendo en /proc/continfo_pr2_so1_202202768`

---

## 6. Configuración de Scripts

### 6.1 Dar permisos de ejecución
```bash
chmod +x load_module.sh spawn_containers.sh
```

### 6.2 Verificar las rutas en `load_module.sh`
Abrir el archivo y confirmar que la ruta al módulo es correcta:
```bash
cat load_module.sh
```

La línea `insmod` debe apuntar a la ubicación real de `continfo.ko`. Si los archivos están en el home:
```bash
# Editar si es necesario
nano load_module.sh
# Verificar que dice: insmod /home/danii/continfo.ko
```

### 6.3 Verificar las rutas en `spawn_containers.sh`
```bash
cat spawn_containers.sh
```

No requiere rutas específicas ya que solo ejecuta comandos `docker run`.

### 6.4 Probar el script de contenedores manualmente
```bash
sudo bash spawn_containers.sh

# Verificar que se crearon contenedores
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
```

Deben aparecer 5 contenedores con prefijo `so1_`.

Limpiar los contenedores de prueba:
```bash
docker rm -f $(docker ps -a --filter "name=so1_" -q)
```

---

## 7. Compilación del Daemon en Go

### 7.1 Ingresar a la carpeta del daemon
```bash
cd daemon/
```

### 7.2 Verificar el archivo `config/config.go`
Confirmar que las rutas son correctas para tu entorno:
```bash
cat config/config.go
```

Verificar especialmente:
- `ProcFile`: debe coincidir con `/proc/continfo_pr2_so1_202202768`
- `CronScript`: debe ser la ruta absoluta a `spawn_containers.sh`

### 7.3 Descargar dependencias
```bash
go mod tidy
```

Este comando descarga todas las dependencias listadas en `go.mod`. Requiere conexión a internet. Puede tardar 1-2 minutos.

### 7.4 Compilar el daemon
```bash
go build -o daemon_so1 .
```

Verificar que se generó el binario:
```bash
ls -la daemon_so1
```

### 7.5 Regresar al directorio principal
```bash
cd ..
```

---

## 8. Configuración de Docker Compose

### 8.1 Verificar el archivo `docker-compose.yml`
```bash
cat docker-compose.yml
```

Debe contener los servicios `valkey` y `grafana` en la red `so1_network`.

### 8.2 Levantar los servicios
```bash
docker compose up -d
```

Salida esperada:
```
✔ Container valkey_so1    Started
✔ Container grafana_so1   Started
```

### 8.3 Verificar que los contenedores están corriendo
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

Debe mostrar:
```
NAMES          STATUS         PORTS
grafana_so1    Up x seconds   0.0.0.0:3000->3000/tcp
valkey_so1     Up x seconds   0.0.0.0:6379->6379/tcp
```

### 8.4 Verificar conectividad con Valkey
```bash
docker exec -it valkey_so1 valkey-cli PING
```

Debe responder: `PONG`

---

## 9. Instalación y Configuración de Grafana

### 9.1 Instalar el plugin de Redis/Valkey
```bash
docker exec -it grafana_so1 grafana-cli plugins install redis-datasource
```

Salida esperada:
```
✔ Downloaded and extracted redis-datasource v4.x.x zip successfully
Please restart Grafana after installing or removing plugins.
```

### 9.2 Reiniciar Grafana para activar el plugin
```bash
docker restart grafana_so1
```

Esperar 10 segundos y verificar que está corriendo:
```bash
docker ps | grep grafana
```

### 9.3 Acceder a Grafana
Abrir un navegador y navegar a:
```
http://localhost:3000
```

Ingresar con:
- **Usuario**: `admin`
- **Contraseña**: `admin`

En el primer acceso, Grafana pedirá cambiar la contraseña. Se puede omitir con **Skip**.

### 9.4 Configurar el Data Source
1. En el menú izquierdo ir a **Connections → Data Sources**
2. Clic en **Add data source**
3. Buscar y seleccionar **Redis**
4. Configurar los campos:
   - **Name**: `Valkey-SO1`
   - **Address**: `valkey_so1:6379`
5. Clic en **Save & Test**

Debe aparecer el mensaje: **Data source connected and labels found**

> **Importante:** Usar `valkey_so1` como host (no `localhost`), ya que Grafana se comunica con Valkey a través de la red interna de Docker.

---

## 10. Configuración del Dashboard

### 10.1 Crear un nuevo dashboard
1. En el menú izquierdo ir a **Dashboards → New Dashboard**
2. Clic en **Add visualization**

### 10.2 Paneles a crear

Crear los siguientes 7 paneles con la configuración indicada:

#### Panel 1 — Total RAM
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Título | Total RAM |
| Data source | Valkey-SO1 |
| Command | GET |
| Key | `so1:ram_total` |
| Unit | megabytes (MB) |

#### Panel 2 — RAM Libre
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Título | Free RAM |
| Data source | Valkey-SO1 |
| Command | GET |
| Key | `so1:ram_free` |
| Unit | megabytes (MB) |

#### Panel 3 — RAM Usada
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Título | RAM Usada |
| Data source | Valkey-SO1 |
| Command | GET |
| Key | `so1:ram_used` |
| Unit | megabytes (MB) |

#### Panel 4 — Uso de RAM en el tiempo
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Graph mode | Area |
| Título | Uso de RAM en el tiempo |
| Data source | Valkey-SO1 |
| Command | GET |
| Key | `so1:ram_used` |
| Streaming | ON |

#### Panel 5 — Total Contenedores Eliminados
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Título | Total Contenedores Eliminados |
| Data source | Valkey-SO1 |
| Command | LLEN |
| Key | `so1:eliminated` |

#### Panel 6 — Top 5 RAM
| Campo | Valor |
|---|---|
| Tipo | Stat |
| Título | Top 5 RAM |
| Data source | Valkey-SO1 |
| Queries | 5 queries: GET `so1:top_ram:1:rss` hasta `so1:top_ram:5:rss` |

#### Panel 7 — Top 5 CPU
| Campo | Valor |
|---|---|
| Tipo | Pie chart |
| Título | Top 5 CPU |
| Data source | Valkey-SO1 |
| Queries | 5 queries: GET `so1:top_cpu:1:cpu` hasta `so1:top_cpu:5:cpu` |

### 10.3 Guardar el dashboard
1. Clic en el ícono de guardar, arriba a la derecha
2. Nombre: `Panel de Contenedores - 202202768`
3. Clic en **Save**

### 10.4 Configurar auto-refresh
En el dashboard guardado, arriba a la derecha junto al botón **Refresh**:
1. Clic en la flecha desplegable
2. Seleccionar **10s**

---

## 11. Verificación de la Instalación

Ejecutar las siguientes verificaciones para confirmar que todo está correctamente instalado:

```bash
# 1. Módulo de kernel cargado
lsmod | grep continfo
echo "✓ Módulo de kernel" || echo "✗ Módulo de kernel"

# 2. Archivo /proc accesible
cat /proc/continfo_pr2_so1_202202768 | head -5
echo "✓ Archivo /proc" || echo "✗ Archivo /proc"

# 3. Docker corriendo
docker ps | grep -E "grafana_so1|valkey_so1"
echo "✓ Contenedores base" || echo "✗ Contenedores base"

# 4. Valkey responde
docker exec -it valkey_so1 valkey-cli PING
echo "✓ Valkey" || echo "✗ Valkey"

# 5. Binario del daemon existe
ls ~/daemon/daemon_so1
echo "✓ Daemon compilado" || echo "✗ Daemon compilado"

# 6. Grafana accesible
curl -s http://localhost:3000/api/health | grep -q "ok"
echo "✓ Grafana" || echo "✗ Grafana"
```

Todos deben mostrar `✓`. Si alguno falla, revisar el paso correspondiente en este manual.

---

## 12. Ejecución del Sistema Completo

Una vez completada la instalación, el sistema se ejecuta con un único comando:

```bash
cd ~/daemon
sudo ./daemon_so1
```

El daemon se encarga automáticamente de:
- Verificar que Grafana y Valkey estén corriendo
- Cargar el módulo de kernel si no está cargado
- Registrar el cronjob de creación de contenedores
- Iniciar el loop de monitorización cada 20 segundos

### Salida esperada al iniciar
```
=== Daemon SO1 Proyecto 2 iniciando ===
[1] Levantando Grafana y Valkey...
[2] Cargando módulo de kernel...
[3] Registrando cronjob...
[✓] Daemon corriendo. Ctrl+C para detener.

--- Ciclo 10:30:00 ---
RAM: 2562 MB usada / 5363 MB total
  [ALTO]  so1_1773894001 (roldyoran/go-client)
  [ALTO]  so1_1773894002 (alpine)
  [BAJO]  so1_1773894003 (alpine)
  [BAJO]  so1_1773894004 (alpine)
  [BAJO]  so1_1773894005 (alpine)
Total: 5 contenedores (2 alto, 3 bajo)
Métricas guardadas en Valkey.
```

### Verificar datos en Grafana
1. Abrir `http://localhost:3000`
2. Ir al dashboard **Panel de Contenedores - 202202768**
3. Los paneles deben mostrar datos después del primer ciclo (20 segundos)
4. Verificar que el auto-refresh está activo (10s)

---

## 13. Detención del Sistema

### Detención normal
```bash
# En la terminal del daemon, presionar:
Ctrl + C
```

El daemon elimina automáticamente el cronjob al recibir la señal de terminación.

### Descargar el módulo de kernel
```bash
sudo rmmod continfo

# Verificar que se descargó
lsmod | grep continfo
# No debe mostrar nada
```

### Detener Grafana y Valkey
```bash
docker compose down
```

### Limpiar contenedores de prueba (opcional)
```bash
docker rm -f $(docker ps -a --filter "name=so1_" -q)
```

### Verificar que el cronjob fue eliminado
```bash
sudo crontab -l
# No debe aparecer la línea de spawn_containers.sh
```

---

*Manual de Instalación — Proyecto 2 — Sistemas Operativos 1 — 1S 2026*
