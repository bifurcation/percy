package main

import (
	"fmt"

	"github.com/bifurcation/percy"
)

type NoopTunnel bool

func (tun NoopTunnel) Send(assoc percy.AssociationID, msg []byte) error {
	return nil
}

func (tun NoopTunnel) SendWithProfiles(assoc percy.AssociationID, msg []byte, profiles []percy.ProtectionProfile) error {
	return nil
}

func main() {
	tunnel := NoopTunnel(false)
	mdd := percy.NewMDD(tunnel)
	err := mdd.Listen(4430)
	if err != nil {
		panic(err)
	}
	defer mdd.Stop()

	fmt.Println("Listening, press <enter> to stop")

	var input string
	fmt.Scanln(&input)
}
