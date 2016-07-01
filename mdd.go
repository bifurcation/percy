package percy

import (
	"log"
	"net"
)

func addrToString(addr *net.UDPAddr) string {
	// XXX: Something fancier?
	return addr.String()
}

type MDD struct {
	name     string
	addr     *net.UDPAddr
	conn     *net.UDPConn
	clients  map[string]*net.UDPAddr
	stopChan chan bool
}

func NewMDD() *MDD {
	mdd := new(MDD)
	mdd.name = "mdd"
	mdd.clients = map[string]*net.UDPAddr{}
	return mdd
}

func (mdd *MDD) Listen(port int) error {
	var err error

	mdd.addr = &net.UDPAddr{Port: port}
	mdd.conn, err = net.ListenUDP("udp", mdd.addr)
	if err != nil {
		return err
	}

	go func(mdd *MDD) {
		buf := make([]byte, 2048)

		for {
			select {
			case <-mdd.stopChan:
				log.Println("Stopping MDD")
			default:
			}

			n, addr, err := mdd.conn.ReadFromUDP(buf)

			if err != nil {
				log.Printf("Recv Error: %v\n", err)
				continue
			}

			log.Printf("Recv [%s] [%d] [%s] [%s]", mdd.name, n, addr.String(), string(buf[:n]))

			// Remember the client if it's new
			// If we can't make a conn to it, don't forward the packet
			// XXX: Could have an interface to add/remove clients, then
			//      just filter unknown clients here.
			clientID := addrToString(addr)
			if _, ok := mdd.clients[clientID]; !ok {
				mdd.clients[clientID] = addr
			}

			// TODO: if (dtlsPacket) { mdd.tunnel.send(buf) }

			// XXX: Here's where you mess with the packet

			// Send the packet out to all the clients except
			// the one that sent it
			for client, addr := range mdd.clients {
				if client == clientID {
					continue
				}

				_, err := mdd.conn.WriteToUDP(buf[:n], addr)
				if err != nil {
					log.Println("Error forwarding packet")
				}
				log.Printf("Send [%s] [%d] [%s] [%s]", mdd.name, n, client, string(buf[:n]))
			}
		}
	}(mdd)

	log.Println("Started MDD")
	return nil
}

func (mdd *MDD) Stop() {
	mdd.stopChan <- true
}
