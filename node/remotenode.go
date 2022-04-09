package node

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/luxingwen/pnet/cache"
	"github.com/luxingwen/pnet/protos"
	"github.com/luxingwen/pnet/transport"
	"github.com/luxingwen/pnet/utils"

	"github.com/luxingwen/pnet/log"

	"github.com/luxingwen/pnet/multiplexer"

	"github.com/gogo/protobuf/proto"
)

const (
	stopGracePeriod = 100 * time.Millisecond

	startRetries = 3
)

type RemoteNode struct {
	*Node
	LocalNode  *LocalNode
	IsOutbound bool
	conn       net.Conn
	rxMsgChan  chan *protos.Message
	txMsgChan  chan *protos.Message
	txMsgCache cache.Cache

	sync.RWMutex
	lastRxTime    time.Time
	roundTripTime time.Duration
}

func NewRemoteNode(localNode *LocalNode, conn net.Conn, isOutbound bool, n *protos.Node) (*RemoteNode, error) {
	if localNode == nil {
		return nil, errors.New("Local node is nil")
	}
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	var node *Node

	if n != nil {
		node = newNode(n)
	} else {
		node = NewNode("", "", "")
	}

	txMsgCache := cache.NewGoCache(localNode.RemoteTxMsgCacheExpiration, localNode.RemoteTxMsgCacheCleanupInterval)

	remoteNode := &RemoteNode{
		Node:       node,
		LocalNode:  localNode,
		conn:       conn,
		IsOutbound: isOutbound,
		rxMsgChan:  make(chan *protos.Message, localNode.RemoteRxMsgChanLen),
		txMsgChan:  make(chan *protos.Message, localNode.RemoteTxMsgChanLen),
		txMsgCache: txMsgCache,
		lastRxTime: time.Now(),
	}

	return remoteNode, nil
}

func (rn *RemoteNode) String() string {
	if !rn.IsReady() {
		return fmt.Sprintf("<%s>", rn.conn.RemoteAddr().String())
	}
	return fmt.Sprintf("%v<%s>", rn.Node, rn.conn.RemoteAddr().String())
}

func (rn *RemoteNode) GetConn() net.Conn {
	return rn.conn
}

func (rn *RemoteNode) GetLastRxTime() time.Time {
	rn.RLock()
	defer rn.RUnlock()
	return rn.lastRxTime
}

func (rn *RemoteNode) setLastRxTime(lastRxTime time.Time) {
	rn.Lock()
	rn.lastRxTime = lastRxTime
	rn.Unlock()
}

func (rn *RemoteNode) setNode(n *protos.Node) error {
	rn.Node.Lock()
	defer rn.Node.Unlock()

	if rn.Id != "" && rn.Id != n.Id {
		return fmt.Errorf("Node id %s is different from expected value %x", n.Id, rn.Id)
	}

	remoteAddr, err := transport.Parse(n.Addr, rn.LocalNode.SupportedTransports)
	if err != nil {
		return fmt.Errorf("Parse node addr %s error: %s", n.Addr, err)
	}

	if remoteAddr.Host == "" {
		connAddr := rn.conn.RemoteAddr().String()
		remoteAddr.Host, _, err = transport.SplitHostPort(connAddr)
		if err != nil {
			return fmt.Errorf("Parse conn remote addr %s error: %s", connAddr, err)
		}
		n.Addr = remoteAddr.String()
	}

	if rn.Addr != "" {
		expectedAddr, err := transport.Parse(rn.Addr, rn.LocalNode.SupportedTransports)
		if err == nil && expectedAddr.Host == "" {
			connAddr := rn.conn.RemoteAddr().String()
			expectedAddr.Host, _, err = transport.SplitHostPort(connAddr)
			if err == nil {
				rn.Addr = expectedAddr.String()
			}
		}

	}

	if !proto.Equal(rn.Node.Node, n) {
		rn.Node.Node = n
	}

	return nil
}

