package ws

type Hub struct {
  Clients   map[chan []byte]bool
  Broadcast chan []byte
}

func NewHub() *Hub {
  return &Hub{
    Clients:   make(map[chan []byte]bool),
    Broadcast: make(chan []byte),
  }
}

func (h *Hub) Run() {
  for msg := range h.Broadcast {
    for client := range h.Clients {
      client <- msg
    }
  }
}
