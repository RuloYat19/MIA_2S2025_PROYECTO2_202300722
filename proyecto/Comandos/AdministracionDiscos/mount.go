package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

type MOUNT struct {
	Ruta   string
	Nombre string
}

func Mount(parametros []string) string {
	fmt.Println("\n======= MOUNT =======")

	var salida1 = ""

	mount := &MOUNT{}

	//Otras variables
	paramC := true
	pathInit := false
	nameInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "MOUNT Error: Valor desconocido del parámetro " + tmp[0] + "\n"
			fmt.Println("MOUNT Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			mount.Ruta = tmp[1]
		} else if strings.ToLower(tmp[0]) == "name" {
			nameInit = true
			mount.Nombre = tmp[1]
		} else {
			salida1 += "MOUNT Error: Parámetro desconocido " + tmp[0] + "\n"
			fmt.Println("MOUNT Error: Parámetro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && nameInit {
		salida2, err := montarParticionesEnElSistema(mount)
		if err != nil {
			salida1 += "MOUNT Error: Hubo problemas al montar la partición en el sistema.\n"
		}
		return salida1 + salida2
	}
	return salida1
}

func montarParticionesEnElSistema(mount *MOUNT) (string, error) {
	fmt.Printf("Montando partición en memoria con %s y %s\n", mount.Ruta, mount.Nombre)

	var salida2 = ""

	archivo, err := os.OpenFile(mount.Ruta, os.O_RDWR, 0644)

	if err != nil {
		fmt.Printf("Error abriendo el archivo del disco: %v", err)
	}
	defer archivo.Close()

	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Printf("Error deserializando el MBR: %v", err)
	}

	indiceDeParticion := globales.ObtenerIndiceParaParticion(mount.Ruta)

	particion := mbr.ObtenerParticionNombre(mount.Nombre)

	//Se valida si existe la partición en el disco
	if particion == nil {
		salida2 += "MOUNT Error: La partición " + mount.Nombre + " no existe en el disco.\n"
		err := fmt.Errorf("error: La partición '%s' no existe en el disco", mount.Nombre)
		return salida2, err
	}

	//Se valida si ya está montada la partición
	for nombre, rutaParticionMontada := range globales.NombresParticionesMontadas {
		fmt.Println(rutaParticionMontada, nombre, mount.Nombre)
		if rutaParticionMontada == mount.Ruta && strings.Contains(nombre, mount.Nombre) {
			salida2 += "MOUNT Error: La partición " + mount.Nombre + " ya está montada en el sistema.\n"
			err := fmt.Errorf("error: La partición '%s' ya está montada en el sistema", mount.Nombre)
			return salida2, err
		}
	}

	idParticion, err := GenerarIdParticion(mount.Ruta, indiceDeParticion)
	if err != nil {
		salida2 += "MOUNT Error: Hubo problemas generando el ID de la partición.\n"
		fmt.Printf("Error generando el ID de la partición: %v", err)
	}

	globales.ParticionesMontadas[idParticion] = mount.Ruta
	globales.NombresParticionesMontadas[mount.Nombre] = mount.Ruta

	particion.MontarParticion(indiceDeParticion, idParticion)

	// Se recorre las particiones del MBR para así asignarle el id correcto a la particion correcta
	for i := 0; i < len(mbr.Particiones); i++ {
		// Se convierte el nombre de la partición de [16]byte a string
		nombreParticion := strings.TrimRight(string(mbr.Particiones[i].Nombre[:]), "\x00")

		// Se compara con el nombre buscado
		if nombreParticion == mount.Nombre {
			copy(mbr.Particiones[i].Id[:], idParticion[:])
		}
	}

	err = mbr.Codificar(archivo)
	if err != nil {
		fmt.Printf("error serializando el MBR de vuelta al disco: %v", err)
	}

	salida2 += "La partición " + mount.Nombre + " fue montada correctamente en el sistema con ID " + idParticion + ".\n"
	fmt.Printf("Partición '%s' montada correctamente con ID: %s\n", mount.Nombre, idParticion)
	fmt.Println("\n=== Particiones Montadas ===")

	for id, ruta := range globales.ParticionesMontadas {
		fmt.Printf("ID: %s | Ruta: %s\n", id, ruta)
	}

	fmt.Println("\n======FIN MOUNT======")

	return salida2, nil
}

func GenerarIdParticion(path string, indiceParticion int) (string, error) {
	ultimosDosDigitos := globales.Carnet[len(globales.Carnet)-2:]
	letra, err := Herramientas.ObtenerLetra(path)
	if err != nil {
		return "", err
	}

	idParticion := fmt.Sprintf("%s%d%s", ultimosDosDigitos, indiceParticion+1, letra)
	return idParticion, nil
}
