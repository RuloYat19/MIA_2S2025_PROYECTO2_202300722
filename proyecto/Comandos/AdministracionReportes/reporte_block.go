package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"html"
	"os"
	"os/exec"
	"strings"
)

func ReporteBloque(superbloque *Structs.Superbloque, rutaDisco string, ruta string) error {
	// Se crean las carpetas padre si no existen
	err := Herramientas.CrearPadreDirs(ruta)

	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Se abre el archivo de disco
	archivo, err := os.Open(rutaDisco)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}

	defer archivo.Close()

	// Se obtiene el nombre base del archivo sin la extensión
	dotFileName, outputImage := Herramientas.ObtenerNombreArchivos(ruta)

	// Inicio del Dot
	dotContent := iniciarDotGraph()

	// Se generan los bloques y sus conexiones
	dotContent, conexiones, err := generarGrafoBloque(dotContent, superbloque, archivo)

	if err != nil {
		return err
	}

	dotContent += conexiones // Se agregan conexiones fuera de las definiciones de nodos
	dotContent += "}"        // Fin del Dot

	// Se crea el archivo DOT
	err = escribirDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Se ejecuta el Graphviz para generar la imagen
	err = GenerarImagenBloque(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los bloques generada:", outputImage)
	return nil
}

// Se genera una imagen a partir del archivo DOT usando Graphviz
func GenerarImagenBloque(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}

// generarGrafoBloque genera el contenido del grafo de bloques en formato DOT
func generarGrafoBloque(dotContent string, superbloque *Structs.Superbloque, archivo *os.File) (string, string, error) {
	visitedBlocks := make(map[int32]bool)
	var conexiones string

	for i := int32(0); i < superbloque.S_inodes_count; i++ {
		inodo := &Structs.Inodo{}
		err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(i*superbloque.S_inode_s)))
		if err != nil {
			return "", "", fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		if inodo.I_uid == -1 || inodo.I_uid == 0 {
			continue
		}

		// Se recorren los bloques asociados al inodo
		for _, block := range inodo.I_block {
			if block != -1 {
				if !visitedBlocks[block] {
					dotContent, conexiones, err = generarEtiquetaBloque(dotContent, conexiones, block, inodo, superbloque, archivo, visitedBlocks)
					if err != nil {
						return "", "", err
					}
					visitedBlocks[block] = true
				}
			}
		}
	}
	return dotContent, conexiones, nil
}

func generarEtiquetaBloque(dotContent, conexiones string, indiceBloque int32, inodo *Structs.Inodo, superbloque *Structs.Superbloque, archivo *os.File, visitedBlocks map[int32]bool) (string, string, error) {
	bloqueOffset := int64(superbloque.S_block_start + (indiceBloque * superbloque.S_block_s))

	if inodo.I_type[0] == '0' { // Bloque de carpeta
		bloqueFolder := &Structs.BloqueFolder{}
		err := bloqueFolder.Decodificar(archivo, bloqueOffset)
		if err != nil {
			return "", "", fmt.Errorf("error al decodificar bloque de carpeta %d: %w", indiceBloque, err)
		}

		// Se genera la etiqueta del bloque de carpeta con bordes
		label := fmt.Sprintf("BLOQUE DE CARPETA %d", indiceBloque)
		hasValidConnections := false // Se verifica si el bloque tiene conexiones válidas

		for i, content := range bloqueFolder.B_contenido {
			name := limpiarNombreBloque(content.B_nombre)

			// Se usa el html.EscapeString para evitar que caracteres especiales rompan el DOT
			name = html.EscapeString(name)

			// Se evitan las conexiones internas (.) y (..)
			if content.B_inodo != -1 && !(i == 0 || i == 1) {
				// Se añade conexiones a otros inodos
				label += fmt.Sprintf("\\nContenido %d: %s (Inodo %d)", i+1, name, content.B_inodo)
				// Se conecta solo si es un bloque de archivo válido
				if content.B_inodo != indiceBloque {
					conexiones += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", indiceBloque, content.B_inodo)
				}
				hasValidConnections = true
			} else {
				if i > 1 { // Se evitan mostrar las referencias internas en la etiqueta
					label += fmt.Sprintf("\\nContenido %d: %s (Inodo no asignado)", i+1, name)
				}
			}
		}

		// Solo se agrega al contenido DOT si el bloque tiene conexiones válidas o contenido significativo
		if hasValidConnections {
			dotContent += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"];\n", indiceBloque, label)
		}

	} else if inodo.I_type[0] == '1' { // Bloque de archivo
		bloqueFile := &Structs.BloqueFile{}
		err := bloqueFile.Decodificar(archivo, bloqueOffset)
		if err != nil {
			return "", "", fmt.Errorf("error al decodificar bloque de archivo %d: %w", indiceBloque, err)
		}

		content := limpiarContenidoBloque(bloqueFile.ObtenerContenido())

		// Solo se genera la tabla si hay contenido
		if len(strings.TrimSpace(content)) > 0 {
			label := fmt.Sprintf("BLOQUE DE ARCHIVO %d\\n%s", indiceBloque, content)
			dotContent += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"];\n", indiceBloque, label)

			// Se conecta con el siguiente bloque de archivo si existe
			nextBlock := encontrarSiguienteBloqueValido(inodo, indiceBloque)
			if nextBlock != -1 {
				conexiones += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", indiceBloque, nextBlock)
			}
		}
	}

	// Se agrega la referencia al bloque padre si existe
	parentBlock := encontrarBloquePadre(inodo, indiceBloque)
	if parentBlock != -1 {
		conexiones += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", parentBlock, indiceBloque)
	}

	return dotContent, conexiones, nil
}

// Se busca el bloque padre del bloque actual
func encontrarBloquePadre(inodo *Structs.Inodo, bloqueActual int32) int32 {
	for i := 0; i < len(inodo.I_block); i++ {
		if inodo.I_block[i] == bloqueActual && i > 0 {
			return inodo.I_block[i-1]
		}
	}
	return -1 // No hay bloque padre
}

// Se busca el siguiente bloque válido en el array I_block del inodo
func encontrarSiguienteBloqueValido(inodo *Structs.Inodo, bloqueActual int32) int32 {
	for i := 0; i < len(inodo.I_block); i++ {
		if inodo.I_block[i] == bloqueActual {
			for j := i + 1; j < len(inodo.I_block); j++ {
				if inodo.I_block[j] != -1 {
					return inodo.I_block[j]
				}
			}
		}
	}
	return -1 // No hay más bloques válidos
}

// Se limpia el nombre del bloque, eliminando los caracteres nulos
func limpiarNombreBloque(nameArray [12]byte) string {
	return strings.TrimRight(string(nameArray[:]), "\x00")
}

// Se limpia el contenido del bloque
func limpiarContenidoBloque(content string) string {
	return strings.ReplaceAll(content, "\n", "\\n")
}
