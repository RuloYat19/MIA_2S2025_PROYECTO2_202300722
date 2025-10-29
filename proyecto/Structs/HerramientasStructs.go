package Structs

import (
	"Proyecto/Herramientas"
	"fmt"
	"os"
	"strings"
	"time"
)

// Se crea una carpeta en el sistema de archivos
func (sb *Superbloque) CrearCarpeta(archivo *os.File, directoriosPadre []string, nombreDirectorio string, p bool) error {
	// Se valida si la carpeta padre existe entre los bloques
	var padreEncontrado bool
	var indiceInodoEncontrado int32
	for i := int32(0); i < sb.S_inodes_count; i++ {
		indiceInodo, encontrado, err := sb.ValidarPadreDirectorio(archivo, i, directoriosPadre, nombreDirectorio, p)
		if encontrado {
			padreEncontrado = true
			indiceInodoEncontrado = int32(indiceInodo)
			break
		}

		if err != nil {
			return err
		}
	}

	// Se valida si la carpeta que se quiere crear existe en los bloques
	var siExisteCarpeta bool
	for i := int32(0); i < sb.S_inodes_count; i++ {
		bandera, err := sb.ValidarExistenciaCarpeta(archivo, i, directoriosPadre, nombreDirectorio, p)

		if bandera {
			siExisteCarpeta = true
			break
		}

		if err != nil {
			return err
		}
	}

	// Se valida si el árbol del directorio coincide con el ya creado
	tamanioDirectoriosPadre := len(directoriosPadre)

	if len(directoriosPadre) == 1 {
		padreEncontrado = true
	}

	// Se hacen las validaciones para crear la carpeta
	if siExisteCarpeta {
		if directoriosPadre[tamanioDirectoriosPadre-1] == nombreDirectorio {
			// Se buscan los índices de los inodos donde están los BloqueCarpetas del nombreDirectorio repetido

			_, indicesInodo, err := sb.BuscarIndiceOIndicesInodoDeDirectorio(archivo, nombreDirectorio, true)
			if err != nil {
				return err
			}

			// Si existe la carpeta pero es el nombre del último directorio se validan los árboles de dichos directorios
			for _, indiceInodo := range indicesInodo {
				otrosPadreRuta, err := sb.ObtenerPadresDirectorioOArchivo(archivo, indiceInodo, nombreDirectorio)
				if err != nil {
					return err
				}

				var rutaOrdenada []string
				for i := len(otrosPadreRuta) - 1; i >= 0; i-- {
					rutaOrdenada = append(rutaOrdenada, otrosPadreRuta[i])
				}

				inodoVerdadero, err := sb.EncontrarInodoVerdadero(archivo, directoriosPadre)

				if err != nil {
					return err
				}

				fmt.Println(inodoVerdadero)

				booleano := false
				if len(directoriosPadre) == len(rutaOrdenada) {
					for i := 0; i < len(directoriosPadre); i++ {
						if directoriosPadre[i] == rutaOrdenada[i] {

						} else {
							booleano = true
						}
					}

					if booleano {
						err := sb.CrearCarpetaEnInodo(archivo, int32(indiceInodoEncontrado), directoriosPadre, nombreDirectorio, p)
						if err != nil {
							return err
						}
					} else {
						fmt.Println("Se está intendo crear una ruta donde los directorios ya existen.")
						return nil
					}
				} else {
					err := sb.CrearCarpetaEnInodo(archivo, int32(indiceInodoEncontrado), directoriosPadre, nombreDirectorio, p)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	} else if padreEncontrado {
		err := sb.CrearCarpetaEnInodo(archivo, int32(indiceInodoEncontrado), directoriosPadre, nombreDirectorio, p)
		if err != nil {
			return err
		}
	} else if len(directoriosPadre) == 0 {
		err := sb.CrearCarpetaEnInodo(archivo, 0, directoriosPadre, nombreDirectorio, p)
		if err != nil {
			return err
		}
	} else if p {
		err := sb.CrearCarpetaEnInodo(archivo, 0, directoriosPadre, nombreDirectorio, p)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("no se encontró la carpeta padre de: %s\n", nombreDirectorio)
		return fmt.Errorf("no se encontró la carpeta padre de: %s", nombreDirectorio)
	}

	return nil
}

// Se encuentra el verdadero inodo para crear el directorio
func (sb *Superbloque) EncontrarInodoVerdadero(archivo *os.File, directoriosPadre []string) (int32, error) {
	// Se valida si la carpeta padre existe entre los bloques
	var padreEncontrado bool
	var indiceInodoEncontrado int32
	tamanioDirectoriosPadre := len(directoriosPadre)
	for i := int32(0); i < sb.S_inodes_count; i++ {
		indiceInodo, encontrado, err := sb.ValidarPadreDirectorio(archivo, i, directoriosPadre, directoriosPadre[tamanioDirectoriosPadre-1], true)
		if encontrado {
			padreEncontrado = true
			indiceInodoEncontrado = int32(indiceInodo)
			break
		}

		if err != nil {
			return 0, err
		}
	}
	fmt.Println(padreEncontrado)
	fmt.Println(indiceInodoEncontrado)

	/*indiceInodoAbuelo, _, err := sb.BuscarPadreEnInodo(archivo, indiceInodoEncontrado, )
	if err != nil {
		return 0, err
	}
	fmt.Println(indiceInodoAbuelo)*/

	return 0, nil
}

// Se valida si la carpeta existe entre los bloques
func (sb *Superbloque) ValidarExistenciaDeDirectorio(archivo *os.File, indiceInodo int32, destDir string) (bool, error) {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		return false, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return false, nil
	}

	// Se itera sobre cada bloque del inodo (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", indiceInodo)
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Crear un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Deserializar el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return false, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indiceContenido, indiceBloque)

			// Si el contenido está vacío, salir
			if contenido.B_inodo == -1 { //-------------------------
				fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", indiceInodo, indiceContenido)
				break
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			nombreDirectorio := strings.Trim(destDir, "\x00 ")
			fmt.Printf("Comparando '%s' con el nombre de la carpeta padre '%s'\n", nombreContenido, nombreDirectorio)

			// Si el nombre del contenido coincide con el nombre de la carpeta padre
			if strings.EqualFold(nombreContenido, nombreDirectorio) {
				return true, nil
			}
		}
	}

	return false, nil
}

