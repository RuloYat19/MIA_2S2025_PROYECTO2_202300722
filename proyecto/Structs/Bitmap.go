package Structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	BloqueBitLibre   = 0
	BloqueBitOcupado = 1
)

func (sb *Superbloque) CrearBitMaps(archivo *os.File) error {
	err := sb.crearBitmap(archivo, sb.S_bm_inode_start, sb.S_inodes_count+sb.S_free_inodes_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de inodos: %w", err)
	}

	err = sb.crearBitmap(archivo, sb.S_bm_block_start, sb.S_blocks_count+sb.S_free_blocks_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de bloques: %w", err)
	}

	return nil
}

func (sb *Superbloque) crearBitmap(archivo *os.File, start int32, contador int32, ocupado bool) error {
	_, err := archivo.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error buscando el inicio del bitmap: %w", err)
	}

	byteContador := (contador + 7) / 8

	llenarByte := byte(0x00)
	if ocupado {
		llenarByte = 0xFF
	}

	buffer := make([]byte, byteContador)
	for i := range buffer {
		buffer[i] = llenarByte
	}

	err = binary.Write(archivo, binary.LittleEndian, buffer)
	if err != nil {
		return fmt.Errorf("error escribiendo el bitmap: %w", err)
	}

	return nil
}

// Se actualiza el bitmap de inodos
func (sb *Superbloque) ActualizarBitmapInodo(archivo *os.File, posicion int32, ocupado bool) error {
	return sb.actualizarBitmap(archivo, sb.S_bm_inode_start, posicion, ocupado)
}

// Se actualiza el bitmap de bloques
func (sb *Superbloque) ActualizarBitmapBloque(archivo *os.File, posicion int32, ocupado bool) error {
	return sb.actualizarBitmap(archivo, sb.S_bm_block_start, posicion, ocupado)
}

func (sb *Superbloque) actualizarBitmap(archivo *os.File, start int32, posicion int32, ocupado bool) error {
	byteIndice := posicion / 8
	bitOffset := posicion % 8

	_, err := archivo.Seek(int64(start)+int64(byteIndice), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap: %w", err)
	}

	var byteValor byte
	err = binary.Read(archivo, binary.LittleEndian, &byteValor)
	if err != nil {
		return fmt.Errorf("error leyendo el byte del bitmap: %w", err)
	}

	if ocupado {
		byteValor |= (1 << bitOffset)
	} else {
		byteValor &= ^(1 << bitOffset)
	}

	_, err = archivo.Seek(int64(start)+int64(byteIndice), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap para escribir: %w", err)
	}

	err = binary.Write(archivo, binary.LittleEndian, &byteValor)
	if err != nil {
		return fmt.Errorf("error escribiendo el byte actualizado del bitmap: %w", err)
	}

	return nil
}

// Verifica si un inodo en el bitmap está libre
func (sb *Superbloque) EstaElInodoLibre(archivo *os.File, start int32, posicion int32) (bool, error) {
	byteIndex := posicion / 8 // Calcular el byte dentro del bitmap
	bitOffset := posicion % 8 // Calcular el bit dentro del byte

	// Leer el byte que contiene el bit correspondiente al inodo
	_, err := archivo.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando el byte en el bitmap de inodos: %w", err)
	}

	var byteVal byte
	err = binary.Read(archivo, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte del bitmap de inodos: %w", err)
	}

	// Verificar si el bit correspondiente está en 0 (libre)
	return (byteVal & (1 << bitOffset)) == 0, nil
}

// Se verifica si un bloque en el bitmap está libre
func (sb *Superbloque) EstaElBloqueLibre(archivo *os.File, start int32, posicion int32) (bool, error) {
	// Se calculan tanto el byte como el bit dentro del byte
	byteIndex := posicion / 8
	bitOffset := posicion % 8

	// Se mueve el puntero al byte correspondiente
	_, err := archivo.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando la posición en el bitmap: %w", err)
	}

	// Se lee el byte actual
	var byteVal byte
	err = binary.Read(archivo, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte del bitmap: %w", err)
	}

	var estaLibre bool

	if (byteVal & (1 << bitOffset)) == 0 {
		estaLibre = true
	} else {
		estaLibre = false
	}

	// Se verifica si el bit está libre (0) u ocupado (1)
	return estaLibre, nil
}
