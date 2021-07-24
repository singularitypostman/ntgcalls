package main

import (
	"encoding/json"
	"fmt"
	"github.com/pion/sdp/v3"
	"strconv"
	"time"
)

func getRemoteDesc(joinResponse Conference, source uint32) string {
	/*
	* This code part is based on this
	* https://github.com/evgeny-nadymov/telegram-react/blob/master/src/Calls/SdpBuilder.js
	 */
	sessionId := uint64(time.Now().Unix())
	remoteDesc := sdp.SessionDescription{
		Version: 0,
		Origin: sdp.Origin{
			Username: "-",
			SessionID: sessionId,
			SessionVersion: 2,
			NetworkType: "IN",
			AddressType: "IP4",
			UnicastAddress: "0.0.0.0",
		},
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
	}
	//Setup attribute
	remoteDesc.WithValueAttribute(sdp.AttrKeyGroup, "BUNDLE 0")
	remoteDesc.WithValueAttribute(sdp.AttrKeyMsidSemantic, sdp.SemanticTokenWebRTCMediaStreams)
	remoteDesc.WithPropertyAttribute(sdp.AttrKeyICELite)

	//Init MediaDescription
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
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address: &sdp.Address{
				Address: joinResponse.Transport.Candidates[1].Ip,
			},
		},
	}
	//Setup of RTCP
	mediaDescription.WithValueAttribute("rtcp","1 IN IP4 0.0.0.0")

	//Setup Candidates
	for cIndex := range joinResponse.Transport.Candidates {
		candidate := joinResponse.Transport.Candidates[cIndex]
		fmt.Println(fmt.Sprintf(
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

	//Setup of Ice Credentials
	mediaDescription.WithICECredentials(joinResponse.Transport.Ufrag, joinResponse.Transport.Pwd)

	//Setup of Fingerprint
	mediaDescription.WithFingerprint(joinResponse.Transport.Fingerprints[0].Hash, joinResponse.Transport.Fingerprints[0].Fingerprint)

	//Setup of DSTRole
	mediaDescription.WithValueAttribute(sdp.AttrKeyConnectionSetup, sdp.ConnectionRolePassive.String())

	//Setup MID
	mediaDescription.WithValueAttribute(sdp.AttrKeyMID, "0")

	//Setup ExtMap
	mediaDescription.WithValueAttribute(sdp.AttrKeyExtMap, fmt.Sprintf("%d %s", sdp.DefExtMapValueABSSendTime, sdp.AudioLevelURI))
	mediaDescription.WithPropertyAttribute(sdp.AttrKeyRecvOnly)
	mediaDescription.WithPropertyAttribute(sdp.AttrKeyRTCPMux)

	//Opus Codec
	mediaDescription.WithCodec(111, "opus", 48000, 2, "minptime=10; useinbandfec=1; usedtx=1")
	mediaDescription.WithValueAttribute("rtcp-fb", "111 transport-cc")

	//PCM16L Codec
	mediaDescription.WithCodec(126, "telephone-event", 8000, 0, "")

	//Apply Media
	remoteDesc.WithMedia(&mediaDescription)
	aa, err := json.Marshal(remoteDesc)
	fmt.Println(string(aa))
	remoteSDP, err := remoteDesc.Marshal()
	if err != nil {
		panic(err)
	}
	return  string(remoteSDP)
}

func getRemoteDescFromJS(joinResponse Conference, source uint32) {
	example_remote := "v=0\r\n" +
		"o=- 1624820738288 2 IN IP4 127.0.0.1\r\n" +
		"s=-\r\n" +
		"t=0 0\r\n" +
		"a=group:BUNDLE 0\r\n" +
		"a=msid-semantic: WMS\r\n" +
		"a=ice-lite\r\n" +
		"m=audio 32001 RTP/SAVPF 111 126\r\n" +
		"c=IN IP4 91.108.9.115\r\n" +
		"a=rtcp:9 IN IP4 0.0.0.0\r\n" +
		"a=candidate:1 1 udp 2130706431 2001:67c:4e8:f102:7:0:285:703 32001 typ host generation 0\r\n" +
		"a=candidate:2 1 udp 2130706431 91.108.9.115 32001 typ host generation 0\r\n" +
		"a=ice-ufrag:9atfa1f97dpq4m\r\n" +
		"a=ice-pwd:3v5ejlciut5vtcrtk9q4u68jrm\r\n" +
		"a=fingerprint:sha-256 BB:06:39:03:AF:C3:82:31:5E:16:20:9B:46:48:0F:21:0A:1D:0C:8B:62:31:79:AC:7F:A0:3C:65:2D:3D:84:44\r\n" +
		"a=setup:passive\r\n" +
		"a=mid:0\r\n" +
		"a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\n" +
		"a=recvonly\r\n" +
		"a=rtcp-mux\r\n" +
		"a=rtpmap:111 opus/48000/2\r\n" +
		"a=rtcp-fb:111 transport-cc\r\n" +
		"a=fmtp:111 minptime=10;usedtx=1;useinbandfec=1\r\n" +
		"a=rtpmap:126 telephone-event/8000\r\n"
	peerConn := sdp.SessionDescription{}
	err := peerConn.Unmarshal([]byte(example_remote))
	if err != nil {
		return 
	}
	aa, err := json.Marshal(peerConn)
	fmt.Println(string(aa))
}