package Structs

import (
	"Proyecto/Herramientas"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type Superbloque struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             float64
	S_umtime            float64
	S_mnt_count         int32
	S_magic             int32
	S_inode_s           int32
	S_block_s           int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

func (sb *Superbloque) ImprimirSuperbloque() {
	fmt.Printf("%-25s %-10s\n", "Campo", "Valor")
	fmt.Printf("%-25s %-10s\n", "-------------------------", "----------")
	fmt.Printf("%-25s %-10d\n", "S_filesystem_type:", sb.S_filesystem_type)
	fmt.Printf("%-25s %-10d\n", "S_inodes_count:", sb.S_inodes_count)
	fmt.Printf("%-25s %-10d\n", "S_blocks_count:", sb.S_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_blocks_count:", sb.S_free_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_inodes_count:", sb.S_free_inodes_count)
	fmt.Printf("%-25s %-10s\n", "S_mtime:", time.Unix(int64(sb.S_mtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10s\n", "S_umtime:", time.Unix(int64(sb.S_umtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10d\n", "S_mnt_count:", sb.S_mnt_count)
	fmt.Printf("%-25s %-10x\n", "S_magic:", sb.S_magic)
	fmt.Printf("%-25s %-10d\n", "S_inode_size:", sb.S_inode_s)
	fmt.Printf("%-25s %-10d\n", "S_block_size:", sb.S_block_s)
	fmt.Printf("%-25s %-10d\n", "S_first_ino:", sb.S_first_ino)
	fmt.Printf("%-25s %-10d\n", "S_first_blo:", sb.S_first_blo)
	fmt.Printf("%-25s %-10d\n", "S_bm_inode_start:", sb.S_bm_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_bm_block_start:", sb.S_bm_block_start)
	fmt.Printf("%-25s %-10d\n", "S_inode_start:", sb.S_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_block_start:", sb.S_block_start)
}

func (sb *Superbloque) ImprimirBloques(ruta string) error {
	archivo, err := os.Open(ruta)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo %s: %w", ruta, err)
	}
	defer archivo.Close()

	fmt.Println("\nBloques\n----------------")
	inodos := make([]Inodo, sb.S_inodes_count)

	for i := int32(0); i < sb.S_inodes_count; i++ {
		inodo := &inodos[i]
		err := Herramientas.LeerDesdeElArchivo(archivo, int64(sb.S_inode_start+(i*int32(binary.Size(Inodo{})))), inodo)
		if err != nil {
			return fmt.Errorf("error al codificar el inodo %d: %w", i, err)
		}
	}

	for _, inode := range inodos {
		for _, indiceBloque := range inode.I_block {
			if indiceBloque == -1 {
				break
			}
			if inode.I_type[0] == '0' {
				block := &BloqueFolder{}
				err := Herramientas.LeerDesdeElArchivo(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)), block)
				if err != nil {
					return fmt.Errorf("error al decodificar el bloque folder %d: %w", indiceBloque, err)
				}
				fmt.Printf("\nBloque %d:\n", indiceBloque)
				block.ImprimirBloqueFolder()
			} else if inode.I_type[0] == '1' {
				block := &BloqueFile{}
				err := Herramientas.LeerDesdeElArchivo(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)), block)
				if err != nil {
					return fmt.Errorf("error al decodificar el bloque file %d: %w", indiceBloque, err)
				}
				fmt.Printf("\nBloque %d:\n", indiceBloque)
				block.ImprimirBloqueArchivo()
			}
		}
	}

	return nil
}

func (sb *Superbloque) ImprimirInodos(ruta string) error {
	archivo, err := os.Open(ruta)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", ruta, err)
	}
	defer archivo.Close()

	fmt.Println("\nInodos\n----------------")
	inodos := make([]Inodo, sb.S_inodes_count)

	for i := int32(0); i < sb.S_inodes_count; i++ {
		inodo := &inodos[i]
		err := Herramientas.LeerDesdeElArchivo(archivo, int64(sb.S_inode_start+(i*int32(binary.Size(Inodo{})))), inodo)
		if err != nil {
			return fmt.Errorf("failed to decode inode %d: %w", i, err)
		}
	}

	for i, inode := range inodos {
		fmt.Printf("\nInodo %d:\n", i)
		inode.ImprimirInodo()
	}

	return nil
}

