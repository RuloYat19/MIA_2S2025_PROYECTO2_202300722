package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
	"time"
)

// ReporteLS genera un reporte tipo 'ls' mostrando el contenido de una carpeta
func ReporteLS(superbloque *Structs.Superbloque, rutaDisco string, rutaReporte string, rutaCarpeta string) error {
	err := Herramientas.CrearPadreDirs(rutaReporte)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	archivo, err := os.Open(rutaDisco)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer archivo.Close()

	inodoIndice, err := encontrarCarpetaInodo(superbloque, archivo, rutaCarpeta)
	if err != nil {
		return fmt.Errorf("error al buscar el inodo de la carpeta: %v", err)
	}

	tabla, err := generarTablaLS(superbloque, archivo, inodoIndice)
	if err != nil {
		return fmt.Errorf("error al leer el contenido de la carpeta: %v", err)
	}

	reporteArchivo, err := os.Create(rutaReporte)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer reporteArchivo.Close()

	_, err = reporteArchivo.WriteString(tabla)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte LS generado:", rutaReporte)
	return nil
}
func generarTablaLS(superbloque *Structs.Superbloque, archivoDisco *os.File, indiceInodoCarpeta int32) (string, error) {
	inodoCarpeta, err := leerInodoLS(superbloque, archivoDisco, indiceInodoCarpeta)
	if err != nil {
		return "", err
	}
	var tabla strings.Builder
	tabla.WriteString("| Permisos | Owner | Grupo | Size (en Bytes) | Fecha | Hora | Tipo | Name |\n")
	tabla.WriteString("|----------|-------|-------|-----------------|-------|------|------|------|\n")
	for _, indiceBloque := range inodoCarpeta.I_block {
		if indiceBloque == -1 {
			continue
		}
		bloque := &Structs.BloqueFolder{}
		offset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		err := bloque.Decodificar(archivoDisco, offset)
		if err != nil {
			continue
		}
		for _, entry := range bloque.B_contenido {
			nombre := strings.Trim(string(entry.B_nombre[:]), "\x00 ")
			if nombre == "" || nombre == "." || nombre == ".." {
				continue
			}
			inodoEntry, err := leerInodoLS(superbloque, archivoDisco, entry.B_inodo)
			if err != nil {
				continue
			}
			permisos := string(inodoEntry.I_perm[:])
			owner := obtenerOwner(inodoEntry.I_uid)
			grupo := obtenerGrupo(inodoEntry.I_gid)
			size := inodoEntry.I_s
			fecha := ""
			hora := ""
			if inodoEntry.I_mtime != 0 {
				fecha = time.Unix(int64(inodoEntry.I_mtime), 0).Format("02/01/2006")
				hora = time.Unix(int64(inodoEntry.I_mtime), 0).Format("15:04")
			}
			tipo := "Archivo"
			if inodoEntry.I_type[0] == '0' {
				tipo = "Carpeta"
			}
			tabla.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s | %s | %s | %s |\n",
				permisos, owner, grupo, size, fecha, hora, tipo, nombre))
		}
	}
	return tabla.String(), nil
}

func obtenerOwner(uid int32) string {
	return fmt.Sprintf("User%d", uid)
}

func obtenerGrupo(gid int32) string {
	return fmt.Sprintf("Grupo%d", gid)
}

func encontrarCarpetaInodo(superbloque *Structs.Superbloque, archivoDisco *os.File, rutaCarpeta string) (int32, error) {
	indiceInodoActual := int32(0) // Inodo raíz
	directories, _ := Herramientas.ObtenerDirectorios(rutaCarpeta)
	for _, directorio := range directories {
		inodo, err := leerInodoLS(superbloque, archivoDisco, indiceInodoActual)
		if err != nil {
			return -1, err
		}
		encontrado, siguienteIndiceInodo := buscarInodoEnDirectorioLS(inodo, archivoDisco, directorio, superbloque)
		if !encontrado {
			return -1, fmt.Errorf("directorio '%s' no encontrado", directorio)
		}
		indiceInodoActual = siguienteIndiceInodo
	}
	return indiceInodoActual, nil
}

func leerContenidoCarpeta(superbloque *Structs.Superbloque, archivoDisco *os.File, indiceInodo int32) (string, error) {
	inodo, err := leerInodoLS(superbloque, archivoDisco, indiceInodo)
	if err != nil {
		return "", err
	}
	var contenido strings.Builder
	contenido.WriteString("Contenido de la carpeta:\n")
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			continue
		}
		bloque := &Structs.BloqueFolder{}
		offset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		err := bloque.Decodificar(archivoDisco, offset)
		if err != nil {
			continue
		}
		for _, entry := range bloque.B_contenido {
			nombre := strings.Trim(string(entry.B_nombre[:]), "\x00 ")
			if nombre != "" && nombre != "." && nombre != ".." {
				contenido.WriteString(fmt.Sprintf("- %s\n", nombre))
			}
		}
	}
	return contenido.String(), nil
}

func leerInodoLS(superbloque *Structs.Superbloque, archivoDisco *os.File, inodeIndex int32) (*Structs.Inodo, error) {
	inodo := &Structs.Inodo{}
	offset := int64(superbloque.S_inode_start + inodeIndex*superbloque.S_inode_s)
	err := inodo.Decodificar(archivoDisco, offset)
	if err != nil {
		return nil, err
	}
	return inodo, nil
}

func buscarInodoEnDirectorioLS(inodo *Structs.Inodo, archivoDisco *os.File, nombre string, superbloque *Structs.Superbloque) (bool, int32) {
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			continue
		}
		bloque := &Structs.BloqueFolder{}
		offset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		err := bloque.Decodificar(archivoDisco, offset)
		if err != nil {
			continue
		}
		for _, contenido := range bloque.B_contenido {
			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			if nombreContenido == nombre {
				return true, contenido.B_inodo
			}
		}
	}
	return false, -1
}
