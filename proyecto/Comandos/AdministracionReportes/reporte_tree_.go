package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ReporteTree(superbloque *Structs.Superbloque, rutaDisco string, ruta string) error {
	// Se crean carpetas padre si no existen
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

	// Se obtiene el nombre base del archivo sin extensión
	dotFileName, outputImage := Herramientas.ObtenerNombreArchivos(ruta)

	// Inicio del Dot
	dotContent := iniciarDotTreeGraph()

	// Se genera el grafo del árbol EXT2
	dotContent, err = generarGrafoTree(dotContent, superbloque, archivo)
	if err != nil {
		return err
	}

	dotContent += "}"

	// Se crea el archivo DOT
	err = escribirDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Se ejecuta el Graphviz para generar la imagen
	err = generarImagenTree(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen del árbol EXT2 generada:", outputImage)
	return nil
}

// Se inicializa el contenido básico del archivo DOT para el árbol EXT2
func iniciarDotTreeGraph() string {
	return `digraph EXT2Tree {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
		edge [fontname="Helvetica,Arial,sans-serif", color="#1976D2", arrowsize=0.8];
		rankdir=TB;
		bgcolor="#FAFAFA";
		node [shape=plaintext];
		inodeHeaderColor="#388E3C";
		blockHeaderColor="#FBC02D";
		cellBackgroundColor="#FFFDE7";
		cellBorderColor="#EEEEEE";
		textColor="#263238";
	`
}

// Se recorre los inodos y bloques para construir el grafo del sistema EXT2
func generarGrafoTree(dotContent string, superbloque *Structs.Superbloque, archivo *os.File) (string, error) {
	conexiones := ""
	inodos := make(map[int32]*Structs.Inodo)
	// Se decodifican todos los inodos válidos
	for i := int32(0); i < superbloque.S_inodes_count; i++ {
		inodo := &Structs.Inodo{}
		err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(i*superbloque.S_inode_s)))
		if err == nil && inodo.I_uid != -1 && inodo.I_uid != 0 {
			inodos[i] = inodo
			dotContent += generarTablaInodoTree(i, inodo)
		}
	}
	// Se recorren los inodos para generar nodos de bloques y conexiones
	for i, inodo := range inodos {
		for _, block := range inodo.I_block {
			if block != -1 {
				dotContent += generarNodoBloqueTree(block, inodo, superbloque, archivo)
				conexiones += fmt.Sprintf("inodo%d -> block%d [color=\"#1976D2\"]\n", i, block)
				// Si es carpeta, conectar con los inodos hijos
				if inodo.I_type[0] == '0' {
					bloqueOffset := int64(superbloque.S_block_start + (block * superbloque.S_block_s))
					bloqueFolder := &Structs.BloqueFolder{}
					err := bloqueFolder.Decodificar(archivo, bloqueOffset)
					if err == nil {
						for j, content := range bloqueFolder.B_contenido {
							// Evitar . y ..
							if j > 1 && content.B_inodo != -1 {
								// Conexión bloque carpeta con inodo hijo
								conexiones += fmt.Sprintf("block%d -> inodo%d [color=\"#388E3C\"]\n", block, content.B_inodo)
							}
						}
					}
				}
			}
		}
	}
	// Se realizan las conexiones entre inodos padres e hijos (carpetas)
	for i, inodo := range inodos {
		if inodo.I_type[0] == '0' {
			for _, block := range inodo.I_block {
				if block != -1 {
					bloqueOffset := int64(superbloque.S_block_start + (block * superbloque.S_block_s))
					bloqueFolder := &Structs.BloqueFolder{}
					err := bloqueFolder.Decodificar(archivo, bloqueOffset)
					if err == nil {
						for j, content := range bloqueFolder.B_contenido {
							// Se evitan . y ..
							if j > 1 && content.B_inodo != -1 {
								// Conexión inodo padre con inodo hijo SIN flecha
								conexiones += fmt.Sprintf("inodo%d -> inodo%d [style=dashed, color=\"#FBC02D\", arrowhead=none]\n", i, content.B_inodo)
							}
						}
					}
				}
			}
		}
	}
	dotContent += conexiones
	return dotContent, nil
}

