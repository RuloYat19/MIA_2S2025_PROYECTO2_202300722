package Comandos

import (
	//cat "Proyecto/Comandos/AdministracionArchivos"
	//globales "Proyecto/Globales"
	//"Proyecto/Herramientas"
	//"Proyecto/Structs"
	"fmt"
	//"os"
	//"regexp"
	"strings"
)

type CHOWN struct {
	ruta    string
	p       bool
	usuario string
}

func Chown(parametros []string) string {
	fmt.Println("\n======= CHOWN =======")

	var salida1 = ""

	chown := &CHOWN{}

	//Opcionales
	chown.p = false

	//Otras variables
	paramC := true
	rutaInit := false
	userInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "p" {

		} else if len(tmp) != 2 {
			salida1 += "CHOWN Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("CHOWN Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			rutaInit = true
			chown.ruta = tmp[1]

			//Asignación de r
		} else if strings.ToLower(tmp[0]) == "p" {
			chown.p = true
		} else if strings.ToLower(tmp[0]) == "usuario" {
			userInit = true
			chown.usuario = tmp[1]
		} else {
			salida1 += "CHOWN Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("CHOWN Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && rutaInit && userInit {
		err := "asdasdasd"
		//err := ejecutarchown(chown)

		if err != "nil" {
			salida1 += "CHOWN Error: Al cambiar los permisos de las carpetas al usuario " + chown.usuario + ".\n"
		} else {
			salida1 += "Se ha cambiado los permisos de los directorios y archivos para el usuario " + chown.usuario + " con éxito.\n"
		}
	}

	fmt.Println("\n======FIN CHOWN======")
	return salida1
}