// Se valida la existencia de la carpeta dentro de los bloques
func (sb *Superbloque) ValidarExistenciaCarpeta(archivo *os.File, indiceInodo int32, padresDir []string, destDir string, p bool) (bool, error) {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		return false, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return false, nil
	}

	// Se itera sobre cada bloque del inodo (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", indiceInodo)
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Crear un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Deserializar el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return false, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indiceContenido, indiceBloque)

			// Si el contenido está vacío, salir
			if contenido.B_inodo == -1 { //-------------------------
				fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", indiceInodo, indiceContenido)
				break
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			fmt.Printf("Comparando '%s' con el nombre de la carpeta existente '%s'\n", nombreContenido, destDir)

			// Si el nombre del contenido coincide con el nombre de la carpeta padre
			if strings.EqualFold(nombreContenido, destDir) {
				return true, nil
			}
		}
	}

	return false, nil
}

// Se valida si la carpeta padre existe entre los bloques
func (sb *Superbloque) ValidarPadreDirectorio(archivo *os.File, indiceInodo int32, padresDir []string, destDir string, p bool) (int, bool, error) {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		return 0, false, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return 0, false, nil
	}

	// Se itera sobre cada bloque del inodo (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", indiceInodo)
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Crear un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Deserializar el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return 0, false, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indiceContenido, indiceBloque)

			// Si el contenido está vacío, salir
			if contenido.B_inodo == -1 { //-------------------------
				//fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", indiceInodo, indiceContenido)
				break
			}

			padreDir, _ := Herramientas.PadreCarpeta(padresDir, destDir)

			if padreDir == "" {
				padreDir = padresDir[0]
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			nombreDirPadre := strings.Trim(padreDir, "\x00 ")
			fmt.Printf("Comparando '%s' con el nombre de la carpeta padre '%s'\n", nombreContenido, nombreDirPadre)

			// Si el nombre del contenido coincide con el nombre de la carpeta padre
			if strings.EqualFold(nombreContenido, nombreDirPadre) {
				return int(contenido.B_inodo), true, nil
			}
		}
	}

	return 0, false, nil
}

