package bstest

import (
	. "github.com/djbarber/ipfs-hack/blockservice"
	bitswap "github.com/djbarber/ipfs-hack/exchange/bitswap"
	tn "github.com/djbarber/ipfs-hack/exchange/bitswap/testnet"
	mockrouting "github.com/djbarber/ipfs-hack/routing/mock"
	delay "github.com/djbarber/ipfs-hack/thirdparty/delay"
)

// Mocks returns |n| connected mock Blockservices
func Mocks(n int) []*BlockService {
	net := tn.VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	sg := bitswap.NewTestSessionGenerator(net)

	instances := sg.Instances(n)

	var servs []*BlockService
	for _, i := range instances {
		servs = append(servs, New(i.Blockstore(), i.Exchange))
	}
	return servs
}
