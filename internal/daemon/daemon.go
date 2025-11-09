package daemon

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const SocketPath = "/tmp/flick.sock"

type AppController interface {
	GetStatus() string
}

type Daemon struct {
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
	app      AppController
}

func NewDaemon(controller AppController) (*Daemon, error) {
	if err := os.RemoveAll(SocketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	log.Println("Daemon listening on", SocketPath)
	return &Daemon{
		listener: listener,
		quit:     make(chan struct{}),
		app:      controller,
	}, nil
}

func (d *Daemon) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	d.wg.Add(1)
	go d.acceptConnections()

	log.Println("Daemon started. Press Ctrl+C to stop.")

	<-sigChan

	log.Println("Recieved shutdown signal, stopping daemon...")
	d.Stop()
}

func (d *Daemon) Stop() {
	close(d.quit)
	d.listener.Close()
	d.wg.Wait()
	log.Println("Daemon stopped.")
}

func (d *Daemon) acceptConnections() {
	defer d.wg.Done()

	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.quit:
				return
			default:
				log.Println("Error accepting connection:", err)
			}
			continue
		}

		d.wg.Add(1)
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer d.wg.Done()
	defer conn.Close()

	cmd, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		if err.Error() != "EOF" {
			log.Println("Error reading from client:", err)
		}
		return
	}

	cmd = strings.TrimSpace(cmd) // cleans the command
	log.Printf("Daemon: Recieved command: %s", cmd)

	var response string

	switch cmd {
	case "VERSION":
		response = "flick v0.1.0-alpha\n"

	case "STATUS":
		response = d.app.GetStatus() + "\n"

	default:
		response = "Unknown command\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Println("Error writtin response to the client:", err)
	}
}
