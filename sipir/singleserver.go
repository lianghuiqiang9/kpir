package sipir

import (
	rrand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"

	utils "github.com/local/utils"
)

const InvalidIndex = 0xFFFFFFFF

type Hist struct {
	C []uint32
	M []uint32
}

func (h *Hist) Init(m2 uint32) {
	h.C = make([]uint32, 0, 8192)
	h.M = make([]uint32, m2)
	for i := range h.M {
		h.M[i] = InvalidIndex
	}
}

func (h *Hist) Append(c uint32) {
	h.C = append(h.C, c)
	h.M[c] = uint32(len(h.C)) - 1
}

func (h *Hist) Get(t uint32) (uint32, bool) {
	if t >= uint32(len(h.C)) {
		return 0, false
	}
	return h.C[t], true
}

func (h *Hist) Inv(c uint32) (uint32, bool) {

	if c >= uint32(len(h.M)) {
		return 0, false
	}
	index := h.M[c]
	if index == InvalidIndex {
		return 0, false
	}
	return index, true
}

func (h *Hist) Copy() Hist {
	newC := make([]uint32, len(h.C), cap(h.C))
	copy(newC, h.C)

	newM := make([]uint32, len(h.M))
	copy(newM, h.M)

	return Hist{
		C: newC,
		M: newM,
	}
}

func (h *Hist) Size() float64 {
	if h == nil {
		return 0
	}
	return 48 + float64(len(h.C)*4) + float64(len(h.M)*4)
}

type DS struct {
	T    uint32
	M    uint32
	M2   uint32
	Hist *Hist
	Seed [16]byte
	P    []uint32 //P    [][]int
	Pinv []uint32 //Pinv [][]int
}

func (ds *DS) Size() float64 {
	var total float64
	total += 20

	if ds.Hist != nil {
		total += ds.Hist.Size()
	}
	return total
}

func init() {
	gob.Register(Hist{})
	gob.Register(DS{})
}

func (p *DS) Init(ck [16]byte, t, m, m2 uint32) {
	p.T = t
	p.M = m
	p.M2 = m2
	p.Hist = &Hist{}
	p.Hist.Init(m2)
	p.Seed = ck
}

func (p *DS) InitP() {
	size := p.T * p.M2
	p.P = make([]uint32, size)
	p.Pinv = make([]uint32, size)
	M2 := p.M2

	for i := uint32(0); i < p.T; i++ {
		offset := i * M2
		subP := p.P[offset : offset+M2]
		subPinv := p.Pinv[offset : offset+M2]

		for j := uint32(0); j < M2; j++ {
			subP[j] = j
		}

		hash := sha256.New()
		hash.Write(p.Seed[:])
		tempBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(tempBuf, i)
		hash.Write(tempBuf)

		digest := hash.Sum(nil)

		seedInt64 := int64(binary.LittleEndian.Uint64(digest[:8]))
		r := rand.New(rand.NewSource(seedInt64))
		//r := rand.New(rand.NewSource(int64(p.Seed + uint64(i))))
		r.Shuffle(int(M2), func(a, b int) {
			subP[a], subP[b] = subP[b], subP[a]
		})

		for j := uint32(0); j < M2; j++ {
			subPinv[subP[j]] = j
		}
	}
}

func (p *DS) Access(i, c uint32) (uint32, bool) {
	rowOffset := i * p.M2
	rowPinv := p.Pinv[rowOffset : rowOffset+p.M2]
	M := p.M

	for {
		pos := rowPinv[c]
		if pos < M {
			return pos, true
		}

		val, ok := p.Hist.Get(pos - M)
		if !ok {
			break
		}
		c = val
	}

	return 0, false
}

func (p *DS) Locate(i, e uint32) uint32 {
	rowOffset := i * p.M2
	rowP := p.P[rowOffset : rowOffset+p.M2]

	c := rowP[e]
	M := p.M

	for {
		val, ok := p.Hist.Inv(c)
		if !ok {
			break
		}
		c = rowP[M+val]
	}
	return c
}

func (p *DS) Relocate(c uint32) {
	p.Hist.Append(c)
}

