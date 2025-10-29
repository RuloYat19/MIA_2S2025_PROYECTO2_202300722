package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Fdisk(parametros []string) string {
	fmt.Println("\n======= FDISK =======")
	var salida = ""

	// Datos Comando
	// Obligatorios
	var size int
	var path string
	var name string
	// Opcionales
	fit := "W"
	unit := 1024
	tipoUnit := ""
	typ := "P"
	delete := ""
	add := 0

	// Otras variables
	paramC := true
	sizeInit := false
	pathInit := false
	nameInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida += "FDISK Error: Valor desconocido del parámetro " + tmp[0] + "\n"
			fmt.Println("FDISK Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "size" {
			sizeInit = true
			var err error
			size, err = strconv.Atoi(tmp[1])

			if err != nil {
				salida += "FDISK Error: -size debe ser un valor numérico. Se leyó: " + tmp[1] + "\n"
				fmt.Println("FDISK Error: -size debe ser un valor numerico. se leyo: ", tmp[1])
				paramC = false
				break
			} else if size <= 0 {
				salida += "FDISK Error: -size debe ser un valor positivo mayor a cero (0). Se leyó: " + tmp[1] + "\n"
				fmt.Println("FDISK Error: -size debe ser un valor positivo mayor a cero (0). se leyo: ", tmp[1])
				paramC = false
				break
			}
		} else if strings.ToLower(tmp[0]) == "fit" {
			if strings.ToLower(tmp[1]) == "bf" {
				fit = "B"
			} else if strings.ToLower(tmp[1]) == "ff" {
				fit = "F"
			} else if strings.ToLower(tmp[1]) != "wf" {
				salida += "FDISK Error: Valor incorrecto de -fit. Los valores aceptados son: BF, FF o WF. Se leyó: " + tmp[1] + "\n"
				fmt.Println("FDISK Error en -fit. Valores aceptados: BF, FF o WF. ingreso: ", tmp[1])
				paramC = false
				break
			}
		} else if strings.ToLower(tmp[0]) == "unit" {
			if strings.ToLower(tmp[1]) == "m" {
				tipoUnit = "M"
				unit = 1048576
			} else if strings.ToLower(tmp[1]) == "k" {
				tipoUnit = "K"
				unit = 1024
			} else if strings.ToLower(tmp[1]) == "b" {
				tipoUnit = "B"
				unit = 1
			} else {
				salida += "FDISK Error: Valor incorrecto de -unit. Los valores aceptados son: k, m y b. Se leyó: " + tmp[1] + "\n"
				fmt.Println("FDISK Error en -unit. Valores aceptados: k, m y b. ingreso: ", tmp[1])
				paramC = false
				break
			}
		} else if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			path = tmp[1]
		} else if strings.ToLower(tmp[0]) == "name" {
			nameInit = true
			name = tmp[1]
		} else if strings.ToLower(tmp[0]) == "type" {
			if strings.ToLower(tmp[1]) == "e" {
				typ = "E"
			} else if strings.ToLower(tmp[1]) == "l" {
				typ = "L"
			} else if strings.ToLower(tmp[1]) != "p" {
				salida += "FDISK Error: Valor incorrecto de -type. Los valores aceptados son: e, l o p. Se leyó: " + tmp[1] + "\n"
				fmt.Println("FDISK Error en -type. Valores aceptados: e, l o p. ingreso: ", tmp[1])
				paramC = false
				break
			}
		} else if strings.ToLower(tmp[0]) == "delete" {
			if tmp[1] == "fast" || tmp[1] == "full" {
				delete = tmp[1]
			} else {
				salida += "FDISK Error: Parámetro no válido para delete: " + tmp[1] + ". Los valores válido son fast o full.\n"
				break
			}
		} else if strings.ToLower(tmp[0]) == "add" {
			valorAdd, err := strconv.Atoi(tmp[1])

			if err != nil {
				salida += "FDISK Error: El valor de add tiene que ser un número entero. Se leyó: " + tmp[1] + ".\n"
			} else {
				add = valorAdd
			}

		} else {
			salida += "FDISK Error: Parámetro desconocido: " + tmp[0] + "\n"
			fmt.Println("FDISK Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if delete != "" && nameInit && pathInit {
		err := eliminarParticion(name, path, delete)
		if err != nil {
			salida += "FDISK Error: Hubo problemas al eliminar la partición del disco.\n"
		} else {
			salida += "La partición fue correctamente eliminada del disco con el método " + delete + ".\n"
			return salida
		}
	}

	if add != 0 {
		err := modificarParticion(name, add, path, tipoUnit)
		if err != nil {
			salida += "FDISK Error: Hubo problemas al modificar el tamaño de la partición del disco.\n"
		} else {
			salida += "El tamaño de la partición fue correctamente modificado.\n"
			return salida
		}
	}

	if paramC && sizeInit && pathInit && nameInit {
		fmt.Printf("Creando partición con nombre '%s' y tamaño %d * %d\n", name, size, unit)
		fmt.Println("Detalles internos de la creación de la partición", size, unit, fit, path, typ, name)

		archivo, err := os.OpenFile(path, os.O_RDWR, 0644)

		if err != nil {
			fmt.Printf("Error abriendo el archivo del disco: %v", err)
		}
		defer archivo.Close()

		tam := size * unit

		switch typ {
		case "P":
			err = crearParticionPrimaria(name, fit, typ, tam, archivo)

			if err != nil {
				salida += "FDISK Error: Hubo problemas creando la partición primaria.\n"
				fmt.Println("Error creando partición primaria:", err)
			} else {
				salida += "La partición primaria fue creada exitosamente.\n"
			}
		case "E":
			err = crearParticionExtendida(name, fit, typ, tam, archivo)

			if err != nil {
				salida += "FDISK Error: Hubo problemas creando la partición extendida.\n"
				fmt.Println("Error creando partición extendida:", err)
			} else {
				salida += "La partición extendida fue creada exitosamente.\n"
			}
		case "L":
			err = crearParticionLogica(name, fit, tam, archivo)

			if err != nil {
				salida += "FDISK Error: Hubo problemas creando la partición lógica.\n"
				fmt.Println("Error creando partición lógica:", err)
			} else {
				salida += "La partición lógica fue creada exitosamente.\n"
			}
		}
	}
	fmt.Println("\n======FIN FDISK=======")
	return salida
}

