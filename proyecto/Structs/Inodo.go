package Structs

import (
	"Proyecto/Herramientas"
	"fmt"
	"os"
	"time"
)

type Inodo struct {
	I_uid   int32
	I_gid   int32
	I_s     int32
	I_atime float32
	I_ctime float32
	I_mtime float32
	I_block [15]int32
	I_type  [1]byte //Indica si es archivo o carpeta 1=archivo, 0=carpeta
	I_perm  [3]byte
	// Total: 88 bytes
}

func (inode *Inodo) ImprimirInodo() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_s)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}

func (inode *Inodo) ActualizarAtime() {
	inode.I_atime = float32(time.Now().Unix())
}

func (inode *Inodo) ActualizarMTime() {
	inode.I_mtime = float32(time.Now().Unix())
}

func (inode *Inodo) ActualizarCTime() {
	inode.I_ctime = float32(time.Now().Unix())
}

func (inode *Inodo) Codificar(file *os.File, offset int64) error {
	// Se utiliza la función EscribirAlArchivo del paquete Herramientas
	err := Herramientas.EscribirAlArchivo(file, offset, inode)
	if err != nil {
		fmt.Println("Error al escribir el inodo en el archivo.")
		return fmt.Errorf("error writing Inode to file: %w", err)
	}
	return nil
}

func (inode *Inodo) Decodificar(file *os.File, offset int64) error {
	// Se utiliza la función LeerDesdeElArchivo del paquete Herramientas
	err := Herramientas.LeerDesdeElArchivo(file, offset, inode)
	if err != nil {
		fmt.Println("Error al leer el inodo en el archivo.")
		return fmt.Errorf("error reading Inode from file: %w", err)
	}
	return nil
}

// Se debe respetar el tamaño declarado del archivo
func (inode *Inodo) LeerDatos(archivo *os.File, superbloque *Superbloque) ([]byte, error) {
	// Se obtiene todos los bloques de datos
	indicesBloques, err := inode.ObtenerIndicesBloquesDatos(archivo, superbloque)
	if err != nil {
		return nil, err
	}

	// Se calculan cuántos bytes se deben de leer en total
	bytesALeer := int(inode.I_s)
	resultado := make([]byte, 0, bytesALeer)

	// Leer cada bloque
	for _, indiceBloque := range indicesBloques {
		if bytesALeer <= 0 {
			break // Ya leímos todo lo que necesitábamos
		}

		// Leer el bloque
		bloqueArchivo := &BloqueFile{}
		offsetBloque := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		if err := bloqueArchivo.Decodificar(archivo, offsetBloque); err != nil {
			return nil, err
		}

		// Determinar cuántos bytes leer de este bloque
		bytesDelBloque := TamanioBloque
		if bytesDelBloque > bytesALeer {
			bytesDelBloque = bytesALeer
		}

		// Añadir los bytes al resultado
		resultado = append(resultado, bloqueArchivo.B_contenido[:bytesDelBloque]...)
		bytesALeer -= bytesDelBloque
	}

	return resultado, nil
}

