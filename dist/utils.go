package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pion/randutil"
	"github.com/pion/webrtc/v3"
	"io/ioutil"
	"strings"
)

type JoinVoiceCallParams struct {
	ChatId      int     `json:"chat_id"`
	Fingerprint string  `json:"fingerprint"`
	Hash        string  `json:"hash"`
	Setup       string  `json:"setup"`
	Pwd         string  `json:"pwd"`
	Ufrag       string  `json:"ufrag"`
	Source      uint32  `json:"source"`
	InviteHash  string  `json:"invite_hash"`
}
type Transport struct {
	Candidates   []Candidate   `json:"candidates"`
	Fingerprints []Fingerprint `json:"fingerprints"`
	Pwd          string        `json:"pwd"`
	Ufrag        string        `json:"ufrag"`
}
type Candidate struct {
	Component  string    `json:"component"`
	Foundation string    `json:"foundation"`
	Generation string    `json:"generation"`
	Id         string    `json:"id"`
	Ip         string    `json:"ip"`
	Network    string    `json:"network"`
	Port       string    `json:"port"`
	Priority   string    `json:"priority"`
	Protocol   string    `json:"protocol"`
	Type       string    `json:"type"`
}
type Fingerprint struct {
	Fingerprint string `json:"fingerprint"`
	Hash        string `json:"hash"`
	Setup       string `json:"setup"`
}
type SSRC struct {
	Ssrc uint32
	IsMain bool
}

func SsrcToString(ssrcs []SSRC) string{
	var result []string
	for ssrc := range ssrcs{
		result = append(result, ssrcs[ssrc].toAudioSsrc(ssrcs[ssrc]))
	}
	return strings.Join(result, " ")
}

func BoolToInt(p bool) int{
	if p{
		return 1
	}else{
		return 0
	}
}

func (s *SSRC) toAudioSsrc(ssrc SSRC) string{
	if s.IsMain {
		return "0"
	}
	return fmt.Sprintf("audio%d", ssrc.Ssrc)
}

type Conference struct {
	SessionId uint64
	Transport Transport
	Ssrcs []SSRC
}

func newSessionID() uint64 {
	id, _ := randutil.CryptoUint64()
	return id & (^(uint64(1) << 63))
}

func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
func MustReadStdin() webrtc.SessionDescription {
	dat, err := ioutil.ReadFile("bdescription.txt")
	if err != nil {
		panic(err)
	}
	var obj webrtc.SessionDescription
	Decode(string(dat), &obj)
	return obj
}
func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}
