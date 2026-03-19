package container

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	IsHigh bool // true = alto consumo, false = bajo consumo
}

func NewDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// Lista todos los contenedores activos excepto los protegidos
func ListContainers(cli *client.Client, protected []string) ([]ContainerInfo, error) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")

		// Saltar contenedores protegidos
		skip := false
		for _, p := range protected {
			if name == p {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		result = append(result, ContainerInfo{
			ID:     c.ID[:12],
			Name:   name,
			Image:  c.Image,
			IsHigh: isHighConsumption(c.Image, c.Command),
		})
	}
	return result, nil
}

// Determina si un contenedor es de alto consumo
func isHighConsumption(image, command string) bool {
	if strings.Contains(image, "go-client") {
		return true
	}
	// Alpine con bc en bucle = alto consumo CPU
	if strings.Contains(image, "alpine") && strings.Contains(command, "bc") {
		return true
	}
	return false
}

// Detiene y elimina un contenedor
func RemoveContainer(cli *client.Client, containerID string) error {
	ctx := context.Background()
	// Detener primero
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return err
	}
	// Luego eliminar
	return cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}

// Retorna los IDs eliminados
func EnforcePolicy(cli *client.Client, containers []ContainerInfo, maxHigh, maxLow int) []string {
	var highList, lowList []ContainerInfo

	for _, c := range containers {
		if c.IsHigh {
			highList = append(highList, c)
		} else {
			lowList = append(lowList, c)
		}
	}

	var eliminated []string

	// Eliminar exceso de alto consumo
	for len(highList) > maxHigh {
		target := highList[len(highList)-1]
		RemoveContainer(cli, target.ID)
		eliminated = append(eliminated, target.ID)
		highList = highList[:len(highList)-1]
	}

	// Eliminar exceso de bajo consumo
	for len(lowList) > maxLow {
		target := lowList[len(lowList)-1]
		RemoveContainer(cli, target.ID)
		eliminated = append(eliminated, target.ID)
		lowList = lowList[:len(lowList)-1]
	}

	return eliminated
}
