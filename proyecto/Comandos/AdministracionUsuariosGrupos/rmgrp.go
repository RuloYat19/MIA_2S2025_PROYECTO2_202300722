package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type RMGRP struct {
	Nombre string
}

func Rmgrp(parametros []string) string {
	fmt.Println("\n======= RMGRP =======")

	var salida1 = ""

	// Inicializar la estructura RMGRP
	rmgrp := &RMGRP{}

	// Otras variables
	paramC := true
	nombreInit := false

	//Asignar los parámetros a la estructura
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "RMGRP Error: Valor desconocido del parámetro " + tmp[0] + ".\n"
			fmt.Println("RMGRP Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "name" {
			nombreInit = true
			rmgrp.Nombre = tmp[1]
		} else {
			salida1 += "RMGRP Error: Parámetro desconocido: " + tmp[0] + ".\n"
			fmt.Println("RMGRP Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Se elimina el grupo con los datos asignados
	if paramC && nombreInit {
		err := eliminarGrupo(rmgrp)
		if err != nil {
			salida1 += "RMGRP Error: Hubo problemas eliminando el grupo " + rmgrp.Nombre + ".\n"
			return salida1
		} else {
			salida1 += "El grupo " + rmgrp.Nombre + " fue eliminado correctamente.\n"
		}
	}
	fmt.Println("\n======FIN RMGRP======")
	return salida1
}

func eliminarGrupo(rmgrp *RMGRP) error {
	fmt.Println(rmgrp.Nombre)

	//Verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	//Verifica si el usuario es el root
	if globales.UsuarioActual.Nombre != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Verifica que la partición esté montada en el sistema
	_, ruta, err := globales.ObtenerParticionMontadas(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Abre el archivo donde se encuentra la partición montada
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0755)

	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de la partición: %v", err)
	}

	defer archivo.Close()

	// Se carga el Superblock y la Partición
	mbr, superbloque, _, err := globales.ObtenerParticionMontadaRep(globales.UsuarioActual.UID) //Id de la particion del usuario actual

	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Obtiene la partición asociada al id
	particion, err := mbr.ObtenerParticionPorID(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se pudo obtener la partición: %v", err)
	}

	// Lee el inodo de users.txt
	var usuariosInodo Structs.Inodo

	// Calcula el offset del inodo de users.txt que esta en el inodo 1
	inodoOffset := int64(superbloque.S_inode_start + int32(binary.Size(usuariosInodo)))

	// Decodifica el inodo de users.txt
	err = usuariosInodo.Decodificar(archivo, inodoOffset)
	//usuariosInodo.ActualizarAtime() // Actualizar la última fecha de acceso

	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Verifica si el grupo ya existe en users.txt
	_, err = globales.BuscarEnElArchivoDeUsuario(archivo, superbloque, &usuariosInodo, rmgrp.Nombre, "G")

	if err != nil {
		return fmt.Errorf("el grupo '%s' no existe", rmgrp.Nombre)
	}

	// Cambia el estado del grupo y de los usuarios asociados
	err = ActualizarEstadoEntidadOQuitarUsuarios(archivo, superbloque, &usuariosInodo, rmgrp.Nombre, "G", "0")
	if err != nil {
		return fmt.Errorf("error eliminando el grupo y usuarios asociados: %v", err)
	}

	// Actualiza el inodo de users.txt en el archivo
	err = usuariosInodo.Codificar(archivo, inodoOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Guarda el Superbloque utilizando el Start como el offset
	err = superbloque.Codificar(archivo, int64(particion.Start)) // Usar Start como offset
	if err != nil {
		return fmt.Errorf("error guardando el Superblock: %v", err)
	}

	// Muestra mensajes donde confirman que se eliminó el grupo
	fmt.Printf("Grupo '%s' eliminado exitosamente, junto con sus usuarios.\n", rmgrp.Nombre)
	fmt.Println("Inodos actualizados:")
	superbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques de datos actualizados:")
	superbloque.ImprimirBloques(archivo.Name())

	return nil
}

// Cambia el estado de un grupo/usuario y elimina usuarios asociados a un grupo
func ActualizarEstadoEntidadOQuitarUsuarios(archivo *os.File, superbloque *Structs.Superbloque, inodoUsuarios *Structs.Inodo, Nombre string, tipoEntidad string, nuevoEstado string) error {
	// Lee el contenido actual de users.txt
	contenido, err := globales.LeerBloquesArchivo(archivo, superbloque, inodoUsuarios)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Separa las líneas del archivo
	lineas := strings.Split(contenido, "\n")
	modificado := false

	// Variable para detectar si es un grupo
	var groupName string
	if tipoEntidad == "G" {
		groupName = Nombre
	}

	// Recorre las líneas para actualizar el estado de un grupo/usuario y eliminar usuarios asociados si es un grupo
	for i, linea := range lineas {
		linea = strings.TrimSpace(linea) // Se eliminan los espacios en blanco adicionales

		if linea == "" {
			continue
		}

		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue // Se ignoran las líneas mal formadas
		}

		tipo := partes[1]
		nombre := partes[2]

		// Verifica si coincide el tipo de entidad ya sea el usuario o grupo y también el nombre
		if tipo == tipoEntidad && nombre == Nombre {
			// Cambia el estado del grupo o usuario
			partes[0] = nuevoEstado
			lineas[i] = strings.Join(partes, ",")
			modificado = true

			// Si esta vaina es un grupo, busca y elimina a los usuarios asociados
			if tipoEntidad == "G" {
				// Recorre de nuevo todas las líneas para eliminar usuarios de ese grupo
				for j, lineaUsuario := range lineas {
					lineaUsuario = strings.TrimSpace(lineaUsuario)
					if lineaUsuario == "" {
						continue
					}
					partesUsuario := strings.Split(lineaUsuario, ",")
					if len(partesUsuario) == 5 && partesUsuario[2] == groupName {
						// Marcar el usuario como eliminado
						partesUsuario[0] = "0"
						lineas[j] = strings.Join(partesUsuario, ",")
					}
				}
			}
			break // Solo se necesita modificar una entrada del grupo/usuario para hacer esta vaina
		}
	}

	// Si se modificó alguna línea, guardar los cambios en el archivo
	if modificado {
		contenidoActualizado := strings.Join(lineas, "\n")

		// Limpiar los bloques asignados al archivo antes de escribir
		for _, blockIndex := range inodoUsuarios.I_block {
			if blockIndex == -1 {
				break // No hay más bloques asignados
			}

			blockOffset := int64(superbloque.S_block_start + blockIndex*superbloque.S_block_s)
			var fileBlock Structs.BloqueFile

			// Limpiar el contenido del bloque
			fileBlock.LimpiarContenido()

			// Escribir el bloque vacío de nuevo
			err = fileBlock.Codificar(archivo, blockOffset)
			if err != nil {
				return fmt.Errorf("error escribiendo bloque limpio %d: %w", blockIndex, err)
			}
		}

		// Reescribir todo el contenido en los bloques después de limpiar
		err = globales.EscribirBloquesEnUserstxt(archivo, superbloque, inodoUsuarios, contenidoActualizado)
		if err != nil {
			return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
		}
	} else {
		return fmt.Errorf("%s '%s' no encontrado en users.txt", tipoEntidad, Nombre)
	}

	return nil
}
