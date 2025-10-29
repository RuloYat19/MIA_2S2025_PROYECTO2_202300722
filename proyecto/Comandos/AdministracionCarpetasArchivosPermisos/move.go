package Comandos

import (
	"fmt"
	"strings"
)

type MOVE struct {
	ruta    string
	destino string
}

func Move(parametros []string) string {
	fmt.Println("\n======= MOVE =======")

	var salida1 = ""

	move := &MOVE{}

	//Parametros
	//Obligatorios
	pathInit := false
	destinationInit := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "MOVE Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("MOVE Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			move.ruta = tmp[1]

		} else if strings.ToLower(tmp[0]) == "destino" {
			destinationInit = true
			move.destino = tmp[1]

		} else {
			salida1 += "MVOE Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("MOVE Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && destinationInit {
		err := ejecutarMove(move)

		if err != nil {
			fmt.Println("MOVE Error: Hubo problemas al mover el archivo o carpeta.")
			salida1 += "MOVE Error: Hubo problemas al mover el archivo o carpeta.\n"
		} else {
			fmt.Println("Los directorios y/o archivos se movieron con éxito.")
			salida1 += "Los directorios y/o archivos se movieron con éxito.\n"
		}
	}

	fmt.Println("\n======FIN MOVE======")
	return salida1
}

func ejecutarMove(move *MOVE) error {
	copy := &COPY{}
	copy.ruta = move.ruta
	copy.destino = move.destino

	remove := &REMOVE{}
	remove.ruta = move.ruta

	err := ejecutarCopy(copy)

	if err != nil {
		fmt.Println("Hubo problemas al ejecutar el comando Copy para realizar el Move")
		return err
	}

	err = ejecutarRemove(remove)

	if err != nil {
		fmt.Println("Hubo problemas al ejecutar el comando Remove para realizar el Move")
		return err
	}

	return nil
}
