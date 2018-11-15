package ws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	auth bool
	send chan []byte
}

func (c *Client) log(f string, v ...interface{}) {
	msg := fmt.Sprintf("[srv.cli] c: %s, ", c.conn.RemoteAddr())
	log.Printf(msg+f, v...)
}

func (c *Client) command(cmd string, args string) {
	c.log("command: %s, arglen: %d\n", cmd, len(args))
	switch cmd {
	case "auth":
		if len(args) > 0 {
			// ac := getClientsAuthConf()
			name, ok := c.hub.auth[args]
			if ok {
				c.log("auth as '%s' successful\n", name)
				c.auth = true
			} else {
				c.log("auth unsucessful\n")
			}
		}
	}
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(message)
		cmd := bytes.SplitN(message, space, 2)
		if cmd != nil && len(cmd[0]) > 0 {
			args := ""
			if len(cmd[1:]) > 0 {
				args = string(cmd[1:][0])
			}
			c.command(string(cmd[0]), args)
		}
	}
}

func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