// Rewind
func (p *DS) Rewind(b uint64) {
	if b == 0 || p.Hist == nil {
		return
	}

	currLen := len(p.Hist.C)
	if int(b) > currLen {
		b = uint64(currLen)
	}

	for i := 1; i <= int(b); i++ {
		cValue := p.Hist.C[currLen-i]
		p.Hist.M[cValue] = InvalidIndex
	}

	p.Hist.C = p.Hist.C[:currLen-int(b)]
}

type SingleServerParams struct {
	T              uint32
	M              uint32
	M2             uint32
	Q              uint64
	NumEntries     uint64
	Uint64PerEntry uint64
}

func (p *SingleServerParams) Print() string {
	return fmt.Sprintf(
		"Single-Server PIR Parameters:\n"+
			"- T:              %d\n"+
			"- M (Dimensions): %d\n"+
			"- M2 (Backup):    %d\n"+
			"- Q (Capacity):   %d\n"+
			"- NumEntries:     %d\n"+
			"- Uint64PerEntry:  %d",
		p.T, p.M, p.M2, p.Q, p.NumEntries, p.Uint64PerEntry,
	)
}

type SingleServerHint struct {
	DataState DS
	h         []uint64
}

func (h *SingleServerHint) Size() float64 {
	var total float64

	total += h.DataState.Size()

	total += float64(len(h.h) * 8)

	return total
}

type SingleServerMessage struct {
	Sender  string
	Payload [][]uint64
}

func (m *SingleServerMessage) Size() float64 {
	length := 8
	if m.Sender == "client" {
		length = 4
	}
	var total float64
	for _, entry := range m.Payload {
		total += float64(len(entry) * length)
	}
	return total
}

type SingleServerState struct {
	qflag [][]bool
	c     []uint32
	xi    []uint32
	xj    []uint32
}

func (s *SingleServerState) Print() string {
	rows := len(s.qflag)
	cols := 0
	if rows > 0 {
		cols = len(s.qflag[0])
	}

	return fmt.Sprintf(
		"Single-Server PIR State:\n"+
			"- QFlag Matrix:    %dx%d\n"+
			"- Consumed (c):    %v\n"+
			"- Coordinates (xi): %v\n"+
			"- Coordinates (xj): %v",
		rows, cols, s.c, s.xi, s.xj,
	)
}

type SingleServer struct {
	Params    *SingleServerParams
	Hint      *SingleServerHint
	Batchtype string
}

func (p *SingleServer) Name() string {
	fmt.Println("SingleServer-PIR")
	return "SingleServer-PIR"
}

func (p *SingleServer) InitParams(numEntries uint64, bitsPerEntry uint64, batchtype string) {
	if bitsPerEntry%32 != 0 {
		fmt.Println("bitsPerEntry should be 32 * k")
		os.Exit(1)
	}
	uint64PerEntry := (bitsPerEntry + 63) / 64
	// make sure numEntries is a perfect square
	numEntries = utils.NextPerfectSquare(numEntries)

	T := uint32(math.Sqrt(float64(numEntries)))

	M := uint32((numEntries) / uint64(T))

	if T*M != uint32(numEntries) {
		fmt.Println("make sure numEntries is a perfect square")
	}

	M2 := 2 * M

	Q := uint64(M2 - M)

	p.Params = &SingleServerParams{
		T:              T,
		M:              M,
		M2:             M2,
		Q:              Q,
		NumEntries:     numEntries,
		Uint64PerEntry: uint64PerEntry,
	}
	p.Hint = nil
	p.Batchtype = batchtype
}

