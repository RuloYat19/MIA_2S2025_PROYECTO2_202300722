package Structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

type BloqueApuntador struct {
	B_pointers [16]int32
}

// Se leen los bloques indirectos simples
func (pb *BloqueApuntador) LeerBloqueSimpleIndirecto(archivo *os.File, sb *Superbloque) ([]int32, error) {
	var blocks []int32
	for _, pointer := range pb.B_pointers {
		if pointer != -1 {
			blocks = append(blocks, int32(pointer))
		}
	}
	return blocks, nil
}

// Se leen los bloques a través de indirección doble
func (pb *BloqueApuntador) LeerBloquesDoblesIndirectos(archivo *os.File, sb *Superbloque) ([]int32, error) {
	var blocks []int32
	for _, pointer := range pb.B_pointers {
		if pointer != -1 {
			// Leer el bloque de apuntadores secundario
			secondaryPB := &BloqueApuntador{}
			err := secondaryPB.Decodificar(archivo, int64(sb.S_block_start+int32(pointer)*sb.S_block_s))
			if err != nil {
				return nil, err
			}
			// Obtener los bloques de este nivel
			secondaryBlocks, err := secondaryPB.LeerBloqueSimpleIndirecto(archivo, sb)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, secondaryBlocks...)
		}
	}
	return blocks, nil
}

// Se busca el primer apuntador libre en un bloque de apuntadores y devuelve su índice
func (pb *BloqueApuntador) EncontrarApuntadorLibre() (int, error) {
	for i, pointer := range pb.B_pointers {
		if pointer == -1 {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no hay apuntadores libres en el bloque de apuntadores")
}

// Se verifica si todos los apuntadores están libres
func (pb *BloqueApuntador) IsEmpty() bool {
	return pb.ContarApuntadoresLibres() == len(pb.B_pointers)
}

// Se libera el bloque si está vacío y actualiza referencias
func (pb *BloqueApuntador) LibreSiEstaVacio(archivo *os.File, superbloque *Superbloque, indiceBloque int32, inodoPadre *Inodo, indiceApuntador int) error {
	if pb.IsEmpty() {
		// Se marca el bloque como libre en bitmap
		if err := superbloque.ActualizarBitmapBloque(archivo, indiceBloque, false); err != nil {
			return err
		}

		// Se actualiza la referencia en inodo padre
		if inodoPadre != nil && indiceApuntador >= 0 {
			inodoPadre.I_block[indiceApuntador] = -1
			return inodoPadre.Codificar(archivo, superbloque.CalcularOffsetInodo(inodoPadre.I_uid))
		}

		// Se actualiza el contador en el superbloque
		superbloque.ActualizarSuperbloqueDespuesAsignacionBloques()
	}
	return nil
}

// Se coloca un valor en específico en el índice dado
func (pb *BloqueApuntador) ColocarApuntador(indice int, valor int64) error {
	if indice < 0 || indice >= len(pb.B_pointers) {
		fmt.Println("Índice fuera de rango.")
		return fmt.Errorf("índice fuera de rango")
	}
	pb.B_pointers[indice] = int32(valor)
	return nil
}

// Se obtiene el valor de un apuntador en el índice dado
func (pb *BloqueApuntador) ObtenerApuntador(indice int) (int64, error) {
	if indice < 0 || indice >= len(pb.B_pointers) {
		return -1, fmt.Errorf("índice fuera de rango")
	}
	return int64(pb.B_pointers[indice]), nil
}

// Se verifica si todos los apuntadores están ocupados
func (pb *BloqueApuntador) EstaLleno() bool {
	for _, apuntador := range pb.B_pointers {
		if apuntador == -1 {
			return false
		}
	}
	return true
}

// Se cuenta cuántos apuntadores libres hay en el bloque
func (pb *BloqueApuntador) ContarApuntadoresLibres() int {
	contador := 0
	for _, apuntador := range pb.B_pointers {
		if apuntador == -1 {
			contador++
		}
	}
	return contador
}

// Se serializa el PointerBlock en el archivo en la posición dada
func (pb *BloqueApuntador) Codificar(archivo *os.File, offset int64) error {
	_, err := archivo.Seek(offset, 0)
	if err != nil {
		fmt.Println("Error al buscar la posición en el archivo.")
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}

	err = binary.Write(archivo, binary.BigEndian, *pb)

	if err != nil {
		fmt.Println("Error al escribir el BloqueApuntador")
		return fmt.Errorf("error escribiendo el BloqueApuntador: %w", err)
	}

	return nil
}

// Se deserializa el PointerBlock desde el archivo en la posición dada
func (pb *BloqueApuntador) Decodificar(archivo *os.File, offset int64) error {
	// Mover el cursor del archivo a la posición deseada
	_, err := archivo.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}
	err = binary.Read(archivo, binary.BigEndian, pb)
	if err != nil {
		return fmt.Errorf("error leyendo el PointerBlock: %w", err)
	}
	return nil
}

// Se leen los bloques a través de indirección triple
func (pb *BloqueApuntador) LeerBloquesTriplesIndirectos(archivo *os.File, superbloque *Superbloque) ([]int32, error) {
	var bloques []int32
	for _, apuntadorPrim := range pb.B_pointers {
		if apuntadorPrim != -1 {
			// Se lee el bloque de apuntadores secundario
			secPB := &BloqueApuntador{}
			secOffset := int64(superbloque.S_block_start + int32(apuntadorPrim)*superbloque.S_block_s)
			if err := secPB.Decodificar(archivo, secOffset); err != nil {
				fmt.Println("Error al leer el bloque secundario")
				return nil, fmt.Errorf("error leyendo bloque secundario: %w", err)
			}

			for _, apuintadorSec := range secPB.B_pointers {
				if apuintadorSec != -1 {
					// Se lee el bloque de apuntadores terciario
					tercPB := &BloqueApuntador{}
					tercOffset := int64(superbloque.S_block_start + int32(apuintadorSec)*superbloque.S_block_s)
					if err := tercPB.Decodificar(archivo, tercOffset); err != nil {
						fmt.Println("Error al leer el bloque terciario")
						return nil, fmt.Errorf("error leyendo bloque terciario: %w", err)
					}

					// Se obtienen los bloques de este nivel
					terbloques, err := tercPB.LeerBloqueSimpleIndirecto(archivo, superbloque)
					if err != nil {
						return nil, err
					}
					bloques = append(bloques, terbloques...)
				}
			}
		}
	}
	return bloques, nil
}
