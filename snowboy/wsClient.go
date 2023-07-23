package snowboy

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	hub  *Hub
	conn *websocket.Conn
}

func newWsClient(h *Hub) *WsClient {
	// tts
	u := url.URL{Scheme: "ws", Host: "192.168.1.16:8080", Path: "/camera/uploader/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("tts connecting to %s", u.String())
	if err != nil {
		log.Fatal("dial:", err)
	}
	go func() {
		defer c.Close()
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// asr message
			log.Printf("asr recv: %s", message)
			// send asr to  hub
			h.wsMsg <- &Broadcast{sender: HUB_MESSAGE_TYPE_WS, msgType: messageType, data: message}
		}
	}()

	return &WsClient{
		hub:  h,
		conn: c,
	}
}
