package Structs

import (
	"Proyecto/Herramientas"
	"fmt"
	"os"
)

type EBR struct {
	Ebr_mount [1]byte
	Ebr_fit   [1]byte
	Ebr_start int32
	Ebr_size  int32
	Ebr_next  int32
	Ebr_name  [16]byte
}

func ImprimirEBR(data EBR) {
	fmt.Println("part_status ", string(data.Ebr_mount[:]))
	//fmt.Println("part_type ", string(data.Type[:]))
	fmt.Println("part_fit: ", string(data.Ebr_fit[:]))
	fmt.Println("part_start: ", data.Ebr_start)
	fmt.Println("part_s ", data.Ebr_size)
	fmt.Println("part_name: ", string(data.Ebr_name[:]))
	fmt.Println("next_part: ", data.Ebr_next)
}

func CrearYEscribirEBR(start int32, size int32, fit byte, name string, file *os.File) error {
	fmt.Printf("Creando y escribiendo EBR en la posición: %d\n", start)

	ebr := &EBR{}
	ebr.ColocarEBR(fit, size, start, -1, name) // Establecer los valores del EBR

	return ebr.Codificar(file, int64(start))
}

func (e *EBR) Codificar(file *os.File, position int64) error {
	return Herramientas.EscribirAlArchivo(file, position, e)
}

func (e *EBR) ColocarEBR(fit byte, size int32, start int32, next int32, name string) {
	fmt.Println("Estableciendo valores del EBR:")
	fmt.Printf("Fit: %c | Tamaño: %d | Start: %d | Next: %d | Nombre: %s\n", fit, size, start, next, name)

	e.Ebr_mount[0] = '1'
	e.Ebr_fit[0] = fit
	e.Ebr_start = start
	e.Ebr_size = size
	e.Ebr_next = next

	copy(e.Ebr_name[:], name)
	for i := len(name); i < len(e.Ebr_name); i++ {
		e.Ebr_name[i] = 0 // Rellenar con ceros
	}
}

// Para la partición Lógica

func EncontrarUltimoEBR(start int32, archivo *os.File) (*EBR, error) {
	fmt.Printf("Buscando el último EBR a partir de la posición: %d\n", start)

	EBRActual := &EBR{}

	// Decodificar el EBR en la posición inicial
	err := EBRActual.Decodificar(archivo, int64(start))

	if err != nil {
		return nil, err
	}

	for EBRActual.Ebr_next != -1 {
		if EBRActual.Ebr_next < 0 {
			return EBRActual, nil
		}

		fmt.Printf("EBR encontrado - Start: %d, Next: %d\n", EBRActual.Ebr_start, EBRActual.Ebr_next)

		EBRSiguiente := &EBR{}

		err = EBRSiguiente.Decodificar(archivo, int64(EBRActual.Ebr_next))
		if err != nil {
			return nil, err
		}

		EBRActual = EBRSiguiente
	}

	fmt.Printf("Último EBR encontrado en la posición: %d\n", EBRActual.Ebr_start)
	return EBRActual, nil
}

/*func (e *EBR) LeerEBR(start int32, file *os.File) (*EBR, error) {
	fmt.Printf("Leyendo EBR desde el archivo en la posición: %d\n", start)
	return e.Decodificar(file, int64(start))
}*/

func (e *EBR) Decodificar(file *os.File, position int64) error {
	// Se verifica que la posición no sea negativa y esté dentro del rango del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error al obtener información del archivo: %v", err)
	}
	if position < 0 || position >= fileInfo.Size() {
		return fmt.Errorf("posición inválida para EBR: %d", position)
	}

	err = Herramientas.LeerDesdeElArchivo(file, position, e)
	if err != nil {
		return err
	}

	fmt.Printf("EBR decoded from position %d with success.\n", position)
	return nil
}

func (e *EBR) CalcularSiguienteStartEBR(extendedPartitionStart int32, extendedPartitionSize int32) (int32, error) {
	fmt.Printf("Calculando el inicio del siguiente EBR...\nEBR Actual - Start: %d, Size: %d, Next: %d\n",
		e.Ebr_start, e.Ebr_size, e.Ebr_next)

	if e.Ebr_size <= 0 {
		return -1, fmt.Errorf("EBR size is invalid or zero")
	}

	if e.Ebr_start < extendedPartitionStart {
		return -1, fmt.Errorf("EBR start position is invalid")
	}

	nextStart := e.Ebr_start + e.Ebr_size

	if nextStart <= e.Ebr_start || nextStart >= extendedPartitionStart+extendedPartitionSize {
		return -1, fmt.Errorf("error: el siguiente EBR está fuera de los límites de la partición extendida")
	}

	fmt.Printf("Inicio del siguiente EBR calculado con éxito: %d\n", nextStart)
	return nextStart, nil
}

// Se sobrescribe el espacio de la partición lógica (EBR) con ceros
func (e *EBR) Sobreescribir(archivo *os.File) error {
	// Se cerificar si el EBR tiene un tamaño válido
	if e.Ebr_size <= 0 {
		fmt.Println("El tamaño del EBR es inválido o cero.")
		return fmt.Errorf("el tamaño del EBR es inválido o cero")
	}

	// Se posiciona en el inicio del EBR (donde comienza la partición lógica)
	_, err := archivo.Seek(int64(e.Ebr_start), 0)
	if err != nil {
		fmt.Println("Error al mover el punter del archivo a la posición del EBR.")
		return fmt.Errorf("error al mover el puntero del archivo a la posición del EBR: %v", err)
	}

	// Se crea un buffer de ceros del tamaño de la partición lógica
	zeroes := make([]byte, e.Ebr_size)

	// Se escriben los ceros en el archivo
	_, err = archivo.Write(zeroes)
	if err != nil {
		fmt.Println("Error al sobreescribir el espacio del EBR.")
		return fmt.Errorf("error al sobrescribir el espacio del EBR: %v", err)
	}

	fmt.Printf("Espacio de la partición lógica (EBR) en posición %d sobrescrito con ceros.\n", e.Ebr_start)
	return nil
}

func (e *EBR) ColocarSiguienteEBR(newNext int32) {
	fmt.Printf("Estableciendo el siguiente EBR: Actual Start: %d, Nuevo Next: %d\n", e.Ebr_start, newNext)
	e.Ebr_next = newNext
}
