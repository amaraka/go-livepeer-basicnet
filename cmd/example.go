package main

import (
	"flag"
	"io/ioutil"
	"time"

	net "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	peerstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
	"gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"

	"github.com/golang/glog"
	basicnet "github.com/livepeer/go-livepeer-basicnet"
)

var timer time.Time

func main() {
	p := flag.Int("p", 15000, "port")
	id := flag.String("id", "", "id")
	addr := flag.String("addr", "", "addr")
	init := flag.Bool("i", false, "initialize message sending")
	ping := flag.Bool("ping", false, "ping test")

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	priv, pub, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
	node, _ := basicnet.NewNode(*p, priv, pub, &basicnet.BasicNotifiee{})

	pid, _ := peer.IDHexDecode(*id)
	if *id != "" {
		paddr, _ := ma.NewMultiaddr(*addr)
		node.PeerHost.Peerstore().AddAddr(pid, paddr, peerstore.PermanentAddrTTL)
	}

	if *ping {
		pingtest(*init, node, pid)
	}

}

func pingtest(init bool, node *basicnet.NetworkNode, pid peer.ID) {
	if init {
		timer = time.Now()
		strm := node.GetOutStream(pid)
		glog.Infof("Sending message")
		strm.SendMessage(basicnet.SimpleString, "")
	} else {
		setHandler(node)
		glog.Infof("Done setting handler")
	}

	select {}
}

func setHandler(n *basicnet.NetworkNode) {
	n.PeerHost.SetStreamHandler(basicnet.Protocol, func(stream net.Stream) {
		ws := basicnet.NewBasicInStream(stream)

		for {
			if err := streamHandler(ws, n); err != nil {
				glog.Errorf("Error handling stream: %v", err)
				// delete(n.NetworkNode.streams, stream.Conn().RemotePeer())
				stream.Close()
				return
			}
		}
	})
}

func streamHandler(ws *basicnet.BasicInStream, n *basicnet.NetworkNode) error {
	msg, err := ws.ReceiveMessage()
	if err != nil {
		glog.Errorf("Got error decoding msg: %v", err)
		return err
	}

	glog.Infof("%v Recieved msg %v from %v", peer.IDHexEncode(ws.Stream.Conn().LocalPeer()), msg.Op, peer.IDHexEncode(ws.Stream.Conn().RemotePeer()))
	glog.Infof("Time since last message recieved: %v", time.Since(timer))

	timer = time.Now()
	vid, _ := ioutil.ReadFile("./test.ts")
	return n.GetOutStream(ws.Stream.Conn().RemotePeer()).SendMessage(basicnet.StreamDataID, basicnet.StreamDataMsg{Data: vid, SeqNo: 0, StrmID: "test"})
}