func modificarParticion(nombre string, add int, ruta string, tipoUnidad string) error {
	fmt.Printf("Modificando partición '%s', ajustando %d unidades...\n", nombre, add)

	// Se abre el archivo del disco
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error al abrir el archivo del disco.")
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}

	defer archivo.Close()

	// Se lee el MBR del archivo
	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Println("Error al deserializar el MBR.")
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Se busca la partición por nombre
	particion := mbr.ObtenerParticionNombre(nombre)

	if particion == nil {
		return fmt.Errorf("la partición '%s' no existe", nombre)
	}

	// Se convierte add a bytes según la unidad especificada
	addBytes, err := Herramientas.ConvertToBytes(add, tipoUnidad)

	if err != nil {
		fmt.Println("Error al convertir las unidades de -add.")
		return fmt.Errorf("error al convertir las unidades de -add: %v", err)
	}

	// Se calcula el espacio disponible si se está agregando espacio
	var espacioDisponible int32 = 0

	if addBytes > 0 {
		espacioDisponible, err = mbr.CalcularEspacioDisponibleDeLaParticion(particion)
		if err != nil {
			fmt.Println("Error al calcular el espacio disponible para la partición.")
			return fmt.Errorf("error al calcular el espacio disponible para la partición '%s': %v", nombre, err)
		}
	}

	// Se modifica el tamaño de la partición
	err = particion.ModificarTamanioDeLaParticion(int32(addBytes), espacioDisponible)

	if err != nil {
		fmt.Println("Error al modificar el tamaño de la partición.")
		return fmt.Errorf("error al modificar el tamaño de la partición: %v", err)
	}

	// Se actualiza el MBR con la partición ya eliminada
	for i := range mbr.Particiones {
		nombreParticion := mbr.Particiones[i].Nombre[:]
		nombreLimpio := string(bytes.TrimRight(nombreParticion, "\x00"))
		if nombreLimpio == nombre {
			mbr.Particiones[i].Tamanio = mbr.Particiones[i].Tamanio + int32(add)
		}
	}

	// Se actualiza el MBR en el archivo después de la modificación
	err = mbr.Codificar(archivo)

	if err != nil {
		fmt.Println("Error al actualizar el MBR en el disco.")
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	// Mensajes
	fmt.Printf("La partición '%s' ha sido modificada exitosamente.", nombre)
	fmt.Println("\n======= DISCO =======")
	Structs.ImprimirMBR(mbr)
	fmt.Println("\n======FIN FDISK=======")

	return nil
}