func (sb *Superbloque) CrearArchivoUsersEnExt2(archivo *os.File) error {
	// ----------- Se crea el Inodo Raíz -----------
	raizInodo := &Inodo{
		I_uid:   1,
		I_gid:   1,
		I_s:     0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Se escribe el inodo raíz (inodo 0)
	err := Herramientas.EscribirAlArchivo(archivo, int64(sb.S_inode_start), raizInodo)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo raíz: %w", err)
	}

	// Se actualiza el bitmap de inodos (índice 0)
	err = sb.ActualizarBitmapInodo(archivo, 0, true)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos: %w", err)
	}

	// Se actualizan tanto el contador de inodos como el puntero al primer inodo libre
	sb.ActualizarSuperbloqueDespuesAsignacionInodo()

	// ----------- Se crea el Bloque Raíz (/ carpeta) -----------
	raizBloque := &BloqueFolder{
		B_contenido: [4]ContenidoFolder{
			{B_nombre: [12]byte{'.'}, B_inodo: 0},
			{B_nombre: [12]byte{'.', '.'}, B_inodo: 0},
			{B_nombre: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count},
			{B_nombre: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Se escribe el bloque raíz
	err = Herramientas.EscribirAlArchivo(archivo, int64(sb.S_block_start), raizBloque)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque raíz: %w", err)
	}

	// Se actualiza el bitmap de bloques (índice 0)
	err = sb.ActualizarBitmapBloque(archivo, 0, true)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
	}

	// Se actualiza tanto el contador de bloques como el puntero al primer bloque libre
	sb.ActualizarSuperbloqueDespuesAsignacionBloques()

	// ----------- Se crear el Inodo para /users.txt (inodo 1) -----------
	raizGrupo := NuevoGrupo("1", "root")
	raizUsuario := NuevoUsuario("1", "root", "root", "123")
	usuariosTexto := fmt.Sprintf("%s\n%s\n", raizGrupo.ToString(), raizUsuario.ToString())

	usuariosInodo := &Inodo{
		I_uid:   1,
		I_gid:   1,
		I_s:     int32(len(usuariosTexto)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Se escribe el inodo de users.txt (inodo 1)
	err = Herramientas.EscribirAlArchivo(archivo, int64(sb.S_inode_start+int32(binary.Size(usuariosInodo))), usuariosInodo)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo de users.txt: %w", err)
	}

	// Se actualiza el bitmap de inodos (índice 1)
	err = sb.ActualizarBitmapInodo(archivo, 1, true)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos para users.txt: %w", err)
	}

	// Se actualiza tanto el contador de inodos como el puntero al primer inodo libre
	sb.ActualizarSuperbloqueDespuesAsignacionInodo()

	// ----------- Se crea el Bloque para users.txt (bloque 1) -----------
	usuariosBloque := &BloqueFile{}
	copy(usuariosBloque.B_contenido[:], usuariosTexto)

	// Se escribe el bloque de users.txt
	err = Herramientas.EscribirAlArchivo(archivo, int64(sb.S_block_start+int32(binary.Size(usuariosBloque))), usuariosBloque)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de users.txt: %w", err)
	}

	// Se actualiza el bitmap de bloques (índice 1)
	err = sb.ActualizarBitmapBloque(archivo, 1, true)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques para users.txt: %w", err)
	}

	// Se actualiza tanto el contador de bloques como el puntero al primer bloque libre
	sb.ActualizarSuperbloqueDespuesAsignacionBloques()

	// Se imprime como quedaron tanto el superbloque, los bloques y los inodos
	fmt.Println("Archivo users.txt creado correctamente.")
	fmt.Println("Superbloque después de la creación de users.txt:")
	sb.ImprimirSuperbloque()
	fmt.Println("\nBloques:")
	sb.ImprimirBloques(archivo.Name())
	fmt.Println("\nInodos:")
	sb.ImprimirInodos(archivo.Name())

	return nil
}

func (sb *Superbloque) ActualizarSuperbloqueDespuesAsignacionInodo() {
	sb.S_inodes_count++

	sb.S_free_inodes_count--

	sb.S_first_ino += sb.S_inode_s
}

func (sb *Superbloque) Codificar(archivo *os.File, offset int64) error {
	return Herramientas.EscribirAlArchivo(archivo, offset, sb)
}

func (sb *Superbloque) Decodificar(archivo *os.File, offset int64) error {
	return Herramientas.LeerDesdeElArchivo(archivo, offset, sb)
}

// Se asigna un nuevo bloque al inodo en el índice especificado si es necesario
func (sb *Superbloque) AsignarNuevoBloque(archivo *os.File, inodo *Inodo, indice int) (int32, error) {
	fmt.Println("=== Iniciando la asignación de un nuevo bloque ===")

	// Se valida que el índice esté dentro del rango de bloques válidos
	if indice < 0 || indice >= len(inodo.I_block) {
		return -1, fmt.Errorf("índice de bloque fuera de rango: %d", indice)
	}

	// Se verifica si ya hay un bloque asignado en ese índice
	if inodo.I_block[indice] != -1 {
		return -1, fmt.Errorf("bloque en el índice %d ya está asignado: %d", indice, inodo.I_block[indice])
	}

	// Se trata de encontrar un bloque libre
	nuevoBloque, err := sb.EncontrarElSiguienteBloqueLibre(archivo)
	if err != nil {
		return -1, fmt.Errorf("error buscando nuevo bloque libre: %w", err)
	}

	// Se verifica si se encontró un bloque libre
	if nuevoBloque == -1 {
		return -1, fmt.Errorf("no hay bloques libres disponibles")
	}

	// Se asigna el nuevo bloque en el índice especificado
	inodo.I_block[indice] = nuevoBloque
	fmt.Printf("Nuevo bloque asignado: %d en I_block[%d]\n", nuevoBloque, indice)

	// Se actualiza el Superbloque después de asignar el bloque
	sb.ActualizarSuperbloqueDespuesAsignacionBloques()

	return nuevoBloque, nil
}

// Se asigna un nuevo inodo y lo marca como ocupado
func (sb *Superbloque) AsignarNuevoInodo(archivo *os.File) (int32, error) {
	// Se intenta encontrar un inodo libre
	inodoNuevo, err := sb.EncontrarElSiguienteInodoLibre(archivo)
	if err != nil {
		fmt.Println("Error al buscar el nuevo inodo libre.")
		return -1, fmt.Errorf("error buscando nuevo inodo libre: %w", err)
	}

	// Se verifica si se encontró un inodo libre
	if inodoNuevo == -1 {
		fmt.Println("No hay inodos libres disponibles.")
		return -1, fmt.Errorf("no hay inodos libres disponibles")
	}

	// Se actualiza el superbloque después de asignar el inodo
	sb.ActualizarSuperbloqueDespuesAsignacionInodo()

	// Se retorna el nuevo inodo asignado
	return inodoNuevo, nil
}

