package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/config"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"sync"
	"time"
)

const (
	SerTypeNorm = iota
	SerTypeBoot
)

type p2pServer struct {
	host        host.Host
	dualDHT     *ddht.DHT
	rtDiscovery *discovery.RoutingDiscovery
}

func NewP2PServer(ctx context.Context, ch map[string]chan string, bindings map[string]interface{}, serType int, bootstrapPeerStr []string) p2pServer {
	s := p2pServer{}
	isLocal := false
	var bootstrapPeers []peer.AddrInfo
	if len(bootstrapPeerStr) == 0 && serType == SerTypeNorm {
		bootstrapPeers, _ = peer.AddrInfosFromP2pAddrs(dht.DefaultBootstrapPeers...)
	} else {
		isLocal = true
		bootstrapPeers = convertPeers(bootstrapPeerStr)
	}
	for _, p := range bootstrapPeers {
		fmt.Printf("%s\n", p.ID)
	}
	privKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		var err error
		if isLocal {
			s.dualDHT, err = ddht.New(ctx, h, ddht.DHTOption(dht.ProtocolPrefix("/myapp"), dht.Mode(dht.ModeServer)))
		}
		s.dualDHT, err = ddht.New(ctx, h, ddht.DHTOption(dht.Mode(dht.ModeServer)))
		if err != nil {
			panic(err)
		}
		s.rtDiscovery = discovery.NewRoutingDiscovery(s.dualDHT)
		return s.dualDHT, err
	})
	opts := []config.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/0")),
		libp2p.Identity(privKey),
		routing,
		libp2p.NATPortMap(),
	}
	s.host, _ = libp2p.New(ctx, opts...)

	s.host.SetStreamHandler("/android/1.0", func(s network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go readData(rw, ch["out"])
		go writeData(rw, ch)
	})
	s.host.ConnManager()
	if serType == SerTypeNorm {
		var wg sync.WaitGroup
		c := make(chan error, len(bootstrapPeers))
		for _, peerInfo := range bootstrapPeers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Printf("尝试连接到:%s\n", peerInfo.ID.Pretty())
				if err := s.host.Connect(ctx, peerInfo); err != nil {
					fmt.Printf("连接失败:%s\n", peerInfo.ID.Pretty())
					c <- err
					return
				}
				fmt.Printf("连接成功:%s\n", peerInfo.ID.Pretty())
			}()
		}
		wg.Wait()
		close(c)
		count := 0
		for err := range c {
			if err != nil {
				count++
			}
		}

		// 至少有一个bootstrap节点成功连接
		if len(bootstrapPeers) > count {
			fmt.Printf("成功连接%d个peer\n", len(bootstrapPeers)-count)

			//ch <- "连接bootstrap成功"
		} else {
			//ch <- "连接bootstrap失败"
			return s
		}

		rendezvous, _ := bindings["rendezvousBS"].(binding.ExternalString).Get()
		discovery.Advertise(ctx, s.rtDiscovery, rendezvous)
		fmt.Printf("成功发送Advertise：%s\n", rendezvous)
		ch["out"] <- fmt.Sprintf("成功发送Advertise：%s\n", rendezvous)
		ch["out"] <- "成功创建server"
		s.host.Network().Peers()
		go func() {
			for {
				time.Sleep(time.Second * 15)

				peerInfos, _ := discovery.FindPeers(ctx, s.rtDiscovery, rendezvous)
				if len(peerInfos) == 0 {
					ch["out"] <- fmt.Sprintf("未找到peer\n")
					continue
				}
				ch["out"] <- fmt.Sprintf("成功找到%d个peer：\n", len(peerInfos))
				fmt.Printf("成功找到%d个peer：\n", len(peerInfos))
				time.Sleep(time.Second * 2)
				for _, peerInfo := range peerInfos {
					ch["out"] <- fmt.Sprintf("找到:%s\n", peerInfo.ID.Pretty())
					fmt.Printf("%s\n", peerInfo.ID.Pretty())
					if peerInfo.ID == s.host.ID() {
						continue
					}

					//isContains := false
					//peers:=s.host.Peerstore().Peers()
					//for _,p:=range peers{
					//	if p == peerInfo.ID{
					//		isContains = true
					//	}
					//}
					//if isContains{
					//	ch["out"] <- fmt.Sprintf("没有找到新peer\n")
					//	fmt.Printf("没有找到新peer\n")
					//	continue
					//}

					fmt.Printf("尝试创建stream:%s<--->%s\n>", s.host.ID().Pretty(), peerInfo.ID.Pretty())
					stream, err := s.host.NewStream(ctx, peerInfo.ID, "/android/1.0")
					if err != nil {

						fmt.Printf("创建失败stream:%s<--->%s\n>", s.host.ID().Pretty(), peerInfo.ID.Pretty())
						ch["out"] <- fmt.Sprintf("创建失败stream:%s<--->%s\n", s.host.ID().Pretty(), peerInfo.ID.Pretty())
						ch["out"] <- fmt.Sprintf("创建失败stream:%s\n", err.Error())
						continue
					}
					ch["out"] <- fmt.Sprintf("成功连接到peer:%s\n", peerInfo.ID.Pretty())
					fmt.Printf("成功连接到peer:%s\n", peerInfo.ID.Pretty())

					rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
					_, err = rw.WriteString("你好，我向你创建了一个流\n")
					if err != nil {
						ch["out"] <- err.Error()
						return
					}
					_ = rw.Flush()
					go readData(rw, ch["out"])
					go writeData(rw, ch)

				}
			}
		}()
	}

	return s
}
func convertPeers(peers []string) []peer.AddrInfo {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, addr := range peers {
		maddr := ma.StringCast(addr)
		p, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatalln(err)
		}
		pinfos[i] = *p
	}
	return pinfos
}

func readData(rw *bufio.ReadWriter, ch chan string) {
	for {
		s, err := rw.ReadString('\n')
		if err != nil {
			ch <- err.Error()
			continue
		}
		if s == "" {
			continue
		}
		if s != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			ch <- s
			fmt.Printf("\x1b[32m%s\x1b[0m> ", s)
		}
	}
}

func writeData(rw *bufio.ReadWriter, ch map[string]chan string) {
	for {
		select {
		case s := <-ch["in"]:
			fmt.Printf("准备发送：%s\n", s)
			_, err := rw.WriteString(s)
			if err != nil {
				ch["out"] <- err.Error()
			}
			err = rw.Flush()
			if err != nil {
				ch["out"] <- err.Error()
			}
		}
	}
}
