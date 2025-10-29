package Comandos

import (
	AA "Proyecto/Comandos/AdministracionArchivos"
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type MKFILE struct {
	ruta    string
	r       bool
	tamanio int
	cont    string
}

func Mkfile(parametros []string) string {
	fmt.Println("\n======= MKFILE =======")

	var salida1 = ""

	mkfile := &MKFILE{}

	//Opcionales
	mkfile.r = false
	mkfile.tamanio = 0
	mkfile.cont = ""

	//Otras variables
	paramC := true
	rutaInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if tmp[0] == "r" {

		} else if len(tmp) != 2 {
			salida1 += "MKFILE Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("MKFILE Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			rutaInit = true
			mkfile.ruta = tmp[1]

			//Asignación de r
		} else if strings.ToLower(tmp[0]) == "r" {
			mkfile.r = true

			//Asignación de size
		} else if strings.ToLower(tmp[0]) == "size" {
			tamanioParametro := tmp[1]

			i, err := strconv.Atoi(tamanioParametro)
			if err != nil {
				salida1 += "MKFILE Error: Hubo problemas asignando el valor de -size.\n"
				fmt.Println("MKFILE Error: Hubo problemas asignando el valor de -size ", err)
				return salida1
			}

			mkfile.tamanio = i

			// Asignación de cont
		} else if strings.ToLower(tmp[0]) == "cont" {
			mkfile.cont = tmp[1]
		} else {
			salida1 += "MKFILE Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("MKFILE Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if mkfile.tamanio <= 0 {
		salida1 += "MKFILE Error: No se puede crear el archivo porque el tamaño indicado es menor o igual a 0.\n"
		fmt.Println("MKFILE Error: No se puede crear el archivo porque el tamaño indicado es menor o igual a 0.")
		return salida1
	}

	// Si no hubo errores y si están los parámetros obligatorios
	if paramC && rutaInit {
		err := ejecutarMkfile(mkfile)

		if err != nil {
			salida1 += "MKFILE Error: Hubo problemas al crear el archivo " + mkfile.ruta + ".\n"
		} else {
			salida1 += "El archivo " + mkfile.ruta + " fue creado exitosamente.\n"
		}
	}

	fmt.Println("\n======FIN MKFILE======")
	return salida1
}

func ejecutarMkfile(mkfile *MKFILE) error {
	// Verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, particionMontada, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Generar el contenido del archivo si no se proporcionó
	var contenidoArchivo string
	if mkfile.cont == "" {
		contenidoArchivo = generarContenido(mkfile.tamanio)
	} else {
		contenido, err := os.ReadFile(mkfile.cont)
		if err != nil {
			return fmt.Errorf("error al leer el archivo: %v", err)
		}
		contenidoArchivo = string(contenido)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}

	defer archivo.Close()

	//Se indican los datos que tiene el mkfile
	fmt.Println("Los datos del mkfile son: ", mkfile.ruta, mkfile.r, mkfile.tamanio, mkfile.cont)

	// Se obtienen los directorios y el nombre del archivo
	dirRuta, _ := ObtenerDirectorioYArchivo(mkfile.ruta)

	// Se verifica solo la existencia del directorio (sin incluir el archivo)
	fmt.Printf("Verificando la existencia del directorio: %s\n", dirRuta)

	existe, _, err := AA.ComprobacionDirectorio(particionSuperbloque, archivo, 0, dirRuta) // Usamos el inodo raíz (0) para empezar la búsqueda

	if err != nil {
		return fmt.Errorf("error al verificar directorio: %w", err)
	}

	// Si -r está habilitado y el directorio no existe, creamos los directorios intermedios
	if mkfile.r && !existe {
		err = CrearDirectorio(dirRuta, mkfile.r, particionSuperbloque, archivo, particionMontada)
		if err != nil {
			return fmt.Errorf("error al crear directorios intermedios: %w", err)
		}
	}

	// Se crea el archivo usando el archivo de partición abierto
	err = CrearArchivo(mkfile.ruta, mkfile.tamanio, contenidoArchivo, particionSuperbloque, archivo, particionMontada, mkfile.r)

	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	return nil
}

// Se genera una cadena de números del 0 al 9 hasta cumplir el tamaño ingresado del mkfile
func generarContenido(size int) string {
	content := ""
	for len(content) < size {
		content += "0123456789"
	}
	return content[:size] // Lo recorta al tamaño exacto
}

// Se obtiene la carpeta (directorio) y el nombre del archivo (file)
func ObtenerDirectorioYArchivo(ruta string) (string, string) {
	// Se obtiene la carpeta donde se creará el archivo
	dir := filepath.Dir(ruta)
	// Se obtiene el nombre del archivo
	archivo := filepath.Base(ruta)
	return dir, archivo
}

// Se usa ahora el archivo de partición ya abierto
func CrearArchivo(rutaArchivo string, tamanio int, contenido string, superbloque *Structs.Superbloque, archivo *os.File, particionMontada *Structs.Particion, r bool) error {
	fmt.Printf("Creando archivo en la ruta: %s\n", rutaArchivo)

	// Se obtiene los directorios padres y el destino
	padreDirs, _ := Herramientas.ObtenerDirectorios(rutaArchivo)

	// Se obtiene contenido por chunks
	chunks := Herramientas.DividirCadenaEnChunks(contenido)
	fmt.Printf("Contenido generado: %v\n", chunks)

	// Se iteran las carpetas para crearse pero se validan antes
	for _, parentDir := range padreDirs {
		err := superbloque.CrearArchivo(archivo, padreDirs, parentDir, tamanio, chunks, r)
		if err != nil {
			fmt.Println("Error al crear las carpetas.")
			return fmt.Errorf("error al crear las carpetas")
		}
	}

	// Se serializa el superbloque
	err := superbloque.Codificar(archivo, int64(particionMontada.Start))

	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	// Se muestran las estructuras en consola
	fmt.Println("\nInodos:")
	superbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques de datos:")
	superbloque.ImprimirBloques(archivo.Name())

	return nil
}
