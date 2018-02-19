package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/bifurcation/percy"
)

var (
	port         = 4430
	keyFilename  = "../static/key.pem"
	certFilename = "../static/cert.pem"
	htmlFilename = "../static/index.html"
	jsFilename   = "../static/index.js"
	hostField    = "RELAY_IP_FROM_GO_SERVER"
	portField    = "RELAY_PORT_FROM_GO_SERVER"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type NoopTunnel bool

func (tun NoopTunnel) Send(assoc percy.AssociationID, msg []byte) error {
	return nil
}

func (tun NoopTunnel) SendWithProfiles(assoc percy.AssociationID, msg []byte, profiles []percy.ProtectionProfile) error {
	return nil
}

//////////

func localIP() string {
	ifaces, err := net.Interfaces()
	panicOnError(err)

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		panicOnError(err)

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP.To4()
			case *net.IPAddr:
				ip = v.IP.To4()
			}

			if ip == nil || ip[0] == 127 {
				continue
			}

			return ip.String()
		}
	}

	return "127.0.0.1"
}

func httpServer() *http.Server {
	// Read HTML file
	file, err := os.Open(htmlFilename)
	panicOnError(err)

	htmlData, err := ioutil.ReadAll(file)
	panicOnError(err)

	html := string(htmlData)

	// Read JS
	file, err = os.Open(jsFilename)
	panicOnError(err)

	jsData, err := ioutil.ReadAll(file)
	panicOnError(err)

	js := string(jsData)

	// TODO Substitute in address/port in JS
	hostVal := localIP()
	portVal := fmt.Sprintf("%d", port)

	js = strings.Replace(js, hostField, hostVal, -1)
	js = strings.Replace(js, portField, portVal, -1)

	// Start up a web server
	srv := &http.Server{Addr: ":" + portVal}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		io.WriteString(w, html)
	})

	http.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/javascript")
		io.WriteString(w, js)
	})

	go func() {
		srv.ListenAndServeTLS(certFilename, keyFilename)
	}()

	return srv
}

//////////

func main() {
	tunnel := NoopTunnel(false)
	mdd := percy.NewMDD(tunnel)
	err := mdd.Listen(4430)
	panicOnError(err)

	srv := httpServer()

	fmt.Println("Listening, press <enter> to stop")
	var input string
	fmt.Scanln(&input)

	mdd.Stop()
	srv.Shutdown(nil)
}
