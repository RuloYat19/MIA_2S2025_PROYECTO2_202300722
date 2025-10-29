package Structs

import (
	"Proyecto/Herramientas"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

type MBR struct {
	MbrSize     int32
	FechaC      [16]byte
	Id          int32
	Fit         [1]byte
	Particiones [4]Particion
}

type Particion struct {
	Estado      [1]byte
	Tipo        [1]byte
	Fit         [1]byte
	Start       int32
	Tamanio     int32
	Nombre      [16]byte
	Correlativo int32
	Id          [4]byte
}

func (p *Particion) SetInfo(newType string, fit string, newStart int32, newSize int32, name string, correlativo int32) {
	p.Tamanio = newSize
	p.Start = newStart
	p.Correlativo = correlativo
	copy(p.Nombre[:], name)
	copy(p.Fit[:], fit)
	copy(p.Estado[:], "I")
	copy(p.Tipo[:], newType)
}

func ImprimirMBR(data MBR) {
	fmt.Println("\n     Disco")
	fmt.Printf("FechaCreacion: %s, fit: %s, size: %d, id: %d\n", string(data.FechaC[:]), string(data.Fit[:]), data.MbrSize, data.Id)
	for i := 0; i < 4; i++ {
		fmt.Printf("Particion %d: %s, %s, %d, %d, %s, %d\n", i, string(data.Particiones[i].Nombre[:]), string(data.Particiones[i].Tipo[:]), data.Particiones[i].Start, data.Particiones[i].Tamanio, string(data.Particiones[i].Fit[:]), data.Particiones[i].Correlativo)
	}
}

func (p *Particion) ImprimirParticion() {
	fmt.Printf("Estado: %c | Tipo: %c | Fit: %c | Start: %d | Tamaño: %d | Nombre: %s | Correlativo: %d | ID: %s\n",
		p.Estado[0], p.Tipo[0], p.Fit[0], p.Start, p.Tamanio,
		string(p.Nombre[:]), p.Correlativo, string(p.Id[:]))
}

func (mbr *MBR) Codificar(file *os.File) error {
	return Herramientas.EscribirAlArchivo(file, 0, mbr)
}

func (mbr *MBR) Decodificar(file *os.File) error {
	return Herramientas.LeerDesdeElArchivo(file, 0, mbr)
}

func (mbr *MBR) ObtenerPrimerParticionDisponible() (*Particion, int, int) {
	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.Particiones); i++ {
		if mbr.Particiones[i].Start == 0 {
			return &mbr.Particiones[i], offset, i
		} else {
			offset += int(mbr.Particiones[i].Tamanio)
		}
	}
	return nil, 0, 0
}

func (mbr *MBR) CalcularEspacioDisponible() (int32, error) {
	tamanioTotal := mbr.MbrSize
	espacioUsado := int32(binary.Size(MBR{}))

	particiones := mbr.Particiones[:]
	for _, part := range particiones {
		if part.Tamanio != 0 {
			espacioUsado += part.Tamanio
		}
	}

	if espacioUsado >= tamanioTotal {
		return 0, fmt.Errorf("there is no available space on the disk")
	}

	return tamanioTotal - espacioUsado, nil
}

// Se modifica el tamaño de una partición
func (p *Particion) ModificarTamanioDeLaParticion(addTamanio int32, espacioDisponible int32) error {
	nuevoEspacio := p.Tamanio + addTamanio

	if nuevoEspacio < 0 {
		fmt.Println("El tamanño de la partición no puede ser negativo.")
		return fmt.Errorf("el tamaño de la partición no puede ser negativo")
	}

	if addTamanio > 0 && espacioDisponible < addTamanio {
		fmt.Println("No hay suficiente espacio disponible para agregar a la partición.")
		return fmt.Errorf("no hay suficiente espacio disponible para agregar a la partición")
	}
	p.Tamanio = nuevoEspacio

	fmt.Printf("El tamaño de la partición '%s' ha sido modificado. Nuevo tamaño: %d bytes.\n", string(p.Nombre[:]), p.Tamanio)
	return nil
}

// Se calcula el espacio disponible a partir del final de la partición actual
func (mbr *MBR) CalcularEspacioDisponibleDeLaParticion(partition *Particion) (int32, error) {
	startOfPartition := partition.Start
	endOfPartition := startOfPartition + partition.Tamanio
	var nextPartitionStart int32 = -1
	for _, p := range mbr.Particiones {
		if p.Start > endOfPartition && (nextPartitionStart == -1 || p.Start < nextPartitionStart) {
			nextPartitionStart = p.Start
		}
	}
	if nextPartitionStart == -1 {
		nextPartitionStart = mbr.MbrSize
	}

	availableSpace := nextPartitionStart - endOfPartition
	if availableSpace < 0 {
		return 0, fmt.Errorf("el cálculo de espacio disponible resultó en un valor negativo")
	}

	return availableSpace, nil
}

