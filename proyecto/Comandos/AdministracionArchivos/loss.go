package Comandos

import (
	"fmt"
	"strconv"
	"strings"
)

type LOSS struct {
	id int
}

func Loss(parametros []string) string {
	fmt.Println("\n======= LOSS =======")

	var salida1 = ""

	loss := &LOSS{}

	//Otras variables
	paramC := true
	idInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "p" {

		} else if len(tmp) != 2 {
			salida1 += "LOSS Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("LOSS Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de id
		if strings.ToLower(tmp[0]) == "id" {
			idInit = true
			loss.id, _ = strconv.Atoi(tmp[1])
		} else {
			salida1 += "LOSS Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("LOSS Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && idInit {
		err := 1
		//err := ejecutarChmod(chmod)

		if err != 1 {
			salida1 += "LOSS Error: Hubo problemas al intentar perder los datos.\n"
		} else {
			salida1 += "Los datos se han perdido con éxito.\n"
		}
	}

	fmt.Println("\n======FIN LOSS======")
	return salida1
}
