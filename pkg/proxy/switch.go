package proxy

import (
	"cocogo/pkg/transport"
	"cocogo/pkg/userhome"
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

var Manager = &manager{
	container: new(sync.Map),
}

type manager struct {
	container *sync.Map
}

func (m *manager) add(uHome userhome.SessionHome) {
	m.container.Store(uHome.SessionID(), uHome)

}

func (m *manager) delete(roomID string) {
	m.container.Delete(roomID)

}

func (m *manager) search(roomID string) (userhome.SessionHome, bool) {
	if uHome, ok := m.container.Load(roomID); ok {
		return uHome.(userhome.SessionHome), ok
	}
	return nil, false
}

func (m *manager) JoinShareRoom(roomID string, uConn userhome.Conn) {
	if userHome, ok := m.search(roomID); ok {
		userHome.AddConnection(uConn)
	}
}

func (m *manager) ExitShareRoom(roomID string, uConn userhome.Conn) {
	if userHome, ok := m.search(roomID); ok {
		userHome.RemoveConnection(uConn)
	}

}

func (m *manager) Switch(ctx context.Context, uHome userhome.SessionHome, agent transport.Agent) error {
	m.add(uHome)
	defer m.delete(uHome.SessionID())

	subCtx, cancelFunc := context.WithCancel(ctx)
	userSendRequestStream := uHome.SendRequestChannel(subCtx)
	userReceiveStream := uHome.ReceiveResponseChannel(subCtx)
	nodeRequestChan := agent.ReceiveRequestChannel(subCtx)
	nodeSendResponseStream := agent.SendResponseChannel(subCtx)

	for userSendRequestStream != nil || nodeSendResponseStream != nil {
		select {
		case buf1, ok := <-userSendRequestStream:
			if !ok {
				log.Warn("userSendRequestStream close")
				userSendRequestStream = nil
				close(nodeRequestChan)
				continue
			}
			nodeRequestChan <- buf1
		case buf2, ok := <-nodeSendResponseStream:
			if !ok {
				log.Warn("nodeSendResponseStream close")
				nodeSendResponseStream = nil
				close(userReceiveStream)
				cancelFunc()
				continue
			}
			userReceiveStream <- buf2
		case <-ctx.Done():
			log.Info("proxy end by context done")
			cancelFunc()
			return nil
		}
	}
	log.Info("proxy end")
	return nil
}