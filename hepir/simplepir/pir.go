package simplepir

import (
	utils "github.com/local/utils"
)

type HEPIR interface {
	// 基础配置
	InitParams(N, bitsPerEntry uint64) Params
	Name() string

	MakeInternalDB(db *utils.EncodedDB) *InternalDB

	Setup(DB *InternalDB) (State, State)

	Query(indexes []uint64) (Msg, State)

	QueryOffline(batchSize uint64) (Msg, State)

	QueryOnline(indexes []uint64, offlineMsgs Msg) Msg

	Answer(DB *InternalDB, queries Msg, serverHint State) Msg

	Reconstruct(indexes []uint64, clientHint State, query Msg, answer Msg, clientState State) [][]uint64
}
