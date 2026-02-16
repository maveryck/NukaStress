package utils

import (
	"fmt"
	"time"
)

func LogBoot() {
	fmt.Printf("[%s] Iniciando protocolo de estrés nuclear...\n", timestamp())
}

func LogCriticalTemp(temp float64) {
	fmt.Printf("[%s] ¡Radiación crítica en el núcleo! Temp: %.1f°C – ¡Evacua el Vault!\n", timestamp(), temp)
}

func LogSurvivor(name string) {
	fmt.Printf("[%s] %s ha sobrevivido al Día del Juicio. ¡Thumbs up, wastelander!\n", timestamp(), name)
}

func timestamp() string {
	return time.Now().Format("15:04:05")
}
