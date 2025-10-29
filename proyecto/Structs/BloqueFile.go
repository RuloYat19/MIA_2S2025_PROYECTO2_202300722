package Structs

import (
	"Proyecto/Herramientas"
	"fmt"
	"os"
	"strings"
)

const TamanioBloque = 64

type BloqueFile struct {
	B_contenido [TamanioBloque]byte
}

func (fb *BloqueFile) ImprimirBloqueArchivo() {
	fmt.Printf("Contenido:\n")
	fmt.Print("  ", fb.ObtenerContenido(), "\n")
}

func (fb *BloqueFile) ObtenerContenido() string {
	contenido := string(fb.B_contenido[:])

	contenido = strings.TrimRight(contenido, "\x00")
	return contenido
}

func (fb *BloqueFile) Codificar(archivo *os.File, offset int64) error {
	// Utilizamos la función EscribirAlArchivo del paquete Herramientas
	err := Herramientas.EscribirAlArchivo(archivo, offset, fb.B_contenido)
	if err != nil {
		return fmt.Errorf("error writing FileBlock to file: %w", err)
	}
	return nil
}

func (fb *BloqueFile) Decodificar(archivo *os.File, offset int64) error {
	// Se utiliza la función LeerDesdeElArchivo del paquete Herramientas
	err := Herramientas.LeerDesdeElArchivo(archivo, offset, &fb.B_contenido)
	if err != nil {
		return fmt.Errorf("error reading FileBlock from file: %w", err)
	}
	return nil
}

// Se divide una cadena en bloques de tamaño BlockSize y retorna un slice de FileBlocks
func DividirContenido(contenido string) ([]*BloqueFile, error) {
	var bloques []*BloqueFile
	for len(contenido) > 0 {
		end := TamanioBloque
		if len(contenido) < TamanioBloque {
			end = len(contenido)
		}
		fb, err := NuevoBloqueFile(contenido[:end])
		if err != nil {
			return nil, err
		}
		bloques = append(bloques, fb)
		contenido = contenido[end:]
	}
	return bloques, nil
}

// Se crea un nuevo BloqueFile con contenido opcional
func NuevoBloqueFile(contenido string) (*BloqueFile, error) {
	fb := &BloqueFile{}

	err := fb.EstablecerContenido(contenido)
	if err != nil {
		return nil, err
	}

	return fb, nil
}

// Se copia una cadena en B_contenido, asegurando que no exceda el tamaño máximo
func (fb *BloqueFile) EstablecerContenido(contenido string) error {
	espacioDisponible := fb.EspacioDisponible()
	if len(contenido) > espacioDisponible {
		return fmt.Errorf("no hay suficiente espacio en el bloque para añadir '%d' bytes", len(contenido))
	}

	// Encontrar donde termina el contenido actual
	espacioUsado := fb.EspacioUsado()

	// Copiar el nuevo contenido
	copy(fb.B_contenido[espacioUsado:], contenido)

	// IMPORTANTE: Limpiar el resto del bloque con bytes nulos
	for i := espacioUsado + len(contenido); i < TamanioBloque; i++ {
		fb.B_contenido[i] = 0
	}

	return nil
}

// Limpia el contenido de B_content
func (fb *BloqueFile) LimpiarContenido() {
	for i := range fb.B_contenido {
		fb.B_contenido[i] = 0
	}
}

func (fb *BloqueFile) EspacioDisponible() int {
	return TamanioBloque - fb.EspacioUsado()
}

func (fb *BloqueFile) EspacioUsado() int {
	contenido := fb.ObtenerContenido()
	return len(contenido)
}
