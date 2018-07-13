package percy

import (
	"github.com/fluffy/rtp"
)

const (
	NumSpeakers = 3
)

/* The SessionID uniquely identifiers each session from each endpoint connected to the SFU. If a single user is connected with more than one endpoingpint, they will have differnt ClientID values */
type ClientID uint64

/* This uniquely identifies the confernce the Client is in */
type ConfID uint32

type SFUClient struct {
     lastEnergy  float32
     lastEnergyTime float32
     energy   float32
}


type SFUConf struct {
	clientList     map[ClientID]SFUClient
	activeSpeaker  ClientID
	prevSpeaker    ClientID
	otherSpeakers  []ClientID
	tryingSpeakers []ClientID
}

type SFU struct {
	confIdMap map[ClientID]ConfID
	muteMap   map[ClientID]bool
	confMap   map[ConfID]SFUConf
        audioPTList []uint8 // first one is primary speaker, 2nd the secondary and so on - for now assumes all are opus 
}

func NewSFU( audioPTList []uint8 ) *SFU {
     sfu := new( SFU )
     sfu.audioPTList = audioPTList
     
     return sfu
}

// adds a clients to a confernence
func (sfu *SFU) AddClient(conf ConfID, client ClientID) {
	// TODO
}

// removes a client from a conference
func (sfu *SFU) RemoveClient(conf ConfID, client ClientID) {
	// TODO
}

// removes all clients from a confernence
func (sfu *SFU) StopConf(conf ConfID) {
	// TODO
}

// Mute or Unmute a client in a congference
func (sfu *SFU) Mute(client ClientID, mute bool) {
	sfu.muteMap[client] = mute
}

// Get list of active speakers - first one will be main one, last will be previous speaker
func (sfu *SFU) ActiveSpeakers(conf ConfID) []ClientID {
	// TODO
	return nil
}

type SendPacket struct {
	destination ClientID
	rtpPacket   *rtp.RTPPacket
}

// this processing an incoming packet and returns a list of packet to send to clients
func (sfu *SFU) ProcessPacket(client ClientID, p *rtp.RTPPacket) []SendPacket {
	// TODO
	return nil
}
