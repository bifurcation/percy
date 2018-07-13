package percy

import (
	"github.com/fluffy/rtp"
	"time"
)

const (
	NumSpeakers = 3
)

/* The SessionID uniquely identifiers each session from each endpoint connected to the SFU. If a single user is connected with more than one endpoingpint, they will have differnt ClientID values */
type ClientID uint64

/* This uniquely identifies the confernce the Client is in */
type ConfID uint32

// this keeps track of the energy levels from singl speaker in a confernces 
type SFUClient struct {
	lastEnergy     float32 // in dB below zero
	lastEnergyTime time.Time
	energy         float32 // in dB below zero
}

// this keep strack of all the clients in a confernce 
type SFUConf struct {
	clientList     map[ClientID]*SFUClient
	activeSpeaker  ClientID
	prevSpeaker    ClientID
	otherSpeakers  []ClientID
	tryingSpeakers []ClientID
}

// this is a singleton to keep track of all the conferences 
type SFU struct {
	confIdMap map[ClientID]ConfID
	confMap   map[ConfID]*SFUConf

	muteMap map[ClientID]bool

	audioPTList []uint8 // first one is primary speaker, 2nd the secondary and so on - for now assumes all are opus
}

func NewSFU(audioPTList []uint8) *SFU {
	sfu := new(SFU)
	sfu.audioPTList = audioPTList

	return sfu
}

// removes a client from a conference
func (sfu *SFU) RemoveClient(confID ConfID, clientID ClientID) {
	// remove client from any existing confernces
	conf, ok := sfu.confMap[confID]
	if ok {
		delete(conf.clientList, clientID)
	}
	delete(sfu.confIdMap, clientID)
}

// adds a clients to a confernence
func (sfu *SFU) AddClient(confID ConfID, clientID ClientID) {
	//  create confernce if it does not eist
	conf, ok := sfu.confMap[confID]
	if !ok {
		sfu.confMap[confID] = new(SFUConf)
	}

	sfu.RemoveClient(confID, clientID)

	// add client to this this confernce
	sfu.confIdMap[clientID] = confID
	conf.clientList[clientID] = new(SFUClient)
}

// removes all clients from a confernence
func (sfu *SFU) StopConf(confID ConfID) {
	conf := sfu.confMap[confID]
	if conf != nil {
		for clientID := range conf.clientList {
			sfu.RemoveClient(confID, clientID)
		}
	}
	sfu.confMap[confID] = nil
}

// Mute or Unmute a client in a congference
func (sfu *SFU) Mute(clientID ClientID, mute bool) {
	sfu.muteMap[clientID] = mute
}

// Get list of active speakers - first one will be main one, last will be previous speaker
func (sfu *SFU) ActiveSpeakers(confID ConfID) []ClientID {
	// TODO
	var ret []ClientID
	conf, ok := sfu.confMap[confID]
	if ok {
		ret = append(ret, conf.activeSpeaker)
		ret = append(ret, conf.otherSpeakers...)
		ret = append(ret, conf.prevSpeaker)
	}
	return ret
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
