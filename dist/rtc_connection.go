package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/webrtc/v3"
	"strconv"
	"strings"
)

//export joinVoiceCall
func joinVoiceCall(chatId int, inviteHash string) bool{
	if joinCallResult != nil{
		joinCallResult[chatId] = make(chan string)
		tgCallResponse[chatId] = make(chan string)
		go func() {
			iceConnected := make(chan bool)
			opusParams, err := opus.NewParams()
			if err != nil {
				panic(err)
			}
			opusParams.Latency = opus.Latency20ms
			codecSelector := mediadevices.NewCodecSelector(
				mediadevices.WithAudioEncoders(&opusParams),
			)
			mediaEngine := webrtc.MediaEngine{}
			codecSelector.Populate(&mediaEngine)
			api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
			peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
			if err != nil{
				panic(err)
			}

			peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
				if connectionState == webrtc.ICEConnectionStateConnected{
					iceConnected <- true
				}else if connectionState == webrtc.ICEConnectionStateFailed{
					iceConnected <- false
				}
				fmt.Printf("Connection State has changed %s \n", connectionState.String())
			})
			audioSource, err := NewAudioFile("audio.raw")
			if err != nil{
				panic(err)
			}
			audioTrack := mediadevices.NewAudioTrack(audioSource, codecSelector)
			audioTrack.OnEnded(func(err error) {
				fmt.Printf("Track (ID: %s) ended with error: %v\n",
					audioTrack.ID(), err)
			})
			_, err = peerConnection.AddTransceiverFromTrack(audioTrack,
				webrtc.RTPTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionSendonly,
				},
			)
			/*s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
				Audio: func(c *mediadevices.MediaTrackConstraints) {},
				Codec: codecSelector,
			})
			if err != nil{
				panic(err)
			}
			for _, audioTrack := range s.GetTracks(){
				audioTrack.OnEnded(func(err error) {
					fmt.Printf("Track (ID: %s) ended with error: %v\n",
						audioTrack.ID(), err)
				})
				_, err = peerConnection.AddTransceiverFromTrack(audioTrack,
					webrtc.RTPTransceiverInit{
						Direction: webrtc.RTPTransceiverDirectionSendonly,
					},
				)
			}*/
			offer, err := peerConnection.CreateOffer(nil)
			err = peerConnection.SetLocalDescription(offer)

			sdpConn, err := offer.Unmarshal()
			//goland:noinspection GoNilness
			source, err := strconv.Atoi(strings.Split(sdpConn.MediaDescriptions[0].Attributes[8].Value, " ")[0])
			if err != nil {
				panic(err)
			}
			payload, err := json.Marshal(JoinVoiceCallParams{
				ChatId: chatId,
				Fingerprint: strings.Split(sdpConn.Attributes[0].Value, " ")[1],
				Hash: strings.Split(sdpConn.Attributes[0].Value, " ")[0],
				Setup: "active",
				Pwd: sdpConn.MediaDescriptions[0].Attributes[3].Value,
				Ufrag: sdpConn.MediaDescriptions[0].Attributes[2].Value,
				Source: uint32(source),
				InviteHash: inviteHash,
			})
			if err != nil {
				panic(err)
			}
			joinCallResult[chatId]<-string(payload)
			var joinResponse Transport
			err = json.Unmarshal([]byte(<-tgCallResponse[chatId]), &joinResponse)
			if err != nil {
				panic(err)
			}

			sdpString := fromConference(Conference{
				SessionId: newSessionID(),
				Transport: joinResponse,
				Ssrcs: []SSRC{{Ssrc: uint32(source), IsMain: true}},
			}, true)
			err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP: sdpString,
			})
			/*err = peerConnection.SetRemoteDescription(MustReadStdin())
			if err != nil {
				panic(err)
			}*/
			/*answer, err := peerConnection.CreateAnswer(nil)
			err = peerConnection.SetLocalDescription(answer)
			fmt.Println(Encode(*peerConnection.LocalDescription()))*/
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
			fmt.Println("TRYING")
			<-gatherComplete
			fmt.Println("FINISHED")
			//getRemoteDescFromJS(joinResponse, uint32(source))
			<-iceConnected
			fmt.Println("CONNECTED")
			fmt.Println(peerConnection.RemoteDescription())
		}()
		return true
	}else{
		return false
	}
}


//export sendResponseCall
func sendResponseCall(chatId int, result string) bool{
	if joinCallResult != nil && joinCallResult[chatId] != nil{
		tgCallResponse[chatId] <- result
		return true
	}else{
		return false
	}
}

//export waitRequestJoin
func waitRequestJoin(chatId int) *C.char{
	if joinCallResult != nil && joinCallResult[chatId] != nil{
		return C.CString(<-joinCallResult[chatId])
	}else{
		return C.CString("{}")
	}
}