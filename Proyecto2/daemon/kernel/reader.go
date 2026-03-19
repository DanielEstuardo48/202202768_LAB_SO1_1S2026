package kernel

import (
	"encoding/json"
	"os"
)

type Process struct {
	PID        int    `json:"pid"`
	Name       string `json:"name"`
	VSZ        uint64 `json:"vsz"`
	RSS        uint64 `json:"rss"`
	MemPercent uint64 `json:"mem_percent"`
	CPUTime    uint64 `json:"cpu_time"`
}

type SystemInfo struct {
	TotalRAMMB uint64    `json:"total_ram_mb"`
	FreeRAMMB  uint64    `json:"free_ram_mb"`
	UsedRAMMB  uint64    `json:"used_ram_mb"`
	Processes  []Process `json:"processes"`
}

func ReadProcFile(path string) (*SystemInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
