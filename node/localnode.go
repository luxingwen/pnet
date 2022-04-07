package node

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/luxingwen/pnet/cache"
	"github.com/luxingwen/pnet/config"
	"github.com/luxingwen/pnet/protos"
	"github.com/luxingwen/pnet/transport"

	"github.com/luxingwen/pnet/log"
)

type LocalNode struct {
	*Node
	*config.Config

	*middlewareStore
	port     uint16
	listener net.Listener

	handleMsgChan chan *RemoteMessage
	rxMsgChan     map[protos.RoutingType]chan *RemoteMessage

	rxMsgCache     cache.Cache
	replyChanCache cache.Cache
	replyTimeout   time.Duration
	neighbors      sync.Map
	address        *transport.Address
}

func NewLocalNode(id string, conf *config.Config) (*LocalNode, error) {
	if id == "" {
		return nil, errors.New("node id is empty")
	}

	address, err := transport.NewAddress(conf.Transport, conf.Hostname, conf.Port, conf.SupportedTransports)
	if err != nil {
		return nil, err
	}

	node := NewNode(id, conf.Name, address.String())

	handleMsgChan := make(chan *RemoteMessage, conf.LocalHandleMsgChanLen)
	rxMsgCache := cache.NewGoCache(conf.LocalRxMsgCacheExpiration, conf.LocalRxMsgCacheCleanupInterval)

	replyChanCache := cache.NewGoCache(conf.DefaultReplyTimeout, conf.ReplyChanCleanupInterval)

	rxMsgChan := make(map[protos.RoutingType]chan *RemoteMessage)

	middlewareStore := newMiddlewareStore()
	localNode := &LocalNode{
		Node:            node,
		Config:          conf,
		address:         address,
		rxMsgCache:      rxMsgCache,
		replyChanCache:  replyChanCache,
		middlewareStore: middlewareStore,
		port:            address.Port,
		rxMsgChan:       rxMsgChan,
		handleMsgChan:   handleMsgChan,
	}

	for routingType := range protos.RoutingType_name {
		localNode.RegisterRoutingType(protos.RoutingType(routingType))
	}

	return localNode, nil
}

func (ln *LocalNode) Start() error {
	ln.StartOnce.Do(func() {
		for _, mw := range ln.middlewareStore.localNodeWillStart {
			if !mw.Func(ln) {
				break
			}
		}

		for i := uint32(0); i < ln.LocalHandleMsgWorkers; i++ {
			go ln.handleMsg()
		}

		go ln.listen()

		for _, mw := range ln.middlewareStore.localNodeStarted {
			if !mw.Func(ln) {
				break
			}
		}
	})

	return nil
}

func (ln *LocalNode) Stop(err error) {
	ln.StopOnce.Do(func() {
		for _, mw := range ln.middlewareStore.localNodeWillStop {
			if !mw.Func(ln) {
				break
			}
		}

		if err != nil {
			log.Warningf("Local node %v stops because of error: %s", ln, err)
		} else {
			log.Infof("Local node %v stops", ln)
		}

		ln.neighbors.Range(func(key, value interface{}) bool {
			remoteNode, ok := value.(*RemoteNode)
			if ok {
				remoteNode.Stop(err)
			}
			return true
		})

		time.Sleep(stopGracePeriod)

		ln.LifeCycle.Stop()

		if ln.listener != nil {
			ln.listener.Close()
		}

		for _, mw := range ln.middlewareStore.localNodeStopped {
			if !mw.Func(ln) {
				break
			}
		}
	})
}

func (ln *LocalNode) handleMsg() {
	var remoteMsg *RemoteMessage
	var err error

	for {
		if ln.IsStopped() {
			return
		}

		remoteMsg = <-ln.handleMsgChan

		err = ln.handleRemoteMessage(remoteMsg)
		if err != nil {
			log.Errorf("Handle remote message error: %v", err)
			continue
		}

	}
}