// Se crea una carpeta en un inodo específico
func (sb *Superbloque) CrearCarpetaEnInodo(archivo *os.File, indiceInodo int32, directoriosPadre []string, nombreDirectorio string, p bool) error {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return nil
	}

	seCreoArchivo := false
	// Se itera sobre cada bloque del inodo (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			//fmt.Printf("El inodo %d no tiene más bloques, saliendo.\n", indiceInodo)
			if !seCreoArchivo {
				fmt.Printf("Creando nuevo Bloque Carpeta para poder crear el directorio\n")
				err := sb.CrearNuevoBloqueCarpeta(archivo, int32(indiceInodo), directoriosPadre, nombreDirectorio, 0, nil, p, 0)
				if err != nil {
					fmt.Printf("Error con la creación de la carpeta que contendrá el directorio: %s", nombreDirectorio)
					return err
				}
			}
			return nil
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Se crea un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Se deserializa el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Se itera sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indiceContenido, indiceBloque)

			// Cuando se llega al directorio destino (destDir)
			if contenido.B_inodo != -1 {
				fmt.Printf("El inodo %d ya está ocupado con otro contenido, saltando al siguiente.\n", contenido.B_inodo)
				continue
			}

			fmt.Printf("Asignando el nombre del directorio '%s' al bloque en la posición %d\n", nombreDirectorio, indiceContenido)

			// Se actualiza el contenido del bloque con el nuevo directorio
			copy(contenido.B_nombre[:], nombreDirectorio)
			contenido.B_inodo = sb.S_inodes_count

			// Se actualiza el bloque con el nuevo contenido
			bloque.B_contenido[indiceContenido] = contenido

			// Serializar el bloque
			err = bloque.Codificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

			if err != nil {
				return fmt.Errorf("error al serializar el bloque %d: %v", indiceBloque, err)
			}

			fmt.Printf("Bloque %d actualizado con éxito.\n", indiceBloque)

			// Se crea el inodo de la nueva carpeta
			inodoCarpeta := &Inodo{
				I_uid:   1,
				I_gid:   1,
				I_s:     0,
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				I_type:  [1]byte{'0'},
				I_perm:  [3]byte{'6', '6', '4'},
			}

			fmt.Printf("Serializando el inodo de la carpeta '%s' (inodo %d)\n", nombreDirectorio, sb.S_inodes_count) // Depuración

			// Serializar el inodo de la nueva carpeta
			err = inodoCarpeta.Codificar(archivo, int64(sb.S_first_ino))

			if err != nil {
				return fmt.Errorf("error al serializar el inodo del directorio '%s': %v", nombreDirectorio, err)
			}

			// Se actualiza el bitmap de inodos
			err = sb.ActualizarBitmapInodo(archivo, sb.S_inodes_count, true)

			if err != nil {
				return fmt.Errorf("error al actualizar el bitmap de inodos para el directorio '%s': %v", nombreDirectorio, err)
			}

			// Se actualiza el superbloque con los nuevos valores de inodos
			sb.ActualizarSuperbloqueDespuesAsignacionInodo()

			// Se crea el bloque para la nueva carpeta
			bloqueCarpeta := &BloqueFolder{
				B_contenido: [4]ContenidoFolder{
					{B_nombre: [12]byte{'.'}, B_inodo: contenido.B_inodo},
					{B_nombre: [12]byte{'.', '.'}, B_inodo: indiceInodo},
					{B_nombre: [12]byte{'-'}, B_inodo: -1},
					{B_nombre: [12]byte{'-'}, B_inodo: -1},
				},
			}

			fmt.Printf("Serializando el bloque de la carpeta '%s'\n", nombreDirectorio)

			// Se serializa el bloque de la carpeta
			err = bloqueCarpeta.Codificar(archivo, int64(sb.S_first_blo))

			if err != nil {
				return fmt.Errorf("error al serializar el bloque del directorio '%s': %v", nombreDirectorio, err)
			}

			// Se actualiza el bitmap de bloques
			err = sb.ActualizarBitmapBloque(archivo, sb.S_blocks_count, true)

			if err != nil {
				return fmt.Errorf("error al actualizar el bitmap de bloques para el directorio '%s': %v", nombreDirectorio, err)
			}

			// Se actualiza el superbloque con los nuevos valores de bloques
			sb.ActualizarSuperbloqueDespuesAsignacionBloques()

			// Se agrega la entrada al journaling si es necesario
			if sb.S_filesystem_type == 3 {
				journaling_start := int64(sb.JournalStart())

				// Se construye la ruta completa para el journal de forma más robusta
				var rutaEntera string
				if len(directoriosPadre) > 0 {
					// Si hay directorios padres, incluirlos en la ruta
					rutaEntera = "/" + strings.Join(directoriosPadre, "/")
					if len(nombreDirectorio) > 0 {
						rutaEntera += "/" + nombreDirectorio
					}
				} else {
					rutaEntera = "/" + nombreDirectorio
				}

				// Se usa AgregarEntradaJournal para manejar automáticamente los índices y serialización
				if err := AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "mkdir", rutaEntera, "", sb); err != nil {
					fmt.Printf("Advertencia: error registrando operación en journal: %v\n", err)
				} else {
					fmt.Printf("Operación 'mkdir %s' registrada en journal correctamente\n", rutaEntera)
				}
			}

			fmt.Printf("Directorio '%s' creado correctamente en inodo %d.\n", nombreDirectorio, sb.S_inodes_count)
			return nil
		}
	}

	fmt.Printf("No se encontraron bloques disponibles para crear la carpeta '%s' en inodo %d\n", nombreDirectorio, indiceInodo)
	return nil
}

