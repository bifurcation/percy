package percy

type AssociationID [16]byte

type ProtectionProfile uint16

type SRTPKeys struct {
	MasterKeyID []byte
	ClientKey   []byte
	ServerKey   []byte
	ClientSalt  []byte
	ServerSalt  []byte
}

type KMFTunnel interface {
	Send(assoc AssociationID, msg []byte) error
	SendWithProfiles(assoc AssociationID, msg []byte, profiles []ProtectionProfile) error
}

type MDDTunnel interface {
	Send(assoc AssociationID, msg []byte) error
	SendWithKeys(assoc AssociationID, msg []byte, profile ProtectionProfile, keys SRTPKeys) error
}