// Se busca el siguiente inodo libre en el bitmap de inodos y lo marca como ocupado
func (sb *Superbloque) EncontrarElSiguienteInodoLibre(archivo *os.File) (int32, error) {
	totalInodes := sb.S_inodes_count + sb.S_free_inodes_count // Número total de inodos

	// Recorre todos los inodos en el bitmap
	for posicion := int32(0); posicion < totalInodes; posicion++ {
		// Verifica si el inodo está libre
		isFree, err := sb.EstaElInodoLibre(archivo, sb.S_bm_inode_start, posicion)
		if err != nil {
			return -1, fmt.Errorf("error buscando inodo libre en la posición %d: %w", posicion, err)
		}

		// Si encontramos un inodo libre
		if isFree {
			// Marcar el inodo como ocupado
			err = sb.ActualizarBitmapInodo(archivo, posicion, true)
			if err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap del inodo en la posición %d: %w", posicion, err)
			}
			// Devolver la posición del inodo libre encontrado
			fmt.Printf("Inodo libre encontrado y asignado: %d\n", posicion)
			return posicion, nil
		}
	}

	// Si no hay inodos disponibles
	return -1, fmt.Errorf("no hay inodos disponibles")
}

// Se busca el siguiente bloque libre y lo marca como ocupado
func (sb *Superbloque) EncontrarElSiguienteBloqueLibre(archivo *os.File) (int32, error) {
	totalBlocks := sb.S_blocks_count + sb.S_free_blocks_count // Número total de bloques

	for posicion := int32(0); posicion < totalBlocks; posicion++ {
		isFree, err := sb.EstaElBloqueLibre(archivo, sb.S_bm_block_start, posicion)
		if err != nil {
			return -1, fmt.Errorf("error buscando bloque libre: %w", err)
		}

		if isFree {
			// Se marca el bloque como ocupado
			err = sb.ActualizarBitmapBloque(archivo, posicion, true)
			if err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap del bloque: %w", err)
			}

			// Se devuelve el índice del bloque libre encontrado
			fmt.Println("Indice encontrado:", posicion)
			return posicion, nil
		}
	}

	// Si no hay bloques disponibles
	return -1, fmt.Errorf("no hay bloques disponibles")
}

// Se actualiza el Superbloque después de asignar un inodo
func (sb *Superbloque) ActualizarSuperbloqueDespuesAsignacionBloques() {
	// Se incrementa el contador de inodos asignados
	sb.S_blocks_count++

	// Se decrementa el contador de inodos libres
	sb.S_free_blocks_count--

	// Se actualiza el puntero al primer inodo libre
	sb.S_first_blo += sb.S_block_s
}

// Se calcula el inicio del área de journaling
func (sb *Superbloque) JournalStart() int32 {
	// El journal se encuentra justo antes del inicio del bitmap de inodos
	tamanioJournal := int32(binary.Size(Journal{}))
	start := sb.S_bm_inode_start - JOURNAL_ENTRIES*tamanioJournal
	fmt.Printf("Superbloque.JournalStart: bm_inode_start=%d, journalSize=%d, entries=%d -> start=%d\n", sb.S_bm_inode_start, tamanioJournal, JOURNAL_ENTRIES, start)
	return start
}

// Se calcula el final del área de journaling
func (sb *Superbloque) JournalEnd() int32 {
	end := sb.S_bm_inode_start
	fmt.Printf("Superbloque.JournalEnd: bm_inode_start=%d -> end=%d\n", sb.S_bm_inode_start, end)
	return end
}

// Se calcula el offset del Inodo
func (sb *Superbloque) CalcularOffsetInodo(inodeIndex int32) int64 {
	// Calcula el desplazamiento en el archivo basado en el índice del inodo
	return int64(sb.S_inode_start) + int64(inodeIndex)*int64(sb.S_inode_s)
}

// Se elimina un archivo del sistema de archivos
func (sb *Superbloque) EliminarArchivo(archivo *os.File, directoriosPadre []string, nombreArchivo string) error {
	fmt.Printf("Intentando eliminar el archivo '%s'\n", nombreArchivo)

	// Si no hay directorio padre, eliminar desde el directorio raíz
	if len(directoriosPadre) == 0 {
		return sb.eliminarArchivoEnInodo(archivo, 0, nombreArchivo)
	}

	// Se navega recursivamente por la estructura de directorios
	indiceInodoActual := int32(0)

	// Se recorre cada nivel de directorio padre
	for _, nombreDirectorio := range directoriosPadre {
		encontrado := false

		// Se carga el inodo del directorio actual
		inodoActual := &Inodo{}
		if err := inodoActual.Decodificar(archivo, int64(sb.S_inode_start+indiceInodoActual*sb.S_inode_s)); err != nil {
			return fmt.Errorf("error cargando directorio actual (inodo %d): %w", indiceInodoActual, err)
		}

		// Se obtienen los bloques de datos del directorio
		indicesBloques, err := inodoActual.ObtenerIndicesBloquesDatos(archivo, sb)
		if err != nil {
			return fmt.Errorf("error obteniendo bloques de directorio: %w", err)
		}

		// Se busca el siguiente directorio en la ruta
		for _, blockIndex := range indicesBloques {
			if encontrado {
				break
			}

			block := &BloqueFolder{}
			if err := block.Decodificar(archivo, int64(sb.S_block_start+blockIndex*sb.S_block_s)); err != nil {
				return fmt.Errorf("error deserializando bloque %d: %w", blockIndex, err)
			}

			// Se busca la carpeta en este bloque
			for _, contenido := range block.B_contenido {
				nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

				if contenido.B_inodo != -1 && strings.EqualFold(nombreContenido, nombreDirectorio) {
					// Se verifica que sea un directorio
					subDirInode := &Inodo{}
					if err := subDirInode.Decodificar(archivo, int64(sb.S_inode_start+contenido.B_inodo*sb.S_inode_s)); err != nil {
						return fmt.Errorf("error cargando inodo %d: %w", contenido.B_inodo, err)
					}

					if subDirInode.I_type[0] != '0' {
						return fmt.Errorf("la entrada '%s' no es un directorio", nombreDirectorio)
					}

					// Se avanza al siguiente directorio
					indiceInodoActual = contenido.B_inodo
					encontrado = true
					break
				}
			}
		}

		if !encontrado {
			return fmt.Errorf("no se encontró el directorio '%s' en la ruta", nombreDirectorio)
		}
	}
	// Se llega al directorio que debería contener el archivo
	return sb.eliminarArchivoEnInodo(archivo, indiceInodoActual, nombreArchivo)
}

