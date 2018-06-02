package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

var (
	proto = p2p.Protocol{
		Name:    "wexchange",
		Version: 1,
		Length:  1,
		Run:     PumpInstance.GetRun(),
	}
)

var PumpInstance *MsgPump

type Pump interface {
	Receive(*p2p.Peer, []byte) error
	Send(*p2p.Peer, []byte) error
	Broadcast([]byte) error
	GetPeers() (map[discover.NodeID]*p2p.Peer, error)
}

type Peer struct {
	*p2p.Peer
	rw     p2p.MsgReadWriter
	sendCh chan innerMsg
}

type innerMsg struct {
	peer *p2p.Peer
	data []byte
}

type MsgPump struct {
	peers map[discover.NodeID]*Peer
	mu    sync.RWMutex

	server *p2p.Server
	//events chan *p2p.PeerEvent

	// 关播消息，并行发送，不阻塞
	broad event.Feed

	// 聚合收到的信息
	//broadReceive event.Feed
	//chReceive    chan []byte
}

func (mp *MsgPump) init() {
	mp.peers = map[discover.NodeID]*Peer{}
	//mp.events = make(chan *p2p.PeerEvent, 1)
	//mp.chReceive = make(chan []byte, 32)
	//mp.server.SubscribeEvents(mp.events)
	//mp.broadReceive.Subscribe(mp.chReceive)
}

func (mp *MsgPump) getPeer(id discover.NodeID) *Peer {
	mp.mu.RLock()
	peer, ok := mp.peers[id]
	mp.mu.RUnlock()
	if !ok {
		return nil
	}
	return peer
}

func (mp *MsgPump) Receive(p *p2p.Peer, data []byte) error {
	// callback user function
	return nil
}

func (mp *MsgPump) Send(p *p2p.Peer, data []byte) error {
	//peer := mp.getPeer(id)
	//if peer == nil {
	//	return fmt.Errorf("no such id=", id)
	//}

	mp.broad.Send(innerMsg{peer: p, data: data})
	//	return p2p.Send(peer.rw, 0, data)
	return nil
}

func (mp *MsgPump) Broadcast(data []byte) error {
	mp.broad.Send(innerMsg{data: data})
	return nil
}

func (mp *MsgPump) GetPeers() (mm map[discover.NodeID]*Peer, err error) {
	mp.mu.RLock()
	for k, v := range mp.peers {
		mm[k] = v
	}
	mp.mu.RUnlock()
	return
}

func (mp *MsgPump) GetRun() func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

	return func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		peer := Peer{Peer: p, rw: rw}
		peer.sendCh = make(chan innerMsg, 1)
		sub := mp.broad.Subscribe(peer.sendCh)
		quit := make(chan struct{}, 1)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-quit:
					return
				case msg := <-peer.sendCh:
					if msg.peer != nil {
						if msg.peer == p {
							p2p.Send(rw, 0, msg.data)
						} else {
							fmt.Printf("no such id=", msg.peer.ID)
						}
					} else {
						p2p.Send(rw, 0, msg.data)
					}
				}
			}
		}()
		go func() {
			defer wg.Done()
			for {
				msg, err := rw.ReadMsg()
				if err != nil {
					quit <- struct{}{}
					return
				}
				data, err := ioutil.ReadAll(msg.Payload)
				if err != nil {
					fmt.Printf("payload err:%s", err)
					continue
				}
				mp.Receive(p, data)
			}
		}()
		mp.mu.Lock()
		mp.peers[p.ID()] = &peer
		mp.mu.Unlock()

		wg.Wait()
		sub.Unsubscribe()
		close(peer.sendCh)
		close(quit)

		mp.mu.Lock()
		delete(mp.peers, p.ID())
		mp.mu.Unlock()

		return fmt.Errorf("done:%s", p.ID)
	}
}

//go func() {
//	for {
//		select {
//		case evt := <-mp.events:
//			if evt.Type == p2p.PeerEventTypeDrop {
//				// 删除节点
//				mp.mu.Lock()
//				delete(mp.peers, p.ID)
//				mp.mu.Unlock()
//			}
//		}
//	}
//}()
func (mp *MsgPump) Start(listenAddr string, pk *ecdsa.PrivateKey, sn []*discover.Node) error {

	conf := p2p.Config{
		PrivateKey:      pk,
		MaxPeers:        30,
		MaxPendingPeers: 0,
		DialRatio:       2,
		NoDiscovery:     false,
		DiscoveryV5:     false,
		Name:            "f1",
		StaticNodes:     sn,
		Protocols:       []p2p.Protocol{proto},
		ListenAddr:      listenAddr,
		Logger:          logger,
	}

	svr := p2p.Server{
		Config: conf,
	}
	return svr.Start()
}

func GetECDSA(mark int) (pk *ecdsa.PrivateKey, err error) {
	// todo: load from file
	var buf []byte
	if buf, err = ioutil.ReadFile(fmt.Sprintf("p_%d", mark)); err == nil {
		if pk, err = crypto.ToECDSA(buf); err == nil {
			return
		} else {
			logger.Error("crypto.ToECDSA ", "err", err)
			fmt.Println()
		}
	} else {
		logger.Error("readfile ", "err", err)
	}
	pk, err = crypto.GenerateKey()
	buf = crypto.FromECDSA(pk)
	ioutil.WriteFile(fmt.Sprintf("p_%d", mark), buf, 0666)
	return
}
