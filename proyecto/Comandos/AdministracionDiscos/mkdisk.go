package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Mkdisk(parametros []string) string {
	fmt.Println("\n======= MKDISK =======")
	salida1 := ""

	//Datos comando
	//Obligatorios
	var size int
	var path string

	//Opcionales
	fit := "F"
	unit := 1048576

	//Otras variables
	paramC := true
	sizeInit := false
	pathInit := false

	//Se recorren los parametros para asignarlos a las variables
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "MKDISK Error: Valor desconocido del parámetro " + tmp[0] + "\n"
			fmt.Println("MKDISK Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de size
		if strings.ToLower(tmp[0]) == "size" {
			sizeInit = true
			var err error
			size, err = strconv.Atoi(tmp[1])

			if err != nil {
				salida1 += "MKDISK Error: -size debe ser un valor numérico. Se leyó: " + tmp[1] + "\n"
				fmt.Println("MKDISK Error: -size debe ser un valor numerico. se leyo: ", tmp[1])
				paramC = false
				break
			} else if size <= 0 {
				salida1 += "MKDISK Error: -size debe ser un valor positivo mayor a cero (0). Se leyó: " + tmp[1] + "\n"
				fmt.Println("MKDISK Error: -size debe ser un valor positivo mayor a cero (0). se leyo: ", tmp[1])
				paramC = false
				break
			}

			//Asignación de fit
		} else if strings.ToLower(tmp[0]) == "fit" {
			if strings.ToLower(tmp[1]) == "bf" {
				fit = "B"
			} else if strings.ToLower(tmp[1]) == "wf" {
				fit = "W"
			} else if strings.ToLower(tmp[1]) != "ff" {
				salida1 += "MKDISK Error: Valor del fit erróneo. Los valores aceptados son: BF, FF o WF. Se leyó: " + tmp[1] + "\n"
				fmt.Println("MKDISK Error en -fit. Valores aceptados: BF, FF o WF. ingreso: ", tmp[1])
				paramC = false
				break
			}
			//Asignación de unit
		} else if strings.ToLower(tmp[0]) == "unit" {
			if strings.ToLower(tmp[1]) == "k" {
				unit = 1024
			} else if strings.ToLower(tmp[1]) != "m" {
				salida1 += "MKDISK Error: Valor erróneo en -unit. Valores aceptados k, m o b. Se leyó: " + tmp[1] + "\n"
				fmt.Println("MKDISK Error en -unit. Valores aceptados: k, m o b. ingreso: ", tmp[1])
				paramC = false
				break
			}
			// Asignación de path
		} else if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			path = tmp[1]
		} else {
			salida1 += "MKDISK Error: Parámetro desconocido: " + tmp[0] + "\n"
			fmt.Println("MKDISK Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && sizeInit && pathInit {
		salida2, err := crearDisco(size, path, fit, unit)

		if err != nil {
			fmt.Println("MKDISK Error: Hubo problemas creando el disco.")
		}

		return salida1 + salida2
	} else {
		salida1 += "MKDISK Error: No se ingresó el parámetro -size o -path o tuvieron algún error\n"
		fmt.Println("MKDISK Error: No se ingresó -size o -path o tuvieron algún error")
	}

	return salida1
}

func crearDisco(size int, path string, fit string, unit int) (string, error) {
	var salida = ""
	//Tamaño del disco
	tam := size * unit
	//Nombre del disco
	nombreDisco := strings.Split(path, "/")
	disco := nombreDisco[len(nombreDisco)-1]

	//Se crea el disco
	err := Herramientas.CrearDisco(path)
	if err != nil {
		salida += "MKDISK Error: Hubo problemas creando el Disco.\n"
		fmt.Println("MKDISK Error: Hubo problemas creando el Disco.")
		return salida, err
	}

	//Se abre el archivo del disco
	archivo, err := Herramientas.AbrirArchivo(path)
	if err != nil {
		fmt.Println("MKDISK Error: Hubo problemas abriendo el archivo.")
		return salida, err
	}

	//Se escriben los bytes 0 en función del tamaño
	datos := make([]byte, tam)
	newErr := Herramientas.EscribirObjeto(archivo, datos, 0)
	if newErr != nil {
		fmt.Println("MKDISK Error: Hubo problemas escribiendo en el archivo los bytes (0).")
		return salida, newErr
	}

	//Calculo de fecha y hora
	ahora := time.Now()
	segundos := ahora.Second()
	minutos := ahora.Minute()
	cad := fmt.Sprintf("%02d%02d", segundos, minutos)
	idTmp, err := strconv.Atoi(cad)

	if err != nil {
		fmt.Println("MKDISK Error: Hubo problemas convertiendo la fecha en entero para el ID.")
		return salida, err
	}

	//Se crea el MBR
	var nuevoMBR Structs.MBR
	nuevoMBR.MbrSize = int32(tam)
	nuevoMBR.Id = int32(idTmp)
	copy(nuevoMBR.Fit[:], fit)
	copy(nuevoMBR.FechaC[:], ahora.Format("02/01/2006 15:04"))

	//Se escribe el MBR en el disco
	if err := Herramientas.EscribirObjeto(archivo, nuevoMBR, 0); err != nil {
		fmt.Println("MKDISK Error: Hubo problemas escribiendo el MBR en el archivo.")
		return salida, err
	}

	defer archivo.Close()

	fmt.Println("\n Se creo el disco ", disco, " de forma exitosa")
	salida += "Se creó el disco satisfactoriamente\n"
	//Se imprime el MBR para comprobar datos
	var TempMBR Structs.MBR
	if err := Herramientas.LeerObjeto(archivo, &TempMBR, 0); err != nil {
		fmt.Println("MKDISK Error: Hubo problemas leyendo el disco.")
		return salida, err
	}

	// Se imprimir el MBR del disco
	Structs.ImprimirMBR(TempMBR)

	fmt.Println("\n======FIN MKDISK======")

	return salida, nil
}
