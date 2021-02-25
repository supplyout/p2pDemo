package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const gossipSubID = "/meshsub/1.0.0"

func main() {

	//golog.SetAllLoggers(gologging.DEBUG) // Change to DEBUG for extra info
	h1 := newHost(2001)
	h2 := newHost(2002)
	h3 := newHost(2003)
	fmt.Printf("host 1: \n\t-Addr:%s\n\t-ID: %s\n", h1.Addrs()[0], h1.ID().Pretty())
	fmt.Printf("host 2: \n\t-Addr:%s\n\t-ID: %s\n", h2.Addrs()[0], h2.ID().Pretty())
	fmt.Printf("host 3: \n\t-Addr:%s\n\t-ID: %s\n", h3.Addrs()[0], h3.ID().Pretty())

	time.Sleep(100 * time.Millisecond)

	// add h1 to h2's store
	h2.Peerstore().AddAddr(h1.ID(), h1.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h1's store
	h1.Peerstore().AddAddr(h2.ID(), h2.Addrs()[0], pstore.PermanentAddrTTL)
	// add h3 to h2's store
	h2.Peerstore().AddAddr(h3.ID(), h3.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h3's store
	h3.Peerstore().AddAddr(h3.ID(), h3.Addrs()[0], pstore.PermanentAddrTTL)

	// ---- gossip sub part
	topic := "random"
	g1, err := pubsub.NewGossipSub(context.Background(), h1)
	requireNil(err)
	g2, err := pubsub.NewGossipSub(context.Background(), h2)
	requireNil(err)
	g3, err := pubsub.NewGossipSub(context.Background(), h3)
	requireNil(err)

	t1, err := g1.Join(topic)
	requireNil(err)
	t2, err := g2.Join(topic)
	requireNil(err)
	t3, err := g3.Join(topic)
	requireNil(err)

	s2, _ := t2.Subscribe()
	s3, _ := t3.Subscribe()
	time.Sleep(2 * time.Second)

	// 1 connect to 2 and 2 connect to 3
	err = h1.Connect(context.Background(), h2.Peerstore().PeerInfo(h2.ID()))
	requireNil(err)
	err = h2.Connect(context.Background(), h3.Peerstore().PeerInfo(h3.ID()))
	requireNil(err)

	// publish and read
	msg := []byte("Hello Word")
	requireNil(t1.Publish(context.Background(), msg))

	pbMsg, err := s2.Next(context.Background())
	requireNil(err)
	checkEqual(msg, pbMsg.Data)
	fmt.Println(" GOSSIPING WORKS #1")

	pbMsg, err = s3.Next(context.Background())
	requireNil(err)
	checkEqual(msg, pbMsg.Data)
	fmt.Println(" GOSSIPING WORKS #2")
}

func checkEqual(exp, rcvd []byte) {
	if !bytes.Equal(exp, rcvd) {
		panic("not equal")
	}
}

func requireNil(err error) {
	if err != nil {
		panic(err)
	}
}

func newHost(port int) host.Host {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.DisableRelay(),
	}
	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}
	return basicHost
}