// Se crea un archivo en el sistema de archivos
func (sb *Superbloque) CrearArchivo(archivo *os.File, padresDir []string, destCarpeta string, tamanio int, contenido []string, r bool) error {
	fmt.Printf("Creando archivo '%s' con tamaño %d\n", destCarpeta, tamanio)

	// Se valida si la carpeta padre existe entre los bloques
	var padreEncontrado bool
	var indiceInodoEncontrado int32
	for i := int32(0); i < sb.S_inodes_count; i++ {
		indiceInodo, encontrado, err := sb.ValidarPadreDirectorio(archivo, i, padresDir, destCarpeta, r)
		if encontrado {
			padreEncontrado = true
			indiceInodoEncontrado = int32(indiceInodo)
			break
		}

		if err != nil {
			return err
		}
	}

	//Se valida si es carpeta o archivo
	partes := Herramientas.DefinirCarpetaArchivo(destCarpeta)

	if len(partes) == 1 {
		fmt.Printf("Se está validando una carpeta de nombre: %s\n", destCarpeta)
		return nil
	}

	// Se hacen las validaciones para crear el archivo
	if padreEncontrado {
		err := sb.CrearArchivoEnInodo(archivo, int32(indiceInodoEncontrado), padresDir, destCarpeta, tamanio, contenido, r)
		if err != nil {
			return err
		}
	} else if len(padresDir) == 0 {
		err := sb.CrearArchivoEnInodo(archivo, 0, padresDir, destCarpeta, tamanio, contenido, r)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("no se encontró la carpeta padre de: %s\n", destCarpeta)
		return fmt.Errorf("no se encontró la carpeta padre de: %s", destCarpeta)
	}

	if sb.S_filesystem_type == 3 {
		fullPath := "/" + strings.Join(append(padresDir, destCarpeta), "/")
		if err := AgregarEntradaJournal(archivo, int64(sb.JournalStart()), JOURNAL_ENTRIES, "mkfile", fullPath, strings.Join(contenido, ""), sb); err != nil {
			fmt.Printf("WARN journal: %v\n", err)
		}
	}

	return nil
}

