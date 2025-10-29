package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"bytes"
	"fmt"
	"os"
	"strings"
)

type UNMOUNT struct {
	id string
}

func Unmount(parametros []string) string {
	fmt.Println("\n======= UNMOUNT =======")
	salida1 := ""

	unmount := &UNMOUNT{}

	//Otras variables
	paramC := true
	idInit := false

	//Se recorren los parametros para asignarlos a las variables
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "UNMOUNT Error: Valor desconocido del parámetro " + tmp[0] + "\n"
			fmt.Println("UNMOUNT Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "id" {
			idInit = true
			unmount.id = tmp[1]
		} else {
			salida1 += "UNMOUNT Error: Parámetro desconocido: " + tmp[0] + "\n"
			fmt.Println("UNMOUNT Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && idInit {
		err := desmontarParticion(unmount)

		if err != nil {
			salida1 = "UNMOUNT Error: Hubo problemas al desmontar la partición.\n"
		} else {
			salida1 = "La partición se desmontó con éxito.\n"
		}
	}

	fmt.Println("\n======FIN UNMOUNT======")
	return salida1
}

func desmontarParticion(unmount *UNMOUNT) error {
	// Se verifica si el ID de la partición existe en las particiones montadas globales
	rutaParticionMontada, exists := globales.ParticionesMontadas[unmount.id]
	if !exists {
		fmt.Printf("La partición con el ID '%s' no está montada en el sistema.\n", unmount.id)
		return fmt.Errorf("error: la partición con ID '%s' no está montada", unmount.id)
	}

	// Abrir el archivo del disco
	archivo, err := os.OpenFile(rutaParticionMontada, os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error al abrir el archivo del disco.")
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}

	defer archivo.Close()

	// Se leee el MBR del disco
	var mbr Structs.MBR
	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Println("Error al deserializar el MBR.")
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Se busca la partición en el MBR que tiene el ID especificado
	encontrado := false
	for i := range mbr.Particiones {
		particion := &mbr.Particiones[i]
		particionID := strings.TrimSpace(string(particion.Id[:]))
		if particionID == unmount.id {
			// Se desmonta la partición donde se cambia el valor del correlativo a 0
			err = particion.MontarParticion(0, "")
			if err != nil {
				fmt.Println("Error al desmontar la partición.")
				return fmt.Errorf("error desmontando la partición: %v", err)
			}

			// Se actualiza el MBR en el archivo después del desmontaje
			err = mbr.Codificar(archivo)
			if err != nil {
				fmt.Println("Error al actualizar el MBR en el disco.")
				return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
			}

			encontrado = true

			// Se remueve el nombre de la partición del map NombresParticionesMontadas
			nombreParticion := mbr.Particiones[i].Nombre[:]
			nombreLimpio := string(bytes.TrimRight(nombreParticion, "\x00"))
			delete(globales.NombresParticionesMontadas, string(nombreLimpio))

			// Se remueve el ID de la partición del map ParticionesMontadas
			delete(globales.ParticionesMontadas, unmount.id)
			break
		}
	}

	// Si no se encontró la partición con el ID se devuelve el error
	if !encontrado {
		fmt.Printf("No se encontró la partición con ID '%s' en el disco.", unmount.id)
		return fmt.Errorf("error: no se encontró la partición con ID '%s' en el disco", unmount.id)
	}

	//Mensajes
	fmt.Printf("La partición con ID '%s' fue desmontada exitosamente.\n", unmount.id)
	fmt.Println("=== Particiones Montadas ===")
	for id, ruta := range globales.ParticionesMontadas {
		fmt.Printf("ID: %s | Path: %s\n", id, ruta)
	}

	fmt.Println("=== Nombr. de Part. Mont. ===")
	for nombre, ruta := range globales.NombresParticionesMontadas {
		fmt.Printf("Nombre: %s | Path: %s\n", nombre, ruta)
	}

	return nil
}