// Se elimina un archivo en un inodo específico utilizando las funciones avanzadas
func (sb *Superbloque) eliminarArchivoEnInodo(archivo *os.File, indiceInodo int32, nombreArchivo string, rutaPadre ...string) error {
	//Se carga el inodo del directorio
	inodoDirectorio := &Inodo{}

	err := inodoDirectorio.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))

	if err != nil {
		fmt.Printf("Error al deserializar el inodo '%d'", indiceInodo)
		return fmt.Errorf("error al deserializar inodo %d: %w", indiceInodo, err)
	}

	// Se verificar que el inodo sea un directorio
	if inodoDirectorio.I_type[0] != '0' {
		fmt.Printf("el inodo '%d' no es una carpeta", inodoDirectorio.I_s)
		return fmt.Errorf("el inodo %d no es una carpeta", inodoDirectorio.I_s)
	}

	// Se obtienen solo los bloques de datos del directorio (no apuntadores)
	indicesBloques, err := inodoDirectorio.ObtenerIndicesBloquesDatos(archivo, sb)
	if err != nil {
		fmt.Println("Error al obtener los bloques de datos del directorio.")
		return fmt.Errorf("error obteniendo bloques de datos del directorio: %w", err)
	}

	// Se procesa cada bloque del directorio
	for _, indiceBloque := range indicesBloques {
		bloque := &BloqueFolder{}
		offsetBloque := int64(sb.S_block_start + indiceBloque*sb.S_block_s)

		if err := bloque.Decodificar(archivo, offsetBloque); err != nil {
			fmt.Printf("Error al deserializar el bloque '%d'", indiceBloque)
			return fmt.Errorf("error deserializando bloque %d: %w", indiceBloque, err)
		}

		// Se busca el archivo en el bloque
		for i, contenido := range bloque.B_contenido {
			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

			if contenido.B_inodo != -1 && strings.EqualFold(nombreContenido, nombreArchivo) {
				indiceInodoArchivo := contenido.B_inodo
				fmt.Printf("Archivo '%s' encontrado en inodo %d, eliminando.\n", nombreArchivo, indiceInodoArchivo)

				// Se carga el inodo del archivo
				inodoArchivo := &Inodo{}
				offsetInodoArchivo := int64(sb.S_inode_start + (indiceInodoArchivo * sb.S_inode_s))
				if err := inodoArchivo.Decodificar(archivo, offsetInodoArchivo); err != nil {
					fmt.Printf("Error al deserializar el inodo del archivo '%d'", indiceInodoArchivo)
					return fmt.Errorf("error deserializando inodo del archivo %d: %w", indiceInodoArchivo, err)
				}

				// Se verifica que sea efectivamente un archivo
				if inodoArchivo.I_type[0] != '1' {
					return fmt.Errorf("el inodo %d no es un archivo sino de tipo %c", indiceInodoArchivo, inodoArchivo.I_type[0])
				}

				// Se registra en el journal si es necesario
				if sb.S_filesystem_type == 3 {
					// Se construye la ruta completa del archivo
					var rutaEntera string

					// Si se proporciona una ruta de directorio padre
					if len(rutaPadre) > 0 && rutaPadre[0] != "" {
						rutaEntera = rutaPadre[0] + "/" + nombreArchivo
					} else {
						rutaEntera = "/" + nombreArchivo
					}

					journaling_start := int64(sb.JournalStart())

					// Cargar los contenidos del archivo para el journal
					datosArchivo, err := inodoArchivo.LeerDatos(archivo, sb)

					contenidoArchivo := ""

					if err != nil {
						// No fallamos aquí, solo log
						fmt.Printf("Error leyendo contenido del archivo para journal: %v\n", err)
					} else {
						contenidoArchivo = string(datosArchivo)
					}

					// Se usa AgregarEntradaJournal que maneja automáticamente índices y serialización
					if err := AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "rm", rutaEntera, contenidoArchivo, sb); err != nil {
						fmt.Printf("Advertencia: error registrando operación en journal: %v\n", err)
					} else {
						fmt.Printf("Operación 'rm %s' registrada en journal correctamente\n", rutaEntera)
					}
				}

				// Se liberan todos los bloques del archivo usando LiberarTodosLosBloques
				if err := inodoArchivo.LiberarTodosLosBloques(archivo, sb); err != nil {
					return fmt.Errorf("error liberando bloques del archivo: %w", err)
				}

				// Se libera el inodo
				if err := sb.ActualizarBitmapInodo(archivo, indiceInodoArchivo, false); err != nil {
					return fmt.Errorf("error liberando inodo %d: %w", indiceInodoArchivo, err)
				}

				sb.ActualizarSuperbloqueDespuesAsignacionInodo()

				// Se limpia la entrada en el directorio
				bloque.B_contenido[i] = ContenidoFolder{B_nombre: [12]byte{'-'}, B_inodo: -1}
				if err := bloque.Codificar(archivo, offsetBloque); err != nil {
					return fmt.Errorf("error actualizando bloque de directorio: %w", err)
				}

				// 12. Verificar y liberar bloques de apuntadores vacíos
				if err := inodoDirectorio.ChequearYLiberarBloquesIndirectosVacios(archivo, sb); err != nil {
					fmt.Printf("Advertencia: error al verificar bloques indirectos vacíos: %v\n", err)
				}

				fmt.Printf("Archivo '%s' eliminado correctamente.\n", nombreArchivo)
				return nil
			}
		}
	}

	return fmt.Errorf("archivo '%s' no encontrado en directorio (inodo %d)", nombreArchivo, indiceInodo)
}

