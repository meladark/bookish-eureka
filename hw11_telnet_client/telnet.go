package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type TelnetClient interface {
	Connect() error
	Close() error
	Send() error
	Receive() error
}

type client struct {
	address string
	timeout time.Duration
	conn    net.Conn
	in      io.ReadCloser
	out     io.Writer
	// Добавил чтобы проверять на закрытие, так как
	// я понимаю у tcp нет прям мгновенного ответа
	closed atomic.Bool
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &client{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (c *client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *client) Send() error {
	scanner := bufio.NewScanner(c.in)
	for scanner.Scan() {
		if c.closed.Load() {
			return fmt.Errorf("connection closed by peer")
		}
		_, err := c.conn.Write(append(scanner.Bytes(), '\n'))
		if err != nil {
			c.closed.Store(true)
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (c *client) Receive() error {
	_, err := io.Copy(c.out, c.conn)
	c.closed.Store(true)
	return err
}
