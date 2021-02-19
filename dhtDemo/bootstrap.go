package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		port int
	)
	setFlags(ctx, &port)

	//设置host的option
	listenAddr := libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	privKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	identify := libp2p.Identity(privKey)
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		dualDHT, err := ddht.New(ctx, h, ddht.DHTOption(dht.Mode(dht.ModeServer), dht.ProtocolPrefix("/myapp"))) //作为dhtServer
		//,dht.ProtocolPrefix("/myapp")
		_ = dualDHT.Bootstrap(ctx)

		go func() {
			ticker := time.NewTicker(time.Second * 15)
			time.Sleep(time.Second)
			for {
				fmt.Printf("*****定时输出RoutingTable******\n")
				dualDHT.LAN.RoutingTable().Print()
				<-ticker.C
			}
		}()
		return dualDHT, err
	})

	host, err := libp2p.New(
		ctx,
		identify,
		routing,
		listenAddr,
		//libp2p.NATPortMap(),
	)
	if err != nil {
		panic(err)
	}

	for _, addr := range host.Addrs() {
		fmt.Printf("Addr: %s/p2p/%s\n", addr, host.ID().Pretty())
	}

	host.SetStreamHandler(dht.ProtocolDHT, func(stream network.Stream) {
		fmt.Printf("handling %s\n", stream)
	})

	select {}
}

func setFlags(ctx context.Context, port *int) {
	flag.IntVar(port, "port", 6666, "")
}
