package terminal

import (
	_ "embed"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	mgm "github.com/netbirdio/netbird/management/client"
)

//go:embed headline.raw
var headline []byte

type Manager struct {
	terminals map[string]*terminal

	visitorChannel mgm.VisitorChannel
}

func NewManager() *Manager {
	return &Manager{
		terminals: make(map[string]*terminal),
	}
}

// todo: move these functions to a visitor manager layer

func (m *Manager) SetChannel(visitorChannel mgm.VisitorChannel) {
	m.visitorChannel = visitorChannel
}

func (m *Manager) OnNewConnection(sessionId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) OnCloseConnection(sessionId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) OnNewMsg(sessionId string, byteMsg []byte) error {
	msg := &Event{}
	err := json.Unmarshal(byteMsg, &msg)
	if err != nil {
		return err
	}
	switch msg.Action {
	case "open":
		m.openNewTerminal(sessionId, msg.Width, msg.Height)
	case "key":
		data := decodeData(msg.Data)
		m.injectKeyEvent(sessionId, data, msg.Width, msg.Height)
	case "close":
	}
	return nil

}

func decodeData(dataString string) []byte {
	keyCodes := make([]byte, len(dataString)/2)
	pos := 0

	for i := 0; i < len(dataString); i += 2 {
		c0 := dataString[i]
		if c0 < '0' || (c0 > '9' && c0 < 'A') || (c0 > 'F' && c0 < 'a') || c0 > 'f' {
			break
		}

		c1 := dataString[i+1]
		if c1 < '0' || (c1 > '9' && c1 < 'A') || (c1 > 'F' && c1 < 'a') || c1 > 'f' {
			break
		}

		var multiplierc1, multiplierc2 uint8
		if c0 > '9' {
			multiplierc1 = 1
		} else {
			multiplierc1 = 0
		}

		if c1 > '9' {
			multiplierc2 = 1
		} else {
			multiplierc2 = 0
		}
		keyCodes[pos] = 16*((c0&0xF)+9*multiplierc1) + (c1 & 0xF) + 9*multiplierc2
		pos++
	}
	return keyCodes
}

func (m *Manager) LeftSession(sessionId string) error {
	return nil
}

func (m *Manager) openNewTerminal(id string, width int, height int) {
	if _, ok := m.terminals[id]; ok {
		return
	}

	t, err := newTerminal(uint16(width), uint16(height))
	if err != nil {
		// todo close connection
		log.Errorf("failed to create terminal: %s", err)
		return
	}

	m.terminals[id] = t
	m.sendHeadline(id)
	go m.startReadLoop(id, t)
}

func (m *Manager) injectKeyEvent(id string, data []byte, width int, height int) {
	t, ok := m.terminals[id]
	if !ok {
		return
	}

	t.write(data, uint16(width), uint16(height))
}

func (m *Manager) startReadLoop(id string, t *terminal) {
	buf := make([]byte, 1024)
	for {
		n, err := t.read(buf)
		if err != nil {
			log.Error("failed to read from terminal: %s", err)
			return
		}

		data := jsonEscape(buf[:n])

		msg := &Key{
			Action: "key",
			Id:     id,
			Data:   data,
		}

		jsonStr, err := json.Marshal(msg)
		if err != nil {
			log.Error("failed to marshal event: %s", err)
			continue
		}

		err = m.visitorChannel.Send(id, jsonStr)
		if err != nil {
			log.Errorf("failed to send message to session: %s", err)
			return
		}
	}
}

func (m *Manager) sendHeadline(id string) {
	data := jsonEscape(headline)

	msg := &Key{
		Action: "key",
		Id:     id,
		Data:   data,
	}

	jsonStr, err := json.Marshal(msg)
	if err != nil {
		log.Error("failed to marshal headline event: %s", err)
		return
	}

	err = m.visitorChannel.Send(id, jsonStr)
	if err != nil {
		log.Errorf("failed to send message to session: %s", err)
		return
	}
}

func jsonEscape(buf []byte) string {
	const hexDigit = "0123456789ABCDEF"

	result := make([]byte, calcSize(buf))
	dst := 0
	for _, ch := range buf {
		if ch < ' ' {
			result[dst] = '\\'
			dst++
			switch ch {
			case '\b':
				result[dst] = 'b'
			case '\f':
				result[dst] = 'f'
			case '\n':
				result[dst] = 'n'
			case '\r':
				result[dst] = 'r'
			case '\t':
				result[dst] = 't'
			default:
				result[dst] = 'u'
				dst++
				result[dst] = '0'
				dst++
				result[dst] = '0'
				dst++
				result[dst] = hexDigit[ch>>4]
				dst++
				result[dst] = hexDigit[ch&0xF]
			}
			dst++
		} else if ch == '"' || ch == '\\' || ch == '/' {
			result[dst] = '\\'
			dst++
			result[dst] = byte(ch)
			dst++
		} else if ch > '\x7F' {
			result[dst] = '\\'
			dst++
			result[dst] = 'u'
			dst++
			result[dst] = '0'
			dst++
			result[dst] = '0'
			dst++
			result[dst] = hexDigit[ch>>4]
			dst++
			result[dst] = hexDigit[ch&0xF]
			dst++
		} else {
			// Single-byte character
			result[dst] = byte(ch)
			dst++
		}
	}

	return string(result)
}

func calcSize(buf []byte) int {
	count := 0
	for _, ch := range buf {
		if ch < ' ' {
			switch ch {
			case '\b', '\f', '\n', '\r', '\t':
				count += 2
			default:
				count += 6
			}
		} else if ch == '"' || ch == '\\' || ch == '/' {
			count += 2
		} else if ch > '\x7F' {
			count += 6
		} else {
			count++
		}
	}
	return count
}
