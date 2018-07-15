package percy

import (
	// "fmt"
	"testing"
	"time"
)

func TestSpeakerSelection(t *testing.T) {
	sfu := NewSFU([]int8{109})

	var confID ConfID = 1

	sfu.AddClient(confID, 10)
	sfu.AddClient(confID, 11)
	sfu.AddClient(confID, 12)

	sfu.UpdateEnergy(10, -10)
	sfu.UpdateEnergy(11, -12)
	sfu.UpdateEnergy(12, -11)

	time.Sleep(150 * time.Millisecond)

	sfu.UpdateEnergy(10, -10)
	sfu.UpdateEnergy(11, -12)
	sfu.UpdateEnergy(12, -11)

	time.Sleep(150 * time.Millisecond)

	sfu.UpdateEnergy(10, -10)
	sfu.UpdateEnergy(11, -12)
	sfu.UpdateEnergy(12, -11)

	d1 := sfu.GetFibEntry(10, 0)
	t.Logf("d1: %v \n", d1)
	d2 := sfu.GetFibEntry(11, 0)
	t.Logf("d2: %v \n", d2)
	d3 := sfu.GetFibEntry(12, 0)
	t.Logf("d3: %v \n", d3)
}
