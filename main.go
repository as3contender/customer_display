package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var bus *Bus

type Message struct {
	Type string           `json:"type"`
	Body *json.RawMessage `json:"body"`
}

type Check struct {
	Number string `json:"number"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Bus struct {
	register  chan *websocket.Conn
	broadcast chan []byte
	clients   map[*websocket.Conn]bool
}

func (b *Bus) Run() {
	for {
		select {
		case message := <-b.broadcast:
			for client := range b.clients {

				log.Println(message)

				w, err := client.NextWriter(websocket.TextMessage)
				if err != nil {
					delete(b.clients, client)
					continue
				}

				w.Write(message)

			}

		case client := <-b.register:
			{
				log.Println("User registered")
				b.clients[client] = true

				_, err := client.NextWriter(websocket.TextMessage)
				if err != nil {
					delete(b.clients, client)
					continue
				}

			}
		}
	}
}

func NewBus() *Bus {
	return &Bus{
		register:  make(chan *websocket.Conn),
		broadcast: make(chan []byte),
		clients:   make(map[*websocket.Conn]bool),
	}
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {

	f, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
	}

	w.Write(f)

}

func HandleAddCheck(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	params := r.Form

	log.Println(params)

	check, err := json.Marshal(Check{Number: "1504"})
	if err != nil {
		log.Println(err.Error)
	}

	body := json.RawMessage(check)

	var msg Message = Message{Type: "newCheck", Body: &body}

	m, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bus.broadcast <- []byte("hello world")
	w.Write(m)

}

func HandleWS(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("new client ")
	bus.register <- ws

}

func main() {

	bus = NewBus()
	go bus.Run()

	port := "8080"

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/addcheck", HandleAddCheck)
	http.HandleFunc("/ws", HandleWS)

	http.ListenAndServe(":"+port, nil)
}
