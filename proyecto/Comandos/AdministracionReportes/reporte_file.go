package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReporteFile(superbloque *Structs.Superbloque, rutaDisco string, ruta string, rutaArchivo string) error {
	// Se crean las carpetas padre si no existen
	err := Herramientas.CrearPadreDirs(ruta)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Se abre el archivo de disco
	archivo, err := os.Open(rutaDisco)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}

	defer archivo.Close()

	// Se busca el inodo del archivo especificado en rutaArchivo
	inodoIndice, err := encontrarArchivoInodo(superbloque, archivo, rutaArchivo)

	if err != nil {
		return fmt.Errorf("error al buscar el inodo del archivo: %v", err)
	}

	// Se lee el contenido del archivo
	fileContent, err := LeerContenidoArchivo(superbloque, archivo, inodoIndice)

	if err != nil {
		return fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	// Se crea el archivo de salida para el reporte
	reporteArchivo, err := os.Create(ruta)

	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}

	defer reporteArchivo.Close()

	// Se escribe el nombre y el contenido del archivo en el archivo de reporte
	_, nombreArchivo := filepath.Split(rutaArchivo)
	reportContent := fmt.Sprintf("Nombre del archivo: %s\n\nContenido del archivo:\n%s", nombreArchivo, fileContent)

	_, err = reporteArchivo.WriteString(reportContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte del archivo generado:", ruta)
	return nil
}

// Se busca el inodo del archivo especificado a través de su ruta
func encontrarArchivoInodo(superbloque *Structs.Superbloque, archivoDisco *os.File, rutaArchivo string) (int32, error) {
	// Se asume que partimos del inodo raíz
	indiceInodoActual := int32(0) // Inodo raíz

	// Se divide la ruta en directorios y nombre de archivo
	directories, nombreArchivo := Herramientas.ObtenerDirectoriosPadreYArchivo(rutaArchivo)

	// Se navega por cada directorio para encontrar el inodo final
	for _, directorio := range directories {
		inodo, err := leerInodo(superbloque, archivoDisco, indiceInodoActual)

		if err != nil {
			fmt.Printf("Error al leer el inodo: %v\n", err)
			return -1, err
		}

		// Se busca el directorio en el bloque del inodo actual
		encontrado, siguienteIndiceInodo := encontrarInodoEnDirectorio(inodo, archivoDisco, directorio, superbloque)

		if !encontrado {
			fmt.Printf("Directorio '%s' no encontrado\n", directorio)
			return -1, err
		}

		indiceInodoActual = siguienteIndiceInodo
	}

	// Ahora se busca el archivo en el último directorio
	inodo, err := leerInodo(superbloque, archivoDisco, indiceInodoActual)
	if err != nil {
		fmt.Printf("Error al leer el inodo del directorio final: %v\n", err)
		return -1, err
	}

	// Se busca el archivo en el bloque del inodo actual
	encontrado, archivoIndiceInodo := encontrarInodoEnDirectorio(inodo, archivoDisco, nombreArchivo, superbloque)
	if !encontrado {
		fmt.Printf("Archivo '%s' no encontrado\n", nombreArchivo)
		return -1, err
	}

	return archivoIndiceInodo, nil
}

// Se lee el contenido de un archivo dado su inodo
func LeerContenidoArchivo(superbloque *Structs.Superbloque, archivoDisco *os.File, indiceInodo int32) (string, error) {
	inodo, err := leerInodo(superbloque, archivoDisco, indiceInodo)
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo del archivo: %v", err)
	}

	// Se concatena el contenido de los bloques
	var contenido string
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			continue
		}

		// Se lee el bloque de archivo
		bloque, err := leerArchivoBloque(superbloque, archivoDisco, indiceBloque)
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque de archivo: %v", err)
		}

		contenido += string(bloque.B_contenido[:])
	}

	return contenido, nil
}

// Se lee el inodo en la posición dada
func leerInodo(superblock *Structs.Superbloque, archivoDisco *os.File, inodeIndex int32) (*Structs.Inodo, error) {
	inodo := &Structs.Inodo{}
	offset := int64(superblock.S_inode_start + inodeIndex*superblock.S_inode_s)
	err := inodo.Decodificar(archivoDisco, offset)

	if err != nil {
		return nil, fmt.Errorf("error al decodificar el inodo: %v", err)
	}

	return inodo, nil
}

// Se lee un bloque de archivo en la posición dada
func leerArchivoBloque(superblock *Structs.Superbloque, archivoDisco *os.File, indiceBloque int32) (*Structs.BloqueFile, error) {
	bloque := &Structs.BloqueFile{}
	offset := int64(superblock.S_block_start + indiceBloque*superblock.S_block_s)
	err := bloque.Decodificar(archivoDisco, offset)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar el bloque de archivo: %v", err)
	}
	return bloque, nil
}

// Se busca un inodo dentro de un bloque de directorio dado
func encontrarInodoEnDirectorio(inodo *Structs.Inodo, archivoDisco *os.File, nombre string, superbloque *Structs.Superbloque) (bool, int32) {
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			continue
		}

		// Se lee el bloque de carpeta
		bloque := &Structs.BloqueFolder{}
		offset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		err := bloque.Decodificar(archivoDisco, offset)
		if err != nil {
			continue
		}

		// Se buscar el nombre dentro del bloque de carpeta
		for _, contenido := range bloque.B_contenido {
			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			if nombreContenido == nombre {
				return true, contenido.B_inodo
			}
		}
	}
	return false, -1
}
