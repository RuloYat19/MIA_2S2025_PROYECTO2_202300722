package Comandos

import (
	globales "Proyecto/Globales"
	"fmt"
)

func Mounted(parametros []string) string {
	fmt.Println("\n======= MOUNTED =======")
	var salida = ""
	var contador = 0
	for id := range globales.ParticionesMontadas {
		contador++
		if len(globales.ParticionesMontadas) == contador {
			salida += id + "\n"
			fmt.Printf("%s\n", id)
			fmt.Println("\n======FIN MOUNTED======")
			return salida
		}
		salida += id + ","
		fmt.Printf("%s,", id)
	}
	return salida
}
