package Comandos

import (
	cat "Proyecto/Comandos/AdministracionArchivos"
	globales "Proyecto/Globales"
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FIND struct {
	ruta   string
	nombre string
}

func Find(parametros []string) string {
	fmt.Println("\n======= FIND =======")

	var salida1 = ""

	find := &FIND{}

	//Parametros
	//Obligatorios
	pathInit := false
	contentInit := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")

		tmp := strings.Split(tmp2, "=")

		//Si hay un parametro no válido
		if len(tmp) != 2 {
			salida1 += "FIND Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("FIND Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			find.ruta = tmp[1]

		} else if strings.ToLower(tmp[0]) == "name" {
			contentInit = true
			find.nombre = tmp[1]

		} else {
			salida1 += "FIND Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("FIND Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && contentInit {
		resultados, err := ejecutarFind(find)

		if err != nil {
			fmt.Println("FIND Error: Hubo problemas al encontrar el sistema de archivos.")
			salida1 += "FIND Error: Hubo problemas al encontrar el sistema de archivos.\n"
		} else {
			// Escribir resultados al archivo
			err := escribirResultadosArchivo(resultados)
			if err != nil {
				fmt.Println("FIND Error: No se pudieron guardar los resultados en el archivo.")
				salida1 += "FIND Error: No se pudieron guardar los resultados en el archivo.\n"
			} else {
				contenido, err := os.ReadFile("/home/ubuntu/MIA_2S2025_PROYECTO2_202300722/Calificacion_MIA/Reportes/Find.txt")
				if err != nil {
					fmt.Println("Error al leer archivo:", err)
					salida1 += "FIND ERROR: Hubo problemas al leer el archivo del Find.\n"
					return salida1
				}
				texto := string(contenido)
				salida1 += texto
				fmt.Printf("Se encontraron %d resultados y se guardaron en el archivo.\n", len(resultados))
				salida1 += fmt.Sprintf("Se encontraron %d resultados y se guardaron en el archivo.\n", len(resultados))
			}
		}
	}

	fmt.Println("\n======FIN FIND======")
	return salida1
}

func ejecutarFind(find *FIND) ([]string, error) {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return nil, fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, _, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada.")
		return nil, fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de la partición.")
		return nil, fmt.Errorf("error al abrir el archivo de la partición: %w", err)
	}

	defer archivo.Close()

	// Se verifica si el path es la raíz "/"
	var indiceInodoRoot int32
	if find.ruta == "/" {
		indiceInodoRoot = 0 // El inodo raíz es el inodo 0
	} else {
		// Desglosar el path en directorios y obtener el inodo del directorio inicial
		parentDirs, dirName := Herramientas.ObtenerDirectoriosPadreYArchivo(find.ruta)
		indiceInodoRoot, err = cat.EncontrarInodoArchivo(archivo, particionSuperbloque, parentDirs, dirName)
		if err != nil {
			return nil, fmt.Errorf("error al encontrar el directorio inicial: %v", err)
		}
	}

	// Se convierte el nombre con comodines a expresión regular
	pattern, err := wildcardToRegex(find.nombre)
	if err != nil {
		return nil, fmt.Errorf("error al convertir el patrón de búsqueda: %v", err)
	}

	// Slice para almacenar los resultados
	var resultados []string

	// Se inicia la búsqueda recursiva
	err = searchRecursive(archivo, particionSuperbloque, indiceInodoRoot, pattern, find.ruta, &resultados)
	if err != nil {
		return nil, fmt.Errorf("error durante la búsqueda: %v", err)
	}

	return resultados, nil
}

func searchRecursive(archivo *os.File, superbloque *Structs.Superbloque, indiceInodo int32, pattern *regexp.Regexp, rutaActual string, resultados *[]string) error {
	// Se deserializa el inodo del directorio actual
	inodo := &Structs.Inodo{}
	err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(indiceInodo*superbloque.S_inode_s)))
	if err != nil {
		return fmt.Errorf("error al deserializar el inodo %d: %v", indiceInodo, err)
	}

	// Se verifica si el inodo actual coincide con el patrón
	nombreActual := obtenerNombreDesdeRuta(rutaActual)
	if pattern.MatchString(nombreActual) {
		*resultados = append(*resultados, rutaActual)
	}

	// Se verifica que el inodo sea un directorio
	if inodo.I_type[0] != '0' {
		return nil // Si no es un directorio, no hacemos nada
	}

	// Se itera sobre los bloques del inodo del directorio
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			continue // Se usa continue para seguir con otros bloques
		}

		// Se deserializa el bloque de directorio
		block := &Structs.BloqueFolder{}
		err := block.Decodificar(archivo, int64(superbloque.S_block_start+(indiceBloque*superbloque.S_block_s)))
		if err != nil {
			return fmt.Errorf("error al deserializar el bloque %d: %v", indiceBloque, err)
		}

		// Se iterar sobre los contenidos del bloque
		for _, contenido := range block.B_contenido {
			if contenido.B_inodo == -1 {
				continue // Si no hay un inodo válido, lo saltamos
			}

			nombreContenido := strings.Trim(string(contenido.B_nombre[:]), "\x00 ")

			// Se evitan los enlaces "." y ".." que pueden causar loops infinitos
			if nombreContenido == "." || nombreContenido == ".." {
				continue
			}

			// Se contruye la nueva ruta correctamente
			var nuevaRuta string
			if rutaActual == "/" {
				nuevaRuta = "/" + nombreContenido
			} else {
				nuevaRuta = rutaActual + "/" + nombreContenido
			}

			// Se verifica si el nombre coincide con el patrón
			if pattern.MatchString(nombreContenido) {
				*resultados = append(*resultados, nuevaRuta)
			}

			// Si el contenido es un directorio, hacer una búsqueda recursiva
			nuevoIndiceInodo := contenido.B_inodo
			err = searchRecursive(archivo, superbloque, nuevoIndiceInodo, pattern, nuevaRuta, resultados)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func escribirResultadosArchivo(resultados []string) error {
	// Ruta del archivo de salida
	rutaArchivo := "/home/ubuntu/MIA_2S2025_PROYECTO2_202300722/Calificacion_MIA/Reportes/Find.txt"

	// Crear el directorio si no existe
	dir := filepath.Dir(rutaArchivo)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error al crear directorio: %v", err)
	}

	// Eliminar duplicados
	resultadosUnicos := eliminarDuplicados(resultados)

	// Crear o abrir el archivo
	archivo, err := os.Create(rutaArchivo)
	if err != nil {
		return fmt.Errorf("error al crear archivo: %v", err)
	}
	defer archivo.Close()

	// Escribir los resultados únicos en el archivo
	for _, resultado := range resultadosUnicos {
		_, err := archivo.WriteString(resultado + "\n")
		if err != nil {
			return fmt.Errorf("error al escribir en archivo: %v", err)
		}
	}

	fmt.Printf("Se encontraron %d resultados únicos y se guardaron en: %s\n", len(resultadosUnicos), rutaArchivo)
	return nil
}

