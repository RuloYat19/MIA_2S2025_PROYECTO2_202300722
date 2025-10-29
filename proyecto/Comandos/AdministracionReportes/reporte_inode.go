package Comandos

import (
	"Proyecto/Herramientas"
	"Proyecto/Structs"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func ReporteInodo(superbloque *Structs.Superbloque, rutaDisco string, ruta string) error {
	// Se crea las carpetas padre si no existen
	err := Herramientas.CrearPadreDirs(ruta)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Abre el archivo de disco
	archivo, err := os.Open(rutaDisco)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}

	defer archivo.Close()

	// Se obtiene el nombre base del archivo sin la extensión
	dotFileName, outputImage := Herramientas.ObtenerNombreArchivos(ruta)

	// Inicio del Dot
	dotContent := iniciarDotGraph()

	// Si no hay inodos, devolver un error
	if superbloque.S_inodes_count == 0 {
		return fmt.Errorf("no hay inodos en el sistema")
	}

	// Se genera los inodos y sus conexiones
	dotContent, err = generarGrafoInodo(dotContent, superbloque, archivo)
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
	err = generarImagenInodo(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los inodos generada:", outputImage)
	return nil
}

// Se inicializa el contenido básico del archivo DOT
func iniciarDotGraph() string {
	return `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
		edge [fontname="Helvetica,Arial,sans-serif", color="#FF7043", arrowsize=0.8];
		rankdir=LR;
		bgcolor="#FAFAFA";
		node [shape=plaintext];
		inodeHeaderColor="#4CAF50"; 
		blockHeaderColor="#FF9800"; 
		cellBackgroundColor="#FFFDE7";
		cellBorderColor="#EEEEEE";
		textColor="#263238";
	`
}

// Se genera el contenido del grafo de inodos en formato DOT
func generarGrafoInodo(dotContent string, superbloque *Structs.Superbloque, archivo *os.File) (string, error) {
	for i := int32(0); i < superbloque.S_inodes_count; i++ {
		inodo := &Structs.Inodo{}
		err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(i*superbloque.S_inode_s)))
		if err != nil {
			return "", fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		// Se verifica si el inodo está en uso
		if inodo.I_uid == -1 || inodo.I_uid == 0 {
			continue
		}

		// Se genera la tabla del inodo
		dotContent += generarTablaInodo(i, inodo)

		// Se hace la conexión entre inodos
		if i < superbloque.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inodo%d -> inodo%d [color=\"#FF7043\"];\n", i, i+1)
		}
	}
	padres := make(map[int32]int32)
	inodos := make(map[int32]*Structs.Inodo)
	// Decodificar todos los inodos válidos
	for i := int32(0); i < superbloque.S_inodes_count; i++ {
		inodo := &Structs.Inodo{}
		err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(i*superbloque.S_inode_s)))
		if err == nil && inodo.I_uid != -1 && inodo.I_uid != 0 {
			inodos[i] = inodo
			dotContent += generarTablaInodo(i, inodo)
		}
	}
	// Buscar padres
	for hijo := range inodos {
		for padre := range inodos {
			if hijo == padre {
				continue
			}
			if inodos[padre].I_type[0] == '0' {
				for _, block := range inodos[padre].I_block {
					if block != -1 {
						bloqueOffset := int64(superbloque.S_block_start + (block * superbloque.S_block_s))
						bloqueFolder := &Structs.BloqueFolder{}
						err := bloqueFolder.Decodificar(archivo, bloqueOffset)
						if err == nil {
							for k, content := range bloqueFolder.B_contenido {
								if k > 1 && content.B_inodo == hijo {
									padres[hijo] = padre
								}
							}
						}
					}
				}
			}
		}
	}
	// Dibujar flechas padre -> hijo
	for hijo, padre := range padres {
		dotContent += fmt.Sprintf("inodo%d -> inodo%d [color=\"#FF7043\", arrowhead=normal];\n", padre, hijo)
	}
	return dotContent, nil
}

// Se genera la tabla con los atributos y bloques del inodo en formato DOT
func generarTablaInodo(indiceInodo int32, inodo *Structs.Inodo) string {
	// Se convierte tiempos a string
	atime := time.Unix(int64(inodo.I_atime), 0).Format(time.RFC3339)
	ctime := time.Unix(int64(inodo.I_ctime), 0).Format(time.RFC3339)
	mtime := time.Unix(int64(inodo.I_mtime), 0).Format(time.RFC3339)

	// Se genera la tabla del inodo
	table := fmt.Sprintf(`inodo%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#FFFDE7" style="rounded">
			<tr><td colspan="2" bgcolor="#4CAF50" align="center"><b>INODO %d</b></td></tr>
			<tr><td><b>i_uid</b></td><td>%d</td></tr>
			<tr><td><b>i_gid</b></td><td>%d</td></tr>
			<tr><td><b>i_size</b></td><td>%d</td></tr>
			<tr><td><b>i_atime</b></td><td>%s</td></tr>
			<tr><td><b>i_ctime</b></td><td>%s</td></tr>
			<tr><td><b>i_mtime</b></td><td>%s</td></tr>
			<tr><td><b>i_type</b></td><td>%c</td></tr>
			<tr><td><b>i_perm</b></td><td>%s</td></tr>
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUES DIRECTOS</b></td></tr>
	`, indiceInodo, indiceInodo, inodo.I_uid, inodo.I_gid, inodo.I_s, atime, ctime, mtime, rune(inodo.I_type[0]), string(inodo.I_perm[:]))

	// Se agregan los bloques directos
	for j, block := range inodo.I_block[:12] {
		if block != -1 { // Bloques usados
			table += fmt.Sprintf("<tr><td><b>%d</b></td><td>%d</td></tr>", j+1, block)
		}
	}

	// Se agregan bloques indirectos (si existen)
	table += generarBloquesIndirectos(inodo)

	table += "</table>>];"
	return table
}

// Se agregan los bloques indirectos al inodo
func generarBloquesIndirectos(inodo *Structs.Inodo) string {
	result := ""
	// Agregar bloque indirecto simple
	if inodo.I_block[12] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO SIMPLE</b></td></tr>
			<tr><td><b>13</b></td><td>%d</td></tr>
		`, inodo.I_block[12])
	}

	// Agregar bloque indirecto doble
	if inodo.I_block[13] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO DOBLE</b></td></tr>
			<tr><td><b>14</b></td><td>%d</td></tr>
		`, inodo.I_block[13])
	}

	// Agregar bloque indirecto triple
	if inodo.I_block[14] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO TRIPLE</b></td></tr>
			<tr><td><b>15</b></td><td>%d</td></tr>
		`, inodo.I_block[14])
	}

	return result
}

// Se escribe el contenido DOT en un archivo
func escribirDotFile(dotFileName string, dotContent string) error {
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo DOT: %v", err)
	}
	defer dotFile.Close()

	// Se escribe el contenido DOT en el archivo
	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo DOT: %v", err)
	}

	return nil
}

// Se genera una imagen a partir del archivo DOT usando Graphviz
func generarImagenInodo(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}