func (p *Particion) CrearParticion(partStart, partTamanio int, partTipo, partFit, partNombre string) {
	p.Estado[0] = '0'
	p.Start = int32(partStart)
	p.Tamanio = int32(partTamanio)

	if len(partTipo) > 0 {
		p.Tipo[0] = partTipo[0]
	}

	if len(partFit) > 0 {
		p.Fit[0] = partFit[0]
	}

	copy(p.Nombre[:], partNombre)
}

func (mbr *MBR) TieneParticionExtendida() bool {
	for _, particion := range mbr.Particiones {

		if particion.Tipo[0] == 'E' {
			return true
		}
	}
	return false
}

func (mbr *MBR) ObtenerParticionNombre(nombre string) *Particion {
	for i, particion := range mbr.Particiones {
		particionNombre := strings.Trim(string(particion.Nombre[:]), "\x00 ")
		nombreEntrada := strings.Trim(nombre, "\x00 ")

		if strings.EqualFold(particionNombre, nombreEntrada) {
			return &particion
		}

		if i == 1 {

		}
	}
	return nil
}

func (mbr *MBR) ObtenerParticionPorID(id string) (*Particion, error) {
	for i := 0; i < len(mbr.Particiones); i++ {
		particionID := strings.Trim(string(mbr.Particiones[i].Id[:]), "\x00 ")
		entradaID := strings.Trim(id, "\x00 ")

		if strings.EqualFold(particionID, entradaID) {
			return &mbr.Particiones[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}

func (p *Particion) MontarParticion(correlativo int, id string) error {
	if p == nil {
		fmt.Println("Error: estas intentando montar una partición que no existe")
		return fmt.Errorf("error: estas intentando montar una partición que no existe")
	}

	p.Correlativo = int32(correlativo)
	copy(p.Id[:], id)
	return nil
}

func (p *Particion) EliminarParticion(tipoDelete string, archivo *os.File, isExtended bool) error {
	if isExtended {
		err := p.eliminarParticionesLogicas(archivo)
		if err != nil {
			fmt.Println("Error al eliminar las particiones lógicas dentro de la partición.")
			return fmt.Errorf("error al eliminar las particiones lógicas dentro de la partición extendida: %v", err)
		}
	}

	if tipoDelete == "full" {
		err := p.Sobreescribir(archivo)
		if err != nil {
			fmt.Println("Error al sobreescribir en el disco de la partición eliminada.")
			return fmt.Errorf("error al sobrescribir la partición: %v", err)
		}
	}

	p.Start = -1
	p.Tamanio = -1
	p.Nombre = [16]byte{}

	fmt.Printf("La partición ha sido eliminada con el método (%s).\n", tipoDelete)
	return nil
}

// Se eliminan todas las particiones lógicas dentro de una partición extendida
func (p *Particion) eliminarParticionesLogicas(archivo *os.File) error {
	fmt.Println("Eliminando particiones lógicas dentro de la partición extendida...")

	var currentEBR EBR

	start := p.Start

	for {
		err := currentEBR.Decodificar(archivo, int64(start))

		if err != nil {
			fmt.Println("Error al leer el EBR de la partición.")
			return fmt.Errorf("error al leer el EBR: %v", err)
		}

		if currentEBR.Ebr_start == -1 {
			break
		}

		currentEBR.Ebr_start = -1
		currentEBR.Ebr_size = -1
		copy(currentEBR.Ebr_name[:], "")

		err = currentEBR.Sobreescribir(archivo)

		if err != nil {
			return fmt.Errorf("error al sobrescribir el EBR: %v", err)
		}

		start = currentEBR.Ebr_next
	}

	fmt.Println("Particiones lógicas eliminadas exitosamente.")
	return nil
}

// Se sobrescribe el espacio de la partición con \0 (para eliminación Full)
func (p *Particion) Sobreescribir(archivo *os.File) error {
	_, err := archivo.Seek(int64(p.Start), 0)

	if err != nil {
		fmt.Println("Error al hacer el seek.")
		return err
	}

	zeroes := make([]byte, p.Tamanio)
	_, err = archivo.Write(zeroes)
	if err != nil {
		fmt.Println("Error al sobre escribir el espacio de la partición.")
		return fmt.Errorf("error al sobrescribir el espacio de la partición: %v", err)
	}

	fmt.Println("El espacio de la partición se ha sobrescrito con ceros exitosamente.")
	return nil
}
