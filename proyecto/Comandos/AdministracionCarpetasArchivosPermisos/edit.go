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

type EDIT struct {
	ruta      string
	contenido string
}

func Edit(parametros []string) string {
	fmt.Println("\n======= EDIT =======")

	var salida1 = ""

	edit := &EDIT{}

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
			salida1 += "EDIT Error: Valor desconocido del parametro " + tmp[0] + ".\n"
			fmt.Println("EDIT Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			break
		}

		//Asignación de path
		if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			edit.ruta = tmp[1]

		} else if strings.ToLower(tmp[0]) == "contenido" {
			contentInit = true
			edit.contenido = tmp[1]

		} else {
			salida1 += "EDIT Error: Parametro desconocido " + tmp[0] + ".\n"
			fmt.Println("EDIT Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && pathInit && contentInit {
		err := ejecutarEdit(edit)

		if err != nil {
			fmt.Println("EDIT Error: Hubo problemas al modificar el contenido del archivo.")
			salida1 += "EDIT Error: Hubo problemas al modificar el contenido del archivo.\n"
		} else {
			fmt.Println("El contenido del archivo se ha modificado con éxito.")
			salida1 += "El contenido del archivo se ha modificado con éxito.\n"
		}
	}

	fmt.Println("\n======FIN EDIT======")
	return salida1
}

