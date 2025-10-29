package Herramientas

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	rutaArchivo "path/filepath"
	"strings"
)

// Se convierte un tamaño y una unidad a bytes
func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid unit")
	}
}

func CrearDisco(ruta string) error {
	dir := rutaArchivo.Dir(ruta)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Error al crear el disco, path: ", err)
		return err
	}

	if _, err := os.Stat(ruta); os.IsNotExist(err) {
		nuevoArchivo, err := os.Create(ruta)
		if err != nil {
			fmt.Println("Error al crear el disco: ", err)
			return err
		}
		defer nuevoArchivo.Close()
	}
	return nil
}

func AbrirArchivo(nombre string) (*os.File, error) {
	archivo, err := os.OpenFile(nombre, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile ==", err)
		return nil, err
	}
	return archivo, nil
}

func EscribirObjeto(archivo *os.File, data interface{}, posicion int64) error {
	archivo.Seek(posicion, 0)
	err := binary.Write(archivo, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

func LeerObjeto(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

func LeerDesdeElArchivo(archivo *os.File, offset int64, data interface{}) error {
	_, err := archivo.Seek(offset, 0)

	if err != nil {
		return fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	err = binary.Read(archivo, binary.LittleEndian, data)

	if err != nil {
		return fmt.Errorf("failed to read data from file: %w", err)
	}

	return nil
}

func EscribirAlArchivo(archivo *os.File, offset int64, data interface{}) error {
	_, err := archivo.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	err = binary.Write(archivo, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

var rutaALetra = make(map[string]string)

var siguienteIndiceLetra = 0

var alfabeto = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

func ObtenerLetra(ruta string) (string, error) {
	if _, exists := rutaALetra[ruta]; !exists {
		if siguienteIndiceLetra < len(alfabeto) {
			rutaALetra[ruta] = alfabeto[siguienteIndiceLetra]
			siguienteIndiceLetra++
		} else {
			return "", errors.New("no hay más letras disponibles para asignar")
		}
	}

	return rutaALetra[ruta], nil
}

// Se obtienen los directorios
func ObtenerDirectorios(ruta string) ([]string, string) {
	ruta = rutaArchivo.Clean(ruta)

	// Se divide la ruta en sus componentes
	componentes := strings.Split(ruta, string(rutaArchivo.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var padresDirs []string

	// Se construyen las rutas de las carpetas padres
	for i := 1; i < len(componentes); i++ {
		padresDirs = append(padresDirs, componentes[i])
	}

	// La última carpeta es la carpeta de destino
	destDir := componentes[len(componentes)-1]

	return padresDirs, destDir
}

// Se obtienen las carpetas padre
func ObtenerDirectoriosPadre(ruta string) []string {
	ruta = rutaArchivo.Clean(ruta)

	// Se divide la ruta en sus componentes
	componentes := strings.Split(ruta, string(rutaArchivo.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var padresDirs []string

	// Se construyen las rutas de las carpetas padres
	for i := 1; i < len(componentes)-1; i++ {
		padresDirs = append(padresDirs, componentes[i])
	}

	return padresDirs
}

// Se obtienen las carpetas padre
func ObtenerDirectoriosPadreYArchivo(ruta string) ([]string, string) {
	ruta = rutaArchivo.Clean(ruta)

	// Se divide la ruta en sus componentes
	componentes := strings.Split(ruta, string(rutaArchivo.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var padresDirs []string

	// Se construyen las rutas de las carpetas padres
	for i := 1; i < len(componentes)-1; i++ {
		padresDirs = append(padresDirs, componentes[i])
	}

	// La última carpeta es la carpeta de destino
	destDir := componentes[len(componentes)-1]
	return padresDirs, destDir
}

func DefinirCarpetaArchivo(directorio string) []string {
	partes := strings.Split(directorio, ".")
	resultado := make([]string, 0)

	for _, parte := range partes {
		if parte != "" {
			resultado = append(resultado, parte)
		}
	}

	if len(resultado) == 0 {
		return []string{directorio}
	}

	return resultado
}

func PadreCarpeta(slice []string, valor string) (string, bool) {
	for i, v := range slice {
		if v == valor {
			// Verificar que no sea el primer elemento
			if i > 0 {
				return slice[i-1], true
			}
			return "", false // No hay elemento anterior
		}
	}
	return "", false // Valor no encontrado
}

// Se elimina un elemento de un slice en el índice dado
func RemoverElemento[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// Se divide una cadena en partes de tamaño chunkSize y las almacena en una lista
func DividirCadenaEnChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// Se crean las carpetas padre si no existen
func CrearPadreDirs(path string) error {
	dir := filepath.Dir(path)

	// os.MkdirAll no sobrescribe las carpetas existentes, solo crea las que no existen
	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		return fmt.Errorf("error al crear las carpetas padre: %v", err)
	}

	return nil
}

// Se obtiene el nombre del archivo .dot y el nombre de la imagen de salida
func ObtenerNombreArchivos(path string) (string, string) {
	dir := filepath.Dir(path)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dotFileName := filepath.Join(dir, baseName+".dot")
	outputImage := path
	return dotFileName, outputImage
}
