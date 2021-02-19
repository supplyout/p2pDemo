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
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/multiformats/go-multiaddr"
	"strings"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bootstrapPeers addrList
	room := flag.String("room", "", "")
	joinRoom := flag.String("joinRoom", "", "")
	flag.Var(&bootstrapPeers, "b", "")
	flag.Parse()
	if *room == "" && *joinRoom == "" {
		fmt.Printf("请选择加入room(-room [roomName])或者创建room(-joinRoom [roomName])\n")
		return
	}
	if len(bootstrapPeers) == 0 {
		bootstrapPeers = dht.DefaultBootstrapPeers
	}
	//设置host的option

	//设置routingDiscovery
	var dualDHT *ddht.DHT
	var routingDiscovery *discovery.RoutingDiscovery
	routing := libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		var err error
		dualDHT, err = ddht.New(ctx, host, ddht.DHTOption(dht.ProtocolPrefix("/myapp"), dht.Mode(dht.ModeServer)))
		//,dht.ProtocolPrefix("/myapp")
		routingDiscovery = discovery.NewRoutingDiscovery(dualDHT)
		//_ = dualDHT.Bootstrap(ctx)
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

	privKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	identify := libp2p.Identity(privKey)

	listenAddr := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")

	//创建host
	host, err := libp2p.New(
		ctx,
		routing,
		identify,
		listenAddr,
		//libp2p.NATPortMap(),
	)
	if err != nil {
		panic(err)
	}
	host.SetStreamHandler(protocol.ID("/chat/1.0"), handleStream)
	for _, addr := range host.Addrs() {
		fmt.Printf("Address: %s\n", addr)
	}
	fmt.Printf("local id is: %s\n", host.ID().Pretty())

	//连接到bootstrap peer
	//if err := dualDHT.Bootstrap(ctx); err != nil {
	//	panic(err)
	//}

	var wg sync.WaitGroup
	for _, maddr := range bootstrapPeers {
		wg.Add(1)
		peerInfo, _ := peer.AddrInfoFromP2pAddr(maddr)
		go func() {
			defer wg.Done()
			fmt.Printf("尝试连接到:%s\n", peerInfo.ID.Pretty())
			if err := host.Connect(ctx, *peerInfo); err != nil {
				fmt.Printf("连接失败:%s\n", peerInfo.ID.Pretty())
			} else {
				fmt.Printf("成功连接到:%s\n", peerInfo.ID.Pretty())
			}
		}()
	}
	wg.Wait()

	fmt.Printf("成功连接到bootstrapPeer\n")

	fmt.Printf("这里是路由表\n")
	fmt.Printf("LAN:\n")
	dualDHT.LAN.RoutingTable().Print()
	fmt.Printf("WAN:\n")
	dualDHT.WAN.RoutingTable().Print()

	//处理聊天
	if *room != "" {
		//创建房间
		fmt.Printf("%s\n", *room)
		discovery.Advertise(ctx, routingDiscovery, *room)
		fmt.Printf("成功将room:%s 进行广播\n", *room)
		select {}
	}

	if *joinRoom != "" {
		for {
			fmt.Printf("开始寻找peers\n")
			peerInfos, err := discovery.FindPeers(ctx, routingDiscovery, *joinRoom)
			if len(peerInfos) != 0 {
				discovery.Advertise(ctx, routingDiscovery, *joinRoom)
			} else {
				continue
			}

			fmt.Printf("peers:\n")
			for i, pe := range peerInfos {
				fmt.Printf("(%d):%s\n", i, pe.ID.Pretty())
			}
			if err != nil {
				panic(err)
			}
			for _, peerInfo := range peerInfos {
				fmt.Printf("找到peer:%s\n", peerInfo.ID.Pretty())
				if peerInfo.ID == host.ID() {
					continue
				}
				fmt.Printf("尝试连接peer:%s\n", peerInfo.ID.Pretty())
				if err := host.Connect(ctx, peerInfo); err != nil {
					fmt.Printf("连接peer失败:%s\n", peerInfo.ID.Pretty())
					continue
				}
				fmt.Printf("成功连接peer:%s\n", peerInfo.ID.Pretty())
				fmt.Printf("尝试创建stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())
				stream, err := host.NewStream(ctx, peerInfo.ID, "/chat/1.0")
				if err != nil {
					panic(err)
				} else {
					fmt.Printf("成功创建stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())

				}
				go handleStream(stream)
			}
			time.Sleep(time.Minute * 2)
		}
	}

}
func handleStream(stream network.Stream) {
	go fmt.Printf("获得一个新stream\n")
}

// A new type we need for writing a custom flag parser
type addrList []multiaddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (multiaddrs []multiaddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := multiaddr.NewMultiaddr(addrString)
		if err != nil {
			return multiaddrs, err
		}
		multiaddrs = append(multiaddrs, addr)
	}
	return
}
