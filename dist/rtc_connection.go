package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
	"strconv"
	"strings"
	"time"
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
			peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
				if connectionState == webrtc.ICEConnectionStateConnected{
					iceConnected <- true
				}else if connectionState == webrtc.ICEConnectionStateFailed{
					iceConnected <- false
				}
			})
			offer, err := peerConnection.CreateOffer(nil)
			if err != nil {
				panic(err)
			}
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
				Source: source,
				InviteHash: inviteHash,
			})
			if err != nil {
				panic(err)
			}
			joinCallResult[chatId]<-string(payload)
			var joinResponse Conference
			err = json.Unmarshal([]byte(<-tgCallResponse[chatId]), &joinResponse)
			if err != nil {
				panic(err)
			}
			sessionId := uint64(time.Now().Unix())
			/*
			* This code part is based on this
			* https://github.com/evgeny-nadymov/telegram-react/blob/master/src/Calls/SdpBuilder.js
			*/
			remoteDesc := sdp.SessionDescription{
				SessionName: "-",
				TimeDescriptions: []sdp.TimeDescription{
					{
						Timing: sdp.Timing{
							StartTime: 0,
							StopTime:  0,
						},
						RepeatTimes: nil,
					},
				},
				Origin: sdp.Origin{
					Username: "-",
					SessionID: sessionId,
					SessionVersion: 2,
					NetworkType: "IN",
					AddressType: "IP4",
					UnicastAddress: "0.0.0.0",
				},
			}
			intPort, err := strconv.Atoi(joinResponse.Transport.Candidates[0].Port)
			if err != nil {
				panic(err)
			}
			mediaDescription := sdp.MediaDescription{
				MediaName: sdp.MediaName{
					Media: "audio",
					Port: sdp.RangedPort{
						Value: intPort,
					},
					Protos: []string{
						"RTP",
						"SAVPF",
					},
				},
			}
			audioLevelUri := sdp.AudioLevelURI
			mediaDescription.WithExtMap(sdp.ExtMap{
				Value:     sdp.DefExtMapValueABSSendTime,
				Direction: sdp.DirectionSendOnly,
				ExtAttr:   &audioLevelUri,
			})
			mediaDescription.WithMediaSource(uint32(source), fmt.Sprintf("%s%d","stream",uint32(source)), fmt.Sprintf("%s%d","audio",uint32(source)), fmt.Sprintf("%s%d","audio",uint32(source)))
			mediaDescription.WithCodec(111, "opus", 48000, 2, "minptime=10; useinbandfec=1; usedtx=1")
			mediaDescription.WithCodec(126, "telephone-event", 8000, 0, "")
			mediaDescription.WithICECredentials(joinResponse.Transport.Ufrag, joinResponse.Transport.Pwd)
			mediaDescription.WithFingerprint(joinResponse.Transport.Fingerprints[0].Hash, joinResponse.Transport.Fingerprints[0].Fingerprint)
			for cIndex := range joinResponse.Transport.Candidates {
				candidate := joinResponse.Transport.Candidates[cIndex]
				mediaDescription.WithValueAttribute("candidate", fmt.Sprintf(
					"%s %s %s %s %s %s typ %s generation %s",
					candidate.Foundation,
					candidate.Component,
					candidate.Protocol,
					candidate.Priority,
					candidate.Ip,
					candidate.Port,
					candidate.Type,
					candidate.Generation,
				))
			}
			remoteDesc.WithMedia(&mediaDescription)
			remoteDesc.WithValueAttribute(sdp.AttrKeyConnectionSetup, sdp.ConnectionRolePassive.String())
			remoteDesc.WithValueAttribute(sdp.AttrKeyMID, "0")
			remoteDesc.WithValueAttribute(sdp.AttrKeyGroup, "BUNDLE 0")
			remoteDesc.WithPropertyAttribute(sdp.AttrKeyICELite)
			remoteDesc.WithValueAttribute(sdp.AttrKeySSRCGroup, fmt.Sprintf("FID %d", uint32(source)))
			remoteSDP, err := remoteDesc.Marshal()
			fmt.Println(string(remoteSDP))
			err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP: string(remoteSDP),
			})
			if err != nil {
				panic(err)
			}
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
			fmt.Println("TRYING")
			<-gatherComplete
			fmt.Println("FINISHED")
			<-iceConnected
			fmt.Println("CONNECTED")
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