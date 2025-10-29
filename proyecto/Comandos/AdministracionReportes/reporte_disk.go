package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ReporteDisk(mbr *Structs.MBR, ruta string, rutaDisco string) error {
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

	// Se obtiene el nombre base del archivo sin la extensión
	dotFileName, outputImage := Herramientas.ObtenerNombreArchivos(ruta)

	dotContent := `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif"]
		edge [fontname="Helvetica,Arial,sans-serif"]
		concentrate=True;
		rankdir=TB;
		node [shape=record];

		title [label="Reporte DISK" shape=plaintext fontname="Helvetica,Arial,sans-serif"];

		dsk [label="`

	// Se calcula el tamaño total del disco y el tamaño usado
	tamanioTotal := mbr.MbrSize
	tamanioUsado := int32(0)

	// Se agrega MBR al reporte
	dotContent += "{MBR}"

	// Se recorren las particiones del MBR y se genera el contenido DOT
	for _, part := range mbr.Particiones {
		if part.Tamanio > 0 {
			// Se calcula el porcentaje de uso
			porcentaje := (float64(part.Tamanio) / float64(tamanioTotal)) * 100
			tamanioUsado += part.Tamanio

			// Se convierte Nombre a string y se eliminan los caracteres nulos
			nombreParticion := strings.TrimRight(string(part.Nombre[:]), "\x00")

			if part.Tipo[0] == 'P' {
				// Partición primaria
				dotContent += fmt.Sprintf("|{Primaria %s\\n%.2f%%}", nombreParticion, porcentaje)
			} else if part.Tipo[0] == 'E' {
				// Partición extendida
				dotContent += fmt.Sprintf("|{Extendida %.2f%%|{", porcentaje)
				ebrStart := part.Start
				ebrConteo := 0
				ebrTamanioUsado := int32(0)

				for ebrStart != -1 {
					ebr := &Structs.EBR{}
					err := ebr.Decodificar(archivo, int64(ebrStart))
					if err != nil {
						return fmt.Errorf("error al leer EBR: %v", err)
					}

					nombreEBR := strings.TrimRight(string(ebr.Ebr_name[:]), "\x00")
					porcentajeEBR := (float64(ebr.Ebr_size) / float64(tamanioTotal)) * 100
					ebrTamanioUsado += ebr.Ebr_size

					// Se agregan el EBR y la partición lógica al reporte
					if ebrConteo > 0 {
						dotContent += "|"
					}

					dotContent += fmt.Sprintf("{EBR|Lógica %s\\n%.2f%%}", nombreEBR, porcentajeEBR)

					// Se actualiza el inicio para el próximo EBR
					ebrStart = ebr.Ebr_next
					ebrConteo++
				}

				// Calcular espacio libre dentro de la partición extendida
				tamanioLibreEBR := part.Tamanio - ebrTamanioUsado
				if tamanioLibreEBR > 0 {
					porcentajeLibreEBR := (float64(tamanioLibreEBR) / float64(tamanioTotal)) * 100
					dotContent += fmt.Sprintf("|Libre %.2f%%", porcentajeLibreEBR)
				}

				dotContent += "}}"
			}
		}
	}

	// Se calcula el espacio libre restante y se añade si es necesario
	tamanioLibre := tamanioTotal - tamanioUsado
	if tamanioLibre > 0 {
		porcentajeLibre := (float64(tamanioLibre) / float64(tamanioTotal)) * 100
		dotContent += fmt.Sprintf("|Libre %.2f%%", porcentajeLibre)
	}

	// Se cierra el nodo de disco y se completa el DOT
	dotContent += `"];

		title -> dsk [style=invis];
	}`

	// Se crea el archivo DOT
	dotFile, err := os.Create(dotFileName)

	if err != nil {
		return fmt.Errorf("error al crear el archivo DOT: %v", err)
	}

	defer dotFile.Close()

	// Se escribe el contenido DOT en el archivo
	_, err = dotFile.WriteString(dotContent)

	if err != nil {
		return fmt.Errorf("error al escribir en el archivo DOT: %v", err)
	}

	// Se genera la imagen con Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)

	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	fmt.Println("Reporte de disco generado:", outputImage)
	return nil
}
