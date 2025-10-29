package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ReporteMBR(mbr *Structs.MBR, ruta string, archivo *os.File) error {
	err := Herramientas.CrearPadreDirs(ruta)

	if err != nil {
		return err
	}

	// Se obtiene el nombre base del archivo sin la extensión
	dotFileName, outputImage := Herramientas.ObtenerNombreArchivos(ruta)

	// Se define la paleta de colores
	colorParticionPrimaria := "#FFDDC1"
	colorParticionExtendida := "#C1E1C1"
	colorParticionLogica := "#C1D1FF"
	colorEBR := "#FFD1DC"
	colorEspacioNoAsignado := "#FFFFFF"

	// Se define el contenido DOT con una tabla
	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext]
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" bgcolor="#F8D7DA"><b>REPORTE MBR</b></td></tr>
                <tr><td bgcolor="#F5B7B1">mbr_tamano</td><td bgcolor="#F5B7B1">%d</td></tr>
                <tr><td bgcolor="#F5B7B1">mrb_fecha_creacion</td><td bgcolor="#F5B7B1">%s</td></tr>
                <tr><td bgcolor="#F5B7B1">mbr_disk_signature</td><td bgcolor="#F5B7B1">%d</td></tr>
            `, mbr.MbrSize, string(mbr.FechaC[:]), mbr.Id)

	// Se calcula el tamaño total del disco y mantiene un seguimiento del espacio no asignado
	tamanioTotal := mbr.MbrSize
	espacioAsignado := int32(0)

	// Agregar las particiones a la tabla
	for i, particion := range mbr.Particiones {
		// Solo se incluyen particiones válidas
		if particion.Tamanio > 0 && particion.Start > 0 {
			// Se calcula el espacio no asignado antes de esta partición
			if particion.Start > espacioAsignado {
				espacioNoAsignado := particion.Start - espacioAsignado
				dotContent += fmt.Sprintf(`
                    <tr><td colspan="2" bgcolor="%s"><b>ESPACIO NO ASIGNADO (Tamaño: %d bytes)</b></td></tr>
                `, colorEspacioNoAsignado, espacioNoAsignado)
				espacioAsignado += espacioNoAsignado
			}

			// Se convierte Nombre de la partición a string y se eliminan los caracteres nulos
			nombreParticion := strings.TrimRight(string(particion.Nombre[:]), "\x00")

			// Se convierte Estado, Tipo y Fit de la partición a char
			estadoParticion := rune(particion.Estado[0])
			tipoParticion := rune(particion.Tipo[0])
			fitParticion := rune(particion.Fit[0])

			// Se define el color de fondo dependiendo del tipo de partición
			colorFila := ""

			switch tipoParticion {
			case 'P':
				colorFila = colorParticionPrimaria
			case 'E':
				colorFila = colorParticionExtendida
			}

			// Se agrega la partición a la tabla
			dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="%s"><b>PARTICIÓN %d</b></td></tr>
                <tr><td bgcolor="%s">part_status</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_type</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_fit</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_start</td><td bgcolor="%s">%d</td></tr>
                <tr><td bgcolor="%s">part_size</td><td bgcolor="%s">%d</td></tr>
                <tr><td bgcolor="%s">part_name</td><td bgcolor="%s">%s</td></tr>
            `, colorFila, i+1,
				colorFila, colorFila, estadoParticion,
				colorFila, colorFila, tipoParticion,
				colorFila, colorFila, fitParticion,
				colorFila, colorFila, particion.Start,
				colorFila, colorFila, particion.Tamanio,
				colorFila, colorFila, nombreParticion)

			espacioAsignado += particion.Tamanio

			// Si es una partición extendida, se muestran los EBRs y las particiones lógicas
			if tipoParticion == 'E' {
				ebrStart := particion.Start
				dotContent += fmt.Sprintf(`
                    <tr><td colspan="2" bgcolor="%s"><b>PART. EXTENDIDA (Inicio: %d)</b></td></tr>
                `, colorParticionExtendida, ebrStart)

				// Se itera sobre los EBRs
				for ebrStart != -1 {
					ebr := &Structs.EBR{}
					err := ebr.Decodificar(archivo, int64(ebrStart))

					if err != nil {
						return fmt.Errorf("error al leer EBR: %v", err)
					}

					nombreEBR := strings.TrimRight(string(ebr.Ebr_name[:]), "\x00")
					fitEBR := rune(ebr.Ebr_fit[0])

					// Se muestra la información del EBR
					dotContent += fmt.Sprintf(`
                        <tr><td colspan="2" bgcolor="%s"><b>EBR (Inicio: %d)</b></td></tr>
                        <tr><td bgcolor="%s">ebr_fit</td><td bgcolor="%s">%c</td></tr>
                        <tr><td bgcolor="%s">ebr_start</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_size</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_next</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_name</td><td bgcolor="%s">%s</td></tr>
                    `, colorEBR, ebrStart,
						colorEBR, colorEBR, fitEBR,
						colorEBR, colorEBR, ebr.Ebr_start,
						colorEBR, colorEBR, ebr.Ebr_size,
						colorEBR, colorEBR, ebr.Ebr_next,
						colorEBR, colorEBR, nombreEBR)

					// Si hay una partición lógica después del EBR
					if ebr.Ebr_size > 0 {
						dotContent += fmt.Sprintf(`
                            <tr><td colspan="2" bgcolor="%s"><b>PART. LÓGICA (Inicio: %d)</b></td></tr>
                        `, colorParticionLogica, ebr.Ebr_start)
					}

					espacioAsignado += ebr.Ebr_size
					ebrStart = int32(ebr.Ebr_next)
				}
			}
		}

	}

	// Se calcula el espacio no asignado restante al final del disco
	if espacioAsignado < tamanioTotal {
		espacioNoAsignado := tamanioTotal - espacioAsignado
		dotContent += fmt.Sprintf(`
            <tr><td colspan="2" bgcolor="%s"><b>ESPACIO NO ASIGNADO (Tamaño: %d bytes)</b></td></tr>
        `, colorEspacioNoAsignado, espacioNoAsignado)
	}

	// Se cierra la tabla y el contenido DOT
	dotContent += "</table>>] }"

	// Se guarda el contenido DOT en un archivo
	archivo, err = os.Create(dotFileName)

	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}

	defer archivo.Close()

	_, err = archivo.WriteString(dotContent)

	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Se ejecuta el comando Graphviz para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)

	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	fmt.Println("Imagen de la tabla generada:", outputImage)

	return nil
}