func (ln *LocalNode) listen() {

	listener, err := ln.address.Transport.Listen(ln.port)
	if err != nil {
		ln.Stop(fmt.Errorf("failed to listen to port %d", ln.port))
		return
	}
	ln.listener = listener

	if ln.port == 0 {
		_, portStr, err := transport.SplitHostPort(listener.Addr().String())
		if err != nil {
			ln.Stop(err)
			return
		}

		if len(portStr) > 0 {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				ln.Stop(err)
				return
			}

			ln.SetInternalPort(uint16(port))
		}
	}

	if ln.address.Port == 0 {
		ln.address.Port = ln.port
		ln.Addr = ln.address.String()
	}

	for {
		conn, err := listener.Accept()

		if ln.IsStopped() {
			if err == nil {
				conn.Close()
			}
			return
		}

		if err != nil {
			log.Errorf("Error accepting connection: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		_, loaded := ln.neighbors.LoadOrStore(conn.RemoteAddr().String(), nil)
		if loaded {
			log.Errorf("Remote addr %s is already connected, reject connection", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		log.Infof("Remote node connect from %s to local address %s", conn.RemoteAddr().String(), conn.LocalAddr())

		rn, err := ln.StartRemoteNode(conn, false, nil)
		if err != nil {
			log.Errorf("Error creating remote node: %v", err)
			ln.neighbors.Delete(conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		ln.neighbors.Store(conn.RemoteAddr().String(), rn)
	}
}

func (ln *LocalNode) SetInternalPort(port uint16) {
	ln.port = port
}

func (ln *LocalNode) StartRemoteNode(conn net.Conn, isOutbound bool, n *protos.Node) (*RemoteNode, error) {
	remoteNode, err := NewRemoteNode(ln, conn, isOutbound, n)
	if err != nil {
		return nil, err
	}

	for _, mw := range ln.middlewareStore.remoteNodeConnected {
		if !mw.Func(remoteNode) {
			break
		}
	}

	err = remoteNode.Start()
	if err != nil {
		return nil, err
	}

	return remoteNode, nil
}

func (ln *LocalNode) AddToRxCache(msgID []byte) (bool, error) {
	_, found := ln.rxMsgCache.Get(msgID)
	if found {
		//fmt.Println("----1111 :", string(msgID), "found:", found, " v:", v)
		return false, nil
	}

	err := ln.rxMsgCache.Add(msgID, struct{}{})
	if err != nil {
		//	fmt.Println("----222")
		if _, found := ln.rxMsgCache.Get(msgID); found {
			//	fmt.Println("----3333")
			return false, nil
		}
		return false, err
	}

	//fmt.Println("-----222, :", string(msgID))
	return true, nil
}

func (ln *LocalNode) RegisterRoutingType(routingType protos.RoutingType) {
	ln.rxMsgChan[routingType] = make(chan *RemoteMessage, ln.LocalRxMsgChanLen)
}

func (ln *LocalNode) GetRxMsgChan(routingType protos.RoutingType) (chan *RemoteMessage, error) {
	c, ok := ln.rxMsgChan[routingType]
	if !ok {
		return nil, fmt.Errorf("Msg chan does not exist for type %d", routingType)
	}
	return c, nil
}

func (ln *LocalNode) AllocReplyChan(msgID []byte, expiration time.Duration) (chan *RemoteMessage, error) {
	if len(msgID) == 0 {
		return nil, errors.New("Message id is empty")
	}

	replyChan := make(chan *RemoteMessage)

	err := ln.replyChanCache.AddWithExpiration(msgID, replyChan, expiration)
	if err != nil {
		return nil, err
	}

	return replyChan, nil
}

func (ln *LocalNode) GetReplyChan(msgID []byte) (chan *RemoteMessage, bool) {
	value, ok := ln.replyChanCache.Get(msgID)
	if !ok {
		return nil, false
	}

	replyChan, ok := value.(chan *RemoteMessage)
	if !ok {
		return nil, false
	}

	return replyChan, true
}

func (ln *LocalNode) Connect(n *protos.Node) (*RemoteNode, bool, error) {
	if n.Addr == ln.address.String() {
		return nil, false, errors.New("trying to connect to self")
	}

	remoteAddress, err := transport.Parse(n.Addr, ln.SupportedTransports)
	if err != nil {
		return nil, false, err
	}

	key := remoteAddress.ConnRemoteAddr()
	value, loaded := ln.neighbors.LoadOrStore(key, nil)
	if loaded {
		remoteNode, ok := value.(*RemoteNode)
		if ok {
			if remoteNode.IsStopped() {
				log.Warningf("Remove stopped remote node %v from list", remoteNode)
				ln.neighbors.Delete(key)
			} else {
				log.Infof("Load remote node %v from list", remoteNode)
				return remoteNode, remoteNode.IsReady(), nil
			}
		} else {
			log.Infof("Another goroutine is connecting to %s", key)
			return nil, false, nil
		}
	}

	var shouldConnect, shouldCallNextMiddleware bool
	for _, mw := range ln.middlewareStore.willConnectToNode {
		shouldConnect, shouldCallNextMiddleware = mw.Func(n)
		if !shouldConnect {
			return nil, false, nil
		}
		if !shouldCallNextMiddleware {
			break
		}
	}

	conn, err := remoteAddress.Dial(ln.DialTimeout)
	if err != nil {
		ln.neighbors.Delete(key)
		return nil, false, err
	}

	remoteNode, err := ln.StartRemoteNode(conn, true, n)
	if err != nil {
		ln.neighbors.Delete(key)
		conn.Close()
		return nil, false, err
	}

	ln.neighbors.Store(key, remoteNode)

	return remoteNode, false, nil
}

func (ln *LocalNode) HandleRemoteMessage(remoteMsg *RemoteMessage) error {
	select {
	case ln.handleMsgChan <- remoteMsg:
	default:
		log.Warningf("Local node handle msg chan full, discarding msg")
	}
	return nil
}

func (ln *LocalNode) GetNeighbors(filter func(*RemoteNode) bool) (map[string]*RemoteNode, error) {
	nodes := make(map[string]*RemoteNode, 0)
	ln.neighbors.Range(func(key, value interface{}) bool {
		remoteNode, ok := value.(*RemoteNode)
		if ok && remoteNode.IsReady() && !remoteNode.IsStopped() {
			if filter == nil || filter(remoteNode) {
				nodes[remoteNode.Id] = remoteNode
			}
		}
		return true
	})
	return nodes, nil
}