// Se elimina un directorio y su contenido recursivamente en el sistema de archivos
func (sb *Superbloque) EliminarDirectorio(archivo *os.File, directoriosPadre []string, nombreDirectorio string) error {
	fmt.Printf("Intentando eliminar carpeta '%s'\n", nombreDirectorio)

	// Se construye la ruta completa para pasar al journaling
	var rutaEntera string
	if len(directoriosPadre) > 0 {
		rutaEntera = "/" + strings.Join(directoriosPadre, "/") + "/" + nombreDirectorio
	} else {
		rutaEntera = "/" + nombreDirectorio
	}
	fmt.Printf("Ruta completa: %s\n", rutaEntera)

	// Si no hay directorio padre, eliminar desde el directorio raíz
	if len(directoriosPadre) == 0 {
		return sb.EliminarDirectorioDesdeDirectorios(archivo, 0, nombreDirectorio, rutaEntera)
	}

	// Se navega recursivamente por la estructura de directorios
	indiceInodoActual := int32(0)

	// Se recorre cada nivel de directorio padre
	for _, nombreDirectorio := range directoriosPadre {
		encontrado := false

		// Se carga el inodo del directorio actual
		inodoActual := &Inodo{}
		if err := inodoActual.Decodificar(archivo, int64(sb.S_inode_start+indiceInodoActual*sb.S_inode_s)); err != nil {
			fmt.Printf("Error al cargar el directorio actual con el inodo '%d'.\n", indiceInodoActual)
			return fmt.Errorf("error cargando directorio actual (inodo %d): %w", indiceInodoActual, err)
		}

		// Se verifica que sea un directorio
		if inodoActual.I_type[0] != '0' {
			fmt.Printf("El inodo '%d' no es un directorio.\n", indiceInodoActual)
			return fmt.Errorf("el inodo %d no es un directorio", indiceInodoActual)
		}

		// Se obtienen los bloques de datos del directorio
		indicesBloques, err := inodoActual.ObtenerIndicesBloquesDatos(archivo, sb)
		if err != nil {
			fmt.Println("Error al obtener los bloques de directorio.")
			return fmt.Errorf("error obteniendo bloques de directorio: %w", err)
		}

		// Buscar el siguiente directorio en la ruta
		for _, indiceBloque := range indicesBloques {
			if encontrado {
				break
			}

			block := &BloqueFolder{}
			if err := block.Decodificar(archivo, int64(sb.S_block_start+indiceBloque*sb.S_block_s)); err != nil {
				fmt.Printf("Error al deserializar el bloque '%d'.\n", indiceBloque)
				return fmt.Errorf("error deserializando bloque %d: %w", indiceBloque, err)
			}

			// Se busca la carpeta en este bloque
			for _, contenido := range block.B_contenido {
				nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

				if contenido.B_inodo != -1 && strings.EqualFold(nombreContenido, nombreDirectorio) {
					// Se verifica que sea un directorio
					subDirInode := &Inodo{}
					if err := subDirInode.Decodificar(archivo, int64(sb.S_inode_start+contenido.B_inodo*sb.S_inode_s)); err != nil {
						fmt.Printf("Error al cargar el inodo '%d'.\n", contenido.B_inodo)
						return fmt.Errorf("error cargando inodo %d: %w", contenido.B_inodo, err)
					}

					if subDirInode.I_type[0] != '0' {
						fmt.Printf("La entrada '%s' no es un directorio.\n", nombreDirectorio)
						return fmt.Errorf("la entrada '%s' no es un directorio", nombreDirectorio)
					}

					// Se avanza al siguiente directorio
					indiceInodoActual = contenido.B_inodo
					encontrado = true
					break
				}
			}
		}

		if !encontrado {
			fmt.Printf("No se encontró el directorio '%s' en la ruta", nombreDirectorio)
			return fmt.Errorf("no se encontró el directorio '%s' en la ruta", nombreDirectorio)
		}
	}

	// Se llega al directorio que debería contener la carpeta a eliminar
	return sb.EliminarDirectorioDesdeDirectorios(archivo, indiceInodoActual, nombreDirectorio, rutaEntera)
}

