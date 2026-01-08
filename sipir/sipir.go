package sipir

import "github.com/local/utils"

type Parameters interface {
	Print() string
}

type Hint interface {
	Size() float64
}

type Message interface {
	Size() float64
}

type State interface {
	Print() string
}

type SIPIR interface {
	Name() string

	InitParams(numEntries uint64, uint64PerEntry uint64, Type string)

	GenerateHint(db *utils.EncodedDB) Hint

	//LoadHint(id string, filepath string, db *utils.EncodedDB) (hint Hint, hintsize int64)

	//Query(indexes []uint64) (req Message, st State)

	Answer(db *utils.EncodedDB, resp Message) Message

	//Reconstruct(resp Message, st State) [][]uint64

	//Refresh(indexes []uint64, answers [][]uint64, st State)

	//SaveHintState() SIPIR

	QueryAndFakeRefresh(indexes []uint64) (req Message, st State)

	Rewind(state State)

	ReconstructAndRefresh(resp Message, st State, indexes []uint64) [][]uint64

	//Verify(db []uint64, index []uint64, answer [][]uint64) []bool

	GetHintSize() float64

	GetParamQ() uint64
}
