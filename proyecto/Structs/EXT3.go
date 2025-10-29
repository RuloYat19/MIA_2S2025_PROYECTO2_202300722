package Structs

import (
	"fmt"
	"os"
)

func (superbloque *Superbloque) CrearArchivoUsersEnExt3(archivo *os.File, journaling_start int64) error {
	// Se inicializa el área de journaling para la partición si es necesario
	fmt.Println("Inicializando área de journaling para EXT3...")

	err := InicializarAreaJournal(archivo, journaling_start, JOURNAL_ENTRIES)
	if err != nil {
		fmt.Println("Error al inicializar el área de journaling.")
		return fmt.Errorf("error al inicializar el área de journaling: %w", err)
	}

	// Se obtiene el siguiente índice de journal disponible
	siguienteIndiceJournal, err := ObtenerSiguienteIndiceVacioJournal(archivo, journaling_start, JOURNAL_ENTRIES)
	if err != nil {
		fmt.Println("Error al obtener el siguiente índice del journal.")
		return fmt.Errorf("error obteniendo el siguiente índice de journal: %w", err)
	}

	fmt.Printf("Siguiente índice de journal disponible: %d\n", siguienteIndiceJournal)

	// Se crea la entrada de journal para el directorio raíz
	err = AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "mkdir", "/", "", superbloque)

	if err != nil {
		fmt.Println("Error al guardar la entrada de la raíz en el journal.")
		return fmt.Errorf("error al guardar la entrada de la raíz en el journal: %w", err)
	}

	// Se crea el inodo y bloque para la raíz como antes
	bloqueIndiceRoot, err := superbloque.EncontrarElSiguienteBloqueLibre(archivo)
	if err != nil {
		fmt.Println("Error al encontrar el primer bloque libre para la raíz.")
		return fmt.Errorf("error al encontrar el primer bloque libre para la raíz: %w", err)
	}

	bloquesRoot := [15]int32{bloqueIndiceRoot, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

	inodoRoot := &Inodo{}
	err = inodoRoot.CrearInodo(archivo, superbloque, '0', 0, bloquesRoot, [3]byte{'7', '7', '7'})
	if err != nil {
		fmt.Println("Error al crear el inodo raiz.")
		return fmt.Errorf("error al crear el inodo raíz: %w", err)
	}

	bloqueRoot := &BloqueFolder{
		B_contenido: [4]ContenidoFolder{
			{B_nombre: [12]byte{'.'}, B_inodo: 0},
			{B_nombre: [12]byte{'.', '.'}, B_inodo: 0},
			{B_nombre: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: superbloque.S_inodes_count},
			{B_nombre: [12]byte{'-'}, B_inodo: -1},
		},
	}

	err = superbloque.ActualizarBitmapBloque(archivo, bloqueIndiceRoot, true)
	if err != nil {
		fmt.Println("Error al actualizar el bitmap de bloques.")
		return fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
	}

	err = bloqueRoot.Codificar(archivo, int64(superbloque.S_first_blo))
	if err != nil {
		fmt.Println("Error al serializar el bloque raiz.")
		return fmt.Errorf("error serializando el bloque raíz: %w", err)
	}

	superbloque.ActualizarSuperbloqueDespuesAsignacionBloques()

	// Se crea el contenido del archivo de usuarios
	grupoRoot := NuevoGrupo("1", "root")
	usuarioRoot := NuevoUsuario("1", "root", "root", "123")
	textoUsers := fmt.Sprintf("%s\n%s\n", grupoRoot.ToString(), usuarioRoot.ToString())

	// Se crea una segunda entrada en el journal para el archivo users.txt
	err = AgregarEntradaJournal(archivo, journaling_start, JOURNAL_ENTRIES, "mkfile", "/users.txt", textoUsers, superbloque)
	if err != nil {
		fmt.Println("Error al guardar la entrada del archivo /users.txt en el journal.")
		return fmt.Errorf("error al guardar la entrada del archivo /users.txt en el journal: %w", err)
	}

	// Se crea el archivo users.txt
	indiceBloqueUsers, err := superbloque.EncontrarElSiguienteBloqueLibre(archivo)
	if err != nil {
		fmt.Println("Error al encontrar el primer bloque libre para /users.txt")
		return fmt.Errorf("error al encontrar el primer bloque libre para /users.txt: %w", err)
	}
	bloquesArchivo := [15]int32{indiceBloqueUsers, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

	inodoUsers := &Inodo{}
	err = inodoUsers.CrearInodo(archivo, superbloque, '1', int32(len(textoUsers)), bloquesArchivo, [3]byte{'7', '7', '7'})
	if err != nil {
		fmt.Println("Error al crear el inodo de /users.txt.")
		return fmt.Errorf("error al crear el inodo de /users.txt: %w", err)
	}

	bloqueUsers := &BloqueFile{
		B_contenido: [64]byte{},
	}

	bloqueUsers.EstablecerContenido(textoUsers)
	err = bloqueUsers.Codificar(archivo, int64(superbloque.S_first_blo))
	if err != nil {
		fmt.Println("Error al serializar el bloque de /users.txt.")
		return fmt.Errorf("error serializando el bloque de /users.txt: %w", err)
	}

	err = superbloque.ActualizarBitmapBloque(archivo, indiceBloqueUsers, true)
	if err != nil {
		fmt.Println("Error al actualizar el bitmap de bloques para /users.txt.")
		return fmt.Errorf("error actualizando el bitmap de bloques para /users.txt: %w", err)
	}

	superbloque.ActualizarSuperbloqueDespuesAsignacionBloques()

	// Se muestra estado del sistema de archivos
	fmt.Println("Bloques")
	superbloque.ImprimirBloques(archivo.Name())

	// Se muestran las entradas de journal usando los nuevos métodos
	fmt.Println("Journal Entries:")
	entradas, err := EncontrarEntradaValidaJournal(archivo, journaling_start, JOURNAL_ENTRIES)
	if err != nil {
		fmt.Printf("Error leyendo entradas de journal: %v\n", err)
	} else {
		for i, entrada := range entradas {
			fmt.Printf("== Entrada '%d' ==\n", i)
			entrada.ImprimirJournal()
		}
	}

	fmt.Println("Sistema de archivos EXT3 inicializado correctamente con el journaling")
	return nil
}
