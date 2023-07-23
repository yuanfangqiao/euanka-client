package snowboy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"euanka/data"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/gorilla/websocket"
)

const (
	HUB_MESSAGE_TYPE_LISTENER = 100
	HUB_MESSAGE_TYPE_WS       = 200
)

const (
	HUB_SENDER_LISTENER = 10
	HUB_SENDER_WS       = 20
)

var speakerChan chan string = make(chan string, 20)

type Broadcast struct {
	sender  int
	msgType int
	data    []byte
}

type Hub struct {
	msgType  int
	wsClient *WsClient
	listener *Listener

	// Inbound message from the clients
	wsMsg chan *Broadcast
	lMsg  chan *Broadcast
}

func newHub() *Hub {
	return &Hub{
		wsMsg: make(chan *Broadcast),
		lMsg:  make(chan *Broadcast),
	}
}

func (h *Hub) GetListener() *Listener {
	return h.listener
}

func (h *Hub) SendWsClient(msg string) {
	h.wsClient.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (h *Hub) run() {
	for {
		select {
		case lMsg := <-h.lMsg:
			// 发送消息给到ws
			res := converPcm16ToFloat32(lMsg.data)
			h.wsClient.conn.WriteMessage(websocket.BinaryMessage, res)
		case wsMsg := <-h.wsMsg:
			log.Print("ws rev:", string(wsMsg.data))
			var dmRes data.DmData
			json.Unmarshal(wsMsg.data, &dmRes)
			// dm
			if dmRes.Topic == "dm.result" {
				// base64 转码存储
				//decoded, _ := base64.StdEncoding.DecodeString(dmRes.DM.AudioBase64)
				url := dmRes.DM.AudioUrl
				speakerChan <- url
			}
		}
	}
}

func (h *Hub) runSpeaker() {
	for {
		select {
		case url := <-speakerChan:
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("err:", err)
				continue
			}
			if resp == nil {
				log.Printf("resp nill")
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Panicf("resp status is %s", resp.Status)
			}
			decoded, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("err2:", err)
			}

			path := "./tmp/tts01.wav"
			log.Printf("===========", len(decoded))
			err2 := ioutil.WriteFile(path, decoded, 0666)
			if err2 != nil {
				log.Printf("err2:", err2)
			}
			BeepPlayWav(path)
		}
	}
}

func converPcm16ToFloat32(pcmI16 []byte) []byte {

	res := make([]byte, 0)
	end := 0
	for i := 0; i < len(pcmI16); {
		if i+2 < len(pcmI16) {
			end = i + 2
		} else {
			end = len(pcmI16)
		}

		//itemInt16 := binary.LittleEndian.int16(before[i:end])

		//itemInt16 := int(binary.LittleEndian.Uint16(before[i:end]))

		binBuf := bytes.NewBuffer(pcmI16[i:end])

		var x int16
		binary.Read(binBuf, binary.LittleEndian, &x)

		//fmt.Printf("-%X", x)

		itemFloat32 := float32(x) / 32768

		b := Float32ToByte(itemFloat32)

		res = append(res, b...)

		i += 2
	}
	return res
}

func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

var (
	BeepSpeakerInited bool
)

func init() {
	BeepSpeakerInited = false
}

func BeepPlayWav(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("BeepPlayWav err:", err)
		return err
	}
	// Decode the .mp3 File, if you have a .wav file, use wav.Decode(f)
	s, format, _ := wav.Decode(f)

	// Init the Speaker with the SampleRate of the format and a buffer size of 1/10s
	if !BeepSpeakerInited {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			log.Fatalf("BeepPlayWav 2 err:", err)
			return err
		}
		BeepSpeakerInited = true
	}

	log.Printf("==BeepPlayWav=====3344")

	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	// Now we Play our Streamer on the Speaker
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing
	return nil
}