// Función para eliminar duplicados
func eliminarDuplicados(resultados []string) []string {
	// Usar un mapa para trackear elementos únicos
	unicos := make(map[string]bool)
	var resultadosUnicos []string

	for _, resultado := range resultados {
		if !unicos[resultado] {
			unicos[resultado] = true
			resultadosUnicos = append(resultadosUnicos, resultado)
		}
	}

	return resultadosUnicos
}

// Las funciones wildcardToRegex y obtenerNombreDesdeRuta se mantienen igual
func wildcardToRegex(pattern string) (*regexp.Regexp, error) {
	// Si el patrón es solo "*", debe coincidir con todo
	if pattern == "*" {
		return regexp.Compile(".*")
	}

	// Escapar caracteres especiales de expresiones regulares
	pattern = regexp.QuoteMeta(pattern)

	// Reemplazar comodines con sus equivalentes en regex
	pattern = strings.ReplaceAll(pattern, "\\?", ".")  // ? se convierte en un solo carácter
	pattern = strings.ReplaceAll(pattern, "\\*", ".*") // * se convierte en uno o más caracteres

	// Se compila la expresión regular (case insensitive)
	return regexp.Compile("(?i)^" + pattern + "$")
}

func obtenerNombreDesdeRuta(ruta string) string {
	// Se limpia la ruta
	ruta = strings.TrimSuffix(ruta, "/")

	if ruta == "/" {
		return "/"
	}

	partes := strings.Split(ruta, "/")
	if len(partes) > 0 {
		return partes[len(partes)-1]
	}

	return ruta
}
