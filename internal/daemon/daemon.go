package daemon

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SocketPath es la ruta donde se creará el archivo de socket Unix.
const SocketPath = "/tmp/flick.sock"

// Daemon representa el proceso en segundo plano.
// Contiene la lógica para iniciar, detener y manejar conexiones.
type Daemon struct {
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
}

// NewDaemon crea e inicializa una nueva instancia del daemon.
func NewDaemon() (*Daemon, error) {
	// Asegurarse de que el socket no exista antes de empezar.
	// Esto previene errores si el daemon anterior no se cerró correctamente.
	if err := os.RemoveAll(SocketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	log.Println("Daemon escuchando en", SocketPath)
	return &Daemon{
		listener: listener,
		quit:     make(chan struct{}),
	}, nil
}

// Start inicia el daemon, comenzando a aceptar conexiones
// y esperando una señal de interrupción para detenerse.
func (d *Daemon) Start() {
	// Creamos un canal para escuchar las señales del sistema operativo.
	// Esto nos permite atrapar Ctrl+C (SIGINT) o una señal de terminación (SIGTERM)
	// para apagar el daemon de forma segura.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	d.wg.Add(1)
	go d.acceptConnections()

	log.Println("Daemon iniciado. Presiona Ctrl+C para detener.")

	// Esperamos a recibir una señal de apagado.
	<-sigChan

	// Una vez recibida la señal, iniciamos el proceso de apagado.
	log.Println("Recibida señal de apagado, deteniendo el daemon...")
	d.Stop()
}

// Stop detiene el daemon de forma segura.
func (d *Daemon) Stop() {
	// Cerramos el canal 'quit' para señalar a todas las goroutines que deben detenerse.
	close(d.quit)

	// Cerramos el listener para dejar de aceptar nuevas conexiones.
	d.listener.Close()

	// Esperamos a que todas las goroutines en el WaitGroup terminen.
	// En este caso, esperamos a que 'acceptConnections' finalice.
	d.wg.Wait()
	log.Println("Daemon detenido.")
}

// acceptConnections es el bucle principal que acepta nuevas conexiones de clientes.
// Se ejecuta en su propia goroutine.
func (d *Daemon) acceptConnections() {
	defer d.wg.Done()

	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.quit:
				// Si el canal 'quit' está cerrado, es una salida esperada.
				return
			default:
				log.Println("Error al aceptar conexión:", err)
			}
			continue
		}

		// Por cada conexión, iniciamos una nueva goroutine para manejarla.
		// Esto permite al daemon manejar múltiples clientes simultáneamente.
		d.wg.Add(1)
		go d.handleConnection(conn)
	}
}

// handleConnection maneja la lógica para una conexión de cliente individual.
func (d *Daemon) handleConnection(conn net.Conn) {
	defer d.wg.Done()
	defer conn.Close()

	log.Println("Cliente conectado:", conn.RemoteAddr().String())

	// Aquí puedes agregar la lógica para comunicarte con el TUI.
	// Por ahora, solo enviamos un mensaje de bienvenida.
	_, err := conn.Write([]byte("¡Bienvenido al daemon de Flick!\n"))
	if err != nil {
		log.Println("Error al escribir al cliente:", err)
		return
	}

	// Puedes crear un bucle para leer comandos del TUI aquí.
	// buffer := make([]byte, 1024)
	// for {
	//     n, err := conn.Read(buffer)
	//     if err != nil {
	//         log.Println("Cliente desconectado.")
	//         return
	//     }
	//     log.Printf("Recibido: %s", buffer[:n])
	// }
}
