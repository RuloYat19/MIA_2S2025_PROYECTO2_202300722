package globales

import (
	"Proyecto/Structs"
	"errors"
	"fmt"
	"os"
)

const Carnet string = "22" // 202300722

var (
	UsuarioActual              *Structs.Usuario = nil
	ParticionInicioSesion      string
	ParticionesMontadas        map[string]string = make(map[string]string)
	NombresParticionesMontadas map[string]string = make(map[string]string)
)

func ObtenerIndiceParaParticion(ruta string) int {
	var contador = 0
	for i, rutaParticion := range ParticionesMontadas {
		if rutaParticion == ruta {
			contador++
		}

		if i == "" {

		}
	}
	return contador
}

func ObtenerParticionMontadas(id string) (*Structs.Particion, string, error) {
	ruta := ParticionesMontadas[id]
	if ruta == "" {
		return nil, "", errors.New("la partición no está montada")
	}

	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, "", err
	}
	defer archivo.Close()

	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)
	if err != nil {
		return nil, "", err
	}

	particion, err := mbr.ObtenerParticionPorID(id)
	if particion == nil {
		return nil, "", err
	}

	return particion, ruta, nil
}

// Se obtiene el MBR y el SuperBlock de la partición montada con el id especificado
func ObtenerParticionMontadaRep(id string) (*Structs.MBR, *Structs.Superbloque, string, error) {
	ruta := ParticionesMontadas[id]
	//fmt.Println("La ruta de la partición es: " + ruta)

	if ruta == "" {
		fmt.Println("La partición no está montada en el sistema.")
		return nil, nil, "", errors.New("la partición no está montada")
	}

	archivo, err := os.Open(ruta)

	if err != nil {
		fmt.Println("Hubo problemas al abrir el archivo.")
		return nil, nil, "", err
	}

	defer archivo.Close()

	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Println("Hubo problemas al decodificar el MBR del disco.")
		return nil, nil, "", err
	}

	particion, err := mbr.ObtenerParticionPorID(id)
	if err != nil {
		fmt.Println("Hubo problemas al obtener la partición por el ID.")
		return nil, nil, "", err
	}

	var superbloque Structs.Superbloque

	err = superbloque.Decodificar(archivo, int64(particion.Start))
	if err != nil {
		fmt.Println("Hubo problemas el decodificar el superbloque de la partición")
		return nil, nil, "", err
	}

	return &mbr, &superbloque, ruta, nil
}

func HaIniciadoSesion() bool {
	return UsuarioActual != nil && UsuarioActual.Estado
}

// Se obtiene el superbloque de la particion definida con el id determinado
func ObtenerParticionMontadaSuperbloque(id string) (*Structs.Superbloque, *Structs.Particion, string, error) {
	// Se obtiene la ruta de la partición
	ruta := ParticionesMontadas[id]

	if ruta == "" {
		fmt.Println("La partición no está montada")
		return nil, nil, "", errors.New("la partición no está montada")
	}

	// Se abre el archivo para poder leer el MBR
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, nil, "", err
	}

	var mbr Structs.MBR

	// Se deserializa el MBR desde un archivo binario
	err = mbr.Decodificar(archivo)
	if err != nil {
		return nil, nil, "", err
	}

	// Se busca la partición con el id determinado
	particion, err := mbr.ObtenerParticionPorID(id)
	if particion == nil {
		return nil, nil, "", err
	}

	var superbloque Structs.Superbloque

	// Se deserializa el SuperBloque desde un archivo binario
	err = superbloque.Decodificar(archivo, int64(particion.Start))
	if err != nil {
		return nil, nil, "", err
	}

	return &superbloque, particion, ruta, nil
}

func ValidarAcceso(partitionId string) error {
	if !HaIniciadoSesion() {
		return errors.New("no hay un usuario logueado")
	}
	_, _, err := ObtenerParticionMontadas(partitionId)
	if err != nil {
		return errors.New("la partición no está montada")
	}
	return nil
}
