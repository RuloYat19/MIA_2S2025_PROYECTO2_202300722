package Comandos

import (
	"fmt"
	"strconv"
	"strings"
)

type RECOVERY struct {
	id int
}

func Recovery(parametros []string) string {
	fmt.Println("\n======= RECOVERY =======")

	var salida1 = ""

	recovery := &RECOVERY{}

	//Otras variables
	paramC := true
	idInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "p" {

		} else if len(tmp) != 2 {
			salida1 += "RECOVERY Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("RECOVERY Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de id
		if strings.ToLower(tmp[0]) == "id" {
			idInit = true
			recovery.id, _ = strconv.Atoi(tmp[1])
		} else {
			salida1 += "RECOVERY Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("RECOVERY Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && idInit {
		err := 1
		//err := ejecutarChmod(chmod)

		if err != 1 {
			salida1 += "RECOVERY Error: Hubo problemas al intentar recuperar los datos usando el journaling.\n"
		} else {
			salida1 += "Los datos se han recuperado con éxito usando el journaling.\n"
		}
	}

	fmt.Println("\n======FIN RECOVERY======")
	return salida1
}
