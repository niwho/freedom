package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

var (
	logger = log.New()
)

func GetECDSADeprete(mark int) (pk *ecdsa.PrivateKey, err error) {
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

func main() {

	logger.SetHandler(log.StdoutHandler)
	logger.Info("pre start22")

	nodeKey, err := GetECDSA(0)
	if err != nil {
		logger.Crit("get privatekey", "err", err)
	}
	nid, err := discover.HexID("2e3147536036cc38f590ca60b352b391751a47c5e4a6cb0b902e08ea6ca03c67ccbb103fc9f451c0a06c99207fb9c8b1a3e34b35e6f5297e45e5eb69446c2aa2")
	if err != nil {
		logger.Crit("parse node ", "err", err)
	}

	staticNode := discover.Node{
		IP:  net.ParseIP("192.168.31.233"), // len 4 for IPv4 or 16 for IPv6
		TCP: 3227,
		UDP: 3227, // port numbers
		ID:  nid,
	}
	_ = staticNode
	InitPump(func(p *p2p.Peer, data []byte) error {
		logger.Info("receive", "data", string(data), "from", p)
		return nil
	})
	PumpInstance.Start(":3226", nodeKey, []*discover.Node{&staticNode})
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			ps := PumpInstance.server.Peers()
			_ = ps
			//fmt.Println(len(ps), ps)
			PumpInstance.Broadcast([]byte(time.Now().Format(time.RFC3339)))
		}
	}

}
