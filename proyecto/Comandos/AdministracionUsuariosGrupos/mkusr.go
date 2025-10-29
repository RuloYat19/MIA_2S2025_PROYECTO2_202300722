package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type MKUSR struct {
	Usuario     string
	Contrasenia string
	Grupo       string
}

func Mkusr(parametros []string) string {
	fmt.Println("\n======= MKUSR =======")

	var salida1 = ""

	// Se inicializa la estructura MKUSR
	mkusr := &MKUSR{}

	// Otras variables
	paramC := true
	usuarioInit := false
	contraseniaInit := false
	grupoInit := false

	// Se asignan los parámetros a la estructura
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "MKUSR Error: Valor desconocido del parámetro " + tmp[0] + ".\n"
			fmt.Println("MKUSR Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "user" {
			usuarioInit = true
			mkusr.Usuario = tmp[1]
		} else if strings.ToLower(tmp[0]) == "pass" {
			contraseniaInit = true
			mkusr.Contrasenia = tmp[1]
		} else if strings.ToLower(tmp[0]) == "grp" {
			grupoInit = true
			mkusr.Grupo = tmp[1]
		} else {
			salida1 += "MKUSR Error: Parametro desconocido: " + tmp[0] + ".\n"
			fmt.Println("MKUSR Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Se validan las longitudes de los parámetros
	// Parámetro Usuario
	if err := validarLongitudParametros(mkusr.Usuario, 10, "Usuario"); err != nil {
		salida1 += "MKUSR Error: La longitud del usuario " + mkusr.Usuario + " es mayor de 10 caracteres.\n"
		return salida1
	}

	// Parámetro Contraseña
	if err := validarLongitudParametros(mkusr.Contrasenia, 10, "Contraseña"); err != nil {
		salida1 += "MKUSR Error: La longitud de la contraseña " + mkusr.Contrasenia + " es mayor de 10 caracteres.\n"
		return salida1
	}

	// Parámetro Grupo
	if err := validarLongitudParametros(mkusr.Grupo, 10, "Grupo"); err != nil {
		salida1 += "MKUSR Error: La longitud del grupo " + mkusr.Grupo + " es mayor de 10 caracteres.\n"
		return salida1
	}

	if paramC && usuarioInit && contraseniaInit && grupoInit {
		err := crearUsuario(mkusr)
		if err != nil {
			salida1 += "MKUSR Error: Hubo problemas creando el usuario " + mkusr.Usuario + ".\n"
		} else {
			salida1 += "El usuario " + mkusr.Usuario + " fue creado exitosamente en el grupo " + mkusr.Grupo + ".\n"
		}

	}

	fmt.Println("\n======FIN MKUSR======")
	return salida1
}

func validarLongitudParametros(parametro string, maxLongitud int, nombreParametro string) error {
	if len(parametro) > maxLongitud {
		return fmt.Errorf("%s debe tener un máximo de %d caracteres", nombreParametro, maxLongitud)
	}
	return nil
}

func crearUsuario(mkusr *MKUSR) error {
	fmt.Println(mkusr.Usuario, " ", mkusr.Grupo, " ", mkusr.Contrasenia)
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

	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Se verifica si el grupo ya existe en users.txt
	_, err = globales.BuscarEnElArchivoDeUsuario(archivo, superbloque, &usuariosInodo, mkusr.Grupo, "G")

	if err != nil {
		fmt.Printf("El grupo '%s' no existe\n", mkusr.Grupo)
		return fmt.Errorf("el grupo no existe: %v", err)
	}

	// Se verifica si el usuario ya existe en users.txt
	_, err = globales.BuscarEnElArchivoDeUsuario(archivo, superbloque, &usuariosInodo, mkusr.Usuario, "U")

	if err == nil {
		fmt.Printf("El usuario '%s' ya existe\n", mkusr.Usuario)
		return fmt.Errorf("el usuario ya existe: %v", err)
	}

	// Se crea un nuevo objeto de tipo usuario
	usuario := Structs.NuevoUsuario(fmt.Sprintf("%d", superbloque.S_inodes_count+1), mkusr.Grupo, mkusr.Usuario, mkusr.Contrasenia)
	fmt.Printf("El objeto usuario quedó como: %s\n", usuario.ToString())

	// Se inserta la nueva entrada en el archivo users.txt
	err = globales.InsertarEnElArchivoDeUsuarios(archivo, superbloque, &usuariosInodo, usuario.ToString())
	if err != nil {
		return fmt.Errorf("error insertando el usuario '%s': %v", mkusr.Usuario, err)
	}

	// Se actualiza el inodo de users.txt
	err = usuariosInodo.Codificar(archivo, inodoOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Se guarda el Superbloque usando el Start como el offset
	err = superbloque.Codificar(archivo, int64(particion.Start))
	if err != nil {
		return fmt.Errorf("error guardando el Superblock: %v", err)
	}

	// Se muestran los mensajes donde se confirma que se creo el usuario en el grupo indicado
	fmt.Printf("Usuario '%s' agregado exitosamente al grupo '%s'\n", mkusr.Usuario, mkusr.Grupo)
	fmt.Println("\nSuperblock")
	superbloque.ImprimirSuperbloque()
	fmt.Println("\nInodos")
	superbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques")
	superbloque.ImprimirBloques(archivo.Name())

	return nil
}
