package network

import (
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	key "github.com/djbarber/ipfs-hack/blocks/key"
	bsmsg "github.com/djbarber/ipfs-hack/exchange/bitswap/message"
	peer "github.com/djbarber/ipfs-hack/p2p/peer"
	protocol "github.com/djbarber/ipfs-hack/p2p/protocol"
)

var ProtocolBitswap protocol.ID = "/ipfs/bitswap"

// BitSwapNetwork provides network connectivity for BitSwap sessions
type BitSwapNetwork interface {

	// SendMessage sends a BitSwap message to a peer.
	SendMessage(
		context.Context,
		peer.ID,
		bsmsg.BitSwapMessage) error

	// SetDelegate registers the Reciver to handle messages received from the
	// network.
	SetDelegate(Receiver)

	ConnectTo(context.Context, peer.ID) error

	Routing
}

// Implement Receiver to receive messages from the BitSwapNetwork
type Receiver interface {
	ReceiveMessage(
		ctx context.Context,
		sender peer.ID,
		incoming bsmsg.BitSwapMessage)

	ReceiveError(error)

	// Connected/Disconnected warns bitswap about peer connections
	PeerConnected(peer.ID)
	PeerDisconnected(peer.ID)
}

type Routing interface {
	// FindProvidersAsync returns a channel of providers for the given key
	FindProvidersAsync(context.Context, key.Key, int) <-chan peer.ID

	// Provide provides the key to the network
	Provide(context.Context, key.Key) error
}
