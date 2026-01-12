package sipir

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"time"

	utils "github.com/local/utils"
)

type PrfKey [16]byte

func NewPrfKey(rng *rand.Rand) PrfKey {
	var key [16]byte
	binary.LittleEndian.PutUint64(key[0:8], rng.Uint64())
	binary.LittleEndian.PutUint64(key[8:16], rng.Uint64())
	return key
}

func expandKeyAsm(key *byte, enc *uint32)
func aes128MMO(xk *uint32, dst, src *byte)

func (k *PrfKey) Eval(x uint64) uint64 {
	var longKey [44]uint32
	var src [16]byte
	var dsc [16]byte

	expandKeyAsm(&k[0], &longKey[0])

	binary.LittleEndian.PutUint64(src[:], x)

	aes128MMO(&longKey[0], &dsc[0], &src[0])

	return binary.LittleEndian.Uint64(dsc[:])
}

type LocalHint struct {
	Key             PrfKey   // [16]byte
	Parity          []uint64 // 64 bits
	ProgrammedPoint uint64   // 64 bits
	IsProgrammed    bool     // 1 bits
}

func (lh *LocalHint) calculateSize() float64 {
	size := 16.0                        // PrfKey (128 bits)
	size += float64(len(lh.Parity) * 8) // Parity (uint64 array)
	size += 8.0                         // ProgrammedPoint (uint64)
	size += 0.125                       // IsProgrammed (1 bit)
	return size
}

// Copy 实现了 LocalHint 的深拷贝
func (lh *LocalHint) Copy() LocalHint {
	newLH := *lh

	if lh.Parity != nil {
		newLH.Parity = make([]uint64, len(lh.Parity))
		copy(newLH.Parity, lh.Parity)
	}
	return newLH
}

func init() {
	gob.Register(LocalHint{})
	gob.Register(PrfKey{})
}

type PianoParams struct {
	ChunkSize      uint64
	ChunkNum       uint64
	Q              uint64
	M1             uint64
	M2             uint64
	NumEntries     uint64
	Uint64PerEntry uint64
}

func (p *PianoParams) Print() string {
	return fmt.Sprintf(
		"Piano Protocol Parameters:\n"+
			"- ChunkSize: %d\n"+
			"- ChunkNum:  %d\n"+
			"- Q:		  %d\n"+
			"- M1 (Primary): %d\n"+
			"- M2 (Backup):  %d\n"+
			"- NumEntries:   %d\n"+
			"- Uint64PerEntry:  %d",
		p.ChunkSize, p.Q, p.ChunkNum, p.M1, p.M2, p.NumEntries, p.Uint64PerEntry,
	)
}

type PianoHint struct {
	PrimaryHints       []LocalHint
	ReplacementIndices []uint64 // 32bits
	ReplacementValues  []uint64 // 64bits
	BackupHints        []LocalHint
}

func (h *PianoHint) Size() float64 {
	var total float64

	for _, lh := range h.PrimaryHints {
		total += lh.calculateSize()
	}

	for _, lh := range h.BackupHints {
		total += lh.calculateSize()
	}

	total += float64(len(h.ReplacementIndices) * 4)

	total += float64(len(h.ReplacementValues) * 8)

	return total
}

type PianoMessage struct {
	Sender  string
	Payload [][]uint64 // client 32bits, server 64bits
}