func (rn *RemoteNode) Start() error {
	rn.StartOnce.Do(func() {
		if rn.IsStopped() {
			return
		}

		go rn.handleMsg()
		go rn.startMultiplexer()
		go rn.startMeasuringRoundTripTime()

		go func() {
			var n *protos.Node
			var err error

			for i := 0; i < startRetries; i++ {
				n, err = rn.ExchangeNode()
				if err == nil {
					break
				}
			}
			if err != nil {
				rn.Stop(fmt.Errorf("Get node error: %s", err))
				return
			}

			err = rn.setNode(n)
			if err != nil {
				rn.Stop(err)
				return
			}

			var existing *RemoteNode
			rn.LocalNode.neighbors.Range(func(key, value interface{}) bool {
				remoteNode, ok := value.(*RemoteNode)
				if ok && remoteNode.IsReady() && remoteNode.Id == n.Id {
					if remoteNode.IsStopped() {
						log.Warningf("Remove stopped remote node %v from list", remoteNode)
						rn.LocalNode.neighbors.Delete(key)
					} else {
						existing = remoteNode
					}
					return false
				}
				return true
			})
			if existing != nil {
				rn.Stop(fmt.Errorf("Node with id %x is already connected at addr %s", existing.Id, existing.conn.RemoteAddr().String()))
				return
			}

			rn.SetReady(true)

			for _, mw := range rn.LocalNode.middlewareStore.remoteNodeReady {
				if !mw.Func(rn) {
					break
				}
			}
		}()
	})

	return nil
}

func (rn *RemoteNode) Stop(err error) {
	rn.StopOnce.Do(func() {
		if err != nil {
			log.Warningf("Remote node %v stops because of error: %s", rn, err)
		} else {
			log.Infof("Remote node %v stops", rn)
		}

		err = rn.NotifyStop()
		if err != nil {
			log.Warning("Notify remote node %v stop error:", rn, err)
		}

		time.AfterFunc(stopGracePeriod, func() {
			rn.LifeCycle.Stop()

			if rn.conn != nil {
				rn.LocalNode.neighbors.Delete(rn.conn.RemoteAddr().String())
				rn.conn.Close()
			}

			for _, mw := range rn.LocalNode.middlewareStore.remoteNodeDisconnected {
				if !mw.Func(rn) {
					break
				}
			}
		})
	})
}

func (rn *RemoteNode) handleMsg() {
	var msg *protos.Message
	var remoteMsg *RemoteMessage
	var msgChan chan *RemoteMessage
	var added, ok bool
	var err error
	keepAliveTimeoutTimer := time.NewTimer(rn.LocalNode.KeepAliveTimeout)

NEXT_MESSAGE:
	for {
		if rn.IsStopped() {
			utils.StopTimer(keepAliveTimeoutTimer)
			return
		}

		select {
		case msg, ok = <-rn.rxMsgChan:
			if !ok {
				utils.StopTimer(keepAliveTimeoutTimer)
				return
			}

			for _, routingType := range rn.LocalNode.LocalRxMsgCacheRoutingType {
				if routingType == msg.RoutingType {

					added, err = rn.LocalNode.AddToRxCache(msg.MessageId)
					if err != nil {
						log.Errorf("Add msg id %x to rx cache error: %v", msg.MessageId, err)
						continue NEXT_MESSAGE
					}
					if !added {
						fmt.Println("added: ", added, "msgid : ", string(msg.MessageId), "routingType:", routingType)
						continue NEXT_MESSAGE
					}
				}
			}

			remoteMsg, err = NewRemoteMessage(rn, msg)
			if err != nil {
				log.Errorf("New remote message error: %v", err)
				continue
			}

			msgChan, err = rn.LocalNode.GetRxMsgChan(msg.RoutingType)
			if err != nil {
				log.Errorf("Get rx msg chan for routing type %v error: %v", msg.RoutingType, err)
				continue
			}

			select {
			case msgChan <- remoteMsg:
			default:
				log.Warningf("Msg chan full for routing type %d, discarding msg", msg.RoutingType)
			}
		case <-keepAliveTimeoutTimer.C:
			if time.Since(rn.GetLastRxTime()) > rn.LocalNode.KeepAliveTimeout {
				rn.Stop(errors.New("keepalive timeout"))
			}
		}

		utils.ResetTimer(keepAliveTimeoutTimer, rn.LocalNode.KeepAliveTimeout)
	}
}

