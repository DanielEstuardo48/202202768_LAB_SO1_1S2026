package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"daemon/kernel"

	"github.com/redis/go-redis/v9"
)

type ValkeyClient struct {
	rdb *redis.Client
}

func NewValkeyClient(addr string) *ValkeyClient {
	return &ValkeyClient{
		rdb: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

// Guarda snapshot de RAM en una time series list
func (v *ValkeyClient) SaveRAMSnapshot(info *kernel.SystemInfo) error {
	ctx := context.Background()

	v.rdb.Set(ctx, "so1:ram_total", info.TotalRAMMB, 0)
	v.rdb.Set(ctx, "so1:ram_free", info.FreeRAMMB, 0)
	v.rdb.Set(ctx, "so1:ram_used", info.UsedRAMMB, 0)

	// Para la gráfica
	entry := fmt.Sprintf("%d,%d", time.Now().UnixMilli(), info.UsedRAMMB)
	v.rdb.LPush(ctx, "so1:ram_history", entry)

	// Limitar la lista a 200 entradas para no crecer infinito
	v.rdb.LTrim(ctx, "so1:ram_history", 0, 199)

	return nil
}

// Registra contenedores eliminados
func (v *ValkeyClient) SaveEliminatedContainers(ids []string) error {
	ctx := context.Background()
	ts := time.Now().Unix()

	for _, id := range ids {
		entry := map[string]interface{}{
			"timestamp":    ts,
			"container_id": id,
		}
		data, _ := json.Marshal(entry)
		v.rdb.LPush(ctx, "so1:eliminated", data)
	}
	return nil
}

// Actualiza el top 5 de contenedores por RAM
func (v *ValkeyClient) UpdateTopRAM(processes []kernel.Process) error {
	ctx := context.Background()
	sorted := sortByRSS(processes)
	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i, p := range sorted[:limit] {
		// Guarda nombre y valor
		v.rdb.Set(ctx, fmt.Sprintf("so1:top_ram:%d:name", i+1), p.Name, 0)
		v.rdb.Set(ctx, fmt.Sprintf("so1:top_ram:%d:rss", i+1), p.RSS, 0)
	}
	return nil
}

// Actualiza el top 5 por CPU
func (v *ValkeyClient) UpdateTopCPU(processes []kernel.Process) error {
	ctx := context.Background()
	sorted := sortByCPU(processes)
	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i, p := range sorted[:limit] {
		v.rdb.Set(ctx, fmt.Sprintf("so1:top_cpu:%d:name", i+1), p.Name, 0)
		v.rdb.Set(ctx, fmt.Sprintf("so1:top_cpu:%d:cpu", i+1), p.CPUTime, 0)
	}
	return nil
}

func sortByRSS(p []kernel.Process) []kernel.Process {
	sorted := make([]kernel.Process, len(p))
	copy(sorted, p)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].RSS > sorted[i].RSS {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func sortByCPU(p []kernel.Process) []kernel.Process {
	sorted := make([]kernel.Process, len(p))
	copy(sorted, p)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CPUTime > sorted[i].CPUTime {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
