package Comandos

import (
	cat "Proyecto/Comandos/AdministracionArchivos"
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

type REMOVE struct {
	ruta string
}

func Remove(parametros []string) string {
	fmt.Println("\n======= REMOVE =======")

	var salida1 = ""

	remove := &REMOVE{}

	//Parametros
	//Obligatorios
	pathInit := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "REMOVE Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("REMOVE Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			remove.ruta = tmp[1]

		} else {
			salida1 += "REMOVE Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("REMOVE Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit {
		err := ejecutarRemove(remove)

		if err != nil {
			fmt.Println("REMOVE Error: Hubo problemas al borrar el archivo o carpeta.")
			salida1 += "REMOVE Error: Hubo problemas al borrar el archivo o carpeta.\n"
		} else {
			fmt.Println("La carpeta o archivo se ha eliminado con éxito.")
			salida1 += "La carpeta o archivo se ha eliminado con éxito.\n"
		}
	}

	fmt.Println("\n======FIN REMOVE======")
	return salida1
}

func ejecutarRemove(remove *REMOVE) error {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, particionMontada, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada.")
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de la partición.")
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}

	defer archivo.Close()

	// Se llamaa a la función refactorizada para eliminar archivo/carpeta
	err = eliminarDirectorioOArchivo(remove.ruta, particionSuperbloque, archivo)

	if err != nil {
		fmt.Println("Error al eliminar el archivo o directorio.")
		return fmt.Errorf("error al eliminar archivo o directorio: %v", err)
	}

	// Se serializa el superbloque para guardar los cambios
	err = particionSuperbloque.Codificar(archivo, int64(particionMontada.Start))

	if err != nil {
		fmt.Println("Error al serializar el superbloque después de la eliminación de la carpeta o directorio.")
		return fmt.Errorf("error al serializar el superbloque después de la eliminación: %v", err)
	}

	fmt.Printf("Archivo o carpeta '%s' eliminado exitosamente.\n", remove.ruta)
	return nil
}

// Se elije si se elimina un archivo o carpeta en función de la ruta
func eliminarDirectorioOArchivo(ruta string, superbloque *Structs.Superbloque, archivo *os.File) error {
	// Se convierte el path del archivo o carpeta en un array de carpetas
	directoriosPadre, nombreArchivoODirectorio := Herramientas.ObtenerDirectoriosPadreYArchivo(ruta)

	// Se intenta eliminar el archivo
	err := eliminarArchivo(superbloque, archivo, directoriosPadre, nombreArchivoODirectorio)
	if err == nil {
		// Si el archivo se eliminó correctamente, regresar
		return nil
	}

	// Si no es un archivo, intentar eliminarlo como carpeta
	err = eliminarDirectorio(superbloque, archivo, directoriosPadre, nombreArchivoODirectorio)
	if err != nil {
		fmt.Printf("Error al eliminar el directorio '%s'.\n", ruta)
		return fmt.Errorf("error al eliminar el directorio '%s': %v", ruta, err)
	}

	return nil
}

// Se intenta eliminar un archivo dado su ruta
func eliminarArchivo(superbloque *Structs.Superbloque, archivo *os.File, directoriosPadre []string, nombreArchivo string) error {
	// Se busca el inodo del archivo
	_, err := cat.EncontrarInodoArchivo(archivo, superbloque, directoriosPadre, nombreArchivo)
	if err != nil {
		// No se encontró el archivo
		fmt.Printf("El archivo '%s' no fue encontrado.\n", nombreArchivo)
		return fmt.Errorf("archivo '%s' no encontrado: %v", nombreArchivo, err)
	}

	// Se llama a la función que elimina el archivo
	err = superbloque.EliminarArchivo(archivo, directoriosPadre, nombreArchivo)
	if err != nil {
		return fmt.Errorf("error al eliminar el archivo '%s': %v", nombreArchivo, err)
	}

	fmt.Printf("Archivo '%s' eliminado correctamente.\n", nombreArchivo)
	return nil
}

// Se intenta eliminar una carpeta dada su path
func eliminarDirectorio(superbloque *Structs.Superbloque, archivo *os.File, directoriosPadre []string, nombreDirectorio string) error {
	// Se busca el inodo de la carpeta
	_, err := cat.EncontrarInodoDirectorio(archivo, superbloque, directoriosPadre)
	if err != nil {
		fmt.Printf("La carpeta '%s' no fue encontrada.\n", nombreDirectorio)
		return fmt.Errorf("carpeta '%s' no encontrada: %v", nombreDirectorio, err)
	}

	// Se llama a la función que elimina la carpeta
	err = superbloque.EliminarDirectorio(archivo, directoriosPadre, nombreDirectorio)
	if err != nil {
		fmt.Printf("Error al eliminar la carpeta '%s'.\n", nombreDirectorio)
		return fmt.Errorf("error al eliminar la carpeta '%s': %v", nombreDirectorio, err)
	}

	fmt.Printf("Carpeta '%s' eliminada correctamente.\n", nombreDirectorio)
	return nil
}
