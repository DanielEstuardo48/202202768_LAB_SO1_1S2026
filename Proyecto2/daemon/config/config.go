package config

const (
	ProcFile       = "/proc/continfo_pr2_so1_202202768"
	ValKeyAddr     = "localhost:6379"
	LoopInterval   = 20 // segundos entre cada ciclo del daemon
	CronScript     = "/ruta/absoluta/proyecto2/scripts/spawn_containers.sh"
	GrafanaName    = "grafana_so1"
	ValkeyName     = "valkey_so1"
	MaxLowConsume  = 3
	MaxHighConsume = 2
)
