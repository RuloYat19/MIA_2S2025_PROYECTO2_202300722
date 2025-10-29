package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func ReporteBMBloque(superbloque *Structs.Superbloque, rutaDisco string, rutaSalida string) error {
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

	// Se calcula el número total de bloques
	totalBlocks := superbloque.S_blocks_count + superbloque.S_free_blocks_count

	// Se calculan cuántos bytes necesita el bitmap ya que cada byte tiene 8 bits
	byteCount := (totalBlocks + 7) / 8

	// Variable para almacenar el contenido del reporte del bitmap de bloques
	var bitmapContent strings.Builder

	for byteIndex := int32(0); byteIndex < byteCount; byteIndex++ {
		// Se mueve el puntero al byte correspondiente en el bitmap de bloques
		_, err := archivo.Seek(int64(superbloque.S_bm_block_start+byteIndex), 0)
		if err != nil {
			return fmt.Errorf("error al posicionar el archivo: %v", err)
		}

		// Se lee un byte del bitmap
		var byteVal byte
		err = binary.Read(archivo, binary.LittleEndian, &byteVal)
		if err != nil {
			return fmt.Errorf("error al leer el byte del bitmap: %v", err)
		}

		// Se procesa cada bit del byte ya que cada bit representa un bloque
		for bitOffset := 0; bitOffset < 8; bitOffset++ {
			// Se verifica si estamos fuera del rango total de bloques
			if byteIndex*8+int32(bitOffset) >= totalBlocks {
				break
			}

			// Si el bit es 1, el bloque está ocupado ('1'), si es 0, está libre ('0')
			if (byteVal & (1 << bitOffset)) != 0 {
				bitmapContent.WriteByte('1') // Bloque ocupado
			} else {
				bitmapContent.WriteByte('0') // Bloque libre
			}

			// Se añade el salto de línea cada 20 bloques
			if (byteIndex*8+int32(bitOffset)+1)%20 == 0 {
				bitmapContent.WriteString("\n")
			}
		}
	}

	// Se guarda el reporte en el archivo especificado
	archivoTXT, err := os.Create(rutaSalida)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer archivoTXT.Close()

	_, err = archivoTXT.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte del bitmap de bloques generado correctamente:", rutaSalida)
	return nil
}
