package main

import (
	"fmt"
	//"net/http"

	"github.com/bifurcation/percy"
)

var (
	port     = 4430
	htmlFile = "../www/index.html"
	jsFile   = "../www/index.js"
)

type NoopTunnel bool

func (tun NoopTunnel) Send(assoc percy.AssociationID, msg []byte) error {
	return nil
}

func (tun NoopTunnel) SendWithProfiles(assoc percy.AssociationID, msg []byte, profiles []percy.ProtectionProfile) error {
	return nil
}

/*
func httpServer() *http.Server {
	// TODO Read HTML file
	html := "html"

	// TODO Read JS, sub in address / port
	js := "js"

	// TODO Generate a self-signed certificate

	// Start up a web server
	return nil
}
*/

func main() {
	tunnel := NoopTunnel(false)
	mdd := percy.NewMDD(tunnel)
	err := mdd.Listen(4430)
	if err != nil {
		panic(err)
	}
	defer mdd.Stop()

	//srv := httpServer()
	//defer srv.Shutdown(nil)

	fmt.Println("Listening, press <enter> to stop")

	var input string
	fmt.Scanln(&input)
}