// Se maneja la eliminación de particiones
func eliminarParticion(nombre string, ruta string, delete string) error {
	fmt.Printf("Eliminando partición con nombre '%s' usando el método %s...\n", nombre, delete)

	// Se abre el archivo del disco
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error abriendo el archivo del disco.")
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}

	defer archivo.Close()

	// Se lee el MBR del archivo
	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Println("Error al deserializar el MBR del disco.")
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Se busca la partición por nombre
	particion := mbr.ObtenerParticionNombre(nombre)

	if particion == nil {
		fmt.Printf("La partiión '%s' no existe en el disco.\n", nombre)
		return fmt.Errorf("la partición '%s' no existe", nombre)
	}

	// Se verifica si es extendida para eliminar particiones lógicas
	isExtended := particion.Tipo[0] == 'E'

	err = particion.EliminarParticion(delete, archivo, isExtended)

	if err != nil {
		fmt.Println("Error al eliminar la partición.")
		return fmt.Errorf("error al eliminar la partición: %v", err)
	}

	// Se actualiza el MBR con la partición ya eliminada
	for i := range mbr.Particiones {
		nombreParticion := mbr.Particiones[i].Nombre[:]
		nombreLimpio := string(bytes.TrimRight(nombreParticion, "\x00"))
		if nombreLimpio == nombre {
			mbr.Particiones[i].Nombre = [16]byte{}
			mbr.Particiones[i].Tipo = [1]byte{}
			mbr.Particiones[i].Start = 0
			mbr.Particiones[i].Tamanio = 0
			mbr.Particiones[i].Fit = [1]byte{}
			mbr.Particiones[i].Correlativo = 0
		}
	}

	// Se actualiza el MBR en el archivo después de la eliminación
	err = mbr.Codificar(archivo)
	if err != nil {
		fmt.Println("Error al actualizar el MBR en el disco.")
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	// Mensajes
	//fmt.Printf("Partición '%s' fue eliminada exitosamente.\n", nombre)
	fmt.Println("\n======= DISCO =======")
	Structs.ImprimirMBR(mbr)
	fmt.Println("\n======FIN FDISK=======")

	return nil
}

func crearParticionPrimaria(name string, fit string, typ string, tam int, file *os.File) error {
	var mbr Structs.MBR

	err := mbr.Decodificar(file)

	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	particionDisponible, particionStart, particionIndice := mbr.ObtenerPrimerParticionDisponible()

	if particionDisponible == nil {
		return errors.New("no hay espacio disponible en el MBR para una nueva partición")
	}

	espacioDisponible, err := mbr.CalcularEspacioDisponible()
	if err != nil {
		return err
	}

	if int32(tam) > espacioDisponible {
		return errors.New("no hay suficiente espacio para la partición primaria")
	}

	particionDisponible.CrearParticion(particionStart, tam, typ, fit, name)
	mbr.Particiones[particionIndice] = *particionDisponible

	err = mbr.Codificar(file)
	if err != nil {
		fmt.Println("Error actualizando el MBR en el disco:", err)
		return err
	}

	fmt.Println("Partición primaria creada exitosamente.")

	Structs.ImprimirMBR(mbr)

	return nil
}