// Se crea un archivo en un inodo específico
func (sb *Superbloque) CrearArchivoEnInodo(archivo *os.File, indiceInodo int32, padresDir []string, destArchivo string, tamanioArchivo int, contenidoArchivo []string, r bool) error {
	// Se crea un nuevo inodo
	fmt.Printf("Intentando crear archivo '%s' en inodo con índice %d\n", destArchivo, indiceInodo)

	inodo := &Inodo{}

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] == '1' {
		fmt.Printf("El inodo %d es una carpeta, omitiendo.\n", indiceInodo)
		return nil
	}

	seCreoArchivo := false

	// Se itera sobre cada bloque del inodo o (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			//fmt.Printf("El inodo %d no tiene más bloques, saliendo.\n", indiceInodo)
			if !seCreoArchivo {
				fmt.Printf("Creando nuevo Bloque Carpeta para poder crear el archivo\n")
				err := sb.CrearNuevoBloqueCarpeta(archivo, int32(indiceInodo), padresDir, destArchivo, tamanioArchivo, contenidoArchivo, r, 1)
				if err != nil {
					fmt.Printf("Error con la creación de la carpeta que contendrá el archivo: %s", destArchivo)
					return err
				}
			}
			return nil
		}

		// Se crea un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Se deserializa el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s))) // posición actual del bloque

		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		// Se iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son "." y ".."
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			// Se obtiene el contenido del bloque
			contenido := bloque.B_contenido[indiceContenido]

			// Si el apuntador al inodo está ocupado, se continua con el siguiente
			if contenido.B_inodo != -1 {
				fmt.Printf("El inodo %d ya está ocupado, continuando.\n", contenido.B_inodo)
				continue
			}

			// Se actualiza el contenido del bloque
			copy(contenido.B_nombre[:], []byte(destArchivo))
			contenido.B_inodo = sb.S_inodes_count

			// Se actualiza el bloque
			bloque.B_contenido[indiceContenido] = contenido

			// Se serializa el bloque
			err = bloque.Codificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

			if err != nil {
				return fmt.Errorf("error al serializar bloque %d: %v", indiceBloque, err)
			}

			fmt.Printf("Bloque actualizado para el archivo '%s' en el inodo %d\n", destArchivo, sb.S_inodes_count)

			// Se crea el inodo del archivo
			inodoArchivo := &Inodo{
				I_uid:   1,
				I_gid:   1,
				I_s:     int32(tamanioArchivo),
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				I_type:  [1]byte{'1'},
				I_perm:  [3]byte{'6', '6', '4'},
			}

			// Se crean los bloques del archivo
			for i := 0; i < len(contenidoArchivo); i++ {
				inodoArchivo.I_block[i] = sb.S_blocks_count

				// Se crea el bloque del archivo
				bloqueArchivo := &BloqueFile{
					B_contenido: [64]byte{},
				}

				copy(bloqueArchivo.B_contenido[:], contenidoArchivo[i])

				// Se serializa el bloque
				err = bloqueArchivo.Codificar(archivo, int64(sb.S_first_blo))

				if err != nil {
					return fmt.Errorf("error al serializar bloque de archivo: %v", err)
				}

				fmt.Printf("Bloque de archivo '%s' serializado correctamente.\n", destArchivo)

				// Se actualiza el bitmap de bloques
				err = sb.ActualizarBitmapBloque(archivo, sb.S_blocks_count, true)

				if err != nil {
					return fmt.Errorf("error al actualizar bitmap de bloque: %v", err)
				}

				// Se actualiza el superbloque
				sb.ActualizarSuperbloqueDespuesAsignacionBloques()
			}

			// Se serializa el inodo del archivo
			err = inodoArchivo.Codificar(archivo, int64(sb.S_first_ino))

			if err != nil {
				return fmt.Errorf("error al serializar inodo del archivo: %v", err)
			}

			fmt.Printf("Inodo del archivo '%s' serializado correctamente.\n", destArchivo)

			// Se actualiza el bitmap de inodos
			err = sb.ActualizarBitmapInodo(archivo, sb.S_inodes_count, true)

			if err != nil {
				return fmt.Errorf("error al actualizar bitmap de inodo: %v", err)
			}

			// Se actualiza el superbloque
			sb.ActualizarSuperbloqueDespuesAsignacionInodo()

			fmt.Printf("Archivo '%s' creado correctamente en el inodo %d.\n", destArchivo, sb.S_inodes_count)

			return nil

		}
	}
	fmt.Println("Ya no hubo espacio en el inodo para crear el archivo")
	return nil
}

