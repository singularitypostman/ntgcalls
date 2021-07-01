package main

type JoinVoiceCallParams struct {
	ChatId      int     `json:"chat_id"`
	Fingerprint string  `json:"fingerprint"`
	Hash        string  `json:"hash"`
	Setup       string  `json:"setup"`
	Pwd         string  `json:"pwd"`
	Ufrag       string  `json:"ufrag"`
	Source      int     `json:"source"`
	InviteHash  string  `json:"invite_hash"`
}
type Conference struct {
	Transport Transport `json:"transport"`
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