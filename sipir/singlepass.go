package sipir

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	utils "github.com/local/utils"
)

type SinglePassParams struct {
	T              uint32
	M              uint32
	Q              uint64
	NumEntries     uint64
	Uint64PerEntry uint64
}

func (p *SinglePassParams) Print() string {
	return fmt.Sprintf(
		"SinglePass PIR Parameters:\n"+
			"- T (Rows):        %d\n"+
			"- M (Columns):     %d\n"+
			"- Q (Max Batch):   %d\n"+
			"- NumEntries:      %d\n"+
			"- Uint64PerEntry:   %d",
		p.T, p.M, p.Q, p.NumEntries, p.Uint64PerEntry,
	)
}

type SinglePassHint struct {
	ck uint64
	p  []uint32 // generate by ck, do not compute in hintsize
	h  []uint64
}

func (h *SinglePassHint) Size() float64 {
	var total float64
	total += 8
	total += float64(len(h.h) * 8)

	return total
}

type SinglePassMessage struct {
	Sender         string
	Payload        [][]uint64
	UpdatedSender  string
	UpdatedPayload [][]uint64
}

func (m *SinglePassMessage) Size() float64 {
	var total float64

	length := 8
	if m.Sender == "client" {
		length = 4
	}

	for _, entry := range m.Payload {
		total += float64(len(entry) * length)
	}

	length2 := 8
	if m.UpdatedSender == "client" {
		length2 = 4
	}

	for _, entry := range m.UpdatedPayload {
		total += float64(len(entry) * length2)
	}

	return total
}

type SinglePassState struct {
	rv  [][]uint32
	xi  []uint32
	ind []uint32
}

func (s *SinglePassState) Print() string {
	return fmt.Sprintf(
		"SinglePass Client State:\n"+
			"- Random Vectors (rv): %d sets\n"+
			"- Indices (ind):       %d entries\n"+
			"- Mask (xi):           %d elements",
		len(s.rv), len(s.ind), len(s.xi),
	)
}

type SinglePass struct {
	Params    *SinglePassParams
	Hint      *SinglePassHint
	batchtype string
}

func (p *SinglePass) Name() string {
	fmt.Println("SinglePass-PIR")
	return "SinglePass-PIR"
}

func (p *SinglePass) InitParams(numEntries uint64, uint64PerEntry uint64, batchtype string) {
	numEntries = utils.NextPerfectSquare(numEntries)
	T := uint32(math.Sqrt(float64(numEntries)))

	M := uint32((numEntries) / uint64(T))

	if T*M != uint32(numEntries) {
		fmt.Println("make sure numEntries is a perfect square")
	}

	// Q is unlimited, max is numEntries, suitable is T
	Q := uint64(T)

	p.Params = &SinglePassParams{
		T:              T,
		M:              M,
		Q:              Q,
		NumEntries:     numEntries,
		Uint64PerEntry: uint64PerEntry,
	}
	p.Hint = nil
	p.batchtype = batchtype
}
func (p *SinglePass) GenerateHint(db *utils.EncodedDB) Hint {
	ck := uint64(time.Now().UnixNano())
	p_vals := p.InitP(ck)

	uint64PerEntry := p.Params.Uint64PerEntry
	M := uint64(p.Params.M)
	T := uint64(p.Params.T)
	h_vals := make([]uint64, M*uint64PerEntry)

	for j := uint64(0); j < T; j++ {
		rowOffset := j * M

		for i := uint64(0); i < M; i++ {
			dbIndex := rowOffset + uint64(p_vals[rowOffset+i])

			dbEntry := db.GetEntry(dbIndex)

			hEntry := h_vals[i*uint64PerEntry : (i+1)*uint64PerEntry]

			db.XORInplace(hEntry, dbEntry)
		}
	}

	p.Hint = &SinglePassHint{
		ck: ck,
		p:  p_vals,
		h:  h_vals,
	}

	return p.Hint
}

