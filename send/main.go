package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pion/webrtc/v2"
)

var (
	sendPeer *webrtc.PeerConnection
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

func spamData(dc *webrtc.DataChannel) {
	randomLargeData := make([]byte, 14000)
	rand.Read(randomLargeData)

	for count := 0; ; count++ {
		binary.LittleEndian.PutUint64(randomLargeData, uint64(time.Now().UnixNano()))
		err := dc.Send(randomLargeData)
		checkError(err)

		if count%1000 == 0 {
			time.Sleep(time.Second)
		}
	}
}

func main() {
	sendPeer, err := webrtc.NewPeerConnection(peerConnectionConfig)
	checkError(err)

	dc, err := sendPeer.CreateDataChannel("testing", &webrtc.DataChannelInit{})
	checkError(err)

	dc.OnOpen(func() {
		go spamData(dc)
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		sendTimeNanos := int64(binary.LittleEndian.Uint64(msg.Data))
		fmt.Printf("E2E Latency: %f ms \n", float64(time.Now().UnixNano()-sendTimeNanos)/float64(10e6))
	})

	offer, err := sendPeer.CreateOffer(nil)
	checkError(err)
	err = sendPeer.SetLocalDescription(offer)
	checkError(err)
	offerBuffer := new(bytes.Buffer)
	err = json.NewEncoder(offerBuffer).Encode(offer)
	checkError(err)
	fmt.Printf("Offer:\n%s", offerBuffer.String())

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("Answer> ")
	answerRaw, err := buf.ReadBytes('\n')
	checkError(err)

	var answer webrtc.SessionDescription
	err = json.NewDecoder(bytes.NewBuffer(answerRaw)).Decode(&answer)
	checkError(err)
	sendPeer.SetRemoteDescription(answer)

	dc1, err := sendPeer.CreateDataChannel("testing-1", &webrtc.DataChannelInit{})
	checkError(err)

	dc1.OnOpen(func() {
		go spamData(dc1)
	})
	dc1.OnMessage(func(msg webrtc.DataChannelMessage) {
		sendTimeNanos := int64(binary.LittleEndian.Uint64(msg.Data))
		fmt.Printf("E2E Latency: %f ms \n", float64(time.Now().UnixNano()-sendTimeNanos)/float64(10e6))
	})

	dc2, err := sendPeer.CreateDataChannel("testing-2", &webrtc.DataChannelInit{})
	checkError(err)

	dc2.OnOpen(func() {
		go spamData(dc2)
	})
	dc2.OnMessage(func(msg webrtc.DataChannelMessage) {
		sendTimeNanos := int64(binary.LittleEndian.Uint64(msg.Data))
		fmt.Printf("E2E Latency: %f ms \n", float64(time.Now().UnixNano()-sendTimeNanos)/float64(10e6))
	})

	select {}
}
