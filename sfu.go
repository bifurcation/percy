package percy

import (
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
	speakers       []ClientID // first is active, second is previos, aditional are extra
	tryingSpeakers []ClientID // clients trying to become active - TODO - do we need this
}

// this is a singleton to keep track of all the conferences
type SFU struct {
	confIdMap map[ClientID]ConfID
	confMap   map[ConfID]*SFUConf

	muteMap map[ClientID]bool

	audioPTList []int8 // first one is primary speaker, 2nd the secondary and so on - for now assumes all are opus
}

func NewSFU(audioPTList []int8) *SFU {
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

// Get list of active speakers - first one will be main one, second will be previous speaker
func (sfu *SFU) ActiveSpeakers(confID ConfID) []ClientID {
	conf, ok := sfu.confMap[confID]
	if !ok {
		return nil
	}

	return conf.speakers
}

type SendPacket struct {
	destClientID ClientID
	pt           int8
}

// this processing an incoming packet and returns a list of packet to send to clients
func (sfu *SFU) ProcessPacket(clientID ClientID, audio bool, dBov int8) []SendPacket {
	var ret []SendPacket

	// is the client in a conference
	confId, okClient := sfu.confIdMap[clientID]
	if !okClient {
		return nil
	}

	// does the confernces exist
	conf, okConf := sfu.confMap[confId]
	if !okConf {
		return nil
	}

	// it this client in that confernce
	client, okClient := conf.clientList[clientID]
	if !okClient {
		return nil
	}

	// is it an audio packet
	if audio {
		// this packet is audio

		// update the energy
		if dBov != 0 {
			client.updateEnergy(dBov)
		}

		// update active speaker list
		conf.updateSpeakers()

		// if it is from an active speaker, send it to others clients
		for i := range conf.speakers {
			if conf.speakers[i] == clientID {
				// send it to all others
				for destClientID := range conf.clientList {
					if destClientID != clientID {
						// send to destClientID
						var dest SendPacket
						dest.destClientID = destClientID
						dest.pt = sfu.audioPTList[i]

						ret = append(ret, dest)
					}
				}
			}

		}

	} else {
		// assume it is video
                
		// if from active speaker, sent to everyone else
		if conf.speakers[0] == clientID {
			// send it to all others
			for destClientID := range conf.clientList {
				if destClientID != clientID {
					// send to destClientID
					var dest SendPacket
					dest.destClientID = destClientID
					dest.pt = 0

					ret = append(ret, dest)
				}
			}
		}

		// if from prev speaker, send to active speaker
		if conf.speakers[1] == clientID {
			destClientID := conf.speakers[0]
			if destClientID != clientID {
				// send to destClientID
				var dest SendPacket
				dest.destClientID = destClientID
				dest.pt = 0

				ret = append(ret, dest)
			}

		}
	}

	return nil
}

func (client *SFUClient) updateEnergy(dBov int8) {
	// TODO
}

func (conf *SFUConf) updateSpeakers() {
	// TODO
}
