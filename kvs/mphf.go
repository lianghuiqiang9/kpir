package kvs

import (
	bbhash "github.com/local/bbhash"
	consensusrecsplit "github.com/local/consensusrecsplit"
	pthash "github.com/local/pthash"
)

type MPHFInterface interface {
	Lookup(uint64) uint64
	Bits() uint64
	Free()
}

type TMPHF uint32

type PTHashKVS struct {
	MPHFKVS[*pthash.PTHash, TMPHF]
}

func NewPTHashKVS() *PTHashKVS {
	return &PTHashKVS{
		MPHFKVS: MPHFKVS[*pthash.PTHash, TMPHF]{
			NameStr:    "pthashkvs",
			NewMPHFunc: func(keys []uint64) *pthash.PTHash { return pthash.New(keys) },
		},
	}
}

type BBHashKVS struct {
	MPHFKVS[*bbhash.BBHash, TMPHF]
}

func NewBBHashKVS() *BBHashKVS {
	return &BBHashKVS{
		MPHFKVS: MPHFKVS[*bbhash.BBHash, TMPHF]{
			NameStr:    "bbhashkvs",
			NewMPHFunc: func(keys []uint64) *bbhash.BBHash { return bbhash.New(keys) },
		},
	}
}

type ConsensusRecSplitKVS struct {
	MPHFKVS[*consensusrecsplit.MPH, TMPHF]
}

func NewConsensusRecSplitKVS() *ConsensusRecSplitKVS {
	return &ConsensusRecSplitKVS{
		MPHFKVS: MPHFKVS[*consensusrecsplit.MPH, TMPHF]{
			NameStr:    "consensusrecsplitkvs",
			NewMPHFunc: func(keys []uint64) *consensusrecsplit.MPH { return consensusrecsplit.New(keys) },
		},
	}
}
