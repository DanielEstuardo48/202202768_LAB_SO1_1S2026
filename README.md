<div align="center">

# 202202768_LAB_SO1_1S2026

**Sistemas Operativos 1 — Sección P**  
Universidad de San Carlos de Guatemala · Facultad de Ingeniería  

**Daniel Estuardo Salvatierra Macajola**  
Carnet: `202202768` · Primer Semestre 2026

</div>

---

## Contenido del repositorio

| Carpeta | Descripción |
|---------|-------------|
| `API_VM1/` | Código fuente de API1 y API2 (desplegadas en VM1) |
| `API_VM2/` | Código fuente de API3 (desplegada en VM2) |
| `imagenes/` | Capturas de pantalla de la guía de instalación |

## Documentación

- [Guía de Instalación](Guia_de_instalacion.md)
- [Manual Técnico](Manual_Tecnico.md)

## Descripción del Proyecto

Despliegue de tres APIs distribuidas en un entorno virtualizado con KVM. Las APIs corren como contenedores en máquinas virtuales usando Containerd, y se distribuyen mediante un registro privado Zot alojado en una tercera VM con Docker.
