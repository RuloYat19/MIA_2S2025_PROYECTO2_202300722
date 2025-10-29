package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

type MKDIR struct {
	ruta string
	p    bool
}

func Mkdir(parametros []string) string {
	fmt.Println("\n======= MKDIR =======")

	var salida1 = ""

	mkdir := &MKDIR{}

	//Opcionales
	mkdir.p = false

	//Otras variables
	paramC := true
	rutaInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "p" {

		} else if len(tmp) != 2 {
			salida1 += "MKDIR Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("MKDIR Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			rutaInit = true
			mkdir.ruta = tmp[1]

			//Asignación de r
		} else if strings.ToLower(tmp[0]) == "p" {
			mkdir.p = true
		} else {
			salida1 += "MKDIR Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("MKDIR Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && rutaInit {
		err := ejecutarMkdir(mkdir)

		if err != nil {
			salida1 += "MKDIR Error: Hubo problemas creando el directorio " + mkdir.ruta + ".\n"
		} else {
			salida1 += "El directorio " + mkdir.ruta + " fue creado exitosamente.\n"
		}
	}

	fmt.Println("\n======FIN MKDIR======")
	return salida1
}

func ejecutarMkdir(mkdir *MKDIR) error {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, particionMontada, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}

	defer archivo.Close()

	fmt.Printf("Creando directorio: %s\n", mkdir.ruta)

	// Se crea el directorio usando el archivo abierto, pasando la opción -p
	err = CrearDirectorio(mkdir.ruta, mkdir.p, particionSuperbloque, archivo, particionMontada)

	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	return nil
}

func CrearDirectorio(dirRuta string, crearPadres bool, superbloque *Structs.Superbloque, archivo *os.File, particionMontada *Structs.Particion) error {
	// Se obtienen los directoriosPadre
	directoriosPadre, _ := Herramientas.ObtenerDirectorios(dirRuta)

	//Se obtienen los directorios padres para poder validar su existencia en los bloques
	dirPadres := directoriosPadre[:len(directoriosPadre)-1]

	siExistenTodosLosPadres := false
	contador := 0

	if len(directoriosPadre) == 1 {
		siExistenTodosLosPadres = true
	} else {

		for _, parentDir := range dirPadres {
			for i := int32(0); i < superbloque.S_inodes_count; i++ {
				bandera, err := superbloque.ValidarExistenciaDeDirectorio(archivo, i, parentDir)

				if err != nil {
					return err
				}

				if bandera {
					contador++
				}

				if contador == len(dirPadres) {
					siExistenTodosLosPadres = true
					break
				}
			}
		}
	}

	if siExistenTodosLosPadres || crearPadres {
		// Se iteran las carpetas para crearse pero se validan antes
		for _, nombreDirectorio := range directoriosPadre {
			err := superbloque.CrearCarpeta(archivo, directoriosPadre, nombreDirectorio, crearPadres)
			if err != nil {
				fmt.Println("Error al crear las carpetas")
				return err
			}
		}
	} else {
		fmt.Println("No se pudieron crear los directorios ya que hacen falta carpeta padre")
		return fmt.Errorf("no se pudieron crear los directorios ya que hacen falta carpeta padre")
	}

	// Se serializar el superbloque en el archivo de partición abierto
	err := superbloque.Codificar(archivo, int64(particionMontada.Start))

	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	// Se muestran las estructuras en consola
	fmt.Println("\nInodos:")
	superbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques:")
	superbloque.ImprimirBloques(archivo.Name())

	return nil
}
