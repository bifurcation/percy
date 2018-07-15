package percy

import (
	// "fmt"
	"testing"
)

func TestSpeakerSelection(t *testing.T) {
	sfu := NewSFU([]int8{109})

	var confID ConfID = 1
	var clientID ClientID = 10

	sfu.AddClient(confID, clientID)
	sfu.AddClient(confID, 11)
	sfu.AddClient(confID, 12)

	sfu.UpdateEnergy(10, -10)
	sfu.UpdateEnergy(11, -12)
	sfu.UpdateEnergy(12, -12)

	d1 := sfu.GetFibEntry(10, 109)
	t.Logf("d1: %xv", d1)

}
