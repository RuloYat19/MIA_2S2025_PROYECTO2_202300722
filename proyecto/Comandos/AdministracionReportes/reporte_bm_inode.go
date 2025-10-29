package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func ReporteBMInodo(superbloque *Structs.Superbloque, rutaDisco string, rutaSalida string) error {
	// Se crean las carpetas padre si no existen
	err := Herramientas.CrearPadreDirs(rutaSalida)

	if err != nil {
		return fmt.Errorf("error creando carpetas padre: %v", err)
	}

	// Se abre el archivo de disco
	archivo, err := os.Open(rutaDisco)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}

	defer archivo.Close()

	// Se calcula el número total de inodos
	totalInodos := superbloque.S_inodes_count + superbloque.S_free_inodes_count

	// Se calcula cuántos bytes necesita el bitmap ya que cada byte tiene 8 bits
	conteoBytes := (totalInodos + 7) / 8

	// Variable para almacenar el contenido del reporte del bitmap de inodos
	var contenidoBitmap strings.Builder

	for indiceByte := int32(0); indiceByte < conteoBytes; indiceByte++ {
		// Se mueve el puntero al byte correspondiente en el bitmap de inodos
		_, err := archivo.Seek(int64(superbloque.S_bm_inode_start+indiceByte), 0)
		if err != nil {
			return fmt.Errorf("error al posicionar el archivo: %v", err)
		}

		// Se lee un byte del bitmap
		var byteVal byte
		err = binary.Read(archivo, binary.LittleEndian, &byteVal)
		if err != nil {
			return fmt.Errorf("error al leer el byte del bitmap: %v", err)
		}

		// Se procesa cada bit del byte (cada bit representa un inodo)
		for bitOffset := 0; bitOffset < 8; bitOffset++ {
			// Se verifica si estamos fuera del rango total de inodos
			if indiceByte*8+int32(bitOffset) >= totalInodos {
				break
			}

			// Si el bit es 1, el inodo está ocupado, si es 0, está libre
			if (byteVal & (1 << bitOffset)) != 0 {
				contenidoBitmap.WriteByte('1') // Inodo ocupado
			} else {
				contenidoBitmap.WriteByte('0') // Inodo libre
			}

			// Se añade salto de línea cada 20 inodos
			if (indiceByte*8+int32(bitOffset)+1)%20 == 0 {
				contenidoBitmap.WriteString("\n")
			}
		}
	}

	// Se guarda el reporte en el archivo especificado
	archivoTXT, err := os.Create(rutaSalida)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer archivoTXT.Close()

	_, err = archivoTXT.WriteString(contenidoBitmap.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte del bitmap de inodos generado correctamente:", rutaSalida)

	return nil
}
