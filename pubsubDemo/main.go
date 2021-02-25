package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"sync"
	"time"
)

type Message struct {
	Content string
}

func main() {
	ctx := context.Background()

	var dhtNew *dht.IpfsDHT
	var rtDiscovery *discovery.RoutingDiscovery
	routing := libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		var err error
		dhtNew, err = dht.New(ctx, host, dht.Mode(dht.ModeServer))
		rtDiscovery = discovery.NewRoutingDiscovery(dhtNew)
		//_ = dhtNew.Bootstrap(ctx)
		return dhtNew, err
	})

	h, err := libp2p.New(ctx, routing, libp2p.NATPortMap())
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				fmt.Printf("%s\n", err.Error())
			} else {
				fmt.Printf("Connection established with bootstrap node:%s\n", *peerinfo)
			}
		}()
	}
	wg.Wait()

	_, _ = rtDiscovery.Advertise(ctx, "supplyout111")

	fmt.Printf("我的peerID：%s\n", h.ID().Pretty())
	gossip, err := pubsub.NewGossipSub(ctx, h)

	mTopic := "2521796328"

	topic, err := gossip.Join(mTopic)
	if err == nil {
		fmt.Printf("join成功\n")
	}
	sub, err := topic.Subscribe()
	if err == nil {
		fmt.Printf("subscribe成功\n")
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				continue
			}

			if msg.ReceivedFrom == h.ID() {
				continue
			}
			fmt.Printf("接收msg成功\n")

			message := new(Message)
			fmt.Printf("raw:%s", string(msg.Data))
			err = json.Unmarshal(msg.Data, &message)
			if err != nil {
				panic(err)
			}
			fmt.Printf("我从%s收到了：%s\n", msg.ReceivedFrom[0:6], message.Content)
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	ticker2 := time.NewTicker(time.Second * 10)
	ticker3 := time.NewTicker(time.Second * 8)

	for {
		select {
		case <-ticker.C:
			msg := Message{
				Content: "定时发送",
			}
			sendData, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			fmt.Printf("sendData:%s\n", sendData)
			if err != nil {
				panic(err)
			}
			err = topic.Publish(ctx, sendData)
			if err != nil {
				fmt.Print(err.Error())
			} else {
				fmt.Printf("已发送\n")
			}

		case <-ticker2.C:
			peers := gossip.ListPeers(mTopic)
			fmt.Printf("该主题共有%d个peer\n", len(peers))
			for _, p := range peers {
				fmt.Printf("peer %s\n", p)
			}
		case <-ticker3.C:
			peerInfos, err := discovery.FindPeers(ctx, rtDiscovery, "supplyout111")
			if err != nil {
				fmt.Printf(err.Error())
			}
			fmt.Printf("通过findPeers找到：\n")
			for _, p := range peerInfos {
				if p.ID == h.ID() {
					continue
				}
				fmt.Printf("peer %s\n", p.ID.Pretty()[0:6])
				if h.Network().Connectedness(p.ID) == network.Connected {
					fmt.Printf("已经建立连接：%s\n", p.ID.Pretty()[0:6])
				} else {
					err := h.Connect(ctx, p)
					if err != nil {
						fmt.Printf("连接失败：%s,err:%s\n", p.ID.Pretty()[0:6], err.Error())
					} else {
						fmt.Printf("连接成功：%s\n", p.ID.Pretty()[0:6])
					}
				}

			}
			//dhtNew.RoutingTable().Print()
		}
	}

}
