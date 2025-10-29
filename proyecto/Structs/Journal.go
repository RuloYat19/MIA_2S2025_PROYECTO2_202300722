package Structs

import (
	"Proyecto/Herramientas"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

const JOURNAL_ENTRIES = 50

type Journal struct {
	J_count   int32
	J_content Information
}

type Information struct {
	I_operation [10]byte // 10 bytes - Tipo de operación (mkdir, mkfile, etc)
	I_path      [32]byte // 32 bytes - Ruta del recurso afectado
	I_content   [64]byte // 64 bytes - Contenido afectado
	I_date      uint32   // 4 bytes - Timestamp - Mejor precisión que float32
	// Total: 110 bytes
}

// Se guarda el journal en el archivo en la posición indicada
func (journal *Journal) Codificar(archivo *os.File, offset int64) error {
	// Se muestra el conteo y el offset
	fmt.Printf("Journal.Encode: count=%d, offset=%d\n", journal.J_count, offset)

	err := Herramientas.EscribirAlArchivo(archivo, offset, journal)
	if err != nil {
		fmt.Println("Error al escribir el journal en el archivo.")
		return fmt.Errorf("error al escribir el journal en el archivo: %w", err)
	}

	return nil
}

// Se lee el journal desde el archivo en la posición indicada
func (journal *Journal) Decodificar(archivo *os.File, offset int64) error {
	// Se muestra el offset
	fmt.Printf("Journal.Decodificar: offset=%d\n", offset)
	err := Herramientas.LeerDesdeElArchivo(archivo, offset, journal)
	if err != nil {
		fmt.Println("Error al leer el journar del archivo.")
		return fmt.Errorf("error al leer el journal del archivo: %w", err)
	}

	return nil
}

// Se imprime la información del journal en formato legible
func (journal *Journal) ImprimirJournal() {
	date := time.Unix(int64(journal.J_content.I_date), 0)
	fmt.Println("Journal:")
	fmt.Printf("J_count: %d\n", journal.J_count)
	fmt.Println("Information:")
	fmt.Printf("I_operation: %s\n", strings.TrimSpace(string(journal.J_content.I_operation[:])))
	fmt.Printf("I_path: %s\n", strings.TrimSpace(string(journal.J_content.I_path[:])))
	fmt.Printf("I_content: %s\n", strings.TrimSpace(string(journal.J_content.I_content[:])))
	fmt.Printf("I_date: %s\n", date.Format(time.RFC3339))
}

// Se crea una nueva entrada de journal en memoria
func (j *Journal) CrearEntradaJournal(op, ruta, contenido string) {
	*j = Journal{}
	copy(j.J_content.I_operation[:], op)
	copy(j.J_content.I_path[:], ruta)
	copy(j.J_content.I_content[:], contenido)
	j.J_content.I_date = uint32(time.Now().Unix())
}

// Se genera una tabla en formato dot para el journal
func (journal *Journal) GenerarTablaJournal(indiceJournal int32) string {
	fecha := time.Unix(int64(journal.J_content.I_date), 0).Format(time.RFC3339)
	operacion := strings.TrimSpace(string(journal.J_content.I_operation[:]))
	ruta := strings.TrimSpace(string(journal.J_content.I_path[:]))
	contenido := strings.TrimSpace(string(journal.J_content.I_content[:]))

	tabla := fmt.Sprintf(`journal_table_%d [label=<
        <TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">
            <TR>
                <TD COLSPAN="2" BGCOLOR="#4CAF50"><FONT COLOR="#FFFFFF">Journal Entry %d</FONT></TD>
            </TR>
            <TR>
                <TD BGCOLOR="#FF7043">Operation:</TD>
                <TD>%s</TD>
            </TR>
            <TR>
                <TD BGCOLOR="#FF7043">Path:</TD>
                <TD>%s</TD>
            </TR>
            <TR>
                <TD BGCOLOR="#FF7043">Content:</TD>
                <TD>%s</TD>
            </TR>
            <TR>
                <TD BGCOLOR="#FF7043">Date:</TD>
                <TD>%s</TD>
            </TR>
        </TABLE>
    >];`, indiceJournal, indiceJournal, operacion, ruta, contenido, fecha)

	return tabla
}

// Se genera el contenido del grafo de las entradas del Journal en formato DOT
func (journal *Journal) GenerarGrafo(journalStart int64, journalConteo int32, archivo *os.File) (string, error) {
	dotContent := ""
	tamanioEntrada := int64(binary.Size(Journal{}))

	fmt.Println("Generando grafo de Journal...")

	for i := int32(0); i < journalConteo; i++ {
		offset := journalStart + int64(i)*tamanioEntrada
		fmt.Printf("Leyendo entrada del Journal en offset: %d\n", offset)

		err := journal.Decodificar(archivo, offset)
		if err != nil {
			fmt.Printf("Error al deserializar el journal '%d' con offser de '%d'.\n", i, offset)
			return "", fmt.Errorf("error al deserializar el journal %d en offset %d: %v", i, offset, err)
		}

		operacion := strings.TrimSpace(string(journal.J_content.I_operation[:]))
		if operacion == "" {
			fmt.Printf("La entrada del journal está vacía con índice '%d'.\n", i)
			fmt.Printf("Entrada de Journal vacía encontrada en índice %d, deteniendo la lectura.\n", i)
			break
		}

		fmt.Printf("Generando tabla para la entrada de Journal %d con operación: %s\n", i, operacion)
		dotContent += journal.GenerarTablaJournal(i)
	}

	return dotContent, nil
}

// Se guarda una nueva entrada en el journal y la serializa en el archivo
func (journal *Journal) GuardarEntradaJournal(archivo *os.File, journaling_start int64, operacion string, ruta string, contenido string) error {
	journal.CrearEntradaJournal(operacion, ruta, contenido)

	// Se calcula el offset correcto basado en J_count
	tamanioEntrada := int64(binary.Size(Journal{}))
	offset := journaling_start + int64(journal.J_count)*tamanioEntrada

	// Se muestra el cálculo de offset
	fmt.Printf("SaveJournalEntry: index=%d, entrySize=%d, journaling_start=%d, offset=%d\n", journal.J_count, tamanioEntrada, journaling_start, offset)

	err := journal.Codificar(archivo, offset)
	if err != nil {
		fmt.Println("Error al guardar la entrada del journal.")
		return fmt.Errorf("error al guardar la entrada de journal: %w", err)
	}
	return nil
}

// Se calcula el espacio total necesario para el journaling
func CalculateJournalingSpace(n int32) int64 {
	return int64(n) * int64(binary.Size(Journal{}))
}

// Se inicializa el área completa de journaling con entradas vacías
func InicializarAreaJournal(archivo *os.File, journalStart int64, n int32) error {
	fmt.Println("=== Inicializando área de journaling ===")

	tamanioEntrada := int64(binary.Size(Journal{}))
	journalEnd := journalStart + tamanioEntrada*int64(n)

	fmt.Printf("Rango del Journal : [%d, %d)  (%d slots · %d bytes c/u)\n", journalStart, journalEnd, n, tamanioEntrada)

	nullJournal := &Journal{
		J_content: Information{
			I_operation: [10]byte{},
			I_path:      [32]byte{},
			I_content:   [64]byte{},
			I_date:      0,
		},
	}

	for i := int32(0); i < n; i++ {
		nullJournal.J_count = i
		offset := journalStart + tamanioEntrada*int64(i)

		if err := Herramientas.EscribirAlArchivo(archivo, offset, nullJournal); err != nil {
			fmt.Printf("Error inicializando el journal con slot '%d' y con offset '%d'.\n", i, offset)
			return fmt.Errorf("error inicializando journal slot %d (off %d): %w", i, offset, err)
		}
		fmt.Printf("slot=%02d | off=%d | ok\n", i, offset)
	}

	fmt.Printf("Journal inicializado correctamente con %d entradas\n", n)
	return nil
}

// Se buscan y se devuelven todas las entradas de journal válidas
func EncontrarEntradaValidaJournal(archivo *os.File, journalStart int64, maximoEntradas int32) ([]Journal, error) {
	var entradas []Journal
	var conteoValido int32 = 0
	tamanioEntrada := int64(binary.Size(Journal{}))

	fmt.Println("Buscando entradas válidas de journal...")

	// Se definen las operaciones válidas
	operacionesValidas := map[string]bool{
		"mkdir": true, "mkfile": true, "rm": true, "rmdir": true,
		"edit": true, "cat": true, "rename": true, "copy": true,
	}

	// Se muestran todas las entradas
	fmt.Println("Contenido actual del journal:")

	for i := int32(0); i < maximoEntradas; i++ {
		offset := journalStart + int64(i)*tamanioEntrada
		journal := &Journal{}

		if err := journal.Decodificar(archivo, offset); err != nil {
			fmt.Printf("Error leyendo journal en offset %d: %v\n", offset, err)
			break
		}

		// Se extrae y se limpia la operación
		operacionEnBruto := journal.J_content.I_operation[:]

		// Se encuentra el primer byte nulo
		posicionNula := 0
		for ; posicionNula < len(operacionEnBruto); posicionNula++ {
			if operacionEnBruto[posicionNula] == 0 {
				break
			}
		}

		operacion := string(operacionEnBruto[:posicionNula])
		operacion = strings.TrimSpace(operacion)

		// Se muestra cada entrada
		fmt.Printf("-- Entrada %d --\n", i)
		journal.ImprimirJournal()

		// Si no hay operación, se llega al final
		if operacion == "" {
			break
		}

		// Se usa el valor float32 directamente para obtener la fecha
		fechaUnix := float64(journal.J_content.I_date)
		fecha := time.Unix(int64(fechaUnix), 0)

		// Se muestran los detalles de validación
		fmt.Printf("Validando entrada %d: operación='%s', fechaUnix=%f, fecha=%s\n", i, operacion, fechaUnix, fecha.Format(time.RFC3339))

		// Se verifica si la operación es válida
		if _, ok := operacionesValidas[operacion]; !ok {
			fmt.Printf("Entrada %d rechazada: operación '%s' no válida\n", i, operacion)
			continue
		}

		// La entrada pasó todas las validaciones
		fmt.Printf("Entrada '%d' ACEPTADA.\n", i)
		entradas = append(entradas, *journal)
		conteoValido++
	}

	fmt.Printf("Se encontraron '%d' entradas válidas de journal.\n", conteoValido)
	return entradas, nil
}

// Se verifica si una entrada de journal está vacía
func EstaVacioJournal(j *Journal) bool {
	// Sólo se comprueba si la operación está vacía
	op := strings.TrimSpace(
		string(bytes.TrimRight(j.J_content.I_operation[:], "\x00")),
	)
	return op == ""
}

func AgregarEntradaJournal(archivo *os.File, journalStart int64, maximoEntradas int32, operacion string, ruta string, contenido string, superbloque *Superbloque) error {
	// Se verifica la consistencia usando el superbloque
	startEsperado := int64(superbloque.JournalStart())
	if journalStart != startEsperado {
		fmt.Printf("Se está ajustando el inicio del journal de '%d' a '%d'.\n", journalStart, startEsperado)
		journalStart = startEsperado
	}

	siguienteIndice, err := ObtenerSiguienteIndiceVacioJournal(archivo, journalStart, maximoEntradas)
	if err != nil {
		fmt.Println("Error buscando el siguiente índice disponible.")
		return fmt.Errorf("error buscando el siguiente índice disponible: %w", err)
	}

	if siguienteIndice >= maximoEntradas {
		fmt.Printf("Journal lleno, sobreescribiendo desde el principio (índice 0)\n")
		siguienteIndice = 0
	}

	journal := &Journal{
		J_count: siguienteIndice,
	}

	journal.CrearEntradaJournal(operacion, ruta, contenido)
	offset := journalStart + int64(siguienteIndice)*int64(binary.Size(Journal{}))

	// Se verifican los límites usando el superbloque
	journalEnd := int64(superbloque.JournalEnd())
	if offset >= journalEnd {
		fmt.Printf("Error en el intento de escritura fuera del área de journal (%d >= %d).\n", offset, journalEnd)
		return fmt.Errorf("error: intento de escritura fuera del área de journal (%d >= %d)", offset, journalEnd)
	}

	// Usar Encode para ser consistentes
	if err := journal.Codificar(archivo, offset); err != nil {
		fmt.Println("Error escribiendo la nueva entrada del journal.")
		return fmt.Errorf("error escribiendo nueva entrada de journal: %w", err)
	}

	if err := archivo.Sync(); err != nil {
		fmt.Println("Error sincronizando el archivo.")
		return fmt.Errorf("error sincronizando archivo: %w", err)
	}

	fmt.Printf("Nueva entrada de journal agregada en índice %d: %s %s\n", siguienteIndice, operacion, ruta)

	return nil
}

// Se encuentra el próximo índice realmente vacío
func ObtenerSiguienteIndiceVacioJournal(archivo *os.File, journalStart int64, maximoEntradas int32) (int32, error) {
	tamanioEntrada := int64(binary.Size(Journal{}))

	fmt.Printf("Scan Journal | start=%d | slots=%d | entrySize=%d bytes\n",
		journalStart, maximoEntradas, tamanioEntrada)

	for i := int32(0); i < maximoEntradas; i++ {
		offset := journalStart + tamanioEntrada*int64(i)

		j := &Journal{}
		if err := j.Decodificar(archivo, offset); err != nil {
			return -1, fmt.Errorf("leer journal[%d] en off %d: %w", i, offset, err)
		}

		op := strings.TrimSpace(string(j.J_content.I_operation[:]))
		date := j.J_content.I_date

		fmt.Printf("slot=%02d | off=%d | op='%s' | date=%d\n", i, offset, op, date)

		// Reutilizamos la lógica centralizada:
		if EstaVacioJournal(j) {
			fmt.Printf("Slot libre encontrado en índice %d\n", i)
			return i, nil
		}
	}

	fmt.Println("Journal lleno: se usará sobreescritura circular (idx 0)")
	return 0, nil
}