func ejecutarEdit(edit *EDIT) error {
	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		fmt.Println("No hay ninguna sesión activa")
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se obtiene mediante el usuario logueado el id de la partición
	idParticion := globales.UsuarioActual.UID

	// Se obtiene la partición montada en función del usuario logueado
	particionSuperbloque, _, rutaPartition, err := globales.ObtenerParticionMontadaSuperbloque(idParticion)

	if err != nil {
		fmt.Println("Error al obtener la partición montada.")
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Se abre el archivo de la partición para trabajar en él
	archivo, err := os.OpenFile(rutaPartition, os.O_RDWR, 0666)

	if err != nil {
		fmt.Println("Error al abrir el archivo de la partición.")
		return fmt.Errorf("error al abrir el archivo de la partición: %w", err)
	}

	defer archivo.Close()

	// Se desglosa la ruta en directorios y el archivo a editar
	directoriosPadre, nombreArchivo := Herramientas.ObtenerDirectoriosPadreYArchivo(edit.ruta)

	// Se busca el inodo del archivo a editar
	indiceInodo, err := cat.EncontrarInodoArchivo(archivo, particionSuperbloque, directoriosPadre, nombreArchivo)

	if err != nil {
		fmt.Println("Error al encontrar el archivo.")
		return fmt.Errorf("error al encontrar el archivo: %v", err)
	}

	// Se lee el contenido del archivo desde el sistema operativo real
	nuevoContenido, err := os.ReadFile(edit.contenido)
	if err != nil {
		fmt.Printf("Error al leer el archivo con el contenido '%s'.\n", edit.contenido)
		return fmt.Errorf("error al leer el archivo de contenido '%s': %v", edit.contenido, err)
	}

	// Se edita el contenido del archivo en el sistema de archivos simulado
	err = editarContenidoArchivo(archivo, particionSuperbloque, indiceInodo, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error al editar el contenido del archivo: %v", err)
	}

	fmt.Printf("Contenido del archivo '%s' editado exitosamente\n", nombreArchivo)

	return nil
}

func editarContenidoArchivo(archivo *os.File, superbloque *Structs.Superbloque, indiceInodo int32, nuevoContenido []byte) error {
	inodo := &Structs.Inodo{}
	err := inodo.Decodificar(archivo, int64(superbloque.S_inode_start+(indiceInodo*superbloque.S_inode_s)))
	if err != nil {
		fmt.Printf("Error al deserializar el inodo '%d'.\n", indiceInodo)
		return fmt.Errorf("error al deserializar el inodo %d: %v", indiceInodo, err)
	}

	if inodo.I_type[0] != '1' {
		fmt.Printf("El inodo '%d' no correspondo a un archivo.\n", indiceInodo)
		return fmt.Errorf("el inodo %d no corresponde a un archivo", indiceInodo)
	}

	// Se limpian los bloques actuales
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque != -1 {
			bloqueArchivo := &Structs.BloqueFile{}
			bloqueArchivo.LimpiarContenido()
			err := bloqueArchivo.Codificar(archivo, int64(superbloque.S_block_start+(indiceBloque*superbloque.S_block_s)))
			if err != nil {
				fmt.Printf("Error al intentar limpiar el bloque '%d'.\n", indiceBloque)
				return fmt.Errorf("error al limpiar el bloque %d: %v", indiceBloque, err)
			}
		}
	}

	// Se divide el nuevo contenido en bloques de 64 bytes
	bloques, err := Structs.DividirContenido(string(nuevoContenido))
	if err != nil {
		fmt.Println("Error al dividir el contenido en bloques.")
		return fmt.Errorf("error al dividir el contenido en bloques: %v", err)
	}

	// Se escriben los nuevos bloques
	conteoBloques := len(bloques)
	for i := 0; i < conteoBloques; i++ {
		if i < len(inodo.I_block) {
			// Si ya existe un bloque asignado, solo escribe en ese bloque
			indiceBloque := inodo.I_block[i]
			if indiceBloque == -1 {
				// Si no hay bloque asignado, asignar un nuevo bloque
				indiceBloque, err = superbloque.AsignarNuevoBloque(archivo, inodo, i)
				if err != nil {
					fmt.Println("Error al signar un nuevo bloque.")
					return fmt.Errorf("error asignando un nuevo bloque: %v", err)
				}
			}

			// Se escribe el contenido en el bloque actual
			err := bloques[i].Codificar(archivo, int64(superbloque.S_block_start+(indiceBloque*superbloque.S_block_s)))
			if err != nil {
				fmt.Printf("Error al escribir el bloque '%d'.\n", indiceBloque)
				return fmt.Errorf("error al escribir el bloque %d: %v", indiceBloque, err)
			}
		} else {
			// Si se excede el espacio de I_block, usar un PointerBlock para manejar más bloques
			indiceBloqueApuntador := inodo.I_block[len(inodo.I_block)-1]
			if indiceBloqueApuntador == -1 {
				// Si no hay un bloque de punteros asignado, asignar uno nuevo
				indiceBloqueApuntador, err = superbloque.AsignarNuevoBloque(archivo, inodo, len(inodo.I_block)-1)
				if err != nil {
					fmt.Println("Error al asignar un nuevo bloque de apuntadores.")
					return fmt.Errorf("error asignando un nuevo bloque de apuntadores: %v", err)
				}
			}

			// Se carga el bloque de apuntadores
			bloqueApuntador := &Structs.BloqueApuntador{}
			err := bloqueApuntador.Decodificar(archivo, int64(superbloque.S_block_start+(indiceBloqueApuntador*superbloque.S_block_s)))
			if err != nil {
				fmt.Println("Error al decodificar el bloque de apuntadores.")
				return fmt.Errorf("error al decodificar el bloque de apuntadores: %v", err)
			}

			// Se busca un puntero libre en el bloque de apuntadores
			indiceLibre, err := bloqueApuntador.EncontrarApuntadorLibre()
			if err != nil {
				fmt.Println("No hay apuntadores libres en el bloque de apuntadores")
				return fmt.Errorf("no hay apuntadores libres en el bloque de apuntadores: %v", err)
			}

			// Se asignar un nuevo bloque para el contenido adicional
			nuevoIndiceBloque, err := superbloque.AsignarNuevoBloque(archivo, inodo, indiceLibre)
			if err != nil {
				fmt.Println("Error al asignar un nuevo bloque.")
				return fmt.Errorf("error asignando un nuevo bloque: %v", err)
			}

			// Se actualiza el puntero en el bloque de apuntadores
			err = bloqueApuntador.ColocarApuntador(indiceLibre, int64(nuevoIndiceBloque))
			if err != nil {
				fmt.Println("Error al actualizar el bloque de apuntadores.")
				return fmt.Errorf("error actualizando el bloque de apuntadores: %v", err)
			}

			// Se guarda el bloque de apuntadores actualizado
			err = bloqueApuntador.Codificar(archivo, int64(superbloque.S_block_start+(indiceBloqueApuntador*superbloque.S_block_s)))
			if err != nil {
				fmt.Println("Error al guardar el bloque de apuntadores.")
				return fmt.Errorf("error al guardar el bloque de apuntadores: %v", err)
			}

			// Se escribe el contenido en el nuevo bloque
			err = bloques[i].Codificar(archivo, int64(superbloque.S_block_start+(nuevoIndiceBloque*superbloque.S_block_s)))
			if err != nil {
				return fmt.Errorf("error al escribir el nuevo bloque %d: %v", nuevoIndiceBloque, err)
			}
		}
	}

	// Se actualiza el tamaño del archivo en el inodo
	inodo.I_s = int32(len(nuevoContenido))
	err = inodo.Codificar(archivo, int64(superbloque.S_inode_start+(indiceInodo*superbloque.S_inode_s)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo %d: %v", indiceInodo, err)
	}

	return nil
}
