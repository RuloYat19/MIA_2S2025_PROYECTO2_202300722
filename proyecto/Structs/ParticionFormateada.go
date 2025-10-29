package Structs

import "fmt"

type ParticionFormateada struct {
	ruta            string
	nombreParticion string
	idParticion     string
}

var (
	ParticionesFormateadas []*ParticionFormateada = make([]*ParticionFormateada, 0)
)

func AgregarParticionFormateada(particion *ParticionFormateada) {
	ParticionesFormateadas = append(ParticionesFormateadas, particion)
}

func NuevaParticionFormateada(ruta string, nombreParticion string, idParticion string) *ParticionFormateada {
	return &ParticionFormateada{ruta, nombreParticion, idParticion}
}

func ImprimirParticionesFormateadas() {
	fmt.Println("==== PARTICIONES FORMATEADAS ====")
	for _, part := range ParticionesFormateadas {
		fmt.Printf("Ruta: '%s'\n", part.ruta)
		fmt.Printf("Nombre: '%s'\n", part.nombreParticion)
		fmt.Printf("ID: '%s'\n", part.idParticion)
	}
}

func ObtenerParticionesFormateadas(nombre string) bool {
	for _, part := range ParticionesFormateadas {
		if part.nombreParticion == nombre {
			return true
		}
	}

	return false
}
