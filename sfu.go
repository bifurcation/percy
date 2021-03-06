package percy

import (
	"math"
	"time"
)

const (
	NumSpeakers = 2
)

/* The SessionID uniquely identifiers each session from each endpoint connected to the SFU. If a single user is connected with more than one endpoingpint, they will have differnt ClientID values */
type ClientID uint64

/* This uniquely identifies the confernce the Client is in */
type ConfID uint32

// this keeps track of the energy levels from singl speaker in a confernces
type SFUClient struct {
	lastEnergy     float64 // in dB below zero
	lastEnergyTime time.Time
	energy         float64 // in dB below zero
}

// this keep strack of all the clients in a confernce
type SFUConf struct {
	clientList             map[ClientID]*SFUClient
	speakers               []ClientID // first is active, second is previos, aditional are extra
	activeSpeakerStartTime time.Time
}

type Destination struct {
	clientID ClientID
	pt       int8
}

type Source struct {
	clientID ClientID
	pt       int8
}

// this is a singleton to keep track of all the conferences
type SFU struct {
	confIdMap map[ClientID]ConfID
	confMap   map[ConfID]*SFUConf
	muteMap   map[ClientID]bool

	audioPTList []int8 // first one is primary speaker, 2nd the secondary and so on - for now assumes all are opus

	fibMap map[Source][]Destination
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

// Get Forwarding Map for Packets
func (sfu *SFU) GetFibEntry(clientID ClientID, pt int8) []Destination {

	if pt != sfu.audioPTList[0] {
		pt = 0
	}

	var src Source
	src.pt = pt
	src.clientID = clientID

	return sfu.fibMap[src]
}

// this processing an incoming packet and returns a list of packet to send to clients
func (sfu *SFU) UpdateEnergy(clientID ClientID, dBov int8) {

	// is the client in a conference
	confId, okClient := sfu.confIdMap[clientID]
	if !okClient {
		return
	}

	// does the confernces exist
	conf, okConf := sfu.confMap[confId]
	if !okConf {
		return
	}

	// it this client in that confernce
	client, okClient := conf.clientList[clientID]
	if !okClient {
		return
	}

	// update the energy
	if dBov != 0 {
		client.updateEnergy(dBov)
	}

	// update active speaker list
	sfu.updateSpeakers(conf)

	sfu.updateFIB(conf) // TODO - don't update as often

}

func (client *SFUClient) updateEnergy(dBov int8) {
	if dBov >= 0 {
		return
	}
	db := float64(dBov)

	now := time.Now()
	if now.Sub(client.lastEnergyTime) > time.Duration(1500*time.Millisecond) {
		// just replace old endergy measurements - too old to care
		client.energy = db
	} else {
		// TODO - fix for rolling average later
		client.energy = math.Max(db, 0.2*db+0.8*client.energy)
	}

	client.lastEnergy = db
	client.lastEnergyTime = now
}

func (sfu *SFU) updateSpeakers(conf *SFUConf) {
	if len(conf.speakers) != NumSpeakers {
		conf.speakers = make([]ClientID, NumSpeakers)
	}

	// TODO - roll off old speakers

	// build list of trying speakers
	var trying []ClientID
	var maxEnergy float64 = -1000.0
	var maxEnergyClientID ClientID = 0

	for clientID := range conf.clientList {
		client := conf.clientList[clientID]
		if client.energy > -35.0 {
			trying = append(trying, clientID)
			if client.energy > maxEnergy {
				maxEnergy = client.energy
				maxEnergyClientID = clientID
			}
		}
	}

	// figure out active speaker
	if maxEnergy > -1000.0 {
		now := time.Now()
		if now.Sub(conf.activeSpeakerStartTime) > time.Duration(200*time.Millisecond) {
			if maxEnergyClientID != conf.speakers[0] {
				// switch active speaker
				conf.speakers[1] = conf.speakers[0]  // update previos speaker
				conf.speakers[0] = maxEnergyClientID // new active speaker
				conf.activeSpeakerStartTime = now
			}
		}
	}

	// TODO figure out other speakers to mix in

}

func (sfu *SFU) updateFIB(conf *SFUConf) {
	// do audio forwarnding
	for clientID := range conf.clientList {

		var destList []Destination

		// if it is from a speaker, send it to others clients
		for i := range conf.speakers {
			if conf.speakers[i] == clientID {
				// send it to all others
				for destClientID := range conf.clientList {
					if destClientID != clientID {
						// send to destClientID
						var dest Destination
						dest.clientID = destClientID
						dest.pt = sfu.audioPTList[i]

						destList = append(destList, dest)
					}
				}
			}

		}

		var src Source
		src.pt = 0
		src.clientID = clientID
		sfu.fibMap[src] = destList

	}

	// do video forwarnding
	for clientID := range conf.clientList {
		var destList []Destination

		// if video from active speaker, sent to everyone else
		if conf.speakers[0] == clientID {
			// send it to all others
			for destClientID := range conf.clientList {
				if destClientID != clientID {
					// send to destClientID
					var dest Destination
					dest.clientID = destClientID
					dest.pt = 0

					destList = append(destList, dest)
				}
			}
		}

		// if from prev speaker, send to active speaker
		if conf.speakers[1] == clientID {
			destClientID := conf.speakers[0]
			if destClientID != clientID {
				// send to destClientID
				var dest Destination
				dest.clientID = destClientID
				dest.pt = 0

				destList = append(destList, dest)
			}

		}

		var src Source
		src.pt = 0
		src.clientID = clientID
		sfu.fibMap[src] = destList
	}
}
