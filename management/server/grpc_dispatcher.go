package server

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/netbirdio/netbird/management/proto"
)

type grpcDispatcherSrv struct {
	peerKey     wgtypes.Key
	grpcEncrypt *grpcEncrypt
	stream      proto.ManagementService_DispatcherServer
}

func newGrpcDispatcherSrv(peerKey wgtypes.Key, grpcEncrypt *grpcEncrypt, stream proto.ManagementService_DispatcherServer) *grpcDispatcherSrv {
	return &grpcDispatcherSrv{
		peerKey:     peerKey,
		grpcEncrypt: grpcEncrypt,
		stream:      stream,
	}
}

func (s *grpcDispatcherSrv) Write(id string, buf []byte) error {
	msg := &proto.DispatchSessionData{
		SessionId: id,
		Data:      buf,
	}

	encMsg, err := s.grpcEncrypt.encryptMsg(s.peerKey, msg)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %s", err)
	}
	err = s.stream.SendMsg(encMsg)
	// todo eof handling
	return err
}

func (s *grpcDispatcherSrv) Read() ([]byte, string, error) {
	encryptedMsg, err := s.stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, "", err
		}
		return nil, "", err
	}

	msg := &proto.DispatchSessionData{}
	_, err = s.grpcEncrypt.parseRequest(encryptedMsg, msg)
	if err != nil {
		log.Debugf("failed to parse request: %s", err)
		return nil, "", fmt.Errorf("failed to read message")
	}

	return msg.GetData(), msg.GetSessionId(), nil
}