func crearParticionExtendida(name string, fit string, typ string, tam int, file *os.File) error {
	var mbr Structs.MBR

	err := mbr.Decodificar(file)

	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	if mbr.TieneParticionExtendida() {
		return errors.New("ya existe una partición extendida en este disco")
	}

	espacioDisponible, err := mbr.CalcularEspacioDisponible()

	if err != nil {
		return err
	}

	if int32(tam) > espacioDisponible {
		return errors.New("no hay suficiente espacio para la partición extendida")
	}

	particionDisponible, particionStart, particionIndice := mbr.ObtenerPrimerParticionDisponible()

	if particionDisponible == nil {
		return errors.New("no hay espacio disponible en el MBR para una nueva partición")
	}

	particionDisponible.CrearParticion(particionStart, tam, typ, fit, name)
	mbr.Particiones[particionIndice] = *particionDisponible

	err = Structs.CrearYEscribirEBR(int32(particionStart), 0, fit[0], name, file)

	if err != nil {
		return fmt.Errorf("error al crear el primer EBR en la partición extendida: %v", err)
	}

	err = mbr.Codificar(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	fmt.Println("Partición extendida creada exitosamente.")

	Structs.ImprimirMBR(mbr)

	return nil
}

func crearParticionLogica(name string, fit string, tam int, file *os.File) error {
	var mbr Structs.MBR

	err := mbr.Decodificar(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Se verifica si existe una partición extendida utilizando TieneParticionExtendida
	if !mbr.TieneParticionExtendida() {
		fmt.Println("No se encontró una partición extendida en el disco.")
		return errors.New("no se encontró una partición extendida en el disco")
	}

	// Se identifica la partición extendida específica
	var particionExtendida *Structs.Particion
	for i := range mbr.Particiones {
		if mbr.Particiones[i].Tipo[0] == 'E' {
			particionExtendida = &mbr.Particiones[i]
			break
		}
	}

	// Se busca el último EBR en la partición extendida
	ultimoEBR, err := Structs.EncontrarUltimoEBR(particionExtendida.Start, file)
	if err != nil {
		fmt.Println("Error al buscar el último EBR.")
		return fmt.Errorf("error al buscar el último EBR: %v", err)
	}

	if ultimoEBR.Ebr_size == 0 {
		fmt.Println("Detectado EBR inicial vacío, asignando tamaño a la nueva partición lógica.") // Mensaje de depuración
		ultimoEBR.Ebr_size = int32(tam)
		copy(ultimoEBR.Ebr_name[:], name)

		err = ultimoEBR.Codificar(file, int64(ultimoEBR.Ebr_start))
		if err != nil {
			return fmt.Errorf("error al escribir el primer EBR con la nueva partición lógica: %v", err)
		}

		fmt.Println("Primera partición lógica creada exitosamente.") // Mensaje importante
		return nil
	}

	nuevoStartEBR, err := ultimoEBR.CalcularSiguienteStartEBR(particionExtendida.Start, particionExtendida.Tamanio)
	if err != nil {
		return fmt.Errorf("error calculando el inicio del nuevo EBR: %v", err)
	}

	disponibleSize := particionExtendida.Tamanio - (nuevoStartEBR - particionExtendida.Start)
	if disponibleSize < int32(tam) {
		return errors.New("no hay suficiente espacio en la partición extendida para una nueva partición lógica")
	}

	newEBR := Structs.EBR{}
	newEBR.ColocarEBR(fit[0], int32(tam), nuevoStartEBR, -1, name)

	err = newEBR.Codificar(file, int64(nuevoStartEBR))
	if err != nil {
		return fmt.Errorf("error al escribir el nuevo EBR en el disco: %v", err)
	}

	ultimoEBR.ColocarSiguienteEBR(nuevoStartEBR)

	err = ultimoEBR.Codificar(file, int64(ultimoEBR.Ebr_start))
	if err != nil {
		return fmt.Errorf("error al actualizar el EBR anterior: %v", err)
	}

	fmt.Println("Partición lógica creada exitosamente.")

	return nil
}