// Se crea un Bloque Carpeta para cuando se acaba el espacio en un Bloque Carpeta y se necesita crear el archivo
func (sb *Superbloque) CrearNuevoBloqueCarpeta(archivo *os.File, indiceInodo int32, padresDir []string, destArchivo string, tamanioArchivo int, contenidoArchivo []string, r bool, tipoInodo int) error {
	// Se crea un nuevo inodo
	inodo := &Inodo{}

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] == '1' {
		fmt.Printf("El inodo %d es una carpeta, omitiendo.\n", indiceInodo)
		return nil
	}

	// Se itera sobre cada bloque del inodo o (apuntadores)
	for i, indiceBloque := range inodo.I_block {

		if indiceBloque == -1 {
			inodo.I_block[i] = sb.S_blocks_count

			// Se serializa el inodo del archivo
			err = inodo.Codificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

			if err != nil {
				return fmt.Errorf("error al serializar inodo del archivo: %v", err)
			}

			fmt.Printf("Inodo del archivo '%s' serializado correctamente.\n", destArchivo)

			// Se actualiza el bitmap de inodos
			err = sb.ActualizarBitmapInodo(archivo, sb.S_inodes_count, true)

			if err != nil {
				return fmt.Errorf("error al actualizar bitmap de inodo: %v", err)
			}

			// Se actualiza el superbloque
			sb.ActualizarSuperbloqueDespuesAsignacionInodo()

			indiceInodoPadre, _, err := sb.ValidarPadreDirectorio(archivo, indiceInodo, padresDir, destArchivo, true)

			if err != nil {
				fmt.Printf("Error al conseguir el id del padre para crear la carpeta que contendrá el archivo: %s", destArchivo)
				return err
			}

			// Se crea el bloque para la nueva carpeta
			bloqueCarpeta := &BloqueFolder{
				B_contenido: [4]ContenidoFolder{
					{B_nombre: [12]byte{'.'}, B_inodo: indiceInodo},
					{B_nombre: [12]byte{'.', '.'}, B_inodo: int32(indiceInodoPadre)},
					{B_nombre: [12]byte{'-'}, B_inodo: -1},
					{B_nombre: [12]byte{'-'}, B_inodo: -1},
				},
			}

			fmt.Printf("Serializando el bloque de la carpeta que va a contener el archivo '%s'\n", destArchivo)

			// Se serializa el bloque de la carpeta
			err = bloqueCarpeta.Codificar(archivo, int64(sb.S_first_blo))

			if err != nil {
				return fmt.Errorf("error al serializar el bloque del directorio '%s': %v", destArchivo, err)
			}

			// Se actualiza el bitmap de bloques
			err = sb.ActualizarBitmapBloque(archivo, sb.S_blocks_count, true)

			if err != nil {
				return fmt.Errorf("error al actualizar el bitmap de bloques para el directorio '%s': %v", destArchivo, err)
			}

			// Se actualiza el superbloque con los nuevos valores de bloques
			sb.ActualizarSuperbloqueDespuesAsignacionBloques()

			if tipoInodo == 0 {
				err = sb.CrearCarpetaEnInodo(archivo, int32(indiceInodo), padresDir, destArchivo, r)
				if err != nil {
					return err
				}
			} else if tipoInodo == 1 {
				err = sb.CrearArchivoEnInodo(archivo, int32(indiceInodo), padresDir, destArchivo, tamanioArchivo, contenidoArchivo, r)
				if err != nil {
					return err
				}
			}

			return nil
		}
	}
	return nil
}

// Se obtiene el indice del inodo del bloque
func (sb *Superbloque) ObtenerIndiceInodoDirectorio(archivo *os.File, indiceInodo int32, destDir string) (int, bool, error) {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		return 0, false, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return 0, false, nil
	}

	// Se itera sobre cada bloque del inodo (apuntadores)
	for _, indiceBloque := range inodo.I_block {
		// Si el bloque no existe, salir
		if indiceBloque == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", indiceInodo)
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Crear un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Deserializar el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return 0, false, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indiceContenido, indiceBloque)

			// Si el contenido está vacío, salir
			if contenido.B_inodo == -1 {
				fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", indiceInodo, indiceContenido)
				break
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			fmt.Printf("Comparando '%s' con el nombre de la carpeta '%s'\n", nombreContenido, destDir)

			// Si el nombre del contenido coincide con el nombre de la carpeta
			if strings.EqualFold(nombreContenido, destDir) {
				return int(contenido.B_inodo), true, nil
			}
		}
	}

	return 0, false, nil
}

