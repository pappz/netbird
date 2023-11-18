package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pappz/dispatcher"
	log "github.com/sirupsen/logrus"

	"github.com/netbirdio/netbird/management/server"
	"github.com/netbirdio/netbird/management/server/http/util"
	"github.com/netbirdio/netbird/management/server/jwtclaims"
	"github.com/netbirdio/netbird/management/server/status"
)

type DispatcherHandler struct {
	accountManager  server.AccountManager
	peerStore       *dispatcher.Store
	claimsExtractor *jwtclaims.ClaimsExtractor
	sessionWriter   *dispatcher.SessionWriter
	ws              *websocket.Conn
}

func NewDispatcherHandler(accountManager server.AccountManager, authCfg AuthCfg, peerStore *dispatcher.Store) *DispatcherHandler {
	return &DispatcherHandler{
		accountManager: accountManager,
		peerStore:      peerStore,
		claimsExtractor: jwtclaims.NewClaimsExtractor(
			jwtclaims.WithAudience(authCfg.Audience),
			jwtclaims.WithUserIDClaim(authCfg.UserIDClaim),
		),
	}
}

func (h *DispatcherHandler) Connect(w http.ResponseWriter, r *http.Request) {
	// todo authenticate

	log.Infof("DispatcherHandler: Connect")
	vars := mux.Vars(r)
	peerId := vars["peerid"]
	if len(peerId) == 0 {
		util.WriteError(status.Errorf(status.InvalidArgument, "invalid peerId ID"), w)
		return
	}
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Errorf("ws handshake error %s", err)
		}
		return
	}

	p, ok := h.peerStore.Device(peerId)
	if !ok {
		log.Errorf("peer not found: %s", peerId)
		_ = ws.Close()
		return
	}
	h.ws = ws
	h.sessionWriter = p.OpenNewSession()
	go h.startReadLoop()
	go h.startWriteLoop()
}

func (h *DispatcherHandler) startReadLoop() {
	for {
		_, data, err := h.ws.ReadMessage()
		if err != nil {
			log.Debugf("failed to read from Websocket %s", err)
			h.close()
			return
		}
		err = h.sessionWriter.Write(data)
		if err != nil {
			log.Debugf("session write error %s", err)
			return
		}
	}
}

func (h *DispatcherHandler) startWriteLoop() {
	for {
		data, err := h.sessionWriter.Read()
		if err != nil {
			log.Debugf("failed to read from peer: %s", err)
			return
		}
		err = h.ws.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Debugf("failed to write to webscoket: %s", err)
			return
		}
	}
}

func (h *DispatcherHandler) close() {
	h.sessionWriter.Close()

}
