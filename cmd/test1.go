package main

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/luxingwen/pnet"

	"github.com/luxingwen/pnet/config"
	"github.com/luxingwen/pnet/log"
	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/protos"
)

func newPnet(id string, name string, port uint16) *pnet.PNet {

	cfg := config.DefaultConfig()
	cfg.Hostname = "127.0.0.1"
	cfg.Port = port
	pn, err := pnet.NewPNet(id, cfg)
	if err != nil {
		panic(err)
	}

	pn.ApplyMiddleware(node.LocalNodeStarted{func(lc *node.LocalNode) bool {
		lc.SetReady(true)
		return true
	}, 0})

	pn.ApplyMiddleware(node.BytesReceived{func(msg, msgID []byte, srcID, rpath string, remoteNode *node.RemoteNode) ([]byte, bool) {
		log.Infof("Receive message \"%s\" from %s by %s , path: %s ", string(msg), srcID, remoteNode.Id, rpath)
		pn.SendBytesRelayReply(msgID, []byte("receive send res:"+rpath), srcID)
		return nil, true
	}, 0})

	err = pn.Start()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Second)
		if pn.GetLocalNode().IsReady() {
			return pn
		}
	}

	return pn
}

func main() {
	hostname := "127.0.0.1"

	p1 := newPnet("p1", hostname, 40001)
	p2 := newPnet("p2", hostname, 40002)
	p3 := newPnet("p3", hostname, 40003)
	p4 := newPnet("p4", hostname, 40004)
	p5 := newPnet("p5", hostname, 40005)
	p6 := newPnet("p6", hostname, 40006)

	p2.Join(p1.GetLocalNode().Addr)
	p3.Join(p2.GetLocalNode().Addr)
	p4.Join(p3.GetLocalNode().Addr)
	p5.Join(p4.GetLocalNode().Addr)
	p6.Join(p5.GetLocalNode().Addr)

	time.Sleep(time.Second * 3)

	p2.SendBytesRelayAsync([]byte("p2 send->p1"), p1.GetLocalNode().GetId())
	p3.SendBytesRelayAsync([]byte("p3 send->p1"), p1.GetLocalNode().GetId())
	p4.SendBytesRelayAsync([]byte("p4 send->p1"), p1.GetLocalNode().GetId())
	p5.SendBytesRelayAsync([]byte("p5 send->p1"), p1.GetLocalNode().GetId())
	p6.SendBytesRelayAsync([]byte("p6 send->p1"), p1.GetLocalNode().GetId())

	p3.SendBytesRelayAsync([]byte("p3 send->p5"), p5.GetLocalNode().GetId())
	p2.SendBytesRelayAsync([]byte("p2 send->p6"), p6.GetLocalNode().GetId())

	r, srcid, err := p1.SendBytesRelaySync([]byte("p1 send->p6"), p6.GetLocalNode().GetId())
	if err != nil {
		log.Error("p1 send p6 err:", err)

	} else {
		log.Debugf("get res: %s  srcid:%s", string(r), srcid)
	}

	ok, err := p3.SendBytesRelayAsync([]byte("p3 send->p7"), "p7")
	if err != nil {
		log.Error("p3 send p7 err:", err)
	}
	log.Info("ok:", ok)

	r, srcid, err = p4.SendBytesRelaySync([]byte("p4 send->p7"), "p7")
	if err != nil {
		log.Error("p4 send p7 err:", err)

	} else {
		log.Debugf("p4 get res: %s  srcid:%s", string(r), srcid)
	}

	msg := p2.GetLocalNode().NewGetNeighborsMessage(p5.GetLocalNode().Id)

	relay, _, err := p2.SendMessageSync(msg, protos.RELAY, time.Second*10)
	if err != nil {
		log.Error("p2 send p5 err:", err)
		return
	}

	nodes := &protos.Neighbors{}

	err = proto.Unmarshal(relay.Message, nodes)
	if err != nil {
		log.Error("unmarshal err:", err)
		return
	}

	for _, item := range nodes.Nodes {
		fmt.Println(item.Addr, item.Name, item.Id)
	}

	msg, _ = p2.GetLocalNode().NewConnnetNodeMessage(p5.GetLocalNode().Id, p1.GetLocalNode().Node.Node)

	relay, _, err = p2.SendMessageSync(msg, protos.RELAY, time.Second*10)
	if err != nil {
		log.Error("p2 send p5 err:", err)
		return
	}
	fmt.Println("connect :", string(relay.Message))

	msg = p2.GetLocalNode().NewGetNeighborsMessage(p5.GetLocalNode().Id)

	relay, _, err = p2.SendMessageSync(msg, protos.RELAY, time.Second*10)
	if err != nil {
		log.Error("p2 send p5 err:", err)
		return
	}

	nodes1 := &protos.Neighbors{}

	err = proto.Unmarshal(relay.Message, nodes1)
	if err != nil {
		log.Error("unmarshal err:", err)
		return
	}

	for _, item := range nodes1.Nodes {
		fmt.Println(item.Addr, item.Name, item.Id)
	}

	msg, _ = p2.GetLocalNode().NewStopConnnetNodeMessage(p5.GetLocalNode().Id, p6.GetLocalNode().Node.Node)

	relay, _, err = p2.SendMessageSync(msg, protos.RELAY, time.Second*10)
	if err != nil {
		log.Error("p2 send p5 err:", err)
		return
	}
	fmt.Println("stop connect :", string(relay.Message))

	msg = p2.GetLocalNode().NewGetNeighborsMessage(p5.GetLocalNode().Id)

	relay, _, err = p2.SendMessageSync(msg, protos.RELAY, time.Second*10)
	if err != nil {
		log.Error("p2 send p5 err:", err)
		return
	}

	nodes2 := &protos.Neighbors{}

	err = proto.Unmarshal(relay.Message, nodes2)
	if err != nil {
		log.Error("unmarshal err:", err)
		return
	}

	for _, item := range nodes2.Nodes {
		fmt.Println(item.Addr, item.Name, item.Id)
	}

	select {}

}