func (p *SingleServer) LoadHint(id string, filepath string, db *utils.EncodedDB) (Hint, int64) {
	ckPath := filepath + id + "ck.gob"
	dsPath := filepath + id + "DataStateWithoutP.gob"
	hPath := filepath + id + "h.gob"

	flag1 := utils.FileExists(ckPath)
	flag2 := utils.FileExists(dsPath)
	flag3 := utils.FileExists(hPath)

	var ds DS
	var h []uint64
	var ck [16]byte

	if flag1 && flag2 && flag3 {
		fmt.Printf("[PIR] Loading existing hints for ID: %s\n", id)

		utils.LoadFromFile(&ck, ckPath)
		utils.LoadFromFile(&ds, dsPath)

		ds.InitP()

		utils.LoadFromFile(&h, hPath)

		p.Hint = &SingleServerHint{
			DataState: ds,
			h:         h,
		}
	} else {
		fmt.Printf("[PIR] Hints not found for ID: %s. Generating new ones...\n", id)

		p.GenerateHint(db)

		ds = p.Hint.DataState
		h = p.Hint.h
		ck = ds.Seed

		dsWithoutP := DS{}
		dsWithoutP.Init(ds.Seed, ds.T, ds.M, ds.M2)

		utils.SaveToFile(ck, ckPath)
		utils.SaveToFile(dsWithoutP, dsPath)
		utils.SaveToFile(h, hPath)
	}

	ckStat, _ := os.Stat(ckPath)
	dsStat, _ := os.Stat(dsPath)
	hStat, _ := os.Stat(hPath)

	//pMatrixSize := int64(ds.T*ds.M2) * 2 * 4

	size := ckStat.Size() + dsStat.Size() + hStat.Size()

	return p.Hint, size
}

func (p *SingleServer) GenerateHint(db *utils.EncodedDB) Hint {

	//ck := uint64(time.Now().UnixNano())
	var ck [16]byte
	if _, err := rrand.Read(ck[:]); err != nil {
		panic("failed to generate secure 128-bit seed: " + err.Error())
	}

	ds := DS{}

	T := p.Params.T
	M := p.Params.M
	M2 := p.Params.M2
	uint64PerEntry := db.Uint64PerEntry

	h := make([]uint64, uint64(M2)*uint64PerEntry)

	ds.Init(ck, T, M, M2)
	ds.InitP()

	for i := uint32(0); i < T; i++ {
		for j := uint32(0); j < M; j++ {
			c := ds.Locate(i, j)

			logicalIdx := uint64(i)*uint64(M) + uint64(j)

			if logicalIdx >= db.NumEntries {
				continue
			}
			entry := db.GetEntry(logicalIdx)
			if entry == nil {
				continue
			}

			hBase := uint64(c) * uint64PerEntry
			targetBucket := h[hBase : hBase+uint64PerEntry]

			db.XORInplace(targetBucket, entry)
		}
	}

	p.Hint = &SingleServerHint{
		DataState: ds,
		h:         h,
	}

	return p.Hint
}

/*
func (p *SingleServer) SaveHintState() SingleServer {
	return *p
}
*/

func (p *SingleServer) QueryAndFakeRefresh(indexes []uint64) (Message, State) {
	ds := p.Hint.DataState
	batchsize := len(indexes)
	T := p.Params.T
	M2 := p.Params.M2

	state := &SingleServerState{
		qflag: make([][]bool, batchsize),
		c:     make([]uint32, batchsize),
		xi:    make([]uint32, batchsize),
		xj:    make([]uint32, batchsize),
	}

	payload := make([][]uint64, batchsize)

	for i, x := range indexes {

		xi := uint32(x / uint64(p.Params.M))
		xj := uint32(x % uint64(p.Params.M))
		c := ds.Locate(xi, xj)

		query64 := make([]uint64, T)
		queryflag := make([]bool, T)

		for j := uint32(0); j < T; j++ {
			val, flag := ds.Access(j, c)
			query64[j] = uint64(val)
			queryflag[j] = flag
		}

		var r uint32
		for {
			r = rand.Uint32() % M2
			if _, ok := ds.Hist.Inv(r); !ok {
				break
			}
		}

		valR, flagR := ds.Access(xi, r)
		query64[xi] = uint64(valR)
		queryflag[xi] = flagR

		// FakeRefresh
		if i < batchsize-1 {
			ds.Relocate(c)
		}

		state.qflag[i] = queryflag
		state.c[i] = c
		state.xi[i] = xi
		state.xj[i] = xj
		payload[i] = query64
	}

	return &SingleServerMessage{
		Sender:  "client",
		Payload: payload,
	}, state
}

