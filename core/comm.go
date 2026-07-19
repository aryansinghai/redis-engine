package core

import (
	"bytes"
	"fmt"
	"io"
	"syscall"
)

type Client struct {
	io.ReadWriter
	FD     int
	cqueue RedisCmds
	isTxn  bool
}

func NewClient(fd int) *Client { // #genai: package constructor for per-connection client state
	return &Client{
		FD:     fd,
		cqueue: make(RedisCmds, 0),
	}
}

func (c Client) Read(p []byte) (n int, err error) {
	return syscall.Read(c.FD, p)
}

func (c Client) Write(p []byte) (n int, err error) {
	return syscall.Write(c.FD, p)
}

func (c *Client) TxnBegin(cmd *RedisCmd) {
	c.isTxn = true
}

func (c *Client) TxnExec() []byte {
	var out []byte
	buf := bytes.NewBuffer(out)

	buf.WriteString(fmt.Sprintf("*%d\r\n", len(c.cqueue)))
	for _, cmd := range c.cqueue {
		buf.Write(executeCommand(cmd, c))
	}
	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false
	return buf.Bytes()
}

func (c *Client) TxnQueue(cmd *RedisCmd) {
	c.cqueue = append(c.cqueue, cmd)
}

func (c *Client) TxnDiscard() []byte {
	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false
	return RESP_OK
}
