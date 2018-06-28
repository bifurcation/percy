package percy

import (
	"net"
	"fmt"
	"log"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha1"
	"hash/crc32"
	"github.com/bifurcation/mint/syntax"
)

func u32intToBytes(val uint32) []byte{
	return []byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
}

const STUN_COOKIE = 0x2112A442

type STUNMessageType uint16
const (
	MSG_BINDING STUNMessageType = 0x001
	MSG_ALLOCATE STUNMessageType = 0x003
	MSG_REFRESH STUNMessageType = 0x004
	MSG_SEND STUNMessageType = 0x006
	MSG_DATA STUNMessageType = 0x007
	MSG_CREATE_PERMISSION STUNMessageType = 0x008
	MSG_CHANNEL_BIND STUNMessageType = 0x009
	MSG_CONNECT STUNMessageType = 0x00A
	MSG_CONNECTION_BIND STUNMessageType = 0x00B
	MSG_CONNECTION_ATTEMPT STUNMessageType = 0x00C
)

func (smt STUNMessageType) String() string {
	switch smt {
		case MSG_BINDING:
			return "Binding"
		case MSG_ALLOCATE:
			return "Allocate"
		case MSG_REFRESH:
			return "Refresh"
		case MSG_SEND:
			return "Send"
		case MSG_DATA:
			return "Data"
		case MSG_CREATE_PERMISSION:
			return "CreatePermission"
		case MSG_CHANNEL_BIND:
			return "ChannelBind"
		case MSG_CONNECT:
			return "Connect"
		case MSG_CONNECTION_BIND:
			return "ConnectionBind"
		case MSG_CONNECTION_ATTEMPT:
			return "ConnectionAttempt"
		default:
			return fmt.Sprintf("<0x%x>", uint16(smt))
	}
}


type STUNAttrType uint16
const (
	ATTR_MAPPED_ADDRESS STUNAttrType = 0x0001
	ATTR_CHANGE_REQUEST STUNAttrType = 0x0003
	ATTR_USERNAME STUNAttrType = 0x0006
	ATTR_MESSAGE_INTEGRITY STUNAttrType = 0x0008
	ATTR_ERROR_CODE STUNAttrType = 0x0009
	ATTR_UNKNOWN_ATTRIBUTES STUNAttrType = 0x000A
	ATTR_CHANNEL_NUMBER STUNAttrType = 0x000C
	ATTR_LIFETIME STUNAttrType = 0x000D
	ATTR_XOR_PEER_ADDRESS STUNAttrType = 0x0012
	ATTR_DATA STUNAttrType = 0x0013
	ATTR_REALM STUNAttrType = 0x0014
	ATTR_NONCE STUNAttrType = 0x0015
	ATTR_XOR_RELAYED_ADDRESS STUNAttrType = 0x0016
	ATTR_REQUESTED_ADDRESS_FAMILY STUNAttrType = 0x0017
	ATTR_EVEN_PORT STUNAttrType = 0x0018
	ATTR_REQUESTED_TRANSPORT STUNAttrType = 0x0019
	ATTR_DONT_FRAGMENT STUNAttrType = 0x001A
	ATTR_ACCESS_TOKEN STUNAttrType = 0x001B
	ATTR_XOR_MAPPED_ADDRESS STUNAttrType = 0x0020
	ATTR_RESERVATION_TOKEN STUNAttrType = 0x0022
	ATTR_PRIORITY STUNAttrType = 0x0024
	ATTR_USE_CANDIDATE STUNAttrType = 0x0025
	ATTR_PADDING STUNAttrType = 0x0026
	ATTR_RESPONSE_PORT STUNAttrType = 0x0027
	ATTR_CONNECTION_ID STUNAttrType = 0x002A
	ATTR_SOFTWARE STUNAttrType = 0x8022
	ATTR_ALTERNATE_SERVER STUNAttrType = 0x8023
	ATTR_TRANSACTION_TRANSMIT_COUNTER STUNAttrType = 0x8025
	ATTR_CACHE_TIMEOUT STUNAttrType = 0x8027
	ATTR_FINGERPRINT STUNAttrType = 0x8028
	ATTR_ICE_CONTROLLED STUNAttrType = 0x8029
	ATTR_ICE_CONTROLLING STUNAttrType = 0x802A
	ATTR_RESPONSE_ORIGIN STUNAttrType = 0x802B
	ATTR_OTHER_ADDRESS STUNAttrType = 0x802C
	ATTR_ECN_CHECK STUNAttrType = 0x802D
	ATTR_THIRD_PARTY_AUTHORIZATION STUNAttrType = 0x802E
	ATTR_UNASSIGNED STUNAttrType = 0x802F
	ATTR_MOBILITY_TICKET STUNAttrType = 0x8030
	ATTR_CISCO_STUN_FLOWDATA STUNAttrType = 0xC000
	ATTR_ENF_FLOW_DESCRIPTION STUNAttrType = 0xC001
	ATTR_ENF_NETWORK_STATUS STUNAttrType = 0xC002
)