// Se devuelven todos los índices de bloques utilizados por este inodo
func (inode *Inodo) ObtenerTodosLosIndicesBloque(archivo *os.File, superbloque *Superbloque) ([]int32, error) {
	var indicesBloque []int32

	// Se obtienen los bloques directos
	for i := 0; i < 12; i++ {
		if inode.I_block[i] != -1 {
			indicesBloque = append(indicesBloque, inode.I_block[i])
		}
	}

	// Se procesan el bloque indirecto simple
	if inode.I_block[12] != -1 {
		// Agregar el índice del bloque de apuntadores simple
		indicesBloque = append(indicesBloque, inode.I_block[12])

		// Cargar el bloque de apuntadores
		pb := &BloqueApuntador{}
		pbOffset := int64(superbloque.S_block_start + inode.I_block[12]*superbloque.S_block_s)
		err := pb.Decodificar(archivo, pbOffset)
		if err != nil {
			return nil, fmt.Errorf("error leyendo bloque indirecto simple: %w", err)
		}

		// Obtener los bloques apuntados por este bloque
		for _, pointer := range pb.B_pointers {
			if pointer != -1 {
				indicesBloque = append(indicesBloque, int32(pointer))
			}
		}
	}

	// Se procesa el bloque indirecto doble
	if inode.I_block[13] != -1 {
		indicesBloque = append(indicesBloque, inode.I_block[13])

		// Se carga el bloque de apuntadores primario
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[13]*superbloque.S_block_s)
		err := primPB.Decodificar(archivo, primOffset)
		if err != nil {
			return nil, fmt.Errorf("error leyendo bloque indirecto doble: %w", err)
		}

		// Para cada apuntador válido en el bloque primario
		for _, primPointer := range primPB.B_pointers {
			if primPointer != -1 {
				// Se agrega el índice del bloque secundario
				indicesBloque = append(indicesBloque, int32(primPointer))

				// Se carga el bloque secundario
				secPB := &BloqueApuntador{}
				secOffset := int64(superbloque.S_block_start + int32(primPointer)*superbloque.S_block_s)
				err := secPB.Decodificar(archivo, secOffset)
				if err != nil {
					return nil, fmt.Errorf("error leyendo bloque secundario: %w", err)
				}

				// Se agregan todos los bloques apuntados por el secundario
				for _, secPointer := range secPB.B_pointers {
					if secPointer != -1 {
						indicesBloque = append(indicesBloque, int32(secPointer))
					}
				}
			}
		}
	}

	// Se procesa el bloque indirecto triple
	if inode.I_block[14] != -1 {
		indicesBloque = append(indicesBloque, inode.I_block[14])
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[14]*superbloque.S_block_s)
		err := primPB.Decodificar(archivo, primOffset)
		if err != nil {
			return nil, fmt.Errorf("error leyendo bloque indirecto triple: %w", err)
		}

		// Para cada apuntador válido en el bloque primario
		for _, primPointer := range primPB.B_pointers {
			if primPointer != -1 {
				// Agregamos el índice del bloque secundario
				indicesBloque = append(indicesBloque, int32(primPointer))

				// Cargamos el bloque secundario (nivel 2)
				secPB := &BloqueApuntador{}
				secOffset := int64(superbloque.S_block_start + int32(primPointer)*superbloque.S_block_s)
				err := secPB.Decodificar(archivo, secOffset)
				if err != nil {
					return nil, fmt.Errorf("error leyendo bloque secundario en triple: %w", err)
				}

				// Para cada apuntador válido en el bloque secundario
				for _, secPointer := range secPB.B_pointers {
					if secPointer != -1 {
						indicesBloque = append(indicesBloque, int32(secPointer))
						tercPB := &BloqueApuntador{}
						tercOffset := int64(superbloque.S_block_start + int32(secPointer)*superbloque.S_block_s)
						err := tercPB.Decodificar(archivo, tercOffset)
						if err != nil {
							return nil, fmt.Errorf("error leyendo bloque terciario: %w", err)
						}

						// Agregamos todos los bloques apuntados por el terciario
						for _, tercPointer := range tercPB.B_pointers {
							if tercPointer != -1 {
								indicesBloque = append(indicesBloque, int32(tercPointer))
							}
						}
					}
				}
			}
		}
	}

	return indicesBloque, nil
}

// Se libera un bloque en específico y se actualiza el bitmap
func (inode *Inodo) LiberarBloque(archivo *os.File, superbloque *Superbloque, indiceBloque int32) error {
	if err := superbloque.ActualizarBitmapBloque(archivo, indiceBloque, false); err != nil {
		return fmt.Errorf("error liberando bloque %d: %w", indiceBloque, err)
	}
	superbloque.ActualizarSuperbloqueDespuesAsignacionBloques()
	return nil
}

// Se liberan todos los bloques asociados a un inodo
func (inode *Inodo) LiberarTodosLosBloques(archivo *os.File, superbloque *Superbloque) error {
	// Se obtienen todos los bloques
	bloques, err := inode.ObtenerTodosLosIndicesBloque(archivo, superbloque)
	if err != nil {
		return err
	}

	// Liberar cada bloque
	for _, indiceBloque := range bloques {
		if err := inode.LiberarBloque(archivo, superbloque, indiceBloque); err != nil {
			return err
		}
	}

	// Reiniciar todos los apuntadores del inodo
	for i := range inode.I_block {
		inode.I_block[i] = -1
	}

	inode.I_s = 0
	inode.ActualizarMTime()
	return nil
}

