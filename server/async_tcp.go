package server

import (
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"server/config"
	"server/core"
)

var con_clients int = 0
var cronFrequency time.Duration = 1 * time.Second
var cronLastRun time.Time = time.Now()

var connectedClients map[int]*core.Client = make(map[int]*core.Client)

const EngineStatus_WAITING int32 = 1 << 1
const EngineStatus_BUSY int32 = 1 << 2
const EngineStatus_SHUTING_DOWN int32 = 1 << 3

var eStatus int32 = EngineStatus_WAITING

func WaitForSignals(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	for atomic.LoadInt32(&eStatus) == EngineStatus_BUSY {
	}

	atomic.StoreInt32(&eStatus, EngineStatus_SHUTING_DOWN)
	core.Shutdown()
	os.Exit(0)
}

func RunAsyncTCPServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTING_DOWN)
	}()
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

	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTING_DOWN {
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

		// Critical section
		//Here we need to ensure that the engine is not shutting down and that we are not already busy
		//If the engine is shutting down, we need to return immediately
		//If we are already busy, we need to wait for the engine to be available
		//If we are not busy, we can proceed to the next step
		//This is a critical section because we need to ensure that the engine is not shutting down and that we are not already busy

		// mark the engine as busy only when it is in the waiting state
		if !atomic.CompareAndSwapInt32(&eStatus, EngineStatus_WAITING, EngineStatus_BUSY) {
			// if th engine is shutting down, we need to return immediately
			// if swap failed because the engine is not in the waiting state, we need to wait for the engine to be available
			switch eStatus {
			case EngineStatus_SHUTING_DOWN:
				return nil
			}
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
				connectedClients[fd] = core.NewClient(fd)
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
				comm := connectedClients[int(events[i].Ident)]
				if comm == nil {
					log.Printf("Client not found: %d", int(events[i].Ident))
					continue
				}
				cmds, err := readCommands(comm)
				if err != nil {
					syscall.Close(int(kev.Ident))
					con_clients--
					log.Printf("Client disconnected: %d", con_clients)
					delete(connectedClients, int(events[i].Ident))
					continue
				}
				respond(comm, cmds)
			}
		}
		atomic.StoreInt32(&eStatus, EngineStatus_WAITING)
	}
	return nil
}
