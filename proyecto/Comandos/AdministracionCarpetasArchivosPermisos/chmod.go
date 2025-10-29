package Comandos

import (
	globales "Proyecto/Globales"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CHMOD struct {
	ruta string
	r    bool
	ugo  int
}

func Chmod(parametros []string) string {
	fmt.Println("\n======= CHMOD =======")

	var salida1 = ""

	chmod := &CHMOD{}

	//Opcionales
	chmod.r = false

	//Otras variables
	paramC := true
	rutaInit := false
	ugoInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "p" {

		} else if len(tmp) != 2 {
			salida1 += "CHMOD Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("CHMOD Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			rutaInit = true
			chmod.ruta = tmp[1]

			//Asignación de r
		} else if strings.ToLower(tmp[0]) == "p" {
			chmod.r = true
		} else if strings.ToLower(tmp[0]) == "ugo" {
			ugoInit = true
			chmod.ugo, _ = strconv.Atoi(tmp[1])
		} else {
			salida1 += "CHMOD Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("CHMOD Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && rutaInit && ugoInit {
		err := ejecutarChmod(chmod)

		if err != nil {
			salida1 += "CHMOD Error: Hubo problemas al cambiar los permisos de la carpeta: " + chmod.ruta + " con los permisos " + strconv.Itoa(chmod.ugo) + ".\n"
		} else {
			salida1 += "Los permisos de la/s carpeta/s: " + chmod.ruta + " han sido cambiadas con éxito con los permisos " + strconv.Itoa(chmod.ugo) + ".\n"
		}
	}

	fmt.Println("\n======FIN CHMOD======")
	return salida1
}

func ejecutarChmod(chmod *CHMOD) error {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, _, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada.")
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de la partición.")
		return fmt.Errorf("error al abrir el archivo de la partición: %w", err)
	}

	defer archivo.Close()

	err = particionSuperbloque.BusquedaDeIndiceInodoParaCambiarPermisos(archivo, chmod.ruta, chmod.ugo)

	if err != nil {
		fmt.Println("Hubo problemas al intentar de cambiar los permisos del directorio o archivo.")
		return fmt.Errorf("hubo problemas al intentar de cambiar los permisos del directorio o archivo: %w", err)
	}

	return nil
}