// Nuevo método auxiliar para eliminar una carpeta de un directorio específico
func (sb *Superbloque) EliminarDirectorioDesdeDirectorios(archivo *os.File, indiceInodoPadre int32, nombreDirectorio string, rutaEntera string) error {
	// Se carga el inodo del directorio padre
	inodoPadre := &Inodo{}
	if err := inodoPadre.Decodificar(archivo, int64(sb.S_inode_start+indiceInodoPadre*sb.S_inode_s)); err != nil {
		fmt.Printf("Error al deserializar el inodo del directorio padre '%d'.\n", indiceInodoPadre)
		return fmt.Errorf("error deserializando inodo del directorio padre %d: %w", indiceInodoPadre, err)
	}

	// Se verifica que sea un directorio
	if inodoPadre.I_type[0] != '0' {
		fmt.Printf("El inodo '%d' no es un directorio.\n", indiceInodoPadre)
		return fmt.Errorf("el inodo %d no es un directorio", indiceInodoPadre)
	}

	// Se obtienen los bloques de datos del directorio
	indicesBloques, err := inodoPadre.ObtenerIndicesBloquesDatos(archivo, sb)
	if err != nil {
		return fmt.Errorf("error obteniendo bloques de datos: %w", err)
	}

	// Se busca la carpeta a eliminar
	for _, indiceBloque := range indicesBloques {
		bloque := &BloqueFolder{}
		offsetBloque := int64(sb.S_block_start + indiceBloque*sb.S_block_s)

		if err := bloque.Decodificar(archivo, offsetBloque); err != nil {
			fmt.Printf("Error al deserializar el bloque '%d'.\n", indiceBloque)
			return fmt.Errorf("error deserializando bloque %d: %w", indiceBloque, err)
		}

		// Se busca la entrada en el directorio
		for i, contenido := range bloque.B_contenido {
			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

			if contenido.B_inodo != -1 && strings.EqualFold(nombreContenido, nombreDirectorio) {
				// Se verifica que sea un directorio
				inodoFolder := &Inodo{}
				if err := inodoFolder.Decodificar(archivo, int64(sb.S_inode_start+contenido.B_inodo*sb.S_inode_s)); err != nil {
					fmt.Printf("Error al deserialziar el inodo '%d'.\n", contenido.B_nombre)
					return fmt.Errorf("error deserializando inodo %d: %w", contenido.B_inodo, err)
				}

				if inodoFolder.I_type[0] != '0' {
					fmt.Printf("'%s' no es un directorio.\n", nombreDirectorio)
					return fmt.Errorf("'%s' no es un directorio", nombreDirectorio)
				}

				// Se elimina el directorio recursivamente usando la ruta completa
				if err := sb.EliminarDirectorioEnInodo(archivo, contenido.B_inodo, rutaEntera); err != nil {
					return fmt.Errorf("error eliminando carpeta '%s': %w", nombreDirectorio, err)
				}

				// Se limpia la entrada en el directorio padre
				bloque.B_contenido[i] = ContenidoFolder{
					B_nombre: [12]byte{'-'},
					B_inodo:  -1,
				}

				// Se guarda el bloque actualizado
				if err := bloque.Codificar(archivo, offsetBloque); err != nil {
					fmt.Println("Error al actualizar el bloque de directorios.")
					return fmt.Errorf("error actualizando bloque de directorio: %w", err)
				}

				fmt.Printf("Carpeta '%s' eliminada correctamente\n", nombreDirectorio)
				return nil
			}
		}
	}
	fmt.Printf("Carpeta '%s' no encontrada en el directorio.\n", nombreDirectorio)
	return fmt.Errorf("carpeta '%s' no encontrada en el directorio", nombreDirectorio)
}