func (sat STUNAttrType) String() string {
	switch sat {
		case ATTR_MAPPED_ADDRESS:
			return "MAPPED-ADDRESS"
		case ATTR_CHANGE_REQUEST:
			return "CHANGE-REQUEST"
		case ATTR_USERNAME:
			return "USERNAME"
		case ATTR_MESSAGE_INTEGRITY:
			return "MESSAGE-INTEGRITY"
		case ATTR_ERROR_CODE:
			return "ERROR-CODE"
		case ATTR_UNKNOWN_ATTRIBUTES:
			return "UNKNOWN-ATTRIBUTES"
		case ATTR_CHANNEL_NUMBER:
			return "CHANNEL-NUMBER"
		case ATTR_LIFETIME:
			return "LIFETIME"
		case ATTR_XOR_PEER_ADDRESS:
			return "XOR-PEER-ADDRESS"
		case ATTR_DATA:
			return "DATA"
		case ATTR_REALM:
			return "REALM"
		case ATTR_NONCE:
			return "NONCE"
		case ATTR_XOR_RELAYED_ADDRESS:
			return "XOR-RELAYED-ADDRESS"
		case ATTR_REQUESTED_ADDRESS_FAMILY:
			return "REQUESTED-ADDRESS-FAMILY"
		case ATTR_EVEN_PORT:
			return "EVEN-PORT"
		case ATTR_REQUESTED_TRANSPORT:
			return "REQUESTED-TRANSPORT"
		case ATTR_DONT_FRAGMENT:
			return "DONT-FRAGMENT"
		case ATTR_ACCESS_TOKEN:
			return "ACCESS-TOKEN"
		case ATTR_XOR_MAPPED_ADDRESS:
			return "XOR-MAPPED-ADDRESS"
		case ATTR_RESERVATION_TOKEN:
			return "RESERVATION-TOKEN"
		case ATTR_PRIORITY:
			return "PRIORITY"
		case ATTR_USE_CANDIDATE:
			return "USE-CANDIDATE"
		case ATTR_PADDING:
			return "PADDING"
		case ATTR_RESPONSE_PORT:
			return "RESPONSE-PORT"
		case ATTR_CONNECTION_ID:
			return "CONNECTION-ID"
		case ATTR_SOFTWARE:
			return "SOFTWARE"
		case ATTR_ALTERNATE_SERVER:
			return "ALTERNATE-SERVER"
		case ATTR_TRANSACTION_TRANSMIT_COUNTER:
			return "TRANSACTION_TRANSMIT_COUNTER"
		case ATTR_CACHE_TIMEOUT:
			return "CACHE-TIMEOUT"
		case ATTR_FINGERPRINT:
			return "FINGERPRINT"
		case ATTR_ICE_CONTROLLED:
			return "ICE-CONTROLLED"
		case ATTR_ICE_CONTROLLING:
			return "ICE-CONTROLLING"
		case ATTR_RESPONSE_ORIGIN:
			return "RESPONSE-ORIGIN"
		case ATTR_OTHER_ADDRESS:
			return "OTHER-ADDRESS"
		case ATTR_ECN_CHECK:
			return "ECN-CHECK"
		case ATTR_THIRD_PARTY_AUTHORIZATION:
			return "THIRD-PARTY-AUTHORIZATION"
		case ATTR_MOBILITY_TICKET:
			return "MOBILITY-TICKET"
		case ATTR_CISCO_STUN_FLOWDATA:
			return "CISCO-STUN-FLOWDATA"
		case ATTR_ENF_FLOW_DESCRIPTION:
			return "ENF-FLOW-DESCRIPTION"
		case ATTR_ENF_NETWORK_STATUS:
			return "ENF-NETWORK-STATUS"
		default:
			return fmt.Sprintf("<0x%x>", uint16(sat))
	}
}