func (p *SinglePass) InitP(seed uint64) []uint32 {
	T := p.Params.T
	M := p.Params.M

	permMatrix := make([]uint32, T*M)

	for i := uint32(0); i < T; i++ {
		rowOffset := i * M

		row := permMatrix[rowOffset : rowOffset+M]

		for j := uint32(0); j < M; j++ {
			row[j] = j
		}

		r := rand.New(rand.NewSource(int64(seed + uint64(i))))

		for j := M - 1; j > 0; j-- {
			k := uint32(r.Intn(int(j + 1)))
			row[j], row[k] = row[k], row[j]
		}
	}

	return permMatrix
}

func (p *SinglePass) LoadHint(id string, filepath string, db *utils.EncodedDB) (Hint, int64) {
	ckPath := filepath + id + "ck.gob"
	hPath := filepath + id + "h.gob"

	var ck uint64
	var h_vals []uint64
	var p_vals []uint32

	if utils.FileExists(ckPath) && utils.FileExists(hPath) {
		fmt.Printf("[PIR] Loading existing hints for ID: %s\n", id)

		utils.LoadFromFile(&ck, ckPath)
		p_vals = p.InitP(ck)

		utils.LoadFromFile(&h_vals, hPath)

		p.Hint = &SinglePassHint{
			ck: ck,
			p:  p_vals,
			h:  h_vals,
		}
	} else {
		fmt.Printf("Hints [%s] not found. Generating now...\n", id)

		p.GenerateHint(db)

		utils.SaveToFile(p.Hint.ck, ckPath)
		utils.SaveToFile(p.Hint.h, hPath)
	}

	ckStat, _ := os.Stat(ckPath)
	hStat, _ := os.Stat(hPath)

	//pMatrixSize := int64(T*M) * 4

	size := ckStat.Size() + hStat.Size()
	return p.Hint, size
}

func FindIndex(P []uint32, d uint32) (uint32, bool) {
	for i, val := range P {
		if val == d {
			return uint32(i), true
		}
	}
	return 0, false
}

func (p *SinglePass) singleQuery(x uint64) ([]uint32, []uint32, []uint32, uint32, uint32) {
	T := p.Params.T
	M := p.Params.M
	perm := p.Hint.p

	xi := uint32(x / uint64(M))
	xj := uint32(x % uint64(M))

	ind, _ := FindIndex(perm[xi*M:(xi+1)*M], xj)

	query := make([]uint32, T)
	refresh := make([]uint32, T)
	rv := make([]uint32, T)

	r := rand.Uint32() % M

	for i := uint32(0); i < T; i++ {
		rowOffset := i * M
		row := perm[rowOffset : rowOffset+M]

		if i == xi {
			query[i] = r
		} else {
			query[i] = row[ind]
		}

		randCol := rand.Uint32() % M
		rv[i] = randCol
		refresh[i] = row[randCol]
	}

	return query, refresh, rv, xi, ind
}
func (p *SinglePass) RefreshH(ansFromQuery []uint64, ansFromRefresh []uint64, rv []uint32, xi uint32, ind uint32) {
	T := p.Params.T
	W := uint64(p.Params.Uint64PerEntry)
	h := p.Hint.h

	targetBase := uint64(ind) * W
	ht := h[targetBase : targetBase+W]
	_ = ht[W-1]

	for i := uint32(0); i < T; i++ {
		if i == xi {
			continue
		}

		rowOff := uint64(i) * W
		randBase := uint64(rv[i]) * W

		q := ansFromQuery[rowOff : rowOff+W]
		r := ansFromRefresh[rowOff : rowOff+W]
		hr := h[randBase : randBase+W]

		for k := uint64(0); k < W; k++ {
			delta := q[k] ^ r[k]
			ht[k] ^= delta
			hr[k] ^= delta
		}
	}
}

func (p *SinglePass) RefreshP(rv []uint32, xi uint32, ind uint32) {
	T := p.Params.T
	M := p.Params.M
	perm := p.Hint.p
	pos1 := ind

	for i := uint32(0); i < T; i++ {
		if i == xi {
			continue
		}

		rowOff := i * M
		pos2 := rv[i]
		row := perm[rowOff : rowOff+M]
		row[pos1], row[pos2] = row[pos2], row[pos1]
	}
}

/*
func (p *SinglePass) SaveHintState() SinglePass {
	return *p
}
*/

