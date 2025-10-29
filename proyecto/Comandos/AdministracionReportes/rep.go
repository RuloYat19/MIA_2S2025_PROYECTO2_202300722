package Comandos

import (
	globales "Proyecto/Globales"
	"fmt"
	"os"
	"strings"
)

type REP struct {
	id              string
	ruta            string
	nombre          string
	ruta_archivo_ls string
}

func Rep(parametros []string) string {
	fmt.Println("\n===== REP =====")

	var salida1 = ""

	rep := &REP{}

	//Opcionales
	rep.ruta_archivo_ls = ""

	//Otras variables
	paramC := true
	idInit := false
	rutaInit := false
	nombreInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "r" {

		} else if len(tmp) != 2 {
			salida1 += "REP Error: Valor desconocido del parámetro" + tmp[0] + ".\n"
			fmt.Println("REP Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}
		//fmt.Println("El valor del tmp[0] es: ", tmp[0])

		//Asignación de id
		if strings.ToLower(tmp[0]) == "id" {
			idInit = true
			rep.id = tmp[1]

			//Asignación de path
		} else if strings.ToLower(tmp[0]) == "path" {
			rutaInit = true
			rep.ruta = tmp[1]

			//Asignación de name
		} else if strings.ToLower(tmp[0]) == "name" {
			nombreInit = true
			rep.nombre = tmp[1]

			//Asignación de path_file_ls
		} else if strings.ToLower(tmp[0]) == "path_file_ls" {
			rep.ruta_archivo_ls = tmp[1]

		} else {
			salida1 += "REP Error: Parámetro desconcido: " + tmp[0] + ".\n"
			fmt.Println("REP Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && idInit && rutaInit && nombreInit {
		salida2, err := crearReportes(rep)

		if err != nil {
			fmt.Println("REP Error: Hubo problemas generando el reporte.")
			salida2 += "REP Error: Hubo problemas generando el reporte.\n"
		}

		return salida1 + salida2
	}

	return salida1
}

func crearReportes(rep *REP) (string, error) {
	/*fmt.Println("Los valores que están entrando a REP son:")
	fmt.Println("ID: ", rep.id)
	fmt.Println("RUTA: ", rep.ruta)
	fmt.Println("NOMBRE: ", rep.nombre)
	fmt.Println("RUTA ARCHIVO LS: ", rep.ruta_archivo_ls)*/

	var salida2 = ""

	// Se obtiene la partición montada
	MBRMontada, SuperbloqueMontado, rutaDiscoMontado, err := globales.ObtenerParticionMontadaRep(rep.id)

	if err != nil {
		fmt.Println("Hubo problemas obteniendo la partición montada.")
		return salida2, err
	}

	// Se abre el archivo de la partición
	archivo, err := os.Open(rutaDiscoMontado)

	if err != nil {
		fmt.Println("REP Error: Hubo problemas al abrir el disco.")
		return salida2, err
	}

	defer archivo.Close()

	// Mensaje de inicio de generación de reporte
	fmt.Printf("Generando el reporte '%s'.\n", rep.nombre)

	switch rep.nombre {
	case "mbr":
		// Reporte del MBR
		err = ReporteMBR(MBRMontada, rep.ruta, archivo)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte MBR.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte MBR: %v\n", err)
			return salida2, err
		}
	case "disk":
		// Reporte del Disco
		err = ReporteDisk(MBRMontada, rep.ruta, rutaDiscoMontado)

		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del disco.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del disco: %v\n", err)
			return salida2, err
		}
	case "inode":
		// Reporte de Inodos
		err = ReporteInodo(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte de inodos.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte de inodos: %v.\n", err)
			return salida2, err
		}
	case "block":
		// Reporte de Bloques
		err = ReporteBloque(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte de bloques.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte de bloques: %v.\n", err) // Depuración
			return salida2, err
		}
	case "bm_inode":
		// Reporte del Bitmap de Inodos
		err = ReporteBMInodo(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del bitmap de inodos.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del bitmap de inodos: %v.\n", err) // Depuración
			return salida2, err
		}
	case "bm_block":
		// Reporte del Bitmap de Bloques
		err = ReporteBMBloque(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del bitmap de bloques.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del bitmap de bloques: %v.\n", err) // Depuración
			return salida2, err
		}
	case "sb":
		// Reporte del Superbloque
		err = ReporteSb(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del superbloque.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del superbloque: %v.\n", err) // Depuración
			return salida2, err
		}
	case "file":
		// Reporte de Archivo
		err = ReporteFile(SuperbloqueMontado, rutaDiscoMontado, rep.ruta, rep.ruta_archivo_ls)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte de archivo.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte de archivo: %v.\n", err) // Depuración
			return salida2, err
		}
	case "tree":
		// Reporte de Archivo
		err = ReporteTree(SuperbloqueMontado, rutaDiscoMontado, rep.ruta)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del tree.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del tree: %v.\n", err) // Depuración
			return salida2, err
		}
	case "ls":
		// Reporte de Archivo
		err = ReporteLS(SuperbloqueMontado, rutaDiscoMontado, rep.ruta, rep.ruta_archivo_ls)
		if err != nil {
			salida2 += "REP Error: Hubo problemas generando el reporte del ls.\n"
			fmt.Printf("REP Error: Hubo problemas generando el reporte del ls: %v.\n", err) // Depuración
			return salida2, err
		}
	default:
		salida2 += "REP Error: Tipo de reporte incorrecto o mal escrito" + rep.nombre + ".\n"
		fmt.Printf("REP Error: Tipo de reporte incorrecto o mal escrito: %s", rep.nombre)
		return salida2, err
	}

	// Mensaje de éxito en la generación de reporte
	salida2 += "Reporte " + rep.nombre + " generado exitosamente en la ruta: " + rep.ruta + ".\n"
	fmt.Printf("Reporte '%s' generado exitosamente en la ruta: %s\n", rep.nombre, rep.ruta)
	return salida2, nil
}
