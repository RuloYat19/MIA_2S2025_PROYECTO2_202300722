package Comandos

import (
	"fmt"
	"os"

	"strings"

	cat "Proyecto/Comandos/AdministracionArchivos"
	globales "Proyecto/Globales"
	utilidades "Proyecto/Herramientas"
	estructuras "Proyecto/Structs"
)

type DirectoryTree struct {
	Name     string           `json:"name"`
	Children []*DirectoryTree `json:"children,omitempty"`
	IsDir    bool             `json:"isDir"`
}

type DirectoryTreeService struct {
	partitionSuperblock *estructuras.Superbloque
	partitionPath       string
	file                *os.File
}

func NewDirectoryTreeService() (*DirectoryTreeService, error) {
	if !globales.HaIniciadoSesion() {
		fmt.Println("No se ha iniciado sesión.")
		return nil, fmt.Errorf("operación denegada: no se ha iniciado sesión")
	}
	if err := globales.ValidarAcceso(globales.UsuarioActual.UID); err != nil {
		fmt.Println("Los permisos son insuficientes para poder acceder a la partición.")
		return nil, fmt.Errorf("permisos insuficientes para acceder a la partición: %w", err)
	}
	idPartition := globales.UsuarioActual.UID
	partitionSuperblock, _, partitionPath, err := globales.ObtenerParticionMontadaSuperbloque(idPartition)
	if err != nil {
		fmt.Printf("No ha sido posible obtener la partición montada: '%s'", idPartition)
		return nil, fmt.Errorf("imposible obtener la partición montada (ID: %s): %w", idPartition, err)
	}
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Hubo problemas al abrir el archivo de la partición: '%s'", idPartition)
		return nil, fmt.Errorf("fallo al abrir el archivo de la partición en '%s': %w", partitionPath, err)
	}
	return &DirectoryTreeService{
		partitionSuperblock: partitionSuperblock,
		partitionPath:       partitionPath,
		file:                file,
	}, nil
}

func (dts *DirectoryTreeService) Close() {
	dts.file.Close()
}

func (dts *DirectoryTreeService) GetDirectoryTree(path string) (*DirectoryTree, error) {
	var rootInodeIndex int32
	var err error

	if path == "/" {
		rootInodeIndex = 0
	} else {
		parentDirs, dirName := utilidades.ObtenerDirectoriosPadreYArchivo(path)
		rootInodeIndex, err = cat.EncontrarInodoArchivo(dts.file, dts.partitionSuperblock, parentDirs, dirName)
		if err != nil {
			return nil, fmt.Errorf("imposible localizar el directorio inicial '%s': %w", path, err)
		}
	}
	tree, err := dts.buildDirectoryTree(rootInodeIndex, path)
	if err != nil {
		return nil, fmt.Errorf("fallo al construir el árbol de directorios para '%s': %w", path, err)
	}

	return tree, nil
}

func (dts *DirectoryTreeService) buildDirectoryTree(inodeIndex int32, currentPath string) (*DirectoryTree, error) {
	inode := &estructuras.Inodo{}
	offset := int64(dts.partitionSuperblock.S_inode_start) + int64(inodeIndex*dts.partitionSuperblock.S_inode_s)
	err := inode.Decodificar(dts.file, offset)
	if err != nil {
		return nil, fmt.Errorf("fallo al deserializar el inodo %d (offset %d) para '%s': %w", inodeIndex, offset, currentPath, err)
	}
	var currentName string
	if currentPath == "/" {
		currentName = "/"
	} else {
		pathSegments := strings.Split(strings.Trim(currentPath, "/"), "/")
		currentName = pathSegments[len(pathSegments)-1]
	}

	tree := &DirectoryTree{
		Name:     currentName,
		IsDir:    inode.I_type[0] == '0',
		Children: []*DirectoryTree{}, // Inicializar siempre como slice vacío
	}

	if !tree.IsDir {
		return tree, nil
	}

	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		block := &estructuras.BloqueFolder{}
		blockOffset := int64(dts.partitionSuperblock.S_block_start) + int64(blockIndex*dts.partitionSuperblock.S_block_s)
		err := block.Decodificar(dts.file, blockOffset)
		if err != nil {
			return nil, fmt.Errorf("fallo al deserializar el bloque %d (offset %d): %w", blockIndex, blockOffset, err)
		}
		for _, content := range block.B_contenido {
			if content.B_inodo == -1 {
				continue
			}

			contentName := strings.Trim(string(content.B_nombre[:]), "\x00 ")
			if contentName == "." || contentName == ".." {
				continue
			}
			var childPath string
			if currentPath == "/" {
				childPath = "/" + contentName
			} else {
				childPath = currentPath + "/" + contentName
			}

			childNode, err := dts.buildDirectoryTree(content.B_inodo, childPath)
			if err != nil {
				// Log el error pero continúa con otros hijos
				fmt.Printf("Error building child '%s': %v\n", childPath, err)
				continue
			}
			tree.Children = append(tree.Children, childNode)
		}
	}

	return tree, nil
}
