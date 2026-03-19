package cronjob

import (
	"os/exec"
	"strings"
)

func Register() error {
	// Leer crontab actual
	out, _ := exec.Command("sudo", "crontab", "-l").Output()
	current := string(out)

	// Si ya existe no duplicar
	if strings.Contains(current, "spawn_containers") {
		return nil
	}

	// Agregar la línea
	newCron := current + "\n*/2 * * * * /home/danii/spawn_containers.sh >> /var/log/so1_containers.log 2>&1\n"
	cmd := exec.Command("sudo", "crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)
	return cmd.Run()
}

func Remove() error {
	out, _ := exec.Command("sudo", "crontab", "-l").Output()
	current := string(out)

	var newLines []string
	for _, line := range strings.Split(current, "\n") {
		if !strings.Contains(line, "spawn_containers") {
			newLines = append(newLines, line)
		}
	}

	cmd := exec.Command("sudo", "crontab", "-")
	cmd.Stdin = strings.NewReader(strings.Join(newLines, "\n"))
	return cmd.Run()
}
