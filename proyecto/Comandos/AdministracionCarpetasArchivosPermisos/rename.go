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

type RENAME struct {
	ruta   string
	nombre string
}

func Rename(parametros []string) string {
	fmt.Println("\n======= RENAME =======")

	var salida1 = ""

	rename := &RENAME{}

	//Parametros
	//Obligatorios
	pathInit := false
	nameInit := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "RENAME Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("RENAME Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			rename.ruta = tmp[1]

		} else if strings.ToLower(tmp[0]) == "name" {
			nameInit = true
			rename.nombre = tmp[1]

		} else {
			salida1 += "RENAME Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("RENAME Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && nameInit {
		err := ejecutarRename(rename)

		if err != nil {
			fmt.Println("RENAME Error: Hubo problemas al renombrar el archivo o carpeta.")
			salida1 += "RENAME Error: Hubo problemas al renombrar el archivo o carpeta.\n"
		} else {
			fmt.Println("El nombre del archivo o carpeta se ha modificado con éxito.")
			salida1 += "El nombre del archivo o carpeta se ha modificado con éxito.\n"
		}
	}

	fmt.Println("\n======FIN RENAME======")
	return salida1
}

func ejecutarRename(rename *RENAME) error {
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

	// Se desglosa la ruta en directorios y el archivo a editar
	directoriosPadre, nombreAntiguoArchivo := Herramientas.ObtenerDirectoriosPadreYArchivo(rename.ruta)

	// Se busca el inodo del directorio donde está el archivo o carpeta
	indiceInodo, err := cat.EncontrarInodoDirectorio(archivo, particionSuperbloque, directoriosPadre)
	if err != nil {
		fmt.Println("Error al encontrar el directorio padre.")
		return fmt.Errorf("error al encontrar el directorio padre: %v", err)
	}

	// Se carga el BloqueFolder del directorio padre
	BloqueFolder := &Structs.BloqueFolder{}
	err = BloqueFolder.Decodificar(archivo, int64(particionSuperbloque.S_block_start+(indiceInodo*particionSuperbloque.S_block_s)))
	if err != nil {
		fmt.Println("Error al deserializar el bloque de carpeta.")
		return fmt.Errorf("error al deserializar el bloque de carpeta: %v", err)
	}

	// Se verifica que no exista un archivo o carpeta con el nuevo nombre
	for _, content := range BloqueFolder.B_contenido {
		if strings.EqualFold(strings.Trim(string(content.B_nombre[:]), "\x00 "), rename.nombre) {
			fmt.Printf("Ya existe un archivo o carpeta con el nombre '%s'", rename.nombre)
			return fmt.Errorf("ya existe un archivo o carpeta con el nombre '%s'", rename.nombre)
		}
	}

	// Se renombra el archivo o carpeta
	err = BloqueFolder.RenombrarEnBloqueFolder(nombreAntiguoArchivo, rename.nombre)
	if err != nil {
		fmt.Println("Error al renombrar el archivo o carpeta.")
		return fmt.Errorf("error al renombrar el archivo o carpeta: %v", err)
	}

	// Se guarda el bloque modificado de nuevo en el archivo
	err = BloqueFolder.Codificar(archivo, int64(particionSuperbloque.S_block_start+(indiceInodo*particionSuperbloque.S_block_s)))
	if err != nil {
		fmt.Println("Error al guardar el bloqye de carpeta modificado.")
		return fmt.Errorf("error al guardar el bloque de carpeta modificado: %v", err)
	}

	fmt.Printf("Nombre cambiado exitosamente de '%s' a '%s'\n", nombreAntiguoArchivo, rename.nombre)

	return nil
}
