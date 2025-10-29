package Comandos

import (
	"fmt"
	"os"
	"strings"
)

func Rmdisk(parametros []string) string {
	fmt.Println("\n======= RMDISK =======")
	var salida1 = ""

	//Dato comando
	//Único y obligatorio xd
	var path string

	//Otras variables
	paramC := true
	pathInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "RMDISK Error: Valor desconocido del parámetro" + tmp[0] + "\n"
			fmt.Println("RMDISK Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			path = tmp[1]
		} else {
			salida1 += "RMDISK Error: Parámetro desconocido" + tmp[0] + "\n"
			fmt.Println("RMDISK Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit {
		salida2, err := eliminarDisco(path)

		if err != nil {
			fmt.Println("RMDISK Error: Hubo problemas al eliminar el disco.")
		}

		return salida1 + salida2
	}
	return salida1
}

func eliminarDisco(ruta string) (string, error) {
	var salida2 = ""
	fmt.Printf("Eliminando disco en %s...\n", ruta)

	if _, err := os.Stat(ruta); os.IsNotExist(err) {
		salida2 = "RMDISK Error: El archivo " + ruta + " no existe\n"
		return salida2, err
	}

	err := os.Remove(ruta)

	if err != nil {
		salida2 = "RMDISK Error: Hubo problemas al eliminar el archivo " + ruta + "\n"
		return salida2, err
	}

	salida2 = "Disco en " + ruta + " eliminado exitosamente.\n"
	fmt.Printf("Disco en %s eliminado exitosamente.\n", ruta)
	fmt.Println("======FIN RMDISK======")

	return salida2, nil
}
