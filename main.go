package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

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

type CheckItem struct {
	Name  string  `json:"name"`
	Count float64 `json:"count"`
	Cost  float64 `json:"cost"`
	Sum   float64 `json:"sum"`
}

type DisplayString struct {
	Str string `json:"str"`
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

				err := client.WriteMessage(websocket.TextMessage, message)

				if err != nil {
					delete(b.clients, client)
					continue
				}

			}

		case client := <-b.register:
			{
				log.Println("User registered")
				b.clients[client] = true

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

func HandleClear(w http.ResponseWriter, r *http.Request) {

	check, err := json.Marshal("")
	if err != nil {
		log.Println(err.Error)
	}

	body := json.RawMessage(check)

	var msg Message = Message{Type: "clear", Body: &body}

	m, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bus.broadcast <- m

	w.Write(m)

}

func HandleAddCheck(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	params := r.Form

	log.Println("Params: ", params)

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
	bus.broadcast <- m
	w.Write(m)

}

func HandleAddItem(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	params := r.Form

	if len(params) == 0 {
		w.Write([]byte("No params"))
		return
	}

	log.Println(params)

	count, err := strconv.ParseFloat(params.Get("count"), 32)

	if err != nil {
		log.Println(err.Error)
		return
	}

	cost, err := strconv.ParseFloat(params.Get("cost"), 32)

	if err != nil {
		log.Println(err.Error)
		return
	}

	sum, err := strconv.ParseFloat(params.Get("sum"), 32)

	if err != nil {
		log.Println(err.Error)
		return
	}

	strings, err := json.Marshal(
		CheckItem{Name: params.Get("name"),
			Count: count,
			Cost:  cost,
			Sum:   sum})
	if err != nil {
		log.Println(err.Error)
	}

	body := json.RawMessage(strings)

	var msg Message = Message{Type: "addItem", Body: &body}

	m, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bus.broadcast <- m
	w.Write(m)

}

func HandleAddString(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	params := r.Form

	if len(params) == 0 {
		w.Write([]byte("No params"))
		return
	}

	log.Println(params)

	strings, err := json.Marshal(DisplayString{Str: params.Get("strings")})
	if err != nil {
		log.Println(err.Error)
	}

	body := json.RawMessage(strings)

	var msg Message = Message{Type: "addString", Body: &body}

	m, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bus.broadcast <- m
	w.Write(m)

}

func HandleWS(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	bus.register <- ws

}

func PingWS() {

	for {
		<-time.After(5 * time.Second)

		check, err := json.Marshal("")
		if err != nil {
			log.Println(err.Error)
		}

		body := json.RawMessage(check)

		var msg Message = Message{Type: "ping", Body: &body}

		m, err := json.Marshal(msg)
		if err != nil {
			log.Println(err.Error())
			return
		}
		bus.broadcast <- m

	}

}

func main() {

	bus = NewBus()
	go bus.Run()
	go PingWS()

	port := "8080"

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/clear", HandleClear)
	http.HandleFunc("/addcheck", HandleAddCheck)
	http.HandleFunc("/addstring", HandleAddString)
	http.HandleFunc("/additem", HandleAddItem)
	http.HandleFunc("/ws", HandleWS)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))

	http.ListenAndServe(":"+port, nil)

}
