package snowboy

func SetUpHub() *Hub {

	hub := newHub()

	ls := newListener(hub)
	go ls.listening()
	hub.listener = ls

	wsClient := newWsClient(hub)
	hub.wsClient = wsClient

	go hub.run()
	go hub.runSpeaker()
	return hub
}