func (p *SinglePass) QueryAndFakeRefresh(indexes []uint64) (Message, State) {
	batchSize := len(indexes)

	query := make([][]uint64, 0, batchSize)
	refresh := make([][]uint64, 0, batchSize)
	state := &SinglePassState{
		rv:  make([][]uint32, batchSize),
		xi:  make([]uint32, batchSize),
		ind: make([]uint32, batchSize),
	}

	for i, x := range indexes {
		qy, re, rv, xi, ind := p.singleQuery(x)
		query = append(query, uint32SliceToUint64(qy))
		refresh = append(refresh, uint32SliceToUint64(re))

		state.rv[i] = rv
		state.xi[i] = xi
		state.ind[i] = ind

		p.RefreshP(rv, xi, ind)
	}

	msg := &SinglePassMessage{
		Sender:         "client",
		Payload:        query,
		UpdatedSender:  "client",
		UpdatedPayload: refresh,
	}

	return msg, state
}

func uint32SliceToUint64(src []uint32) []uint64 {
	dst := make([]uint64, len(src))
	for i, v := range src {
		dst[i] = uint64(v)
	}
	return dst
}

func (p *SinglePass) Answer(db *utils.EncodedDB, req Message) Message {
	msg, ok := req.(*SinglePassMessage)
	if !ok {
		fmt.Println("[Error] Incoming request is not *SinglePassMessage")
		return nil
	}

	queryResp := p.processVectors(db, msg.Payload)
	refreshResp := p.processVectors(db, msg.UpdatedPayload)

	return &SinglePassMessage{
		Sender:         "server",
		Payload:        queryResp,
		UpdatedSender:  "updatedserver",
		UpdatedPayload: refreshResp,
	}
}

func (p *SinglePass) processVectors(db *utils.EncodedDB, vectors [][]uint64) [][]uint64 {
	if len(vectors) == 0 {
		return nil
	}

	T := uint64(p.Params.T)
	W := uint64(p.Params.Uint64PerEntry)

	results := make([][]uint64, len(vectors))

	for vIdx, vector := range vectors {
		ans := make([]uint64, T*W)

		for i := uint64(0); i < T; i++ {
			offset := (i*uint64(p.Params.M) + vector[i]) * W
			copy(ans[i*W:(i+1)*W], db.Data[offset:offset+W])
		}
		results[vIdx] = ans
	}

	return results
}

func (p *SinglePass) Rewind(state State) {

}

func (p *SinglePass) ReconstructAndRefresh(resp Message, st State, indexes []uint64) [][]uint64 {
	serverResp, ok := resp.(*SinglePassMessage)
	state, okSt := st.(*SinglePassState)
	if !ok || !okSt {
		return nil
	}

	T := uint32(p.Params.T)
	W := uint64(p.Params.Uint64PerEntry)
	h := p.Hint.h
	batchSize := len(indexes)
	results := make([][]uint64, batchSize)

	if len(serverResp.Payload) < batchSize || len(serverResp.UpdatedPayload) < batchSize {
		return nil
	}

	for b := 0; b < batchSize; b++ {
		ansQ := serverResp.Payload[b]
		ansR := serverResp.UpdatedPayload[b]
		xi := state.xi[b]
		ind := state.ind[b]
		rv := state.rv[b]

		target := make([]uint64, W)
		hBase := uint64(ind) * W
		ht := h[hBase : hBase+W]
		copy(target, ht)

		for i := uint32(0); i < T; i++ {
			if i == xi {
				continue
			}

			rowOff := uint64(i) * W
			qPart := ansQ[rowOff : rowOff+W]
			rPart := ansR[rowOff : rowOff+W]
			randBase := uint64(rv[i]) * W
			hr := h[randBase : randBase+W]

			_ = qPart[W-1]
			_ = rPart[W-1]
			_ = hr[W-1]
			_ = ht[W-1]

			for k := uint64(0); k < W; k++ {
				qk := qPart[k]
				rk := rPart[k]
				delta := qk ^ rk

				target[k] ^= qk

				// Hint：H[ind] ^= delta, H[rv] ^= delta
				ht[k] ^= delta
				hr[k] ^= delta
			}
		}
		results[b] = target
	}

	return results
}

func (p *SinglePass) GetHintSize() float64 {
	return p.Hint.Size()
}

func (p *SinglePass) GetParamQ() uint64 {
	return p.Params.Q
}
