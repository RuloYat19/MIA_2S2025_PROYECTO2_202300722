package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

type MKFS struct {
	id  string
	typ string
	fs  string
}

func Mkfs(parametros []string) string {
	fmt.Println("\n======== MKFS ========")
	salida1 := ""

	mkfs := &MKFS{}

	//Datos comando
	//Opcionales
	mkfs.fs = "2fs"
	mkfs.typ = "full"
	//Otras variables
	paramC := true
	idInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "MKFS Error: Valor desconocido del parámetro " + tmp[0] + "\n"
			fmt.Println("MKFS Error: Valor desconocido del parámetro ", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "id" {
			idInit = true
			mkfs.id = tmp[1]
		} else if strings.ToLower(tmp[0]) == "type" {
			mkfs.typ = tmp[1]
		} else if strings.ToLower(tmp[0]) == "fs" {
			if tmp[1] == "2fs" || tmp[1] == "3fs" {
				mkfs.fs = tmp[1]
			} else {
				salida1 += "MKFS Error: Parámetro inválido para -fs. Se leyó: " + tmp[1] + ". Los valores aceptados son 2fs y 3fs.\n"
			}
		} else {
			salida1 += "MKFS Error: Parámetro desconocido " + tmp[0] + "\n"
			fmt.Println("MKFS Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && idInit {
		err := formatearParticion(mkfs)

		if err != nil {
			salida1 += "MKFS Error: Hubo problemas al formatear la partición.\n"
			fmt.Println("MKFS Error:", err)
		} else {
			salida1 += "Partición con ID " + mkfs.id + " formateada exitosamente.\n"
			fmt.Printf("Partición con ID '%s' formateada exitosamente.\n", mkfs.id)
		}
	}
	fmt.Println("\n=======FIN MKFS=======")
	return salida1
}

func formatearParticion(mkfs *MKFS) error {
	particionMontada, rutaParticion, err := globales.ObtenerParticionMontadas(mkfs.id)

	if err != nil {
		fmt.Printf("Error al obtener la partición montada con ID: '%s'.\n", mkfs.id)
		return fmt.Errorf("error al obtener la partición montada con ID %s: %v", mkfs.id, err)
	}

	archivo, err := os.OpenFile(rutaParticion, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error al abrir el archivo de la partición en '%s'.\n", rutaParticion)
		return fmt.Errorf("error abriendo el archivo de la partición en %s: %v", rutaParticion, err)
	}
	defer archivo.Close()

	fmt.Printf("Partición montada correctamente en '%s'.\n", rutaParticion)
	fmt.Println("Partición montada:")
	particionMontada.ImprimirParticion()

	n := calcularN(particionMontada, mkfs.fs)
	fmt.Println("Valor de n:", n)

	superbloque := crearSuperbloque(particionMontada, n, mkfs.fs)
	fmt.Println("\nSuperBlock:")
	superbloque.ImprimirSuperbloque()

	err = superbloque.CrearBitMaps(archivo)
	if err != nil {
		fmt.Println("Error creando los bitmaps.")
		return fmt.Errorf("error creando bitmaps: %v", err)
	}
	fmt.Println("Bitmaps creados correctamente.")

	if mkfs.fs == "3fs" {
		err = superbloque.CrearArchivoUsersEnExt3(archivo, int64(particionMontada.Start+int32(binary.Size(Structs.Superbloque{}))))
	} else {
		err = superbloque.CrearArchivoUsersEnExt2(archivo)
	}
	if err != nil {
		fmt.Println("Error creando el archivo users.txt.")
		return fmt.Errorf("error creando el archivo users.txt: %v", err)
	}
	fmt.Println("Archivo users.txt creado correctamente.")

	err = Herramientas.EscribirAlArchivo(archivo, int64(particionMontada.Start), superbloque)
	if err != nil {
		fmt.Println("Error al escribir el superbloque en el disco.")
		return fmt.Errorf("error al escribir el superbloque en disco: %v", err)
	}

	fmt.Println("Superbloque escrito correctamente en el disco.")

	// Se lee el MBR del archivo
	var mbr Structs.MBR

	err = mbr.Decodificar(archivo)

	if err != nil {
		fmt.Println("Error al deserializar el MBR.")
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	var nomPar string
	// Se actualiza el MBR con la partición ya eliminada
	for i := range mbr.Particiones {
		idParticion := mbr.Particiones[i].Id[:]
		idLimpio := string(bytes.TrimRight(idParticion, "\x00"))
		if idLimpio == mkfs.id {
			nombreParticion := mbr.Particiones[i].Nombre[:]
			nombreLimpio := string(bytes.TrimRight(nombreParticion, "\x00"))
			nomPar = nombreLimpio
		}
	}

	structParticion := Structs.NuevaParticionFormateada(rutaParticion, nomPar, mkfs.id)

	Structs.AgregarParticionFormateada(structParticion)

	Structs.ImprimirParticionesFormateadas()

	return nil
}

func calcularN(particion *Structs.Particion, fs string) int32 {
	numerador := int(particion.Tamanio) - binary.Size(Structs.Superbloque{})
	denominadorBase := 4 + binary.Size(Structs.Inodo{}) + 3*binary.Size(Structs.BloqueFile{})
	tmp := 0
	if fs == "3fs" {
		tmp = binary.Size(Structs.Journal{})
	}
	denominador := denominadorBase + tmp
	n := math.Floor(float64(numerador) / float64(denominador))

	return int32(n)
}

func crearSuperbloque(particion *Structs.Particion, n int32, fs string) *Structs.Superbloque {
	journal_start, bm_inode_start, bm_block_start, inode_start, block_start := calcularPosicionesStart(particion, fs, n)
	//Bitmaps
	fmt.Println("\nInicio del SuperBlock:", particion.Start)
	fmt.Println("\nFin del SuperBlock:", particion.Start+int32(binary.Size(Structs.Superbloque{})))
	fmt.Println("\nInicio del Journal:", journal_start)
	fmt.Println("\nFin del Journal:", journal_start+int32(binary.Size(Structs.Journal{})))
	fmt.Println("\nInicio del Bitmap de Inodos:", bm_inode_start)
	fmt.Println("\nFin del Bitmap de Inodos:", bm_inode_start+n)
	fmt.Println("\nInicio del Bitmap de Bloques:", bm_block_start)
	fmt.Println("\nFin del Bitmap de Bloques:", bm_block_start+(3*n))
	fmt.Println("\nInicio de Inodos:", inode_start)
	var fsType int32

	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	// Crear un nuevo superbloque
	superBlock := &Structs.Superbloque{
		S_filesystem_type:   fsType,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float64(time.Now().Unix()),
		S_umtime:            float64(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_s:           int32(binary.Size(Structs.Inodo{})),
		S_block_s:           int32(binary.Size(Structs.BloqueFile{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}

func calcularPosicionesStart(partition *Structs.Particion, fs string, n int32) (int32, int32, int32, int32, int32) {
	superblockSize := int32(binary.Size(Structs.Superbloque{}))
	journalSize := int32(binary.Size(Structs.Journal{}))
	inodeSize := int32(binary.Size(Structs.Inodo{}))

	// Se ponen los valores por defecto
	journalStart := int32(0)
	bmInodeStart := partition.Start + superblockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	// Se ajustan las posiciones para EXT3 usando el journal
	if fs == "3fs" {
		// Número fijo de entradas de journal
		const journalEntries = 50

		// Se arranca el journal justo después del superbloque
		journalStart = partition.Start + superblockSize

		// Se arranca el bitmap de inodos tras reservar espacio para journalEntries entradas
		bmInodeStart = journalStart + journalEntries*journalSize

		// El resto de estructuras se sigue con el mismo layout
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)
	}

	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart
}
