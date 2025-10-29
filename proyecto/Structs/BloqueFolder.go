package Structs

import (
	"Proyecto/Herramientas"
	"fmt"
	"os"
	"strings"
)

type BloqueFolder struct {
	B_contenido [4]ContenidoFolder
}

type ContenidoFolder struct {
	B_nombre [12]byte
	B_inodo  int32
}

func (fb *BloqueFolder) ImprimirBloqueFolder() {
	for i, content := range fb.B_contenido {
		name := string(content.B_nombre[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)
	}
}

// Se deserializa la estructura BloqueFolder desde un archivo binario en la posición especificada
func (fb *BloqueFolder) Decodificar(archivo *os.File, offset int64) error {
	// Se utiliza la función LeerDesdeElArchivo del paquete Herramientas
	err := Herramientas.LeerDesdeElArchivo(archivo, offset, fb)

	if err != nil {
		return fmt.Errorf("error leyendo el BloqueFolder del archivo: %w", err)
	}

	return nil
}

// Se serializa la estructura BloqueFolder en un archivo binario en la posición especificada
func (fb *BloqueFolder) Codificar(file *os.File, offset int64) error {
	// Utilizamos la función EscribirAlArchivo del paquete Herramientas
	err := Herramientas.EscribirAlArchivo(file, offset, fb.B_contenido)
	if err != nil {
		return fmt.Errorf("error escribiendo el BloqueFolder al archivo: %w", err)
	}
	return nil
}

// RenameInFolderBlock busca un archivo o carpeta dentro del FolderBlock
func (fb *BloqueFolder) RenombrarEnBloqueFolder(nombreAntiguo string, nuevoNombre string) error {
	for i := 2; i < len(fb.B_contenido); i++ {
		contenido := &fb.B_contenido[i]
		nombreActual := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

		// Si encontramos el nombre que queremos cambiar
		if strings.EqualFold(nombreActual, nombreAntiguo) && contenido.B_inodo != -1 {
			if len(nuevoNombre) > 12 {
				fmt.Printf("El nuevo nombre '%s' es demasiado largo, el número máximo de caracteres es de 12.\n", nuevoNombre)
				return fmt.Errorf("el nuevo nombre '%s' es demasiado largo, máximo 12 caracteres", nuevoNombre)
			}
			copy(contenido.B_nombre[:], nuevoNombre)
			for j := len(nuevoNombre); j < 12; j++ {
				contenido.B_nombre[j] = 0
			}

			fmt.Printf("Nombre cambiado de '%s' a '%s' en la posición %d\n", nombreAntiguo, nuevoNombre, i+1)
			return nil
		}
	}
	fmt.Printf("El nombre '%s' no fue encontrado en los inodos 3 o 4.\n", nombreAntiguo)
	return fmt.Errorf("el nombre '%s' no fue encontrado en los inodos 3 o 4", nombreAntiguo)
}
