package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

type CAT struct {
	archivos []string
}

func Cat(parametros []string) string {
	fmt.Println("\n======= CAT =======")

	var salida1 = ""

	cat := &CAT{}

	//Otras variables
	paramC := true
	archivosInit := false

	// Se itera sobre cada coincidencia y extrae los archivos
	for _, parametro := range parametros {
		// Se separa el parámetro en clave y valor
		kv := strings.SplitN(parametro, "=", 2)
		if len(kv) == 2 {
			rutaArchivo := kv[1]
			// Si el valor está entre comillas, eliminarlas
			if strings.HasPrefix(rutaArchivo, "\"") && strings.HasSuffix(rutaArchivo, "\"") {
				rutaArchivo = strings.Trim(rutaArchivo, "\"")
			}
			cat.archivos = append(cat.archivos, rutaArchivo)
			archivosInit = true
		}
	}

	if paramC && archivosInit {
		salida2, err := mostrarContenidoArchivo(cat)

		if err != nil {
			salida1 += "CAT Error: Hubo problemas ejecutando el comando CAT.\n"
			return salida1 + salida2
		} else {
			salida1 += "Se ejecutó correctamente el comando CAT.\n"
			return salida1 + salida2
		}

	}

	return salida1
}

func mostrarContenidoArchivo(cat *CAT) (string, error) {
	var salida2 = ""

	// Verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return salida2, fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	_, _, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada")
		return salida2, err
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de partición")
		return salida2, err
	}

	defer archivo.Close()

	var cadena string

	//fmt.Println("A punto de entrar al for para mostrar el contenido")
	// Se lee y muestra el contenido de cada archivo
	for _, rutaArchivo := range cat.archivos {
		cadena += fmt.Sprintf("Leyendo archivo: %s\n", rutaArchivo)

		// Se lee el contenido del archivo
		contenido, err := LeerContenidoArchivo(rutaArchivo)

		if err != nil {
			cadena += fmt.Sprintf("Error al leer el archivo %s: %v\n", rutaArchivo, err)
			continue
		}

		// Se concatena el contenido del archivo a la cadena
		cadena += contenido
		//cadena += "\n"
		cadena += "======FIN CAT======"

		salida2 += contenido
	}

	fmt.Print(cadena)

	return salida2, nil
}

// Se verifica si un directorio o archivo ya existe en el inodo dado
func ComprobacionDirectorio(superbloque *Structs.Superbloque, archivo *os.File, indiceInodo int32, dirNombre string) (bool, int32, error) {
	fmt.Printf("Verificando si el directorio o archivo '%s' existe en el inodo %d\n", dirNombre, indiceInodo)

	// Se deserializa el inodo indicado
	inodo := &Structs.Inodo{}
	err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(indiceInodo*superbloque.S_inode_s)))
	if err != nil {
		fmt.Printf("Error al deserializar inodo %d: %v\n", indiceInodo, err)
		return false, -1, err
	}

	// Se verifica si el inodo es de tipo carpeta (I_type == '0') para poder continuar
	if inodo.I_type[0] != '0' {
		fmt.Printf("El inodo %d no es una carpeta\n", indiceInodo)
		return false, -1, err
	}

	// Se itera sobre los bloques del inodo para buscar el directorio o archivo
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			break // Si no hay más bloques asignados, terminamos la búsqueda
		}

		// Se deserializa el bloque de directorio
		bloque := &Structs.BloqueFolder{}
		err := bloque.Decodificar(archivo, int64(superbloque.S_block_start+(indiceBloque*superbloque.S_block_s)))
		if err != nil {
			fmt.Printf("Error al deserializar bloque %d: %v\n", indiceBloque, err)
			return false, -1, err
		}

		// Se itera sobre los contenidos del bloque para poder verificar si el nombre coincide
		for _, content := range bloque.B_contenido {
			contentName := strings.Trim(string(content.B_nombre[:]), "\x00 ") // Se convierte el nombre y se eliminan los caracteres nulos
			if strings.EqualFold(contentName, dirNombre) && content.B_inodo != -1 {
				fmt.Printf("Directorio o archivo '%s' encontrado en inodo %d\n", dirNombre, content.B_inodo)
				return true, content.B_inodo, nil
			}
		}
	}

	fmt.Printf("Directorio o archivo '%s' no encontrado en inodo %d\n", dirNombre, indiceInodo)
	return false, -1, nil
}

