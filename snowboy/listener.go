package snowboy

import (
	"log"
)

type Listener struct {
	// 中心
	hub *Hub

	// 听到的声音
	Voice chan []byte

	// 焦点
	Focus bool

	// 静音计算
	StartSlienceCount int
}

// 专注的声音，去除的
func (l *Listener) listening() {
	for {
		select {
		case voice, ok := <-l.Voice:
			if !ok {
				continue
			}
			// 发送音频给到服务
			if l.Focus {
				l.hub.lMsg <- &Broadcast{sender: HUB_MESSAGE_TYPE_LISTENER, msgType: HUB_MESSAGE_TYPE_LISTENER, data: voice}
				log.Print("voice size:", len(voice))
			}
		}
	}
}

func newListener(h *Hub) *Listener {
	return &Listener{hub: h, Voice: make(chan []byte), Focus: false, StartSlienceCount: 0}
}