type TransactionID [12]byte

type STUNHeader struct {
	Type STUNMessageType
	Length uint16
	Cookie uint32
	TxnID TransactionID
}

func (id TransactionID) String() string {
	return hex.EncodeToString(id[:])
}

func (hdr STUNHeader) String() string {
	return fmt.Sprintf("STUN %v, TXN ID = %v", hdr.Type, hdr.TxnID)
}

type STUNAttribute struct {
	Tag STUNAttrType
	Value []byte `tls:"head=2"`
}

// Format those values that are easier to read in forms other than byte arrays
func (attr STUNAttribute) String() string {
	val := fmt.Sprintf("  %v = ", attr.Tag)
	switch attr.Tag {
		case ATTR_ERROR_CODE:
			val += fmt.Sprintf("%d%02.2d %v", attr.Value[2], attr.Value[3], string(attr.Value[4:]))
		case ATTR_USERNAME:
			val += string(attr.Value)
		case ATTR_MESSAGE_INTEGRITY:
			val += hex.EncodeToString(attr.Value)
		case ATTR_FINGERPRINT:
			val += hex.EncodeToString(attr.Value)
		default:
			val += fmt.Sprintf("%v",attr.Value)
	}
	return val
}

type MessageType uint16
const (
	MSG_TYPE_REQUEST MessageType = 0x0000
	MSG_TYPE_INDICATION MessageType = 0x0010
	MSG_TYPE_SUCCESS MessageType = 0x0100
	MSG_TYPE_ERROR MessageType = 0x0110
)
func (mt MessageType) String() string {
	switch mt {
		case MSG_TYPE_REQUEST:
			return "Request"
		case MSG_TYPE_INDICATION:
			return "Indication"
		case MSG_TYPE_SUCCESS:
			return "Success"
		case MSG_TYPE_ERROR:
			return "Error"
		default:
			return fmt.Sprintf("<0x%x>", uint16(mt))
	}
}

type STUNMessage struct {
	header STUNHeader
	msgType MessageType
	attributes []STUNAttribute
	// This is used for proper computation of the MESSAGE-INTEGRITY attribute
	icePassword string
}

func (msg STUNMessage) String() string {
	var val string = fmt.Sprintf("%v: %v", msg.msgType, msg.header)
	for _,v := range msg.attributes {
		val += fmt.Sprintf("\n  %v", v)
	}
	return val
}

func ParseSTUN (msg []byte) (*STUNMessage, error) {
	// TODO: validate MESSAGE-INTEGRITY and FINGERPRINT -- see RFC5245 §7.2
	request := STUNMessage{}

	used, err := syntax.Unmarshal(msg, &request.header)
	msg = msg[used:]

	if err != nil || request.header.Cookie != STUN_COOKIE {
		return &request, err
	}

	// Fixup message type
	request.msgType = MessageType(uint16(request.header.Type) & 0x0110)
	request.header.Type &= 0xFEEF

	for len(msg) > 0 {
		attr := STUNAttribute{}
		_, err = syntax.Unmarshal(msg, &attr)
		if err != nil {
			log.Printf("Error parsing STUN attribute: %v", msg)
			return &request, err
		}
		skip := ((len(attr.Value) + 7) / 4 ) * 4
		msg = msg[skip:]
		request.attributes = append(request.attributes, attr)
	}
	return &request, nil
}

