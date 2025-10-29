package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MKGRP struct {
	Nombre string
}

func Mkgrp(parametros []string) string {
	fmt.Println("\n======= MKGRP =======")

	var salida1 = ""

	// Se inicializa la estructura MKGRP
	mkgrp := &MKGRP{}

	// Otras variables
	paramC := true
	nombreInit := false

	// Se asignan los parámetros a la estructura
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "MKGRP Error: Valor desconocido del parámetro " + tmp[0] + ".\n"
			fmt.Println("MKGRP Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "name" {
			nombreInit = true
			mkgrp.Nombre = tmp[1]
		} else {
			salida1 += "MKGRP Error: Parametro desconocido: " + tmp[0] + ".\n"
			fmt.Println("MKGRP Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Se crea el grupo con los datos asignados
	if paramC && nombreInit {
		err := crearGrupo(mkgrp)
		if err != nil {
			salida1 += "MKGRP Error: Hubo problemas creando el grupo " + mkgrp.Nombre + ".\n"
			fmt.Printf("MKGRP Error: Hubo problemas creando el grupo " + mkgrp.Nombre + ".\n")
		} else {
			salida1 += "El grupo " + mkgrp.Nombre + " fue creado exitosamente.\n"
		}
	}
	fmt.Println("\n======FIN MKGRP======")
	return salida1
}

func crearGrupo(mkgrp *MKGRP) error {
	fmt.Println(mkgrp.Nombre)

	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se verifica si el usuario es el root
	if globales.UsuarioActual.Nombre != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Se verifica que la partición esté montada en el sistema
	_, ruta, err := globales.ObtenerParticionMontadas(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Se abre el archivo donde se encuentra la partición montada
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

	// Se obtiene la partición asociada al id
	particion, err := mbr.ObtenerParticionPorID(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se pudo obtener la partición: %v", err)
	}

	// Se lee el inodo de users.txt
	var usuariosInodo Structs.Inodo

	// Se calcula el offset del inodo de users.txt que esta en el inodo 1
	inodoOffset := int64(superbloque.S_inode_start + int32(binary.Size(usuariosInodo)))

	// Se decodifica el inodo de users.txt
	err = usuariosInodo.Decodificar(archivo, inodoOffset)
	usuariosInodo.ActualizarAtime() // Se actualiza la última fecha de acceso

	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Se verifica si el grupo ya existe en users.txt
	_, err = globales.BuscarEnElArchivoDeUsuario(archivo, superbloque, &usuariosInodo, mkgrp.Nombre, "G")
	if err == nil {
		fmt.Printf("el grupo '%s' ya existe\n", mkgrp.Nombre)
		err = fmt.Errorf("el grupo '%s' ya existe.%v", mkgrp.Nombre, err)
		return err
	}

	// Se obtiene el siguiente ID disponible para el nuevo grupo
	siguienteIDGrupo, err := calcularIDSiguiente(archivo, superbloque, &usuariosInodo)
	if err != nil {
		return fmt.Errorf("error calculando el siguiente ID: %v", err)
	}

	// Se crea la nueva entrada de grupo con el siguiente ID
	nuevaEntradaGrupo := fmt.Sprintf("%d,G,%s", siguienteIDGrupo, mkgrp.Nombre)

	// Se usa la función modular para crear el grupo en users.txt
	err = globales.AgregarEntradaAlArchivoDeUserstxt(archivo, superbloque, &usuariosInodo, nuevaEntradaGrupo, mkgrp.Nombre, "G")
	if err != nil {
		return fmt.Errorf("error creando el grupo '%s': %v", mkgrp.Nombre, err)
	}

	// Se actualiza el inodo de users.txt
	err = usuariosInodo.Codificar(archivo, inodoOffset)

	usuariosInodo.ActualizarAtime() // Se actualiza la última fecha de acceso

	if err != nil {
		fmt.Printf("Error actualizando inodo de users.txt: %v\n", err)
		return err
	}

	// Se guarda el Superbloque utilizando el Start como el offset
	err = superbloque.Codificar(archivo, int64(particion.Start))
	if err != nil {
		return fmt.Errorf("error guardando el Superblock: %v", err)
	} else {
		fmt.Println("\nSuperbloque guardado correctamente")
		superbloque.ImprimirSuperbloque()
		fmt.Println("\nInodos")
		superbloque.ImprimirInodos(archivo.Name())
		superbloque.ImprimirBloques(archivo.Name())

	}

	fmt.Printf("Grupo creado exitosamente: %s\n", mkgrp.Nombre)

	return nil
}

// Se calcula el siguiente ID disponible para un grupo o usuario en users.txt
func calcularIDSiguiente(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo) (int, error) {
	// Se lee el contenido de users.txt
	contenido, err := globales.LeerBloquesArchivo(archivo, superbloque, inodo)
	if err != nil {
		return -1, fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Se busca el mayor ID en el archivo
	lineas := strings.Split(contenido, "\n")
	maxID := 0
	for _, linea := range lineas {
		if linea == "" {
			continue
		}

		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue
		}

		// Se convierte el primer campo (ID) a entero
		id, err := strconv.Atoi(campos[0])
		if err != nil {
			continue
		}

		// Se actualiza el maxID si encontramos uno mayor
		if id > maxID {
			maxID = id
		}
	}

	// Se devuelve el siguiente ID disponible
	return maxID + 1, nil
}