// Se elimina recursivamente el contenido de una carpeta en un inodo específico
func (sb *Superbloque) EliminarDirectorioEnInodo(archivo *os.File, indiceInodo int32, rutaDirectorio ...string) error {
	// Se deserializa el inodo del directorio
	inodoDirectorio := &Inodo{}
	err := inodoDirectorio.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		fmt.Printf("Error al deserializar el inodo '%d'.\n", indiceInodo)
		return fmt.Errorf("error al deserializar inodo %d: %w", indiceInodo, err)
	}

	// Se verifica que el inodo sea de tipo directorio
	if inodoDirectorio.I_type[0] != '0' {
		fmt.Printf("El inodo '%d' no es una carpeta.\n", indiceInodo)
		return fmt.Errorf("el inodo %d no es una carpeta", indiceInodo)
	}

	// Se obtiene la ruta completa del directorio para el journal
	rutaEntera := "/"

	if len(rutaDirectorio) > 0 && rutaDirectorio[0] != "" {
		rutaEntera = rutaDirectorio[0]
	}

	// Se calcula la posición de inicio del journal
	journaling_start := int64(sb.JournalStart())

	// Se registra en el journal si es de tipo EXT3
	if sb.S_filesystem_type == 3 {
		// Se usa el nuevo método AgregarEntradaJournal para registrar la operación
		if err := AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "rmdir", rutaEntera, "", sb); err != nil {
			fmt.Printf("Advertencia: error registrando operación en journal: %v\n", err)
		} else {
			fmt.Printf("Operación 'rmdir %s' registrada en journal correctamente\n", rutaEntera)
		}
	}

	// Se obtienen todos los bloques de datos del directorio
	indicesBloques, err := inodoDirectorio.ObtenerIndicesBloquesDatos(archivo, sb)
	if err != nil {
		fmt.Println("Error al obtener los bloques de datos del directorio.")
		return fmt.Errorf("error obteniendo bloques de datos del directorio: %w", err)
	}

	// Se procesa cada bloque del directorio
	for _, indiceBloque := range indicesBloques {
		// Se carga el bloque de directorio
		bloque := &BloqueFolder{}
		offsetBloque := int64(sb.S_block_start + indiceBloque*sb.S_block_s)
		if err := bloque.Decodificar(archivo, offsetBloque); err != nil {
			fmt.Printf("Error al deserializar el bloque '%d'", indiceBloque)
			return fmt.Errorf("error deserializando bloque %d: %w", indiceBloque, err)
		}

		// Se procesa cada entrada en el bloque
		for _, contenido := range bloque.B_contenido {
			// Saltar entradas vacías o especiales (. y ..)
			if contenido.B_inodo == -1 ||
				string(contenido.B_nombre[:1]) == "." ||
				string(contenido.B_nombre[:2]) == ".." {
				continue
			}

			// Se obtiene el nombre del contenido y se construye su ruta
			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			fmt.Printf("Eliminando contenido '%s' en inodo %d\n", nombreContenido, contenido.B_inodo)

			rutaHijo := rutaEntera
			if !strings.HasSuffix(rutaHijo, "/") {
				rutaHijo += "/"
			}
			rutaHijo += nombreContenido

			// Se carga el inodo del contenido
			inodoHijo := &Inodo{}
			offsetInodoHijo := int64(sb.S_inode_start + (contenido.B_inodo * sb.S_inode_s))
			if err := inodoHijo.Decodificar(archivo, offsetInodoHijo); err != nil {
				return fmt.Errorf("error deserializando inodo hijo %d: %w", contenido.B_inodo, err)
			}

			// Se procesa según el tipo de contenido (directorio o archivo)
			if inodoHijo.I_type[0] == '0' {
				// Se registra la eliminación de carpeta en el journal
				if sb.S_filesystem_type == 3 {
					if err := AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "rmdir", rutaHijo, "", sb); err != nil {
						fmt.Printf("Advertencia: error registrando eliminación de subcarpeta en journal: %v\n", err)
					}
				}

				// Se elimina recursivamente la subcarpeta
				if err := sb.EliminarDirectorioEnInodo(archivo, contenido.B_inodo, rutaHijo); err != nil {
					fmt.Printf("Error al eliminar la subcarpeta '%s'.\n", nombreContenido)
					return fmt.Errorf("error eliminando subcarpeta '%s': %w", nombreContenido, err)
				}
			} else { // Es un archivo
				// Se registra la eliminación del archivo en el journal
				if sb.S_filesystem_type == 3 {
					// Obtener el contenido del archivo para el journal
					datosArchivo, err := inodoHijo.LeerDatos(archivo, sb)
					contenidoArchivo := ""
					if err == nil {
						contenidoArchivo = string(datosArchivo)
					}

					if err := AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "rm", rutaHijo, contenidoArchivo, sb); err != nil {
						fmt.Printf("Advertencia: error registrando eliminación de archivo '%s' en journal: %v\n", nombreContenido, err)
					} else {
						fmt.Printf("Operación 'rm %s' registrada en journal correctamente\n", rutaHijo)
					}
				}

				// Se liberan todos los bloques del archivo
				if err := inodoHijo.LiberarTodosLosBloques(archivo, sb); err != nil {
					fmt.Printf("Error al liberar los bloques del archivo '%s'.\n", nombreContenido)
					return fmt.Errorf("error liberando bloques del archivo '%s': %w", nombreContenido, err)
				}

				// Se libera el inodo del archivo
				if err := sb.ActualizarBitmapInodo(archivo, contenido.B_inodo, false); err != nil {
					fmt.Printf("Error al liberar el inodo '%d'.\n", contenido.B_inodo)
					return fmt.Errorf("error liberando inodo %d: %w", contenido.B_inodo, err)
				}

				sb.ActualizarSuperbloqueDespuesAsignacionInodo()
				fmt.Printf("Archivo '%s' eliminado (inodo %d)\n", nombreContenido, contenido.B_inodo)
			}
		}
	}

	// Se liberan todos los bloques del directorio
	if err := inodoDirectorio.LiberarTodosLosBloques(archivo, sb); err != nil {
		fmt.Println("Error liberando los bloques del directorio.")
		return fmt.Errorf("error liberando bloques del directorio: %w", err)
	}

	// Se verificar y se liberan los bloques de apuntadores indirectos vacíos
	if err := inodoDirectorio.ChequearYLiberarBloquesIndirectosVacios(archivo, sb); err != nil {
		fmt.Printf("Advertencia: error al verificar bloques indirectos vacíos: %v\n", err)
	}

	// Se libera el inodo del directorio
	if err := sb.ActualizarBitmapInodo(archivo, indiceInodo, false); err != nil {
		return fmt.Errorf("error liberando inodo del directorio %d: %w", indiceInodo, err)
	}
	sb.ActualizarSuperbloqueDespuesAsignacionInodo()

	fmt.Printf("Carpeta en inodo %d eliminada correctamente.\n", indiceInodo)

	return nil
}

// Se copia un directorio y su contenido recursivamente en el sistema de archivos
func (sb *Superbloque) NuevasRutas(archivo *os.File, directoriosPadreRuta []string, nombreDirectorioRuta string, directoriosPadreDestino []string, nombreDirectorioDestino string) ([]string, error) {
	indiceInodoEncontrado, _, err := sb.BuscarIndiceOIndicesInodoDeDirectorio(archivo, nombreDirectorioRuta, false)
	if err != nil {
		return nil, err
	}

	var rutas []string
	var indices []int32
	indices = append(indices, int32(indiceInodoEncontrado))
	var indicesProvisionales []int32
	parar := true
	for parar {
		for i := int32(0); i < int32(len(indices)); i++ {
			rutaCompleta, indicesDirectorios, err := sb.CompletarRuta(archivo, nombreDirectorioRuta, indices[i], directoriosPadreDestino, nombreDirectorioDestino)
			if err != nil {
				return nil, err
			}

			for j := int32(0); j < int32(len(rutaCompleta)); j++ {
				rutas = append(rutas, rutaCompleta[j])
			}

			for k := int32(0); k < int32(len(indicesDirectorios)); k++ {
				indicesProvisionales = append(indicesProvisionales, indicesDirectorios[k])
			}
		}

		if len(indicesProvisionales) == 0 {
			parar = false
		} else {
			indices = indicesProvisionales
			indicesProvisionales = nil
		}
	}

	return rutas, nil
}

