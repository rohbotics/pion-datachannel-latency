package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pion/webrtc/v2"
)

var (
	receivePeer *webrtc.PeerConnection
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

var peerConnectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
	SDPSemantics:       webrtc.SDPSemanticsUnifiedPlanWithFallback,
	ICETransportPolicy: webrtc.ICETransportPolicyAll,
}

func main() {
	receivePeer, err := webrtc.NewPeerConnection(peerConnectionConfig)
	checkError(err)

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("Offer> ")
	offerRaw, err := buf.ReadBytes('\n')
	checkError(err)

	var offer webrtc.SessionDescription
	err = json.NewDecoder(bytes.NewBuffer(offerRaw)).Decode(&offer)
	checkError(err)
	receivePeer.SetRemoteDescription(offer)

	answer, err := receivePeer.CreateAnswer(nil)
	checkError(err)
	err = receivePeer.SetLocalDescription(answer)
	checkError(err)
	answerBuffer := new(bytes.Buffer)
	err = json.NewEncoder(answerBuffer).Encode(answer)
	checkError(err)
	fmt.Printf("Answer:\n%s", answerBuffer.String())

	receivePeer.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			sendTimeNanos := int64(binary.LittleEndian.Uint64(msg.Data))
			fmt.Printf("Data Latency: %f ms \n", float64(time.Now().UnixNano()-sendTimeNanos)/float64(10e6))

			dc.Send(msg.Data)
		})
	})

	select {}
}
