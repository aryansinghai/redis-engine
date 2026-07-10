package server

import (
	"log"
	"net"
	"syscall"
	"time"

	"server/config"
	"server/core"
)

var con_clients int = 0
var cronFrequency time.Duration = 1 * time.Second
var cronLastRun time.Time = time.Now()

func RunAsyncTCPServer() error {
	log.Printf("Starting an async TCP server on %s:%d", config.Config.Host, config.Config.Port)

	maxClients := 20000

	var events []syscall.Kevent_t = make([]syscall.Kevent_t, maxClients)

	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0) // #genai: SOCK_STREAM is the socket type; non-blocking is set below
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(config.Config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	if err = syscall.Listen(serverFD, maxClients); err != nil {
		return err
	}

	// Async IO starts here

	epollFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatalf("Error creating kqueue: %v", err)
	}
	defer syscall.Close(epollFD)

	var event syscall.Kevent_t = syscall.Kevent_t{
		Ident:  uint64(serverFD),    // File descriptor to monitor
		Filter: syscall.EVFILT_READ, // Watch for read events
		Flags:  syscall.EV_ADD,
	}

	_, err = syscall.Kevent(
		epollFD,
		[]syscall.Kevent_t{event}, // Changes to apply
		nil,                       // Not waiting for events
		nil,
	)
	if err != nil {
		return err
	}

	for {
		if time.Now().Sub(cronLastRun) >= cronFrequency {
			cronLastRun = time.Now()
			core.CleanupExpiredKeys()
		}

		nevents, err := syscall.Kevent(
			epollFD,
			nil,
			events,
			nil,
		)
		if err != nil {
			continue
		}

		for i := 0; i < nevents; i++ {
			kev := &events[i]
			if int(kev.Ident) == serverFD {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Printf("Error accepting connection: %v", err)
					continue
				}
				con_clients++
				log.Printf("Client connected: %d", con_clients)
				syscall.SetNonblock(fd, true)

				var event syscall.Kevent_t = syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
				}
				_, err = syscall.Kevent(
					epollFD,
					[]syscall.Kevent_t{event},
					nil,
					nil,
				)
				if err != nil {
					log.Printf("Error adding client to kqueue: %v", err)
					continue
				}
			} else {
				comm := core.FDComm{Fd: int(kev.Ident)}
				cmds, err := readCommands(comm)
				if err != nil {
					syscall.Close(int(kev.Ident))
					con_clients--
					log.Printf("Client disconnected: %d", con_clients)
					continue
				}
				respond(comm, cmds)
			}
		}
	}
}
