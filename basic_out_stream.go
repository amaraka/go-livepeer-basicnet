package basicnet

import (
	"bufio"
	"errors"
	"fmt"
	"sync"

	net "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"

	multicodec "github.com/multiformats/go-multicodec"
	mcjson "github.com/multiformats/go-multicodec/json"

	"github.com/golang/glog"
)

var ErrOutStream = errors.New("ErrOutStream")

//BasicStream is a libp2p stream wrapped in a reader and a writer.
type BasicOutStream struct {
	Stream net.Stream
	enc    multicodec.Encoder
	w      *bufio.Writer
	el     *sync.Mutex
}

//NewBasicStream creates a stream from a libp2p raw stream.
func NewBasicOutStream(s net.Stream) *BasicOutStream {
	writer := bufio.NewWriter(s)
	// This is where we pick our specific multicodec. In order to change the
	// codec, we only need to change this place.
	// See https://godoc.org/github.com/multiformats/go-multicodec/json
	enc := mcjson.Multicodec(false).Encoder(writer)

	return &BasicOutStream{
		Stream: s,
		w:      writer,
		enc:    enc,
		el:     &sync.Mutex{},
	}
}

//SendMessage writes a message into the stream.
func (bs *BasicOutStream) SendMessage(opCode Opcode, data interface{}) error {
	// glog.V(common.DEBUG).Infof("Sending msg %v to %v", opCode, peer.IDHexEncode(bs.Stream.Conn().RemotePeer()))
	msg := Msg{Op: opCode, Data: data}
	return bs.encodeAndFlush(msg)
}

//EncodeAndFlush writes a message into the stream.
func (bs *BasicOutStream) encodeAndFlush(n interface{}) error {
	if bs == nil {
		fmt.Println("stream is nil")
	}

	bs.el.Lock()
	defer bs.el.Unlock()
	err := bs.enc.Encode(n)
	if err != nil {
		glog.Errorf("send message encode error for peer %v: %v", peer.IDHexEncode(bs.Stream.Conn().RemotePeer()), err)
		return ErrOutStream
	}

	err = bs.w.Flush()
	if err != nil {
		glog.Errorf("send message flush error for peer %v: %v", peer.IDHexEncode(bs.Stream.Conn().RemotePeer()), err)
		return ErrOutStream
	}

	return nil
}