func (rn *RemoteNode) NotifyStop() error {
	msg, err := rn.LocalNode.NewStopMessage()
	if err != nil {
		return err
	}

	err = rn.SendMessageAsync(msg)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RemoteNode) SendMessage(msg *protos.Message, hasReply bool, replyTimeout time.Duration) (<-chan *RemoteMessage, error) {
	if rn.IsStopped() {
		return nil, errors.New("Remote node has stopped")
	}

	if len(msg.MessageId) == 0 {
		return nil, errors.New("Message ID is empty")
	}

	for _, routingType := range rn.LocalNode.RemoteTxMsgCacheRoutingType {
		if routingType == msg.RoutingType {
			_, found := rn.txMsgCache.Get(msg.MessageId)
			if found {
				return nil, nil
			}

			err := rn.txMsgCache.Add(msg.MessageId, struct{}{})
			if err != nil {
				return nil, err
			}
		}
	}

	select {
	case rn.txMsgChan <- msg:
	default:
		return nil, errors.New("Tx msg chan full, discarding msg")
	}

	if hasReply {
		return rn.LocalNode.AllocReplyChan(msg.MessageId, replyTimeout)
	}

	return nil, nil
}

func (rn *RemoteNode) SendMessageAsync(msg *protos.Message) error {
	_, err := rn.SendMessage(msg, false, 0)
	return err
}

func (rn *RemoteNode) SendMessageSync(msg *protos.Message, replyTimeout time.Duration) (*RemoteMessage, error) {
	if replyTimeout == 0 {
		replyTimeout = rn.LocalNode.DefaultReplyTimeout
	}

	replyChan, err := rn.SendMessage(msg, true, replyTimeout)
	if err != nil {
		return nil, err
	}

	select {
	case replyMsg := <-replyChan:
		return replyMsg, nil
	case <-time.After(replyTimeout):
		return nil, errors.New("Wait for reply timeout")
	}
}

func (rn *RemoteNode) Ping() error {
	msg, err := rn.LocalNode.NewPingMessage()
	if err != nil {
		return err
	}

	_, err = rn.SendMessageSync(msg, 0)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RemoteNode) ExchangeNode() (*protos.Node, error) {
	msg, err := rn.LocalNode.NewExchangeNodeMessage()
	if err != nil {
		return nil, err
	}

	reply, err := rn.SendMessageSync(msg, 0)
	if err != nil {
		return nil, err
	}

	replyBody := &protos.ExchangeNodeReply{}
	err = proto.Unmarshal(reply.Msg.Message, replyBody)
	if err != nil {
		return nil, err
	}

	return replyBody.Node, nil
}

func (rn *RemoteNode) startMeasuringRoundTripTime() {
	var err error
	var txTime, rxTime time.Time
	var lastRoundTripTime, interval time.Duration

	for {
		time.Sleep(interval)

		if rn.IsStopped() {
			return
		}

		txTime = time.Now()
		err = rn.Ping()
		if err != nil {
			log.Debugf("Ping %v error: %v", rn, err)
			lastRoundTripTime = rn.LocalNode.DefaultReplyTimeout * 2
			interval = rn.LocalNode.MeasureRoundTripTimeInterval / 5
		} else {
			rxTime = time.Now()
			lastRoundTripTime = rxTime.Sub(txTime)
			rn.setLastRxTime(rxTime)
			interval = utils.RandDuration(rn.LocalNode.MeasureRoundTripTimeInterval, 1.0/5.0)
		}

		if rn.roundTripTime > 0 {
			rn.setRoundTripTime((rn.roundTripTime + lastRoundTripTime) / 2)
		} else {
			rn.setRoundTripTime(lastRoundTripTime)
		}
	}
}

func (rn *RemoteNode) setRoundTripTime(roundTripTime time.Duration) {
	rn.Lock()
	rn.roundTripTime = roundTripTime
	rn.Unlock()
}

func (rn *RemoteNode) startMultiplexer() {
	mux, err := multiplexer.NewMultiplexer(rn.LocalNode.Multiplexer, rn.conn, rn.IsOutbound)
	if err != nil {
		rn.Stop(fmt.Errorf("Create multiplexer error: %s", err))
		return
	}

	var conn net.Conn
	if rn.IsOutbound {
		for i := uint32(0); i < rn.LocalNode.NumStreamsToOpen; i++ {
			conn, err = mux.OpenStream()
			if err != nil {
				rn.Stop(fmt.Errorf("Open stream error: %s", err))
				return
			}
			go rn.rx(conn, i == 0)
		}
	} else {
		for i := uint32(0); i < rn.LocalNode.NumStreamsToAccept; i++ {
			conn, err = mux.AcceptStream()
			if err != nil {
				rn.Stop(fmt.Errorf("Accept stream error: %s", err))
				return
			}
			go rn.rx(conn, true)
		}
	}
}

