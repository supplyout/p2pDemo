package main

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"time"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (d *discoveryNotifee) HandlePeerFound(peer peer.AddrInfo) {
	d.PeerChan <- peer
}

// 初始化MDNS服务,需要传递ctx,peerHost,rendezvous
func initMDNS(ctx context.Context, peerHost host.Host, rendezvous string) chan peer.AddrInfo {
	ser, err := discovery.NewMdnsService(
		ctx,
		peerHost,
		time.Minute*5,
		rendezvous,
	)
	if err != nil {
		panic(err)
	}
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser.RegisterNotifee(n)
	return n.PeerChan

}