// Se verifica y libera bloques de apuntadores que quedaron vacíos
func (inode *Inodo) ChequearYLiberarBloquesIndirectosVacios(archivo *os.File, superbloque *Superbloque) error {
	// Se verifica e bloque indirecto simple
	if inode.I_block[12] != -1 {
		pb := &BloqueApuntador{}
		pbOffset := int64(superbloque.S_block_start + inode.I_block[12]*superbloque.S_block_s)
		if err := pb.Decodificar(archivo, pbOffset); err != nil {
			return fmt.Errorf("error leyendo bloque indirecto simple: %w", err)
		}

		// Se verificar si está vacío
		estaVacio := true
		for _, pointer := range pb.B_pointers {
			if pointer != -1 {
				estaVacio = false
				break
			}
		}

		if estaVacio {
			// Se liberar el bloque
			fmt.Printf("Liberando bloque de apuntadores simple %d (vacío)\n", inode.I_block[12])
			if err := inode.LiberarBloque(archivo, superbloque, inode.I_block[12]); err != nil {
				return err
			}
			// Se actualiza la referencia en el inodo
			inode.I_block[12] = -1
		}
	}

	// Se verificar el bloque indirecto doble
	if inode.I_block[13] != -1 {
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[13]*superbloque.S_block_s)
		if err := primPB.Decodificar(archivo, primOffset); err != nil {
			return fmt.Errorf("error leyendo bloque indirecto doble: %w", err)
		}

		// Se verifican los bloques secundarios y marcar los vacíos
		emptySecondaryBlocks := make([]int, 0)
		allEmpty := true

		for i, primPointer := range primPB.B_pointers {
			if primPointer != -1 {
				// Se carga el bloque secundario
				secPB := &BloqueApuntador{}
				secOffset := int64(superbloque.S_block_start + int32(primPointer)*superbloque.S_block_s)
				if err := secPB.Decodificar(archivo, secOffset); err != nil {
					return fmt.Errorf("error leyendo bloque secundario: %w", err)
				}

				// Se verifica si el bloque secundario está vacío
				estaVacio := true
				for _, secPointer := range secPB.B_pointers {
					if secPointer != -1 {
						estaVacio = false
						break
					}
				}

				if estaVacio {
					emptySecondaryBlocks = append(emptySecondaryBlocks, i)
					fmt.Printf("Marcando bloque secundario %d para liberación (vacío)\n", primPointer)
				} else {
					allEmpty = false
				}
			}
		}

		// Se liberan los bloques secundarios vacíos y actualizar referencias
		for _, idx := range emptySecondaryBlocks {
			secBlockIndex := primPB.B_pointers[idx]
			// Se libera el bloque
			if err := inode.LiberarBloque(archivo, superbloque, secBlockIndex); err != nil {
				return err
			}
			// Se actualiza la referencia en el bloque primario
			primPB.B_pointers[idx] = -1
		}

		// Si hay bloques secundarios liberados, guardar el bloque primario actualizado
		if len(emptySecondaryBlocks) > 0 {
			if err := primPB.Codificar(archivo, primOffset); err != nil {
				return fmt.Errorf("error actualizando bloque primario: %w", err)
			}
		}

		// Si todos los bloques secundarios están vacíos/liberados, liberar el bloque primario
		if allEmpty {
			fmt.Printf("Liberando bloque de apuntadores doble %d (vacío)\n", inode.I_block[13])
			if err := inode.LiberarBloque(archivo, superbloque, inode.I_block[13]); err != nil {
				return err
			}
			inode.I_block[13] = -1
		}
	}

	// Se verifica el bloque indirecto triple
	if inode.I_block[14] != -1 {
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[14]*superbloque.S_block_s)
		if err := primPB.Decodificar(archivo, primOffset); err != nil {
			return fmt.Errorf("error leyendo bloque indirecto triple: %w", err)
		}

		primEmptyCount := 0
		allPrimEmpty := true

		// Se procesa cada bloque secundario
		for primIdx, primPointer := range primPB.B_pointers {
			if primPointer == -1 {
				primEmptyCount++
				continue
			}

			secPB := &BloqueApuntador{}
			secOffset := int64(superbloque.S_block_start + int32(primPointer)*superbloque.S_block_s)
			if err := secPB.Decodificar(archivo, secOffset); err != nil {
				return fmt.Errorf("error leyendo bloque secundario en triple: %w", err)
			}

			secEmptyCount := 0
			allSecEmpty := true

			// Se procesa cada bloque terciario
			for secIdx, secPointer := range secPB.B_pointers {
				if secPointer == -1 {
					secEmptyCount++
					continue
				}

				tercPB := &BloqueApuntador{}
				tercOffset := int64(superbloque.S_block_start + int32(secPointer)*superbloque.S_block_s)
				if err := tercPB.Decodificar(archivo, tercOffset); err != nil {
					return fmt.Errorf("error leyendo bloque terciario: %w", err)
				}

				// Se verifica si el bloque terciario está vacío
				isEmpty := true
				for _, tercPointer := range tercPB.B_pointers {
					if tercPointer != -1 {
						isEmpty = false
						break
					}
				}

				if isEmpty {
					// Se libera el bloque terciario
					fmt.Printf("Liberando bloque terciario %d (vacío)\n", secPointer)
					if err := inode.LiberarBloque(archivo, superbloque, secPointer); err != nil {
						return err
					}
					secPB.B_pointers[secIdx] = -1
				} else {
					allSecEmpty = false
				}
			}

			// Si el bloque secundario quedó vacío, liberarlo
			if allSecEmpty {
				fmt.Printf("Liberando bloque secundario %d en triple (vacío)\n", primPointer)
				if err := inode.LiberarBloque(archivo, superbloque, primPointer); err != nil {
					return err
				}
				primPB.B_pointers[primIdx] = -1
			} else {
				// Si hubo cambios en el secundario, guardar
				if secEmptyCount > 0 && secEmptyCount < len(secPB.B_pointers) {
					if err := secPB.Codificar(archivo, secOffset); err != nil {
						return fmt.Errorf("error actualizando bloque secundario: %w", err)
					}
				}
				allPrimEmpty = false
			}
		}

		// Si hubo cambios en el primario pero no está completamente vacío, guardarlo
		if !allPrimEmpty && primEmptyCount > 0 {
			if err := primPB.Codificar(archivo, primOffset); err != nil {
				return fmt.Errorf("error actualizando bloque primario triple: %w", err)
			}
		}

		// Si el bloque primario quedó vacío, liberarlo
		if allPrimEmpty {
			fmt.Printf("Liberando bloque de apuntadores triple %d (vacío)\n", inode.I_block[14])
			if err := inode.LiberarBloque(archivo, superbloque, inode.I_block[14]); err != nil {
				return err
			}
			inode.I_block[14] = -1
		}
	}
	return nil
}