// Se completa la nueva ruta con los padres nuevos
func (sb *Superbloque) CompletarRuta(archivo *os.File, nombreDirectorio string, indiceInodo int32, directoriosPadreDestino []string, nombreDirectorioDestino string) ([]string, []int32, error) {
	var rutaCompleta []string
	var indicesDirectorios []int32

	inodo := &Inodo{}
	fmt.Printf("Deserializando el inodo '%d'.\n", indiceInodo)

	// Se deserializar el inodo
	err := inodo.Decodificar(archivo, int64(sb.S_inode_start+(indiceInodo*sb.S_inode_s)))
	if err != nil {
		return nil, nil, fmt.Errorf("error al deserializar inodo %d: %v", indiceInodo, err)
	}

	fmt.Printf("Inodo '%d' fue deserializado. Tipo: %c\n", indiceInodo, inodo.I_type[0])

	// Se verifica si el inodo es de tipo carpeta
	if inodo.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", indiceInodo, inodo.I_type[0])
		return nil, nil, fmt.Errorf("inodo %d no es una carpeta, es de tipo: %c", indiceInodo, inodo.I_type[0])
	}

	for _, indiceBloque := range inodo.I_block {
		fmt.Printf("Deserializando bloque %d del inodo %d\n", indiceBloque, indiceInodo)
		// Se crea un nuevo bloque de carpeta
		bloque := &BloqueFolder{}

		// Se deserializa el bloque
		err := bloque.Decodificar(archivo, int64(sb.S_block_start+(indiceBloque*sb.S_block_s)))

		if err != nil {
			return nil, nil, fmt.Errorf("error al deserializar bloque %d: %v", indiceBloque, err)
		}

		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", indiceBloque, indiceInodo)

		// Se itera sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indiceContenido := 2; indiceContenido < len(bloque.B_contenido); indiceContenido++ {
			contenido := bloque.B_contenido[indiceContenido]

			// Cuando encuentra un directorio o archivo
			if contenido.B_inodo == -1 {
				fmt.Printf("El inodo '%d' no contiene ningún contenido, saltando al siguiente.\n", contenido.B_inodo)
				continue
			}

			var rutaNueva string

			for i := 0; i < len(directoriosPadreDestino); i++ {
				directorio := directoriosPadreDestino[i]
				rutaNueva += "/" + directorio
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")
			if nombreContenido != "" {
				rutaNueva += "/" + nombreDirectorioDestino

				rutaEnteraDeDirectorio, err := sb.ObtenerPadresDirectorioOArchivo(archivo, contenido.B_inodo, nombreContenido)

				var rutaOrdenada []string
				for i := len(rutaEnteraDeDirectorio) - 1; i >= 0; i-- {
					rutaOrdenada = append(rutaOrdenada, rutaEnteraDeDirectorio[i])
				}

				if err != nil {
					return nil, nil, err
				}

				bandera := false

				for i := 0; i < len(rutaOrdenada); i++ {
					if bandera {
						directorio := rutaOrdenada[i]
						rutaNueva += "/" + directorio
					} else {
						if rutaOrdenada[i] == nombreDirectorio {
							bandera = true
							directorio := rutaOrdenada[i]
							rutaNueva += "/" + directorio
						}
					}

				}

				rutaCompleta = append(rutaCompleta, rutaNueva)
			}

			if strings.Contains(strings.ToLower(nombreContenido), ".") {

			} else {
				if contenido.B_inodo != 0 {
					indicesDirectorios = append(indicesDirectorios, int32(contenido.B_inodo))
				}
			}
			rutaNueva = ""
		}
	}

	return rutaCompleta, indicesDirectorios, nil
}

// Se obtiene la antigua ruta del archivo
func (sb *Superbloque) ObtenerRutaAntiguaArchivo(directoriosPadreRuta []string, nombreDirectorioRuta string, rutaNueva string) string {
	var rutaCompletaAntigua string
	bandera := false

	directoriosRutaNueva, nombreArchivo := Herramientas.ObtenerDirectoriosPadreYArchivo(rutaNueva)

	directoriosPadreRuta = append(directoriosPadreRuta, nombreDirectorioRuta)

	for _, directorio := range directoriosRutaNueva {
		if bandera {
			directoriosPadreRuta = append(directoriosPadreRuta, directorio)
		} else {
			if directorio == nombreDirectorioRuta {
				bandera = true
			}
		}
	}

	directoriosPadreRuta = append(directoriosPadreRuta, nombreArchivo)

	for _, dirOarc := range directoriosPadreRuta {
		rutaCompletaAntigua += "/" + dirOarc
	}

	return rutaCompletaAntigua
}

// Se busca uno o más indices de un directorio que está repetido
func (sb *Superbloque) BuscarIndiceOIndicesInodoDeDirectorio(archivo *os.File, nombreDirectorio string, booleano bool) (int32, []int32, error) {
	var indiceInodoEncontrado int32
	var indicesInodoEncontrados []int32

	if booleano {
		for i := int32(0); i < sb.S_inodes_count; i++ {
			indiceInodo, encontrado, err := sb.ObtenerIndiceInodoDirectorio(archivo, i, nombreDirectorio)
			if encontrado {
				indicesInodoEncontrados = append(indicesInodoEncontrados, int32(indiceInodo))
			}

			if err != nil {
				return 0, nil, err
			}
		}
	} else {
		for i := int32(0); i < sb.S_inodes_count; i++ {
			indiceInodo, encontrado, err := sb.ObtenerIndiceInodoDirectorio(archivo, i, nombreDirectorio)
			if encontrado {
				indiceInodoEncontrado = int32(indiceInodo)
				break
			}

			if err != nil {
				return 0, nil, err
			}
		}
	}

	return indiceInodoEncontrado, indicesInodoEncontrados, nil
}