// Se genera la tabla DOT para un inodo
func generarTablaInodoTree(indiceInodo int32, inodo *Structs.Inodo) string {
	atime := time.Unix(int64(inodo.I_atime), 0).Format(time.RFC3339)
	ctime := time.Unix(int64(inodo.I_ctime), 0).Format(time.RFC3339)
	mtime := time.Unix(int64(inodo.I_mtime), 0).Format(time.RFC3339)
	table := fmt.Sprintf("inodo%d [label=<\n        <table border='0' cellborder='1' cellspacing='0' cellpadding='4' bgcolor='#FFFDE7' style='rounded'>\n            <tr><td colspan='2' bgcolor='#388E3C' align='center'><b>INODO %d</b></td></tr>\n            <tr><td><b>i_uid</b></td><td>%d</td></tr>\n            <tr><td><b>i_gid</b></td><td>%d</td></tr>\n            <tr><td><b>i_size</b></td><td>%d</td></tr>\n            <tr><td><b>i_atime</b></td><td>%s</td></tr>\n            <tr><td><b>i_ctime</b></td><td>%s</td></tr>\n            <tr><td><b>i_mtime</b></td><td>%s</td></tr>\n            <tr><td><b>i_type</b></td><td>%c</td></tr>\n            <tr><td><b>i_perm</b></td><td>%s</td></tr>\n        </table>>];\n",
		indiceInodo, indiceInodo, inodo.I_uid, inodo.I_gid, inodo.I_s, atime, ctime, mtime, rune(inodo.I_type[0]), string(inodo.I_perm[:]))
	return table
}

// Se genera el nodo DOT para un bloque (carpeta o archivo)
func generarNodoBloqueTree(indiceBloque int32, inodo *Structs.Inodo, superbloque *Structs.Superbloque, archivo *os.File) string {
	bloqueOffset := int64(superbloque.S_block_start + (indiceBloque * superbloque.S_block_s))
	var dot string
	if inodo.I_type[0] == '0' { // Bloque de carpeta
		bloqueFolder := &Structs.BloqueFolder{}
		err := bloqueFolder.Decodificar(archivo, bloqueOffset)
		if err != nil {
			return "" // Si hay error, no se agrega el nodo
		}
		label := fmt.Sprintf("BLOQUE DE CARPETA %d", indiceBloque)
		for i, content := range bloqueFolder.B_contenido {
			name := limpiarNombreBloqueTree(content.B_nombre)
			if content.B_inodo != -1 {
				label += fmt.Sprintf("\\nContenido %d: %s (Inodo %d)", i+1, name, content.B_inodo)
			} else {
				label += fmt.Sprintf("\\nContenido %d: %s (Sin inodo)", i+1, name)
			}
		}
		// Se usan doble backslash (\) para saltos de línea en DOT
		label = reemplazarNuevasLineas(label)
		dot += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"]\n", indiceBloque, label)
	} else if inodo.I_type[0] == '1' { // Bloque de archivo
		bloqueFile := &Structs.BloqueFile{}
		err := bloqueFile.Decodificar(archivo, bloqueOffset)
		if err != nil {
			return ""
		}
		content := limpiarContenidoBloqueTree(bloqueFile.ObtenerContenido())
		if len(content) > 0 {
			// Se usan doble backslash (\) para saltos de línea en DOT
			content = reemplazarNuevasLineas(content)
			label := fmt.Sprintf("BLOQUE DE ARCHIVO %d\\n%s", indiceBloque, content)
			dot += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"]\n", indiceBloque, label)
		}
	}
	return dot
}

// Se reemplaza saltos de línea por doble backslash para DOT
func reemplazarNuevasLineas(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

// Se elimina caracteres nulos del nombre
func limpiarNombreBloqueTree(nameArray [12]byte) string {
	name := string(nameArray[:])
	for i := range name {
		if name[i] == '\x00' {
			return name[:i]
		}
	}
	return name
}

// Se limpia el contenido del bloque
func limpiarContenidoBloqueTree(content string) string {
	return content
}

// Se genera una imagen a partir del archivo DOT usando Graphviz
func generarImagenTree(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}
	return nil
}