func (msg *STUNMessage) Serialize() ([]byte, error){
	msg.header.Cookie = STUN_COOKIE

	msg.header.Type |= STUNMessageType(msg.msgType)
	result, err := syntax.Marshal(&msg.header)
	msg.header.Type &= 0xFEEF

	for i, a := range msg.attributes {

		// Fixup those attributes whose value relies on the rest of the message
		switch a.Tag {
			case ATTR_MESSAGE_INTEGRITY:
				// Increase the length by 24 to account for the size of this attribute
				// (24 bytes: 2 byte tag, 2 byte length, 20 byte value)
				result[2] = byte((len(result) - 20 + 24) >> 8);
				result[3] = byte((len(result) - 20 + 24) & 0xFF);
				mac := hmac.New(sha1.New, []byte(msg.icePassword))
				mac.Write(result)
				a.Value = mac.Sum(nil)
				msg.attributes[i] = a
			case ATTR_FINGERPRINT:
				// Increase the length by 8 to account for the size of this attribute
				// (8 bytes: 2 byte tag, 2 byte length, 4 byte value)
				result[2] = byte((len(result) - 20 + 8) >> 8);
				result[3] = byte((len(result) - 20 + 8) & 0xFF);
				IEEETable := crc32.MakeTable(crc32.IEEE)
				checksum := crc32.Checksum(result, IEEETable)
				a.Value = u32intToBytes(checksum ^ 0x5354554e)
				msg.attributes[i] = a
		}

		attr, _ := syntax.Marshal(&a)
		result = append(result, attr...)
		// Pad to even 32-bit boundary
		for len(result) % 4 != 0 {
			result = append(result, 0)
		}
	}

	// Fixup stun.header.Length
	result[2] = byte((len(result) - 20) >> 8);
	result[3] = byte((len(result) - 20) & 0xFF);

	return result, err
}

func (msg *STUNMessage) Add(tag STUNAttrType, value []byte) {
	attr := STUNAttribute{ Tag: tag, Value: value }
	msg.attributes = append(msg.attributes, attr)
}

func (msg *STUNMessage) AddErrorCode(code uint, reason string) {
	msg.Add(ATTR_ERROR_CODE, append([]byte{0, 0, byte(code / 100), byte(code % 100)}, []byte(reason)...))
}

func MakeMappedAddress(addr *net.UDPAddr) []byte {
	var family byte
	var address []byte
	if address = addr.IP.To4(); address != nil {
		family = 1
	} else {
	  address = addr.IP
		family = 2
	}
	port := addr.Port
	return append([]byte{0, family, byte(port >> 8), byte(port)}, address...)
}

func (msg *STUNMessage) AddMappedAddress(addr *net.UDPAddr) {
	msg.Add(ATTR_MAPPED_ADDRESS, MakeMappedAddress(addr))
}

func (msg *STUNMessage) AddXorMappedAddress(addr *net.UDPAddr) {
	messageHeader,err := syntax.Marshal(&msg.header)
	if err != nil {
		log.Println("Could not serialize STUN message header when adding XOR Mapped Address")
		return
	}
	mappedAddress := MakeMappedAddress(addr)

	// XOR Port with first two bytes of magic cookie (starts at header byte 4)
	mappedAddress[2] ^= messageHeader[4]
	mappedAddress[3] ^= messageHeader[5]

	// XOR Address with magic cookie (and transaction ID, if necessary)
	for i := 0; i < len(mappedAddress) - 4; i++ {
		mappedAddress[i+4] ^= messageHeader[i+4]
	}

	msg.Add(ATTR_XOR_MAPPED_ADDRESS, mappedAddress)
}

func (msg *STUNMessage) AddMessageIntegrity() {
	// We leave this empty, as it will be calculated during serialization
	msg.Add(ATTR_MESSAGE_INTEGRITY, []byte{})
}

func (msg *STUNMessage) AddFingerprint() {
	// We leave this empty, as it will be calculated during serialization
	msg.Add(ATTR_FINGERPRINT, []byte{})
}
