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
	"github.com/gorilla/websocket"
	"github.com/gortc/sdp"
)

var (
	port         = 4430
	keyFilename  = "../static/key.pem"
	certFilename = "../static/cert.pem"
	htmlFilename = "../static/index.html"
	jsFilename   = "../static/index.js"
	portField    = "RELAY_PORT_FROM_GO_SERVER"
	kdServer     = "localhost:2000"
	sdp_answer   = []byte("{\"type\": \"sdp\", \"data\":\"v=0\\r\\n" +
		"o=percy0.2 2633292546686233323 0 IN IP4 0.0.0.0\\r\\n" +
		"s=-\\r\\n" +
		"t=0 0\\r\\n" +
		"a=fingerprint:sha-256 4E:53:20:94:6D:C6:7E:58:7C:8E:F1:08:2A:38:74:59:BF:73:48:56:AB:4D:3F:48:F1:B4:9F:B4:AF:2E:76:75\\r\\n" +
		"a=group:BUNDLE sdparta_0\\r\\n" +
		"a=ice-options:trickle\\r\\n" +
		"a=ice-lite\\r\\n" +
		"a=msid-semantic:WMS *\\r\\n" +
		"m=video 9 UDP/TLS/RTP/SAVPF 120\\r\\n" +
		"c=IN IP4 0.0.0.0\\r\\n" +
		"a=recvonly\\r\\n" +
		"a=extmap:3 urn:ietf:params:rtp-hdrext:sdes:mid\\r\\n" +
		"a=extmap:4 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\\r\\n" +
		"a=extmap:5 urn:ietf:params:rtp-hdrext:toffset\\r\\n" +
		"a=fmtp:120 max-fs=12288;max-fr=60\\r\\n" +
		"a=ice-pwd:abcdefabcdefabcdefabcdefabcdefab\\r\\n" +
		"a=ice-ufrag:fedcbafe\\r\\n" +
		"a=mid:sdparta_0\\r\\n" +
		"a=rtcp-fb:120 nack\\r\\n" +
		"a=rtcp-fb:120 nack pli\\r\\n" +
		"a=rtcp-fb:120 ccm fir\\r\\n" +
		"a=rtcp-fb:120 goog-remb\\r\\n" +
		"a=rtcp-mux\\r\\n" +
		"a=rtpmap:120 VP8/90000\\r\\n" +
		"a=setup:passive\\r\\n\"}")
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type NoopTunnel bool
type NoopMDDTunnel bool

func (tun NoopTunnel) Send(assoc percy.AssociationID, msg []byte) error {
	return nil
}

func (tun NoopTunnel) SendWithProfiles(assoc percy.AssociationID, msg []byte, profiles []percy.ProtectionProfile) error {
	return nil
}

func (mtun NoopMDDTunnel) Send(assoc percy.AssociationID, msg []byte) error {
	return nil
}

func (mtun NoopMDDTunnel) SendWithKeys(assocID percy.AssociationID, msg []byte, profile percy.ProtectionProfile, keys percy.SRTPKeys) error {
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

var upgrader = websocket.Upgrader{} // use default options

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

	hostVal := localIP()
	portVal := fmt.Sprintf("%d", port)

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

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break
			}
			fmt.Println("received SDP offer")
			var (
				s sdp.Session
			)
			if s, err = sdp.DecodeSession(message, s); err != nil {
				fmt.Println("failed to decode SDP session")
				break
			}

			// for _, v := range s {
			// 	fmt.Println(v)
			// }

			d := sdp.NewDecoder(s)
			m := new(sdp.Message)
			if err = d.Decode(m); err != nil {
				fmt.Println("failed to decode")
				break
			}
			//fmt.Println("Offer Origin:", m.Origin)

			// Read the attributes from the session level
			fingerprint_hash := m.Attributes["fingerprint"][0]
			fingerprint := strings.Split(fingerprint_hash, " ")[1]
			fmt.Println("Session.fingerprint: ", fingerprint)

			// Read the attributes from the media section
			if len(m.Medias) < 1 {
				fmt.Println("No media section found")
				break
			}
			ice_pwd := m.Medias[0].Attributes["ice-pwd"][0]
			ice_ufrag := m.Medias[0].Attributes["ice-ufrag"][0]
			fmt.Println("Media[0].ice-pwd: ", ice_pwd)
			fmt.Println("Media[0].ice-ufrag: ", ice_ufrag)

			err = c.WriteMessage(mt, sdp_answer)
			if err != nil {
				fmt.Println("write:", err)
				break
			}

			ice_candidate_answer := []byte("{\"type\": \"ice\", \"data\":{\"candidate\": \"candidate:0 1 UDP 2122121471 " + hostVal + " " + portVal + " typ host\",\"sdpMid\": \"sdparta_0\",\"sdpMLineIndex\": 0}}")

			err = c.WriteMessage(mt, ice_candidate_answer)
			if err != nil {
				fmt.Println("write:", err)
				break
			}
		}
	})

	go func() {
		srv.ListenAndServeTLS(certFilename, keyFilename)
	}()

	return srv
}

//////////

func main() {
	kmfTunnel := NoopTunnel(false)
	mddTunnel := NoopMDDTunnel(false)
	forwarder, err := percy.NewUDPForwarder(mddTunnel, kdServer)
	if err != nil {
		panic("Error creating forwarder")
	}
	mdd := percy.NewMDD(kmfTunnel, forwarder)
	err = mdd.Listen(4430)
	panicOnError(err)

	srv := httpServer()

	fmt.Println("Listening, press <enter> to stop")
	var input string
	fmt.Scanln(&input)

	mdd.Stop()
	srv.Shutdown(nil)
}
