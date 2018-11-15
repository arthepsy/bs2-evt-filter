package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Hub struct {
	clients    map[*Client]bool
	auth       map[string]string
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		auth:       make(map[string]string),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Broadcast(msg []byte) {
	h.broadcast <- msg
}

func (h *Hub) UpdateAuth(auth map[string]string) {
	h.auth = auth
}

func (h *Hub) log(f string, v ...interface{}) {
	log.Printf("[hub] "+f, v...)
}

func (h *Hub) newClient(conn *websocket.Conn) *Client {
	return &Client{
		hub:  h,
		conn: conn,
		auth: false,
		send: make(chan []byte, 512),
	}
}

func (h *Hub) Client(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log("websocket connection error: %v", err)
		http.Error(w, "could not open websocket connection", http.StatusBadRequest)
		return
	}
	client := h.newClient(conn)
	client.hub.register <- client
	go client.write()
	go client.read()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.log("websocket client connected: %v\n", client.conn.RemoteAddr())
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.log("websocket client disconnected: %v\n", client.conn.RemoteAddr())
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if !client.auth {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