// Se crear y se serializa un inodo donde se actualiza el bitmap de inodos
func (inode *Inodo) CrearInodo(archivo *os.File, superbloque *Superbloque, tipoInodo byte, tamanio int32, bloques [15]int32, permisos [3]byte) error {
	indiceInodo, err := superbloque.AsignarNuevoInodo(archivo)
	if err != nil {
		fmt.Println("Error asignando el nuevo inodo.")
		return fmt.Errorf("error asignando nuevo inodo: %w", err)
	}

	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_s = tamanio
	inode.I_atime = float32(time.Now().Unix())
	inode.I_ctime = float32(time.Now().Unix())
	inode.I_mtime = float32(time.Now().Unix())
	inode.I_block = bloques
	inode.I_type = [1]byte{tipoInodo}
	inode.I_perm = permisos

	err = superbloque.ActualizarBitmapInodo(archivo, indiceInodo, true)
	if err != nil {
		fmt.Println("Error actualizando el bitmap de inodos.")
		return fmt.Errorf("error actualizando el bitmap de inodos: %w", err)
	}

	offsetInodo := int64(superbloque.S_inode_start + (indiceInodo * superbloque.S_inode_s))
	err = inode.Codificar(archivo, offsetInodo)
	if err != nil {
		fmt.Printf("Error serializando el inodo en la ubicación '%d'.", offsetInodo)
		return fmt.Errorf("error serializando el inodo en la ubicación %d: %w", offsetInodo, err)
	}

	return nil
}