func (p *SingleServer) Answer(db *utils.EncodedDB, req Message) Message {
	msg, ok := req.(*SingleServerMessage)
	if !ok {
		return nil
	}

	T := p.Params.T
	M := p.Params.M
	uint64PerEntry := p.Params.Uint64PerEntry

	answerPayload := make([][]uint64, len(msg.Payload))

	for i, query64 := range msg.Payload {
		ansFromQuery := make([]uint64, uint64(T)*uint64PerEntry)

		for row := uint32(0); row < T; row++ {
			colIndex := query64[row]

			dbIndex := uint64(row)*uint64(M) + colIndex

			entry := db.GetEntry(dbIndex)
			if entry == nil {
				continue
			}

			destStart := uint64(row) * uint64PerEntry
			target := ansFromQuery[destStart : destStart+uint64PerEntry]

			copy(target, entry)
		}

		answerPayload[i] = ansFromQuery
	}

	return &SingleServerMessage{
		Sender:  "server",
		Payload: answerPayload,
	}
}

// rewind batchsize-1 times
func (p *SingleServer) Rewind(state State) {
	st, _ := state.(*SingleServerState)
	if p.Hint == nil {
		return
	}
	p.Hint.DataState.Rewind(uint64(len(st.c)) - 1)
}

func (p *SingleServer) ReconstructAndRefresh(resp Message, st State, indexes []uint64) [][]uint64 {
	serverResp, ok := resp.(*SingleServerMessage)
	if !ok || len(serverResp.Payload) == 0 {
		return nil
	}
	state, ok := st.(*SingleServerState)
	if !ok {
		return nil
	}

	batchSize := len(serverResp.Payload)
	T := p.Params.T
	uint64PerEntry := p.Params.Uint64PerEntry

	results := make([][]uint64, batchSize)

	for b := 0; b < batchSize; b++ {
		ansFromQuery := serverResp.Payload[b]
		xi := state.xi[b]
		queryflag := state.qflag[b]
		c := state.c[b]

		// 1. reconstruct
		answer := make([]uint64, uint64PerEntry)
		hStart := uint64(c) * uint64(uint64PerEntry)
		copy(answer, p.Hint.h[hStart:hStart+uint64(uint64PerEntry)])

		for i := uint32(0); i < T; i++ {
			if i != xi && queryflag[i] {
				ansStart := uint64(i) * uint64(uint64PerEntry)
				temp := ansFromQuery[ansStart : ansStart+uint64(uint64PerEntry)]

				for k := uint64(0); k < uint64PerEntry; k++ {
					answer[k] ^= temp[k]
				}
			}
		}
		p.Hint.DataState.Relocate(c)

		// 2. refresh
		p.Refresh(answer, b, state, ansFromQuery)

		results[b] = answer
	}

	return results
}

func (p *SingleServer) Refresh(answer []uint64, b int, state *SingleServerState, ansFromQuery []uint64) {
	T := p.Params.T
	uint64PerEntry := p.Params.Uint64PerEntry
	ds := p.Hint.DataState

	xi := state.xi[b]
	xj := state.xj[b]
	c := state.c[b]
	queryflag := state.qflag[b]

	targetPos := uint64(xi) * uint64PerEntry
	copy(ansFromQuery[targetPos:targetPos+uint64PerEntry], answer)

	for i := uint32(0); i < T; i++ {
		if queryflag[i] || i == xi {
			var col uint32
			if i == xi {
				col = xj
			} else {
				col, _ = ds.Access(i, c)
			}

			targetC := ds.Locate(i, col)

			hOff := uint64(targetC) * uint64PerEntry
			ansOff := uint64(i) * uint64PerEntry

			for k := uint64(0); k < uint64PerEntry; k++ {
				p.Hint.h[hOff+uint64(k)] ^= ansFromQuery[ansOff+uint64(k)]
			}
		}
	}
}

func (p *SingleServer) GetHintSize() float64 {
	return p.Hint.Size()
}

func (p *SingleServer) GetParamQ() uint64 {
	return p.Params.Q
}