// Se busca el archivo en el sistema de archivos y lee su contenido
func LeerContenidoArchivo(rutaArchivo string) (string, error) {
	// Se obtiene el Superbloque y la partición montada asociada
	idParticion := globales.UsuarioActual.UID

	particionSuperbloque, _, rutaParticion, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)
	if err != nil {
		fmt.Printf("Error al obtener la partición montada: %v\n", err)
		return "", err
	}

	// Se abre el archivo de partición para leerlo
	archivo, err := os.OpenFile(rutaParticion, os.O_RDONLY, 0666)

	if err != nil {
		fmt.Printf("Error al abrir el archivo de partición: %v\n", err)
		return "", err
	}

	defer archivo.Close()

	// Se convierte la ruta del archivo en un array de carpetas
	padreDirs, nombreArchivo := Herramientas.ObtenerDirectoriosPadreYArchivo(rutaArchivo)

	// Se busca el archivo en el sistema de archivos
	indiceInodo, err := EncontrarInodoArchivo(archivo, particionSuperbloque, padreDirs, nombreArchivo)

	if err != nil {
		fmt.Printf("Error al encontrar el archivo: %v", err)
		return "", err
	}

	// Se lee el contenido del archivo
	contenido, err := leerArchivoDesdeInodo(archivo, particionSuperbloque, indiceInodo)

	if err != nil {
		return "", fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	return contenido, nil
}

// Se lee el contenido de un archivo desde su inodo
func leerArchivoDesdeInodo(archivo *os.File, superbloque *Structs.Superbloque, indiceInodo int32) (string, error) {
	inode := &Structs.Inodo{}

	err := inode.Decodificar(archivo, int64(superbloque.S_inode_start+(indiceInodo*superbloque.S_inode_s)))

	if err != nil {
		return "", fmt.Errorf("error al deserializar el inodo %d: %v", indiceInodo, err)
	}

	if inode.I_type[0] != '1' {
		return "", fmt.Errorf("el inodo %d no corresponde a un archivo", indiceInodo)
	}

	// Se concatena los bloques de contenido del archivo
	var contenidoBuilder strings.Builder

	for _, indiceBloque := range inode.I_block {
		if indiceBloque == -1 {
			break
		}

		bloqueArchivo := &Structs.BloqueFile{}
		err := bloqueArchivo.Decodificar(archivo, int64(superbloque.S_block_start+(indiceBloque*superbloque.S_block_s)))

		if err != nil {
			return "", fmt.Errorf("error al deserializar el bloque %d: %v", indiceBloque, err)
		}

		contenidoBuilder.WriteString(string(bloqueArchivo.B_contenido[:]))
	}

	return contenidoBuilder.String(), nil
}

// Se busca el inodo de un archivo dado el path
func EncontrarInodoArchivo(archivo *os.File, superbloque *Structs.Superbloque, padresDir []string, nombreArchivo string) (int32, error) {
	// Se empieza buscando el inodo raiz
	indiceInodo := int32(0)

	// Se navega por los directorios padres para llegar al archivo
	for len(padresDir) > 0 {
		dirNombre := padresDir[0]
		found, nuevoIndiceInodo, err := ComprobacionDirectorio(superbloque, archivo, indiceInodo, dirNombre)

		if err != nil {
			fmt.Printf("Hay un error: %v\n", err)
			return -1, err
		}

		if !found {
			fmt.Printf("Directorio '%s' no encontrado\n", dirNombre)
			return -1, fmt.Errorf("directorio '%s' no encontrado", dirNombre)
		}

		indiceInodo = nuevoIndiceInodo
		padresDir = padresDir[1:]
	}

	// Se busca el archivo en el último directorio
	encontrado, indiceInodoArchivo, err := ComprobacionDirectorio(superbloque, archivo, indiceInodo, nombreArchivo)

	if err != nil {
		return -1, err
	}

	if !encontrado {
		fmt.Printf("Archivo '%s' no encontrado\n", nombreArchivo)
		return -1, fmt.Errorf("archivo '%s' no encontrado", nombreArchivo)
	}

	return indiceInodoArchivo, nil
}

func EncontrarInodoDirectorio(archivo *os.File, superbloque *Structs.Superbloque, directoriosPadre []string) (int32, error) {
	indiceInodo := int32(0)

	for len(directoriosPadre) > 0 {
		dirName := directoriosPadre[0]

		encontrado, nuevoIndiceInodo, err := ComprobacionDirectorio(superbloque, archivo, indiceInodo, dirName)
		if err != nil {
			return -1, err // Error durante la búsqueda
		}
		if !encontrado {
			return -1, fmt.Errorf("directorio '%s' no encontrado", dirName)
		}

		indiceInodo = nuevoIndiceInodo
		directoriosPadre = directoriosPadre[1:]
	}

	return indiceInodo, nil
}