func (rn *RemoteNode) handleMsgBuf(buf []byte) {
	msg := &protos.Message{}
	err := proto.Unmarshal(buf, msg)
	if err != nil {
		rn.Stop(fmt.Errorf("unmarshal msg error: %s", err))
		return
	}

	select {
	case rn.rxMsgChan <- msg:
	default:
		log.Warning("Rx msg chan full, discarding msg")
	}
}

func (rn *RemoteNode) rx(conn net.Conn, isActive bool) {
	msgLenBuf := make([]byte, msgLenBytes)
	var readLen uint32

	if isActive {
		go rn.tx(conn)
	}

	for {
		if rn.IsStopped() {
			return
		}

		l, err := conn.Read(msgLenBuf)
		if err != nil {
			rn.Stop(fmt.Errorf("Read msg len error: %s", err))
			continue
		}
		if l != msgLenBytes {
			rn.Stop(fmt.Errorf("Msg len has %d bytes, which is less than expected %d", l, msgLenBytes))
			continue
		}

		if !isActive {
			isActive = true
			go rn.tx(conn)
		}

		msgLen := binary.BigEndian.Uint32(msgLenBuf)
		if msgLen < 0 {
			rn.Stop(fmt.Errorf("Msg len %d overflow", msgLen))
			continue
		}

		if msgLen > rn.LocalNode.MaxMessageSize {
			rn.Stop(fmt.Errorf("Msg size %d exceeds max msg size %d", msgLen, rn.LocalNode.MaxMessageSize))
			continue
		}

		buf := make([]byte, msgLen)

		for readLen = 0; readLen < msgLen; readLen += uint32(l) {
			l, err = conn.Read(buf[readLen:])
			if err != nil {
				break
			}
		}

		if err != nil {
			rn.Stop(fmt.Errorf("Read msg error: %s", err))
			continue
		}

		if readLen > msgLen {
			rn.Stop(fmt.Errorf("Msg has %d bytes, which is more than expected %d", readLen, msgLen))
			continue
		}

		var shouldCallNextMiddleware bool
		for _, mw := range rn.LocalNode.middlewareStore.messageWillDecode {
			buf, shouldCallNextMiddleware = mw.Func(rn, buf)
			if buf == nil || !shouldCallNextMiddleware {
				break
			}
		}

		if buf == nil {
			continue
		}

		rn.handleMsgBuf(buf)
	}
}

func (rn *RemoteNode) tx(conn net.Conn) {
	var msg *protos.Message
	var buf []byte
	var ok bool
	var err error
	msgLenBuf := make([]byte, msgLenBytes)
	txTimeoutTimer := time.NewTimer(time.Second)

	for {
		if rn.IsStopped() {
			utils.StopTimer(txTimeoutTimer)
			return
		}

		select {
		case msg, ok = <-rn.txMsgChan:
			if !ok {
				utils.StopTimer(txTimeoutTimer)
				return
			}

			buf, err = proto.Marshal(msg)
			if err != nil {
				log.Errorf("Marshal msg error: %v", err)
				continue
			}

			var shouldCallNextMiddleware bool
			for _, mw := range rn.LocalNode.middlewareStore.messageEncoded {
				buf, shouldCallNextMiddleware = mw.Func(rn, buf)
				if buf == nil || !shouldCallNextMiddleware {
					break
				}
			}

			if buf == nil {
				continue
			}

			if uint32(len(buf)) > rn.LocalNode.MaxMessageSize {
				log.Errorf("Msg size %d exceeds max msg size %d", len(buf), rn.LocalNode.MaxMessageSize)
				continue
			}

			binary.BigEndian.PutUint32(msgLenBuf, uint32(len(buf)))

			_, err = conn.Write(msgLenBuf)
			if err != nil {
				rn.Stop(fmt.Errorf("Write to conn error: %s", err))
				continue
			}

			_, err = conn.Write(buf)
			if err != nil {
				rn.Stop(fmt.Errorf("Write to conn error: %s", err))
				continue
			}
		case <-txTimeoutTimer.C:
		}

		utils.ResetTimer(txTimeoutTimer, time.Second)
	}
}
