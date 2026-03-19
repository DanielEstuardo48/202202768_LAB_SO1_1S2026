package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"daemon/config"
	"daemon/container"
	"daemon/cronjob"
	"daemon/kernel"
	"daemon/storage"
)

func main() {
	fmt.Println("=== Daemon SO1 Proyecto 2 iniciando ===")

	// 1. Levantar Grafana + Valkey con docker compose
	fmt.Println("[1] Levantando Grafana y Valkey...")
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = "/home/danii/daemon"
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error al levantar docker compose: %v", err)
	}

	// 2. Cargar el módulo de kernel
	fmt.Println("[2] Cargando módulo de kernel...")
	exec.Command("bash", config.CronScript).Run()

	// 3. Registrar el cronjob
	fmt.Println("[3] Registrando cronjob...")
	if err := cronjob.Register(); err != nil {
		log.Printf("Advertencia: no se pudo registrar cronjob: %v", err)
	}

	// 4. Conectar a Valkey
	store := storage.NewValkeyClient(config.ValKeyAddr)

	// 5. Conectar a Docker
	dockerClient, err := container.NewDockerClient()
	if err != nil {
		log.Fatalf("Error conectando a Docker: %v", err)
	}

	// 6. Manejar señales para apagado limpio
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\n[!] Señal recibida. Apagando daemon...")
		cronjob.Remove()
		fmt.Println("[✓] Cronjob eliminado. Daemon finalizado.")
		os.Exit(0)
	}()

	protected := []string{"grafana_so1", "valkey_so1"}

	// 7. Loop principal
	fmt.Println("[✓] Daemon corriendo. Ctrl+C para detener.\n")
	for {
		fmt.Printf("--- Ciclo %s ---\n", time.Now().Format("15:04:05"))

		// Leer /proc
		info, err := kernel.ReadProcFile(config.ProcFile)
		if err != nil {
			log.Printf("Error leyendo /proc: %v", err)
			time.Sleep(time.Second * config.LoopInterval)
			continue
		}

		fmt.Printf("RAM: %d MB usada / %d MB total\n", info.UsedRAMMB, info.TotalRAMMB)

		// Listar y clasificar contenedores
		containers, err := container.ListContainers(dockerClient, protected)
		if err != nil {
			log.Printf("Error listando contenedores: %v", err)
		} else {
			// Contar por tipo
			highCount := 0
			lowCount := 0
			for _, c := range containers {
				if c.IsHigh {
					highCount++
					fmt.Printf("  [ALTO]  %s (%s)\n", c.Name, c.Image)
				} else {
					lowCount++
					fmt.Printf("  [BAJO]  %s (%s)\n", c.Name, c.Image)
				}
			}
			fmt.Printf("Total: %d contenedores (%d alto, %d bajo)\n",
				len(containers), highCount, lowCount)

			// Aplicar política
			eliminated := container.EnforcePolicy(
				dockerClient, containers,
				config.MaxHighConsume,
				config.MaxLowConsume,
			)

			if len(eliminated) > 0 {
				fmt.Printf("Eliminados: %v\n", eliminated)
				store.SaveEliminatedContainers(eliminated)
			}
		}

		/*containers, err := container.ListContainers(dockerClient, protected)
		if err != nil {
			log.Printf("Error listando contenedores: %v", err)
		} else {
			fmt.Printf("Contenedores activos: %d\n", len(containers))

			// Aplicar política y eliminar exceso
			eliminated := container.EnforcePolicy(
				dockerClient, containers,
				config.MaxHighConsume,
				config.MaxLowConsume,
			)

			if len(eliminated) > 0 {
				fmt.Printf("Eliminados: %v\n", eliminated)
				store.SaveEliminatedContainers(eliminated)
			}
		}*/

		// Guardar métricas en Valkey
		store.SaveRAMSnapshot(info)
		store.UpdateTopRAM(info.Processes)
		store.UpdateTopCPU(info.Processes)

		fmt.Println("Métricas guardadas en Valkey.")
		time.Sleep(time.Second * config.LoopInterval)
	}
}
