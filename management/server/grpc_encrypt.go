package server

import (
	"fmt"

	pb "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/netbirdio/netbird/encryption"
	"github.com/netbirdio/netbird/management/proto"
)

type grpcEncrypt struct {
	serverKey wgtypes.Key
}

func newGrpcEncrypt(serverKey wgtypes.Key) *grpcEncrypt {
	return &grpcEncrypt{
		serverKey: serverKey,
	}
}

func (e *grpcEncrypt) encryptMsg(peerKey wgtypes.Key, msg pb.Message) (*proto.EncryptedMessage, error) {
	encryptedResp, err := encryption.EncryptMessage(peerKey, e.serverKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %s", err)
	}

	return &proto.EncryptedMessage{
		WgPubKey: e.serverKey.PublicKey().String(),
		Body:     encryptedResp,
	}, nil
}

func (e *grpcEncrypt) parseRequest(req *proto.EncryptedMessage, parsed pb.Message) (wgtypes.Key, error) {
	peerKey, err := wgtypes.ParseKey(req.GetWgPubKey())
	if err != nil {
		log.Warnf("error while parsing peer's WireGuard public key %s.", req.WgPubKey)
		return wgtypes.Key{}, fmt.Errorf("provided wgPubKey %s is invalid", req.WgPubKey)
	}

	err = encryption.DecryptMessage(peerKey, e.serverKey, req.Body, parsed)
	if err != nil {
		return wgtypes.Key{}, fmt.Errorf("invalid request message")
	}

	return peerKey, nil
}
