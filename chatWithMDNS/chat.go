package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"os"
)

func handlerStream(stream network.Stream) {
	fmt.Println("获得一个stream")
	//新建一个rw用于对stream进行读写
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if len(str) == 0 {
			return
		}
		if str != "\n" {
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			panic(err)
		}
	}
}

func main() {

	config := parseFlags() //新建一个config
	fmt.Printf("[*]listening on %s:%d\n", config.host, config.port)
	ctx := context.Background() //创建context

	//生成用于加密或产生peerid的密钥
	r := rand.Reader
	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	//创建新的监听地址
	sourceMAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", config.host, config.port))
	//创建节点host
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMAddr),
		libp2p.Identity(privKey),
	)
	if err != nil {
		panic(err)
	}

	//设置stream 处理器
	host.SetStreamHandler(protocol.ID(config.Protocol), handlerStream)

	fmt.Printf("MAddr:%s", sourceMAddr)

	peerChan := initMDNS(ctx, host, config.rendezvous)
	peer := <-peerChan

	//先创建连接,再开启流
	if err := host.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.Protocol))

	//这是连接发起方使用
	if err != nil {
		fmt.Println("Stream open failed", err)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go writeData(rw)
		go readData(rw)
		fmt.Println("Connected to:", peer)
	}

	select {} //wait here

}