// Se devuelven solo los índices de bloques que contienen datos (no apuntadores)
func (inode *Inodo) ObtenerIndicesBloquesDatos(archivo *os.File, superbloque *Superbloque) ([]int32, error) {
	bloquesDatos := []int32{}

	// Los bloques directos sone del 0-11 ya que siempre son bloques de datos
	for i := 0; i < 12; i++ {
		if inode.I_block[i] != -1 {
			bloquesDatos = append(bloquesDatos, inode.I_block[i])
		}
	}

	// Los bloques indirectos simple son solo los apuntados, no el bloque 12
	if inode.I_block[12] != -1 {
		pb := &BloqueApuntador{}
		pbOffset := int64(superbloque.S_block_start + inode.I_block[12]*superbloque.S_block_s)
		if err := pb.Decodificar(archivo, pbOffset); err != nil {
			fmt.Println("Error al leer el bloque indirecto simple.")
			return nil, fmt.Errorf("error leyendo bloque indirecto simple: %w", err)
		}

		for _, apuntador := range pb.B_pointers {
			if apuntador != -1 {
				bloquesDatos = append(bloquesDatos, apuntador)
			}
		}
	}

	// Los bloques en indirectos doble (solo los bloques finales, no los apuntadores)
	if inode.I_block[13] != -1 {
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[13]*superbloque.S_block_s)
		if err := primPB.Decodificar(archivo, primOffset); err != nil {
			fmt.Println("Error al leer el bloque indirecto doble.")
			return nil, fmt.Errorf("error leyendo bloque indirecto doble: %w", err)
		}

		for _, primPointer := range primPB.B_pointers {
			if primPointer != -1 {
				secPB := &BloqueApuntador{}
				secOffset := int64(superbloque.S_block_start + primPointer*superbloque.S_block_s)
				if err := secPB.Decodificar(archivo, secOffset); err != nil {
					fmt.Println("Error al leer el bloque indirecto doble secundario.")
					return nil, fmt.Errorf("error leyendo bloque secundario: %w", err)
				}

				for _, secPointer := range secPB.B_pointers {
					if secPointer != -1 {
						bloquesDatos = append(bloquesDatos, secPointer)
					}
				}
			}
		}
	}

	// Los bloques en indirectos triple (solo los bloques finales, no los apuntadores)
	if inode.I_block[14] != -1 {
		primPB := &BloqueApuntador{}
		primOffset := int64(superbloque.S_block_start + inode.I_block[14]*superbloque.S_block_s)
		if err := primPB.Decodificar(archivo, primOffset); err != nil {
			return nil, fmt.Errorf("error leyendo bloque indirecto triple: %w", err)
		}

		for _, primPointer := range primPB.B_pointers {
			if primPointer != -1 {
				secPB := &BloqueApuntador{}
				secOffset := int64(superbloque.S_block_start + primPointer*superbloque.S_block_s)
				if err := secPB.Decodificar(archivo, secOffset); err != nil {
					return nil, fmt.Errorf("error leyendo bloque secundario en indirección triple: %w", err)
				}

				for _, secPointer := range secPB.B_pointers {
					if secPointer != -1 {
						tercPB := &BloqueApuntador{}
						tercOffset := int64(superbloque.S_block_start + secPointer*superbloque.S_block_s)
						if err := tercPB.Decodificar(archivo, tercOffset); err != nil {
							return nil, fmt.Errorf("error leyendo bloque terciario: %w", err)
						}

						for _, tercPointer := range tercPB.B_pointers {
							if tercPointer != -1 {
								bloquesDatos = append(bloquesDatos, tercPointer)
							}
						}
					}
				}
			}
		}
	}

	return bloquesDatos, nil
}
