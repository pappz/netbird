package client

import (
	"context"
	"io"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/netbirdio/netbird/encryption"
	"github.com/netbirdio/netbird/management/proto"
)

type Dispatcher struct {
	stream         proto.ManagementService_DispatcherClient
	serverPubKey   wgtypes.Key
	privateKey     wgtypes.Key
	ctx            context.Context
	visitorHandler VisitorHandler
}

func newDispatcher(ctx context.Context, stream proto.ManagementService_DispatcherClient, serverPubKey wgtypes.Key, privateKey wgtypes.Key, visitorHandler VisitorHandler) *Dispatcher {
	return &Dispatcher{
		ctx:            ctx,
		stream:         stream,
		serverPubKey:   serverPubKey,
		privateKey:     privateKey,
		visitorHandler: visitorHandler,
	}
}

func (d *Dispatcher) sendHelloMsg() error {
	log.Debugf("send hello msg")
	// replace this with auth message
	helloMsg := &proto.DispatchSessionData{
		SessionId: "hello",
		Data:      []byte("hello"),
	}

	req, err := encryption.EncryptMessage(d.serverPubKey, d.privateKey, helloMsg)
	if err != nil {
		log.Errorf("failed to encrypt hello message: %s", err)
		return err
	}

	encMsg := &proto.EncryptedMessage{
		WgPubKey: d.privateKey.PublicKey().String(),
		Body:     req,
	}
	return d.stream.Send(encMsg)
}

func (d *Dispatcher) receiveDispatchEvents() error {
	for {
		update, err := d.stream.Recv()
		if err == io.EOF {
			log.Debugf("dispatcher stream has been closed by server: %s", err)
			return err
		}
		if err != nil {
			log.Debugf("disconnected from dispatcher stream: %v", err)
			return err
		}

		log.Debugf("got an update message from dispatcher stream")
		decryptedResp := &proto.DispatchSessionData{}
		err = encryption.DecryptMessage(d.serverPubKey, d.privateKey, update.Body, decryptedResp)
		if err != nil {
			log.Errorf("failed decrypting dispatcher message from Management Service: %s", err)
			return err
		}

		err = d.visitorHandler.OnNewMsg(decryptedResp.GetSessionId(), decryptedResp.GetData())
		if err != nil {
			log.Errorf("failed handling an dispatcher message received from Management Service: %v", err.Error())
			return err
		}
	}
}

func (d *Dispatcher) Send(session string, data []byte) error {
	msg := &proto.DispatchSessionData{
		SessionId: session,
		Data:      data,
	}
	req, err := encryption.EncryptMessage(d.serverPubKey, d.privateKey, msg)
	if err != nil {
		log.Errorf("failed to encrypt message: %s", err)
		return err
	}

	encMsg := &proto.EncryptedMessage{
		WgPubKey: d.privateKey.PublicKey().String(),
		Body:     req,
	}
	return d.stream.Send(encMsg)
}
