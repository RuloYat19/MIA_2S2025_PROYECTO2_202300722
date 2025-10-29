package Comandos

import (
	cat "Proyecto/Comandos/AdministracionArchivos"
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

type COPY struct {
	ruta    string
	destino string
}

func Copy(parametros []string) string {
	fmt.Println("\n======= COPY =======")

	var salida1 = ""

	copy := &COPY{}

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
			salida1 += "COPY Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("COPY Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			copy.ruta = tmp[1]

		} else if strings.ToLower(tmp[0]) == "destino" {
			destinationInit = true
			copy.destino = tmp[1]

		} else {
			salida1 += "COPY Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("COPY Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && destinationInit {
		err := ejecutarCopy(copy)

		if err != nil {
			fmt.Println("COPY Error: Hubo problemas al borrar el archivo o carpeta.")
			salida1 += "COPY Error: Hubo problemas al borrar el archivo o carpeta.\n"
		} else {
			fmt.Println("Los directorios y/o archivos se copiaron con éxito.")
			salida1 += "Los directorios y/o archivos se copiaron con éxito.\n"
		}
	}

	fmt.Println("\n======FIN COPY======")
	return salida1
}

func ejecutarCopy(copy *COPY) error {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se valida que el root sea el que inició sesión
	/*usuarioLogueado := globales.UsuarioActual.Nombre

	if usuarioLogueado != "root" {
		fmt.Println("Este comando solamente lo puede ejecutar el root")
		return fmt.Errorf("Este comando solamente lo puede ejecutar el root")
	}*/

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, particionMontada, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada.")
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de la partición.")
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}

	defer archivo.Close()

	err = CopiarDirectoriosOArchivos(copy.ruta, copy.destino, particionSuperbloque, *particionMontada, archivo)

	if err != nil {
		fmt.Println("Error al copiar los archivos o directorios.")
		return fmt.Errorf("error al copiar los archivos o directorios: %w", err)
	}

	fmt.Println("\nInodos:")
	particionSuperbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques de datos:")
	particionSuperbloque.ImprimirBloques(archivo.Name())

	return nil
}

// Se copia los archivos o directorios seleccionados
func CopiarDirectoriosOArchivos(ruta string, destino string, superbloque *Structs.Superbloque, particionMontada Structs.Particion, archivo *os.File) error {
	// Se convierte el path del archivo o carpeta en un array de carpetas
	directoriosPadreRuta, nombreArchivoODirectorioRuta := Herramientas.ObtenerDirectoriosPadreYArchivo(ruta)

	// Se convierte el path del archivo o carpeta en un array de carpetas
	directoriosPadreDestino, nombreArchivoODirectorioDestino := Herramientas.ObtenerDirectoriosPadreYArchivo(destino)

	// Si no es un archivo, intentar eliminarlo como carpeta
	err := CopiarDirectorios(superbloque, particionMontada, archivo, directoriosPadreRuta, nombreArchivoODirectorioRuta, directoriosPadreDestino, nombreArchivoODirectorioDestino)
	if err != nil {
		fmt.Printf("Error al copiar el directorio '%s'.\n", ruta)
		return fmt.Errorf("rror al copiar el directorio '%s': %v", ruta, err)
	}

	return nil
}

func CopiarDirectorios(superbloque *Structs.Superbloque, particionMontada Structs.Particion, archivo *os.File, directoriosPadreRuta []string, nombreDirectorioRuta string, directoriosPadreDestino []string, nombreDirectorioDestino string) error {
	// Se busca el inodo de la carpeta de la ruta
	_, err := cat.EncontrarInodoDirectorio(archivo, superbloque, directoriosPadreRuta)
	if err != nil {
		fmt.Printf("La carpeta '%s' no fue encontrada de la ruta.\n", nombreDirectorioRuta)
		return fmt.Errorf("carpeta '%s' no encontrada de la ruta: %v", nombreDirectorioRuta, err)
	}

	// Se busca el inodo de la carpeta del destino
	_, err = cat.EncontrarInodoDirectorio(archivo, superbloque, directoriosPadreRuta)
	if err != nil {
		fmt.Printf("La carpeta '%s' no fue encontrada del destino.\n", nombreDirectorioDestino)
		return fmt.Errorf("carpeta '%s' no encontrada del destino: %v", nombreDirectorioDestino, err)
	}

	// Se llama a la función que copia la carpeta
	var nuevasRutas []string
	nuevasRutas, err = superbloque.NuevasRutas(archivo, directoriosPadreRuta, nombreDirectorioRuta, directoriosPadreDestino, nombreDirectorioDestino)
	if err != nil {
		fmt.Printf("Error al copiar la carpeta '%s'.\n", nombreDirectorioRuta)
		return fmt.Errorf("error al copiar la carpeta '%s': %v", nombreDirectorioRuta, err)
	}

	err = CrearLosNuevosDirectoriosYArchivos(superbloque, archivo, particionMontada, directoriosPadreRuta, nombreDirectorioRuta, directoriosPadreDestino, nombreDirectorioDestino, nuevasRutas)

	if err != nil {
		return nil
	}

	return nil
}

func CrearLosNuevosDirectoriosYArchivos(superbloque *Structs.Superbloque, archivo *os.File, particionMontada Structs.Particion, directoriosPadreRuta []string, nombreDirectorioRuta string, directoriosPadreDestino []string, nombreDirectorioDestino string, nuevasRutas []string) error {
	for _, ruta := range nuevasRutas {
		if strings.Contains(strings.ToLower(ruta), ".") {
			rutaAntiguaArchivo := superbloque.ObtenerRutaAntiguaArchivo(directoriosPadreRuta, nombreDirectorioRuta, ruta)

			contenidoArchivo, err := cat.LeerContenidoArchivo(rutaAntiguaArchivo)

			if err != nil {
				return err
			}

			err = CrearArchivo(ruta, 83, contenidoArchivo, superbloque, archivo, &particionMontada, true)
			if err != nil {
				return err
			}
		} else {
			err := CrearDirectorio(ruta, true, superbloque, archivo, &particionMontada)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
