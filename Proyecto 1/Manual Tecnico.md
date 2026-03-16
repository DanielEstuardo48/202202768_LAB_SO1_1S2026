<div align="center">

**Universidad de San Carlos de Guatemala**  
**Facultad de Ingenieria**  
**Sistemas Operativos 1 - P**  
**Nombre: Daniel Estuardo Salvatierra Macajola**  
**Carnet: 202202768**  
**Guatemala 17 de Febrero 2026**  

</div>

<h1 align="center">Proyecto 1 - Manual Técnico</h1>

## Tabla de contenido
1. [Descripción General del Proyecto](#1-descripción-general-del-proyecto)
2. [Arquitectura del Sistema](#2-arquitectura-del-sistema)
3. [Tecnologías y Herramientas Utilizadas](#3-tecnologías-y-herramientas-utilizadas)
4. [Estructura de las Máquinas Virtuales](#4-estructura-de-las-máquinas-virtuales)
5. [Descripción de las APIs](#5-descripción-de-las-apis)
6. [Configuración de Red](#6-configuración-de-red)
7. [Registro Privado de Contenedores (Zot)](#7-registro-privado-de-contenedores-zot)
8. [Flujo de Despliegue](#8-flujo-de-despliegue)
9. [Comunicación entre Servicios](#9-comunicación-entre-servicios)
10. [Mantenimiento y Solución de Problemas](#10-mantenimiento-y-solución-de-problemas)

---

## 1. Descripción General del Proyecto

Este proyecto consiste en el despliegue de tres APIs distribuidas en un entorno virtualizado utilizando KVM como hipervisor. Las APIs se ejecutan como contenedores sobre máquinas virtuales con distintos runtimes de contenedores (Containerd y Docker), y se comunican entre sí a través de la red virtual proporcionada por KVM.

El objetivo principal es demostrar la interoperabilidad entre servicios contenerizados en diferentes máquinas virtuales, utilizando un registro privado de imágenes (Zot) alojado en una tercera VM para la distribución de los contenedores.

### Resumen del flujo
1. Las imágenes de las APIs se construyen con Docker en VM3.
2. Las imágenes se publican en el registro privado Zot (también en VM3).
3. VM1 y VM2 descargan las imágenes desde Zot y las ejecutan con Containerd/Nerdctl.
4. Las tres APIs se comunican entre sí a través de la red virtual de KVM.

---

## 2. Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────┐
│                      HOST                           │
│              Ubuntu 22.04 LTS                       │
│                   KVM / libvirt                     │
│                                                     │
│  ┌──────────────┐ ┌──────────────┐ ┌─────────────┐ │
│  │     VM1      │ │     VM2      │ │     VM3     │ │
│  │192.168.122.90│ │192.168.122.79│ │192.168.122. │ │
│  │              │ │              │ │    109      │ │
│  │  Containerd  │ │  Containerd  │ │   Docker    │ │
│  │  API1 :8881  │ │  API3 :8883  │ │  Zot :5000  │ │
│  │  API2 :8882  │ │              │ │             │ │
│  └──────┬───────┘ └──────┬───────┘ └──────┬──────┘ │
│         └────────────────┴────────────────┘        │
│                   virbr0 (NAT Bridge)               │
│                  10.0.0.0 / 192.168.122.0/24        │
└─────────────────────────────────────────────────────┘
```

### Componentes principales

| Componente | Descripción |
|------------|-------------|
| KVM + libvirt | Hipervisor tipo 1 para la virtualización del hardware |
| Ubuntu Server 22.04 LTS | Sistema operativo de las VMs guest |
| Containerd + Nerdctl | Runtime de contenedores en VM1 y VM2 |
| Docker Engine | Motor de contenedores en VM3 para build y push |
| Zot Registry | Registro privado OCI para distribución de imágenes |
| CNI Plugins | Plugins de red para Containerd |
| runc | Runtime de bajo nivel para los contenedores |

---

## 3. Tecnologías y Herramientas Utilizadas

### 3.1 KVM (Kernel-based Virtual Machine)
KVM es un módulo del kernel de Linux que permite al sistema operativo funcionar como hipervisor tipo 1. Utiliza las extensiones de virtualización del procesador (Intel VT-x / AMD-V) para ejecutar máquinas virtuales con rendimiento cercano al nativo.

- **Versión requerida**: Kernel Linux con soporte KVM
- **Paquetes**: `qemu-kvm`, `libvirt-daemon-system`, `libvirt-clients`, `bridge-utils`, `virt-manager`
- **Gestión**: A través de `virsh` (CLI) o `virt-manager` (GUI)

### 3.2 Containerd
Containerd es un runtime de contenedores de nivel industrial compatible con la especificación OCI (Open Container Initiative). Es el runtime subyacente de Docker y es usado directamente en VM1 y VM2 para mayor control y eficiencia.

- **Versión**: Instalada desde repositorios APT de Ubuntu
- **Configuración clave**: `SystemdCgroup = true` para integración con systemd
- **Archivo de configuración**: `/etc/containerd/config.toml`

### 3.3 Nerdctl
Nerdctl es una herramienta CLI compatible con Docker que permite interactuar directamente con Containerd. Provee comandos similares a Docker (`pull`, `run`, `ps`, `logs`) sin necesitar el daemon de Docker.

- **Versión utilizada**: `1.7.3`
- **Repositorio**: `github.com/containerd/nerdctl`

### 3.4 runc
runc es el runtime de contenedores de bajo nivel (OCI runtime) que se encarga de la ejecución real de los contenedores a nivel de sistema operativo. Es invocado por Containerd de forma transparente.

### 3.5 CNI Plugins (Container Network Interface)
Los plugins CNI proveen conectividad de red a los contenedores gestionados por Containerd. Se configura una red tipo bridge (`cni0`) con soporte para IP Masquerade y port mappings.

- **Versión**: `1.4.0`
- **Directorio de instalación**: `/opt/cni/bin`
- **Configuración de red**: `/etc/cni/net.d/10-containerd-net.conflist`
- **Subnet configurada**: `10.88.0.0/16`

### 3.6 Docker Engine
Docker se utiliza únicamente en VM3 para construir las imágenes de los contenedores de las APIs y publicarlas en el registro Zot. No se usa como runtime de producción.

- **Paquetes**: `docker-ce`, `docker-ce-cli`, `containerd.io`, `docker-buildx-plugin`, `docker-compose-plugin`
- **Repositorio oficial**: `https://download.docker.com/linux/ubuntu`

### 3.7 Zot Registry
Zot es un registro de imágenes OCI ligero, de código abierto y compatible con la especificación de distribución OCI. Se ejecuta como contenedor Docker en VM3 y actúa como registro privado para toda la infraestructura.

- **Imagen**: `ghcr.io/project-zot/zot:latest`
- **Puerto expuesto**: `5000`
- **Almacenamiento**: Volumen Docker `zot-data` montado en `/var/lib/zot`
- **API compatible**: Docker Registry API v2

---

## 4. Estructura de las Máquinas Virtuales

### 4.1 Especificaciones de cada VM

| Parámetro | Valor recomendado |
|-----------|-------------------|
| RAM | 2 GB por VM (6 GB total) |
| vCPUs | 2 por VM |
| Disco | 10 GB por VM |
| SO Guest | Ubuntu Server 22.04 LTS |
| Red | Red NAT de libvirt (virbr0) |
| Acceso remoto | OpenSSH Server |

### 4.2 Asignación de IPs (entorno de desarrollo)

| VM | IP Asignada (DHCP) | Usuario SSH |
|----|-------------------|-------------|
| VM1 | 192.168.122.90 | danni1 |
| VM2 | 192.168.122.79 | danni2 |
| VM3 | 192.168.122.109 | danni3 |

> **Nota**: Las IPs son asignadas por DHCP mediante la red virtual de libvirt. En entornos de producción se recomienda asignar IPs estáticas para garantizar estabilidad en la comunicación entre VMs.

### 4.3 Software instalado por VM

**VM1 y VM2 (nodos de ejecución):**
- Ubuntu Server 22.04 LTS
- Containerd (con `SystemdCgroup = true`)
- runc
- Nerdctl v1.7.3
- CNI Plugins v1.4.0

**VM3 (nodo de build y registro):**
- Ubuntu Server 22.04 LTS
- Docker Engine (docker-ce)
- Docker Buildx Plugin
- Docker Compose Plugin
- Zot Registry (contenedor Docker)

---

## 5. Descripción de las APIs

El proyecto contiene tres APIs independientes que se despliegan en las VMs. Cada una expone un endpoint de health y endpoints de comunicación hacia las otras APIs.

### 5.1 API1

| Atributo | Valor |
|----------|-------|
| VM de despliegue | VM1 |
| Puerto | `:8881` |
| Imagen en Zot | `localhost:5000/api1-202202768:v1` |
| Runtime | Containerd / Nerdctl |

**Variables de entorno:**

| Variable | Valor |
|----------|-------|
| `PORT` | `:8881` |
| `API2_URL` | `http://192.168.122.90:8882` |
| `API3_URL` | `http://192.168.122.79:8883` |

**Endpoints disponibles:**
```
GET /                              → Health check
GET /api1/202202768/call-api2      → Llama a API2 y retorna su respuesta
GET /api1/202202768/call-api3      → Llama a API3 y retorna su respuesta
```

### 5.2 API2

| Atributo | Valor |
|----------|-------|
| VM de despliegue | VM1 |
| Puerto | `:8882` |
| Imagen en Zot | `localhost:5000/api2-202202768:v1` |
| Runtime | Containerd / Nerdctl |

**Variables de entorno:**

| Variable | Valor |
|----------|-------|
| `PORT` | `:8882` |
| `API1_URL` | `http://192.168.122.90:8881` |
| `API3_URL` | `http://192.168.122.79:8883` |

**Endpoints disponibles:**
```
GET /                              → Health check
GET /api2/202202768/call-api1      → Llama a API1 y retorna su respuesta
GET /api2/202202768/call-api3      → Llama a API3 y retorna su respuesta
```

### 5.3 API3

| Atributo | Valor |
|----------|-------|
| VM de despliegue | VM2 |
| Puerto | `:8883` |
| Imagen en Zot | `localhost:5000/api3-202202768:v1` |
| Runtime | Containerd / Nerdctl |

**Variables de entorno:**

| Variable | Valor |
|----------|-------|
| `PORT` | `:8883` |
| `API1_URL` | `http://192.168.122.90:8881` |
| `API2_URL` | `http://192.168.122.90:8882` |

**Endpoints disponibles:**
```
GET /                              → Health check
GET /api3/202202768/call-api1      → Llama a API1 y retorna su respuesta
GET /api3/202202768/call-api2      → Llama a API2 y retorna su respuesta
```

### 5.4 Formato de respuesta del health check
```json
{
  "VM": "VM1",
  "carnet": "202202768",
  "message": "API1 is Ready",
  "status": "UP",
  "timestamp": "2026-02-17T19:00:00Z"
}
```

### 5.5 Formato de respuesta de comunicación entre APIs
```json
{
  "apiname": "API2",
  "carnet": "202202768",
  "connection": true,
  "message": "The API2 located on the VM1 is working"
}
```

---

## 6. Configuración de Red

### 6.1 Red Virtual de KVM (virbr0)
libvirt crea automáticamente una red NAT llamada `default` que utiliza el bridge `virbr0`. Todas las VMs se conectan a esta red y obtienen IPs en el rango `192.168.122.0/24` mediante el servidor DHCP integrado de libvirt.

```
Host (virbr0): 192.168.122.1
DHCP range:    192.168.122.2 - 192.168.122.254
VM1:           192.168.122.90
VM2:           192.168.122.79
VM3:           192.168.122.109
```

### 6.2 Red de Contenedores (CNI Bridge en VM1 y VM2)
Containerd utiliza CNI para proveer red a los contenedores. La configuración del bridge `cni0` se realiza en `/etc/cni/net.d/10-containerd-net.conflist`:

```json
{
  "cniVersion": "1.0.0",
  "name": "containerd-net",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "cni0",
      "isGateway": true,
      "ipMasq": true,
      "promiscMode": true,
      "ipam": {
        "type": "host-local",
        "ranges": [
          [{ "subnet": "10.88.0.0/16" }]
        ],
        "routes": [
          { "dst": "0.0.0.0/0" }
        ]
      }
    },
    {
      "type": "portmap",
      "capabilities": {"portMappings": true}
    }
  ]
}
```

> Los contenedores en VM1 y VM2 se ejecutan con `--network host` para exponer sus puertos directamente en la interfaz de la VM y facilitar la comunicación entre las APIs.

### 6.3 Registro inseguro de Docker (VM3)
Para que Docker en VM3 pueda hacer push al registro Zot local, se configura como registro inseguro en `/etc/docker/daemon.json`:
```json
{
  "insecure-registries": ["192.168.122.109:5000"]
}
```

---

## 7. Registro Privado de Contenedores (Zot)

### 7.1 Descripción
Zot actúa como intermediario entre el proceso de construcción de imágenes (en VM3 con Docker) y el despliegue en los nodos de ejecución (VM1 y VM2 con Containerd). Expone la API estándar de Docker Registry v2, por lo que es compatible tanto con `docker push` como con `nerdctl pull`.

### 7.2 Ejecución de Zot
```bash
sudo docker run -d \
  -p 5000:5000 \
  --name zot \
  -v zot-data:/var/lib/zot \
  ghcr.io/project-zot/zot:latest
```

### 7.3 Imágenes almacenadas

| Imagen | Tag | Origen |
|--------|-----|--------|
| `api1-202202768` | `v1` | Construida desde `~/API_VM1/API1/` en VM3 |
| `api2-202202768` | `v1` | Construida desde `~/API_VM1/API2/` en VM3 |
| `api3-202202768` | `v1` | Construida desde `~/API_VM2/` en VM3 |

### 7.4 Verificar el catálogo del registro
```bash
curl http://192.168.122.109:5000/v2/_catalog
```
Respuesta esperada:
```json
{
  "repositories": [
    "api1-202202768",
    "api2-202202768",
    "api3-202202768"
  ]
}
```

### 7.5 Acceso desde VM1 y VM2
Desde VM1 y VM2, el pull de imágenes se realiza apuntando a la IP de VM3 con el flag `--insecure-registry`:
```bash
sudo nerdctl pull --insecure-registry 192.168.122.109:5000/<nombre-imagen>:<tag>
```

---

## 8. Flujo de Despliegue

El siguiente diagrama describe el flujo completo desde la construcción de las imágenes hasta el despliegue final de las APIs:

```
Código fuente (Host)
        │
        │ scp -r ./API_VM1 ./API_VM2 danni3@192.168.122.109:~/
        ▼
     VM3 (Docker)
        │
        │ docker build -t localhost:5000/api1-202202768:v1 .
        │ docker push localhost:5000/api1-202202768:v1
        │ (repetir para API2 y API3)
        ▼
  Zot Registry (VM3:5000)
     ┌──────────────────────────────┐
     │ api1-202202768:v1            │
     │ api2-202202768:v1            │
     │ api3-202202768:v1            │
     └──────────────┬───────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
     VM1 (Containerd)        VM2 (Containerd)
     nerdctl pull api1        nerdctl pull api3
     nerdctl pull api2
        │                       │
        ▼                       ▼
  nerdctl run api1          nerdctl run api3
  nerdctl run api2
        │                       │
        ▼                       ▼
   API1 :8881              API3 :8883
   API2 :8882
```

---

## 9. Comunicación entre Servicios

### 9.1 Matriz de comunicación

| Origen | Destino | Endpoint | Puerto |
|--------|---------|----------|--------|
| API1 | API2 | `/api1/202202768/call-api2` | 8881 → 8882 |
| API1 | API3 | `/api1/202202768/call-api3` | 8881 → 8883 |
| API2 | API1 | `/api2/202202768/call-api1` | 8882 → 8881 |
| API2 | API3 | `/api2/202202768/call-api3` | 8882 → 8883 |
| API3 | API1 | `/api3/202202768/call-api1` | 8883 → 8881 |
| API3 | API2 | `/api3/202202768/call-api2` | 8883 → 8882 |

### 9.2 Protocolo de comunicación
- Todas las comunicaciones se realizan mediante **HTTP** sobre la red virtual de KVM.
- No se utiliza TLS en el entorno de desarrollo.
- Las URLs de las APIs vecinas se inyectan como variables de entorno al momento del despliegue del contenedor.

### 9.3 Comandos de verificación desde el Host
```bash
# Health checks
curl http://192.168.122.90:8881    # API1
curl http://192.168.122.90:8882    # API2
curl http://192.168.122.79:8883    # API3

# Pruebas de comunicación entre APIs
curl http://192.168.122.90:8881/api1/202202768/call-api2
curl http://192.168.122.90:8881/api1/202202768/call-api3
curl http://192.168.122.90:8882/api2/202202768/call-api1
curl http://192.168.122.90:8882/api2/202202768/call-api3
curl http://192.168.122.79:8883/api3/202202768/call-api1
curl http://192.168.122.79:8883/api3/202202768/call-api2
```

---

## 10. Mantenimiento y Solución de Problemas

### 10.1 Comandos de monitoreo

**Ver estado de los servicios en VMs:**
```bash
# En VM1 o VM2
sudo systemctl status containerd
sudo nerdctl ps
sudo nerdctl logs <nombre-contenedor>
```

**Ver estado de Zot en VM3:**
```bash
sudo docker ps | grep zot
sudo docker logs zot
```

**Ver estado de KVM desde el Host:**
```bash
virsh list --all
sudo systemctl status libvirtd
```

### 10.2 Problemas comunes y soluciones

**Problema: `nerdctl pull` falla con error de certificado o conexión rechazada**

Verificar que el flag `--insecure-registry` esté presente en el comando y que el registro Zot esté activo en VM3:
```bash
# Verificar que Zot responde
curl http://192.168.122.109:5000/v2/_catalog
# Verificar que el contenedor de Zot está activo
sudo docker ps | grep zot
```

**Problema: Contenedor no arranca o termina inmediatamente**

Revisar los logs del contenedor para identificar el error:
```bash
sudo nerdctl logs <nombre-contenedor>
```

**Problema: `docker build` falla con "no space left on device"**

Limpiar imágenes y contenedores no utilizados en VM3:
```bash
sudo docker container prune -f
sudo docker image prune -a -f
```

**Problema: Las APIs no se comunican entre sí**

Verificar la conectividad de red entre VMs:
```bash
# Desde VM1, hacer ping a VM2 y VM3
ping 192.168.122.79
ping 192.168.122.109
```

Verificar que las variables de entorno de las APIs tienen las IPs correctas:
```bash
sudo nerdctl inspect <nombre-contenedor> | grep -A 10 "Env"
```

**Problema: `containerd` no inicia**

Verificar la configuración en `/etc/containerd/config.toml` y reiniciar el servicio:
```bash
sudo containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
```

### 10.3 Reinicio controlado de servicios

```bash
# Reiniciar todos los contenedores en VM1
sudo nerdctl restart api1 api2

# Reiniciar todos los contenedores en VM2
sudo nerdctl restart api3

# Reiniciar Zot en VM3
sudo docker restart zot

# Reiniciar libvirtd en el Host
sudo systemctl restart libvirtd
```

### 10.4 Verificación rápida del sistema

```bash
# Versiones instaladas (en VM1 y VM2)
containerd --version
nerdctl --version
runc --version

# Verificar plugins CNI
ls /opt/cni/bin

# Verificar configuración de red CNI
cat /etc/cni/net.d/10-containerd-net.conflist
```

---

*Manual Técnico generado para el Proyecto 1 - Sistemas Operativos 1, Primer Semestre 2026.*  
*Universidad de San Carlos de Guatemala - Facultad de Ingeniería.*