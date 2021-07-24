package main

import (
	"fmt"
	"strings"
)

type SdpBuilder struct {
	lines []string
	newLine []string
}

func (sdp *SdpBuilder) join() string{
	return strings.Join(sdp.lines, "\n")
}

func (sdp *SdpBuilder) finalize() string{
	return sdp.join() + "\n"
}

func (sdp *SdpBuilder) add(line string){
	sdp.lines = append(sdp.lines, line)
}

func (sdp *SdpBuilder) push(line string){
	sdp.newLine = append(sdp.newLine, line)
}

func (sdp *SdpBuilder) addJoined(separator string){
	sdp.add(strings.Join(sdp.newLine, separator))
	sdp.newLine = []string{}
}

func (sdp *SdpBuilder) addCandidate(c Candidate) {
	sdp.push("a=candidate:")
	sdp.push(fmt.Sprintf(
		"%s %s %s %s %s %s typ %s generation %s",
		c.Foundation,
		c.Component,
		c.Protocol,
		c.Priority,
		c.Ip,
		c.Port,
		c.Type,
		c.Generation,
	))
	sdp.addJoined("")
}

func (sdp *SdpBuilder) addHeader(SessionId uint64, ssrcs []SSRC){
	sdp.add("v=0")
	sdp.add(fmt.Sprintf("o=- %d 2 IN IP4 0.0.0.0", SessionId))
	sdp.add("s=-")
	sdp.add("t=0 0")
	sdp.add(fmt.Sprintf("a=group:BUNDLE %s",SsrcToString(ssrcs)))
	sdp.add("a=ice-lite")
}

func (sdp *SdpBuilder) addTransport(transport Transport){
	sdp.add(fmt.Sprintf("a=ice-ufrag:%s", transport.Ufrag))
	sdp.add(fmt.Sprintf("a=ice-pwd:%s", transport.Pwd))
	for fingerprint := range transport.Fingerprints {
		sdp.add(
			fmt.Sprintf("a=fingerprint:%s %s", transport.Fingerprints[fingerprint].Hash, transport.Fingerprints[fingerprint].Fingerprint),
		)
		sdp.add("a=setup:passive")
	}
	candidates := transport.Candidates
	for candidate := range candidates {
		sdp.addCandidate(candidates[candidate])
	}
}

func (sdp *SdpBuilder) addSsrcEntry(entry SSRC, transport Transport, isAnswer bool){
	ssrc := entry.Ssrc
	sdp.add(fmt.Sprintf("m=audio %d RTP/SAVPF 111 126", BoolToInt(entry.IsMain)))
	if entry.IsMain {
		sdp.add("c=IN IP4 0.0.0.0")
	}
	sdp.add(fmt.Sprintf("a=mid:%s", entry.toAudioSsrc(entry)))
	if entry.IsMain {
		sdp.addTransport(transport)
	}
	sdp.add("a=rtpmap:111 opus/48000/2")
	sdp.add("a=rtpmap:126 telephone-event/8000")
	sdp.add("a=fmtp:111 minptime=10; useinbandfec=1; usedtx=1")
	sdp.add("a=rtcp:1 IN IP4 0.0.0.0")
	sdp.add("a=rtcp-mux")
	sdp.add("a=rtcp-fb:111 transport-cc")
	sdp.add("a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level")
	if isAnswer {
		sdp.add("a=recvonly")
		return
	} else if entry.IsMain {
		sdp.add("a=sendrecv")
	} else {
		sdp.add("a=sendonly")
		sdp.add("a=bundle-only")
	}
	sdp.add(fmt.Sprintf("a=ssrc-group:FID %d", ssrc))
	sdp.add(fmt.Sprintf("a=ssrc:%d cname:stream%d", ssrc, ssrc))
	sdp.add(fmt.Sprintf("a=ssrc:%d msid:stream%d audio%d", ssrc, ssrc, ssrc))
	sdp.add(fmt.Sprintf("a=ssrc:%d mslabel:audio%d",ssrc, ssrc))
	sdp.add(fmt.Sprintf("a=ssrc:%d label:audio%d", ssrc, ssrc))
}

func (sdp *SdpBuilder) addConference(conference Conference, isAnswer bool){
	ssrcs := conference.Ssrcs
	if isAnswer {
		for ssrc := range ssrcs {
			if ssrcs[ssrc].IsMain {
				ssrcs = []SSRC{ssrcs[ssrc]}
				break
			}
		}
	}
	sdp.addHeader(conference.SessionId, ssrcs)
	for entry := range ssrcs {
		sdp.addSsrcEntry(ssrcs[entry], conference.Transport, isAnswer)
	}
}

func fromConference(conference Conference, isAnswer bool) string{
	sdp := SdpBuilder{}
	sdp.addConference(conference, isAnswer)
	return sdp.finalize()
}