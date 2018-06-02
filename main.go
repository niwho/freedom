package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"net"

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
	nid, err := discover.HexID("53d856ee0dc1f9e723835223830a2221f355df23cb9c809ed0065fbbcb82dc30bf4650d294e8f6adca581d10338fa4255e4105794a1d4aacc34ac00a3515972a")
	if err != nil {
		logger.Crit("parse node ", "err", err)
	}

	staticNode := discover.Node{
		IP:  net.ParseIP("192.168.0.228"), // len 4 for IPv4 or 16 for IPv6
		TCP: 3227,
		UDP: 3227, // port numbers
		ID:  nid,
	}
	_ = staticNode
	conf := p2p.Config{
		PrivateKey:      nodeKey,
		MaxPeers:        30,
		MaxPendingPeers: 0,
		DialRatio:       2,
		NoDiscovery:     false,
		DiscoveryV5:     false,
		Name:            "f1",
		//StaticNodes:     []*discover.Node{&staticNode},
		Protocols:  []p2p.Protocol{},
		ListenAddr: ":3226",
		Logger:     logger,
	}

	svr := p2p.Server{
		Config: conf,
	}
	if err := svr.Start(); err != nil {
		logger.Crit("start server ", "err", err)
	}
	logger.Info("server", "id", svr.Self().ID.String())
	select {}

}