// Se encuentran los padres de un directorio o archivo y se agrega el directorio o archivo
func (sb *Superbloque) ObtenerPadresDirectorioOArchivo(archivo *os.File, indiceApuntadorInodo int32, destDir string) ([]string, error) {
	var rutaEnteraDeDirectorio []string
	parar := true
	for parar {
		for i := int32(0); i < sb.S_inodes_count; i++ {
			directorioPadre, indicePadreInodo, err := sb.BuscarPadreEnInodo(archivo, i, indiceApuntadorInodo, destDir)

			if err != nil {
				return nil, err
			}

			if indicePadreInodo != 0 && directorioPadre != "" {
				rutaEnteraDeDirectorio = append(rutaEnteraDeDirectorio, directorioPadre)
				indiceApuntadorInodo = int32(indicePadreInodo)
			} else if indicePadreInodo == 0 && directorioPadre != "" {
				rutaEnteraDeDirectorio = append(rutaEnteraDeDirectorio, directorioPadre)
				parar = false
				break
			} else if indicePadreInodo == 0 && directorioPadre == "" {

			} else {
				parar = false
				break
			}
		}
	}

	return rutaEnteraDeDirectorio, nil
}

// Se busca el padre de un directorio o archivo mediante el inodo
func (sb *Superbloque) BuscarPadreEnInodo(archivo *os.File, indiceInodo int32, indiceApuntadorInodo int32, destDir string) (string, int, error) {
	inodo := &Inodo{}
	fmt.Printf("Deserializando el inodo '%d'.\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		return "", 0, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}

	fmt.Printf("Inodo '%d' fue deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	/*if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return "", 0, fmt.Errorf("inodo %d no es una carpeta, es de tipo: %c", indiceInodo, inodo.I_type[0])
	}*/

	for _, indiceBloque := range inodo.I_block {
		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Se crea un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Se deserializa el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return "", 0, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		if indiceBloque == -1 {
			fmt.Printf("El inodo '%d' no contiene ningún contenido, saltando al siguiente inodo.\n", indiceBloque)
			continue
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Se itera sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]

			// Cuando encuentra un directorio o archivo
			if contenido.B_inodo == -1 {
				fmt.Printf("El contenido del bloque '%d' no contiene nada, saltando al siguiente bloque.\n", contenido.B_inodo)
				continue
			}

			if contenido.B_inodo == indiceApuntadorInodo {
				nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
				return nombreContenido, int(bloque.B_contenido[0].B_inodo), err
			}

		}
	}
	return "", 0, nil
}

// Se cambian los permisos en el inodo
func (sb *Superbloque) BusquedaDeIndiceInodoParaCambiarPermisos(archivo *os.File, destDir string, ugo int) error {
	// Se valida si la carpeta padre existe entre los bloques
	var indiceInodoEncontrado int32
	for i := int32(0); i < sb.S_inodes_count; i++ {
		indiceInodo, encontrado, err := sb.ObtenerIndiceInodoDirectorio(archivo, i, destDir)
		if encontrado {
			indiceInodoEncontrado = int32(indiceInodo)
			break
		}

		if err != nil {
			return err
		}
	}

	err := sb.CambiarPermisosEnInodo(archivo, indiceInodoEncontrado, ugo)

	if err != nil {
		fmt.Printf("Hubo problemas al cambiar los permisos en el inodo '%d'.\n", indiceInodoEncontrado)
		return fmt.Errorf("hubo problemas al cambiar los permisos en el inodo %d: %v", indiceInodoEncontrado, err)
	}

	return nil
}

func (sb *Superbloque) CambiarPermisosEnInodo(archivo *os.File, indiceInodo int32, ugo int) error {
	// Crear un nuevo inodo
	inodo := &Inodo{}
	fmt.Printf("Deserializando inodo %d\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se convierten a string y se toman los primeros 3 caracteres
	ugoStr := fmt.Sprintf("%03d", ugo)

	inodo.I_perm = [3]byte{ugoStr[0], ugoStr[1], ugoStr[2]}

	return nil
}
