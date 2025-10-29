package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type JOURNALING struct {
	Id string `json:"id"`
}

type EntradaJournaling struct {
	Operation string `json:"operation"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Date      string `json:"date"`
}

func Journaling(parametros []string) string {
	fmt.Println("\n======= JOURNALING =======")

	var salida1 = ""

	journaling := &JOURNALING{}

	//Otras variables
	paramC := true
	rutaInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Asignación de id
		if strings.ToLower(tmp[0]) == "id" {
			rutaInit = true
			journaling.Id = tmp[1]

		} else {
			salida1 += "JOURNALING Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("JOURNALING Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && rutaInit {
		salidaJournaling, err := ejecutarJournaling(journaling)

		if err != nil {
			salida1 += "JOURNALING Error: Hubo problemas al mostrar el journaling.\n"
		} else {

			rutaJournaling := "/home/rauly/USAC/Archivos/MIA_2S2025_P2_202300722/Calificacion_MIA/Reportes/Journaling.md"

			reporteJournaling, err := os.Create(rutaJournaling)

			if err != nil {
				salida1 += "JOURNALING ERROR: Hubo problemas creando el archivo del reporte del journaling.\n"
				return salida1
			}

			defer reporteJournaling.Close()

			_, err = reporteJournaling.WriteString(salidaJournaling)
			if err != nil {
				salida1 += "JOURNALING ERROR: Hubo problemas escribiendo en el archivo del reporte del journaling.\n"
				return salida1
			}

			salida1 += "Se ha creado el reporte del journaling con éxito.\n"
		}
	}

	fmt.Println("\n======FIN JOURNALING======")
	return salida1
}

func ejecutarJournaling(journaling *JOURNALING) (string, error) {
	// Se obtiene el superbloque de la partición
	superbloque, particion, ruta, err := globales.ObtenerParticionMontadaSuperbloque(journaling.Id)
	if err != nil {
		fmt.Println("Hubo problemas al obtener la partición.")
		return "", fmt.Errorf("error obteniendo la partición: %w", err)
	}

	// Se verifica que la partición sea de tipo EXT3
	if superbloque.S_filesystem_type != 3 {
		fmt.Println("La partición no es de tipo EXT3, por lo tanto no tiene journaling.")
		return "", fmt.Errorf("la partición no es de tipo EXT3, no tiene journaling: %w", err)
	}

	// Se abre el archivo para lectura
	archivo, err := os.OpenFile(ruta, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Hubo problemas al abrir el archivo.")
		return "", fmt.Errorf("error abriendo el archivo: %w", err)
	}
	defer archivo.Close()

	// Se calcula la posición de inicio del journal
	journalStart := int64(particion.Start) + int64(binary.Size(Structs.Superbloque{}))

	fmt.Printf("Leyendo journal desde posición %d\n", journalStart)

	// Se cuscan todas las entradas válidas del journal
	entradas, err := Structs.EncontrarEntradaValidaJournal(archivo, journalStart, Structs.JOURNAL_ENTRIES)
	if err != nil {
		fmt.Println("Hubo problemas al buscar las entradas del journal.")
		return "", fmt.Errorf("error buscando entradas de journal: %w", err)
	}

	// Si no hay entradas, devolver un mensaje
	if len(entradas) == 0 {
		return "No hay entradas de journal para mostrar", nil
	}

	// Se convierten las entradas a un formato más amigable para la interfaz
	var salida string
	salida += "| Operación |    Ruta    |    Contenido    |    Fecha   |\n"
	salida += "|-----------|------------|-----------------|------------|\n"
	var resultado []EntradaJournaling
	for _, entrada := range entradas {
		operation := cleanCString(entrada.J_content.I_operation[:])
		path := cleanCString(entrada.J_content.I_path[:])
		content := cleanCString(entrada.J_content.I_content[:])

		date := time.Unix(int64(entrada.J_content.I_date), 0)
		dateStr := date.Format(time.RFC3339)

		resultado = append(resultado, EntradaJournaling{
			Operation: operation,
			Path:      path,
			Content:   content,
			Date:      dateStr,
		})

		salida += " | " + operation + " | " + path + " | " + content + " | " + dateStr + " | " + "\n"
	}

	fmt.Printf("Se encontraron %d entradas válidas de journal\n", len(resultado))

	return salida, nil
}

// Se eliminan los bytes nulos al final de un array
func cleanCString(buf []byte) string {
	return strings.TrimSpace(
		string(bytes.TrimRight(buf, "\x00")),
	)
}