func (m *PianoMessage) Size() float64 {
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

type PianoState struct {
	chunkId         []uint64
	hitId           []uint64
	replacementVal  [][]uint64
	oldPrimaryHints []LocalHint
}

func (s *PianoState) Print() string {
	return fmt.Sprintf(
		"Piano Batch State:\n"+
			"- Batch Size:      %d\n"+
			"- ChunkIds:        %v\n"+
			"- HitIds (per q):  %v\n"+
			"- ReplacementVals: %v",
		len(s.chunkId), s.chunkId, s.hitId, s.replacementVal,
	)
}

type Piano struct {
	Params             *PianoParams
	Hint               *PianoHint
	consumedReplaceNum []uint64 // 32bits is enough
	consumedHintNum    []uint64 // 32bits is enough
	Batchtype          string
}

func (p *Piano) Copy() *Piano {
	if p == nil {
		return nil
	}

	newP := &Piano{
		Params:    p.Params,
		Batchtype: p.Batchtype,
	}

	if p.Hint != nil {
		newP.Hint = &PianoHint{
			ReplacementIndices: make([]uint64, len(p.Hint.ReplacementIndices)),
			ReplacementValues:  make([]uint64, len(p.Hint.ReplacementValues)),
		}
		copy(newP.Hint.ReplacementIndices, p.Hint.ReplacementIndices)
		copy(newP.Hint.ReplacementValues, p.Hint.ReplacementValues)

		newP.Hint.PrimaryHints = make([]LocalHint, len(p.Hint.PrimaryHints))
		for i := range p.Hint.PrimaryHints {
			newP.Hint.PrimaryHints[i] = p.Hint.PrimaryHints[i].Copy()
		}

		newP.Hint.BackupHints = make([]LocalHint, len(p.Hint.BackupHints))
		for i := range p.Hint.BackupHints {
			newP.Hint.BackupHints[i] = p.Hint.BackupHints[i].Copy()
		}
	}

	if p.consumedReplaceNum != nil {
		newP.consumedReplaceNum = make([]uint64, len(p.consumedReplaceNum))
		copy(newP.consumedReplaceNum, p.consumedReplaceNum)
	}

	if p.consumedHintNum != nil {
		newP.consumedHintNum = make([]uint64, len(p.consumedHintNum))
		copy(newP.consumedHintNum, p.consumedHintNum)
	}

	return newP
}

func (p *Piano) Name() string {
	fmt.Println("Piano-PIR")
	return "Piano-PIR"
}

func (p *Piano) InitParams(numEntries uint64, bitsPerVal uint64, batchtype string) {
	if bitsPerVal%32 != 0 {
		fmt.Println("bitsPerVal should be 32 * k")
		os.Exit(1)
	}

	uint64PerEntry := (bitsPerVal + 63) / 64
	numEntries = utils.NextPerfectSquare(numEntries)

	sqrtNumEntries := math.Sqrt(float64(numEntries))
	logeNumEntries := math.Log(float64(numEntries))
	q := uint64(sqrtNumEntries * logeNumEntries) //\sqrt{DBsize} * log_e^DBsize=4*2.9,
	m1 := 4*q + 2                                //4 * \sqrt{DBsize} * log_e^DBsize=4*4*2.9
	m2 := uint64(4 * logeNumEntries)

	chunkSize := uint64(sqrtNumEntries)
	chunkNum := numEntries / chunkSize
	p.Params = &PianoParams{
		ChunkSize:      chunkSize,
		ChunkNum:       chunkNum,
		Q:              q,
		M1:             m1,
		M2:             m2,
		NumEntries:     numEntries,
		Uint64PerEntry: uint64PerEntry,
	}
	p.Hint = nil
	p.consumedReplaceNum = make([]uint64, chunkNum)
	p.consumedHintNum = make([]uint64, chunkNum)
	p.Batchtype = batchtype
}

func (p *Piano) elem(hint *LocalHint, chunkId uint64) uint64 {
	chunkSize := p.Params.ChunkSize

	if hint.IsProgrammed && chunkId == hint.ProgrammedPoint/chunkSize {
		return hint.ProgrammedPoint
	}

	offset := hint.Key.Eval(chunkId) % chunkSize
	return offset + chunkId*chunkSize
}

func (p *Piano) GenerateHint(db *utils.EncodedDB) Hint {
	m1 := p.Params.M1
	m2 := p.Params.M2
	chunkSize := p.Params.ChunkSize
	chunkNum := p.Params.ChunkNum
	uint64PerEntry := p.Params.Uint64PerEntry

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	primaryHints := make([]LocalHint, m1)
	backupHints := make([]LocalHint, m2*chunkNum)
	replacementIndices := make([]uint64, m2*chunkNum)
	replacementValues := make([]uint64, m2*chunkNum*uint64(uint64PerEntry))

	for i := uint64(0); i < m1; i++ {
		primaryHints[i] = LocalHint{
			Key:    NewPrfKey(rng),
			Parity: make([]uint64, uint64PerEntry),
		}
	}

	for i := uint64(0); i < m2*chunkNum; i++ {
		backupHints[i] = LocalHint{
			Key:    NewPrfKey(rng),
			Parity: make([]uint64, uint64PerEntry),
		}
	}

	for i := uint64(0); i < chunkNum; i++ {

		// A. Primary Hints
		for j := uint64(0); j < m1; j++ {
			index := p.elem(&primaryHints[j], i)

			entry := db.GetEntry(index)

			db.XORInplace(primaryHints[j].Parity, entry)

		}

		// B. Backup Hints
		for j := uint64(0); j < m2*chunkNum; j++ {
			if j/m2 != i {
				index := p.elem(&backupHints[j], i)
				entry := db.GetEntry(index)
				db.XORInplace(backupHints[j].Parity, entry)

			}
		}

		// C.  Replacement Values
		for j := i * m2; j < (i+1)*m2; j++ {
			ind := (rng.Uint64() % chunkSize) + (i * chunkSize)

			if ind >= db.NumEntries {
				ind = db.NumEntries - 1
			}

			replacementIndices[j] = ind
			entry := db.GetEntry(ind)

			offset := j * uint64(uint64PerEntry)
			copy(replacementValues[offset:offset+uint64(uint64PerEntry)], entry)
		}
	}

	p.Hint = &PianoHint{
		PrimaryHints:       primaryHints,
		ReplacementIndices: replacementIndices,
		ReplacementValues:  replacementValues,
		BackupHints:        backupHints,
	}
	return p.Hint
}

func (p *Piano) LoadHint(id string, filepath string, db *utils.EncodedDB) (Hint, int64) {

	basePath := filepath + id + "hints-"
	filePrimary := basePath + "primaryHints.gob"
	fileIndices := basePath + "replacementIndices.gob"
	fileValues := basePath + "replacementValues.gob"
	fileBackup := basePath + "backupHints.gob"

	allExist := utils.FileExists(filePrimary) &&
		utils.FileExists(fileIndices) &&
		utils.FileExists(fileValues) &&
		utils.FileExists(fileBackup)

	if allExist {
		fmt.Printf("Loading hints for ID [%s] from disk...\n", id)

		h1 := []LocalHint{}
		ri := []uint64{}
		rv := []uint64{}
		h2 := []LocalHint{}

		utils.LoadFromFile(&h1, filePrimary)
		utils.LoadFromFile(&ri, fileIndices)
		utils.LoadFromFile(&rv, fileValues)
		utils.LoadFromFile(&h2, fileBackup)

		p.Hint = &PianoHint{
			PrimaryHints:       h1,
			ReplacementIndices: ri,
			ReplacementValues:  rv,
			BackupHints:        h2,
		}
	} else {
		fmt.Printf("Hints for ID [%s] not found. Generating new ones...\n", id)

		p.GenerateHint(db)

		utils.SaveToFile(p.Hint.PrimaryHints, filePrimary)
		utils.SaveToFile(p.Hint.ReplacementIndices, fileIndices)
		utils.SaveToFile(p.Hint.ReplacementValues, fileValues)
		utils.SaveToFile(p.Hint.BackupHints, fileBackup)
	}

	var totalSize int64
	files := []string{filePrimary, fileIndices, fileValues, fileBackup}
	for _, f := range files {
		if s, err := os.Stat(f); err == nil {
			totalSize += s.Size()
		}
	}

	return p.Hint, totalSize
}

func (p *Piano) contains(arr []uint64, target uint64) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func (p *Piano) findHitId(x uint64, hitIdState *[]uint64) (uint64, uint64) {
	chunkId := x / p.Params.ChunkSize

	hitId := uint64(math.MaxUint64)

	for i := uint64(0); i < p.Params.M1; i++ {
		if p.elem(&p.Hint.PrimaryHints[i], chunkId) == x {
			if p.Batchtype != "skip" || !p.contains(*hitIdState, i) {
				hitId = i
				break
			}
		}
	}

	if hitId == uint64(math.MaxUint64) {
		log.Fatalf("Error: cannot find an available uncollided hitId for index %d", x)
	}

	*hitIdState = append(*hitIdState, hitId)
	return chunkId, hitId
}

func (p *Piano) generateExpandSet(chunkId uint64, hitId uint64) ([]uint64, []uint64) {
	expandedSet := make([]uint64, p.Params.ChunkNum)
	for i := uint64(0); i < p.Params.ChunkNum; i++ {
		expandedSet[i] = p.elem(&p.Hint.PrimaryHints[hitId], i)
	}

	replacementVal := make([]uint64, p.Params.Uint64PerEntry)

	if p.consumedReplaceNum[chunkId] < p.Params.M2 {

		tmp := p.consumedReplaceNum[chunkId] + chunkId*p.Params.M2
		replacementInd := p.Hint.ReplacementIndices[tmp]

		offset := tmp * p.Params.Uint64PerEntry

		copy(replacementVal, p.Hint.ReplacementValues[offset:offset+p.Params.Uint64PerEntry])

		p.consumedReplaceNum[chunkId]++
		expandedSet[chunkId] = replacementInd
	} else {
		log.Fatalf("Critical: Not enough replacement values for chunk %d", chunkId)
	}

	return replacementVal, expandedSet
}

func (p *Piano) Query(indexes []uint64) (Message, State) {
	batchSize := len(indexes)

	state := &PianoState{
		chunkId:        make([]uint64, batchSize),
		hitId:          []uint64{},
		replacementVal: make([][]uint64, batchSize),
	}

	queryPayload := make([][]uint64, batchSize)

	for i, x := range indexes {
		cId, hId := p.findHitId(x, &state.hitId)

		rVal, expandedSet := p.generateExpandSet(cId, hId)

		state.chunkId[i] = cId
		state.replacementVal[i] = rVal
		queryPayload[i] = expandedSet
	}

	msg := &PianoMessage{
		Sender:  "client",
		Payload: queryPayload,
	}

	return msg, state
}

func (p *Piano) Answer(db *utils.EncodedDB, req Message) Message {

	msg := req.(*PianoMessage)

	answerPayload := make([][]uint64, len(msg.Payload))

	for i, expandedSet := range msg.Payload {

		parity := make([]uint64, db.Uint64PerEntry)

		for _, index := range expandedSet {
			temp := db.GetEntry(index)
			db.XORInplace(parity, temp)

		}
		answerPayload[i] = parity
	}

	return &PianoMessage{
		Sender:  "server",
		Payload: answerPayload,
	}
}

func (p *Piano) Reconstruct(resp Message, state State) [][]uint64 {
	st := state.(*PianoState)
	msg := resp.(*PianoMessage)

	batchResults := make([][]uint64, len(msg.Payload))

	for i, serverParity := range msg.Payload {
		hId := st.hitId[i]
		rVal := st.replacementVal[i]

		originalHintParity := p.Hint.PrimaryHints[hId].Parity

		uint64PerEntry := len(serverParity)
		reconstructedEntry := make([]uint64, uint64PerEntry)

		// 3. Result = ServerParity ⊕ OriginalHintParity ⊕ ReplacementValue
		for k := 0; k < uint64PerEntry; k++ {
			reconstructedEntry[k] = serverParity[k] ^ originalHintParity[k] ^ rVal[k]
		}

		batchResults[i] = reconstructedEntry
	}

	return batchResults
}

func (p *Piano) Refresh(indexes []uint64, answers [][]uint64, state State) {
	st := state.(*PianoState)
	Params := p.Params
	m2 := Params.M2

	for i, x := range indexes {
		chunkId := st.chunkId[i]
		hitId := st.hitId[i]
		answer := answers[i]

		if p.consumedHintNum[chunkId] < m2 {
			backupIdx := chunkId*m2 + p.consumedHintNum[chunkId]

			p.Hint.PrimaryHints[hitId] = p.Hint.BackupHints[backupIdx]

			p.Hint.PrimaryHints[hitId].IsProgrammed = true
			p.Hint.PrimaryHints[hitId].ProgrammedPoint = x

			for k := 0; k < len(answer); k++ {
				p.Hint.PrimaryHints[hitId].Parity[k] ^= answer[k]
			}

			p.consumedHintNum[chunkId]++
		} else {
			log.Fatalf("Critical: Not enough backup hints for chunk %d. Preprocessing required.", chunkId)
		}
	}
}

func (p *Piano) QueryAndFakeRefresh(indexes []uint64) (Message, State) {
	batchSize := len(indexes)
	m2 := p.Params.M2

	state := &PianoState{
		chunkId:         make([]uint64, batchSize),
		hitId:           []uint64{},
		replacementVal:  make([][]uint64, batchSize),
		oldPrimaryHints: make([]LocalHint, 0, batchSize-1), // init
	}

	queryPayload := make([][]uint64, batchSize)

	for i, x := range indexes {
		// --- A. Query ---
		cId, hId := p.findHitId(x, &state.hitId)

		rVal, expandedSet := p.generateExpandSet(cId, hId)

		state.chunkId[i] = cId
		state.replacementVal[i] = rVal
		queryPayload[i] = expandedSet

		// --- B. FakeRefresh ---
		if i < batchSize-1 {

			if p.consumedHintNum[cId] < m2 {
				backupIdx := cId*m2 + p.consumedHintNum[cId]

				// save
				state.oldPrimaryHints = append(state.oldPrimaryHints, p.Hint.PrimaryHints[hId])

				p.Hint.PrimaryHints[hId] = p.Hint.BackupHints[backupIdx].Copy()
				p.Hint.PrimaryHints[hId].IsProgrammed = true
				p.Hint.PrimaryHints[hId].ProgrammedPoint = x

				p.consumedHintNum[cId]++
			} else {
				log.Fatalf("Critical: Not enough backup hints for chunk %d.", cId)
			}
		}

	}

	msg := &PianoMessage{
		Sender:  "client",
		Payload: queryPayload,
	}

	return msg, state
}

func (p *Piano) Rewind(state State) {
	st := state.(*PianoState)

	// batchsize - 1
	for i := len(st.chunkId) - 2; i >= 0; i-- {
		cId := st.chunkId[i]
		hId := st.hitId[i]

		p.Hint.PrimaryHints[hId] = st.oldPrimaryHints[i]

		p.consumedHintNum[cId]--

		p.consumedReplaceNum[cId]--
	}
}

func (p *Piano) ReconstructAndRefresh(resp Message, state State, indexes []uint64) [][]uint64 {

	st := state.(*PianoState)
	msg := resp.(*PianoMessage)
	m2 := p.Params.M2

	batchSize := len(msg.Payload)
	batchResults := make([][]uint64, batchSize)

	for i, serverParity := range msg.Payload {
		// --- A. Reconstruct ---
		hId := st.hitId[i]
		rVal := st.replacementVal[i]
		x := indexes[i]

		originalHint := &p.Hint.PrimaryHints[hId]
		uint64PerEntry := len(serverParity)
		reconstructedEntry := make([]uint64, uint64PerEntry)

		// Result = ServerParity ⊕ OriginalHintParity ⊕ ReplacementValue
		for k := 0; k < uint64PerEntry; k++ {
			reconstructedEntry[k] = serverParity[k] ^ originalHint.Parity[k] ^ rVal[k]
		}
		batchResults[i] = reconstructedEntry

		// --- B. Refresh ---
		chunkId := st.chunkId[i]
		if p.consumedHintNum[chunkId] < m2 {
			backupIdx := chunkId*m2 + p.consumedHintNum[chunkId]
			p.Hint.PrimaryHints[hId] = p.Hint.BackupHints[backupIdx].Copy()

			p.Hint.PrimaryHints[hId].IsProgrammed = true
			p.Hint.PrimaryHints[hId].ProgrammedPoint = x
			for k := 0; k < uint64PerEntry; k++ {
				p.Hint.PrimaryHints[hId].Parity[k] ^= reconstructedEntry[k]
			}

			p.consumedHintNum[chunkId]++
		} else {
			log.Fatalf("Critical: Not enough backup hints for chunk %d.", chunkId)
		}
	}

	return batchResults
}

/*
func (p *Piano) SaveHintState() Piano {
	return *p
}
*/
/*
	func (p *Piano) Rewind(b uint64, p2 *Piano) {
		*p = *p2
	}
*/

func (p *Piano) Verify(db *utils.EncodedDB, indexes []uint64, answers [][]uint64) bool {
	if len(indexes) != len(answers) {
		log.Printf("Verify Error: count mismatch, indexes(%d) vs answers(%d)", len(indexes), len(answers))
		return true // true
	}

	hasError := false

	for i, x := range indexes {
		answer := answers[i]
		truth := db.GetEntry(x)

		for k := uint64(0); k < db.Uint64PerEntry; k++ {
			if answer[k] != truth[k] {
				fmt.Printf("Verification Failed at index [%d], batch position [%d]\n", x, i)
				fmt.Printf("Expected: %v\nReceived: %v\n", truth, answer)
				hasError = true
				break
			}
		}

		if hasError {
			break
		}
	}

	if hasError {
		log.Fatalf("Error: Batch answer is not correct! PIR logic might be broken.")
	}

	return hasError
}

func (p *Piano) GetHintSize() float64 {
	return p.Hint.Size()
}

func (p *Piano) GetParamQ() uint64 {
	return p.Params.Q
}
