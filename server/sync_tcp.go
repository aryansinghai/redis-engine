package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"server/core"
)

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	buffer := make([]byte, 1024)
	n, err := c.Read(buffer[:])
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buffer[:n])
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 { // #genai
		return nil, fmt.Errorf("empty command")
	}

	return &core.RedisCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}, nil
}

func toArrayString(value []interface{}) ([]string, error) {
	tokens := make([]string, len(value))
	for i := range value {
		tokens[i] = value[i].(string)
	}
	return tokens, nil
}

func readCommands(c io.ReadWriter) (core.RedisCmds, error) {
	buffer := make([]byte, 1024)
	n, err := c.Read(buffer[:])
	if err != nil {
		return nil, err
	}

	values, err := core.Decode(buffer[:n])
	if err != nil {
		return nil, err
	}

	var cmds core.RedisCmds = make(core.RedisCmds, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}
	return cmds, nil
}

func respondError(conn io.ReadWriter, err error) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
}

func respond(c io.ReadWriter, cmds core.RedisCmds) {
	err := core.EvalAndRespond(cmds, c)
	if err != nil {
		respondError(c, err)
	}
}

func RunSyncTCPServer(host string, port int) {
	log.Printf("Starting a sync TCP server on %s:%d", host, port)

	var con_clients int = 0

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Printf("Error listening: %v", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
		}

		con_clients++
		log.Printf("New client connected: %d", con_clients)
		log.Printf("client connected with address: %s", conn.RemoteAddr()) // #genai

		for {
			cmd, err := readCommands(conn)
			if err != nil {
				conn.Close()
				con_clients--
				log.Printf("Client disconnected: %d", con_clients)
				if err == io.EOF {
					break
				}
				log.Printf("Error reading command: %v", err)
			}
			log.Printf("Received command: %s", cmd)

			respond(conn, cmd)
		}
	}
}
