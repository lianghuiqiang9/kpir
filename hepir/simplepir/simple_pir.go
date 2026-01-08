package simplepir

// #cgo CFLAGS: -O3 -march=native
// #include "pir.h"
import "C"
import (
	"crypto/rand"
	"fmt"
	"math"

	utils "github.com/local/utils"
)

type SimplePIR struct {
	Params Params
	seed   [16]byte // generate the state A
	A      State
	DBInfo DBinfo
}

func (pi *SimplePIR) Name() string {
	return "SimplePIR"
}

func (pi *SimplePIR) InitParams(numEntries, uint64PerEntry uint64) Params {
	bitsPerEntry := uint64PerEntry * 64
	n := uint64(1 << 10)
	logq := uint64(32)

	good_p := Params{}
	found := false

	// Iteratively refine p and DB dims, until find tight values
	for mod_p := uint64(2); ; mod_p += 1 {

		l, m := ApproxSquareDatabaseDims(numEntries, bitsPerEntry, mod_p)
		p := Params{
			N:    n,
			Logq: logq,
			L:    l,
			M:    m,
		}
		p.PickParams(false, m)
		if p.P < mod_p {
			if !found {
				panic("Error; should not happen")
			}
			good_p.Print()
			pi.Params = good_p
			pi.DBInfo.NumEntries = numEntries
			pi.DBInfo.BitsPerEntry = bitsPerEntry
			return good_p
		}

		good_p = p
		found = true
	}
}

func (pi *SimplePIR) MakeRandomInteranlDB() *InternalDB {
	p := pi.Params
	D := pi.SetupDB()
	//fmt.Println("p.L : ", p.L, ", p.M : ", p.M, ", p.P : ", p.P)
	var seed [16]byte
	if _, err := rand.Read(seed[:]); err != nil {
		panic(err)
	}
	D.Data = MatrixRand(p.L, p.M, p.Logp, 0, seed)
	D.Data.Sub(p.P / 2)
	return D
}

func (pi *SimplePIR) MakeInternalDB(db *utils.EncodedDB) *InternalDB {
	D := pi.SetupDB()
	p := pi.Params
	D.Data = MatrixZeros(p.L, p.M)

	for i := uint64(0); i < pi.DBInfo.NumEntries; i++ {
		temp := db.GetEntry(i)
		temp2 := UintPToUintQ(temp, 64, p.Logp)
		for j := uint64(0); j < D.Info.Ne; j++ {
			D.Data.Set(temp2[j], (uint64(i)/p.M)*D.Info.Ne+j, uint64(i)%p.M)
		}
	}

	D.Data.Sub(p.P / 2)

	return D
}

func (pi *SimplePIR) SetupDB() *InternalDB {
	p := pi.Params
	numEntries := pi.DBInfo.NumEntries
	bitsPerEntry := pi.DBInfo.BitsPerEntry
	if (numEntries == 0) || (bitsPerEntry == 0) {
		panic("Empty database!")
	}

	D := new(InternalDB)

	D.Info.NumEntries = numEntries
	D.Info.BitsPerEntry = bitsPerEntry
	D.Info.P = p.P
	D.Info.Logq = p.Logq

	db_elems, elems_per_entry, entries_per_elem := Num_DB_entries(numEntries, bitsPerEntry, p.P)
	//fmt.Println("db_elems : ", db_elems, " elems_per_entry : ", elems_per_entry, " entries_per_elem : ", entries_per_elem)
	D.Info.Ne = elems_per_entry
	D.Info.X = D.Info.Ne
	D.Info.Packing = entries_per_elem

	for D.Info.Ne%D.Info.X != 0 {
		D.Info.X += 1
	}

	D.Info.Basis = 0
	D.Info.Squishing = 0

	fmt.Printf("Total packed DB size is ~%f MB\n",
		float64(p.L*p.M)*math.Log2(float64(p.P))/(1024.0*1024.0*8.0))
	//fmt.Printf("Real packed DB size is   %d MB\n",
	//	uint64(Num*row_length/(1024.0*1024.0*8.0)))

	if db_elems > p.L*p.M {
		panic("Params and database size don't match")
	}

	if p.L%D.Info.Ne != 0 {
		panic("Number of DB elems per entry must divide DB height")
	}
	pi.DBInfo = D.Info
	return D
}

func (pi *SimplePIR) PickParamsGivenDimensions(l, m, n, logq uint64) Params {
	p := Params{
		N:    n,
		Logq: logq,
		L:    l,
		M:    m,
	}
	p.PickParams(false, m)
	return p
}

func (pi *SimplePIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.L*p.N*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.M*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.L*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
}

func (pi *SimplePIR) Init() State {
	var seed [16]byte
	if _, err := rand.Read(seed[:]); err != nil {
		panic(err)
	}
	pi.seed = seed
	A := MatrixRand(pi.Params.M, pi.Params.N, pi.Params.Logq, 0, pi.seed)
	pi.A = MakeState(A)
	return pi.A
}

func (pi *SimplePIR) Setup(DB *InternalDB) (State, State) {
	pi.Init()
	A := pi.A.Data[0]
	H := MatrixMul(DB.Data, A)
	DB.Data.Add(pi.Params.P / 2)
	DB.Squish()
	pi.DBInfo = DB.Info
	return MakeState(), MakeState(H)
}

func (pi *SimplePIR) FakeSetup(DB *InternalDB) (State, float64) {
	offline_download := float64(pi.Params.L*pi.Params.N*uint64(pi.Params.Logq)) / (8.0 * 1024.0)
	//fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))
	DB.Data.Add(pi.Params.P / 2)
	DB.Squish()
	pi.DBInfo = DB.Info

	return MakeState(), offline_download
}

func (pi *SimplePIR) Query(indexes []uint64) (Msg, State) {
	//batchSize := len(indexes)
	msgs := Msg{}
	states := State{}

	info := pi.DBInfo
	A := pi.A.Data[0]
	p := pi.Params
	delta := p.Delta()

	for _, x := range indexes {

		var seed128 [16]byte
		if _, err := rand.Read(seed128[:]); err != nil {
			panic("Query seed generation failed: " + err.Error())
		}

		secret := MatrixRand(p.N, 1, p.Logq, 0, seed128)
		errorTerm := MatrixGaussian(p.M, 1)

		query := MatrixMul(A, secret)
		query.MatrixAdd(errorTerm)

		query.Data[x%p.M] += C.Elem(delta)

		if p.M%info.Squishing != 0 {
			padding := info.Squishing - (p.M % info.Squishing)
			query.AppendZeros(padding)
		}

		//states[i] = MakeState(secret)
		//msgs[i] = MakeMsg(query)
		msgs.Data = append(msgs.Data, query)
		states.Data = append(states.Data, secret)
	}

	return msgs, states
}

func (pi *SimplePIR) QueryOffline(batchSize uint64) (Msg, State) {
	//states := make([]State, batchSize)
	//msgs := make([]Msg, batchSize)
	msgs := Msg{}
	states := State{}

	A := pi.A.Data[0]
	p := pi.Params

	for i := uint64(0); i < batchSize; i++ {
		var seed128 [16]byte
		if _, err := rand.Read(seed128[:]); err != nil {
			panic(err)
		}

		secret := MatrixRand(p.N, 1, p.Logq, 0, seed128)
		errorTerm := MatrixGaussian(p.M, 1)

		query := MatrixMul(A, secret)
		query.MatrixAdd(errorTerm)

		//states[i] = MakeState(secret)
		//msgs[i] = MakeMsg(query)
		msgs.Data = append(msgs.Data, query)
		states.Data = append(states.Data, secret)
	}

	return msgs, states
}

func (pi *SimplePIR) QueryOnline(indexes []uint64, offlineMsgs Msg) Msg {
	//batchSize := len(indexes)
	//if batchSize != len(offlineMsgs) {
	//	panic("Batch size mismatch between indexes and offline messages")
	//}

	onlineMsgs := Msg{}
	//make([]Msg, batchSize)
	info := pi.DBInfo
	p := pi.Params
	delta := p.Delta()

	for i, idx := range indexes {
		query := offlineMsgs.Data[i]

		query.Data[idx%p.M] += C.Elem(delta)

		if p.M%info.Squishing != 0 {
			padding := info.Squishing - (p.M % info.Squishing)
			query.AppendZeros(padding)
		}

		onlineMsgs.Data = append(onlineMsgs.Data, query)
	}

	return onlineMsgs
}

func (pi *SimplePIR) Answer(DB *InternalDB, queries Msg, serverHint State) Msg {
	//batchSize := len(queries)
	answers := Msg{}
	//make([]Msg, batchSize)

	basis := DB.Info.Basis
	squishing := DB.Info.Squishing

	for _, query := range queries.Data {
		//queryVec := queries.Data[i]
		ans := MatrixMulVecPacked(
			DB.Data,
			query,
			basis,
			squishing,
		)

		answers.Data = append(answers.Data, ans)
	}

	return answers
}

func (pi *SimplePIR) Reconstruct(indexes []uint64, clientHint State, query Msg, answer Msg, clientState State) [][]uint64 {
	batchSize := len(indexes)
	results := make([][]uint64, batchSize)

	info := pi.DBInfo
	p := pi.Params
	ratio := p.P / 2

	for i := 0; i < batchSize; i++ {
		secret := clientState.Data[i]
		H := clientHint.Data[0]
		ans := answer.Data[i]

		offset := uint64(0)
		for j := uint64(0); j < p.M; j++ {
			offset += ratio * query.Data[i].Get(j, 0)
		}
		offset %= (1 << p.Logq)
		offset = (1 << p.Logq) - offset

		row := indexes[i] / p.M
		interm := MatrixMul(H, secret) // interm = hint_c * s
		ans.MatrixSub(interm)          // ans - interm

		var vals []uint64
		// Recover each Z_p element that makes up the desired database entry
		for j := row * info.Ne; j < (row+1)*info.Ne; j++ {
			noised := uint64(ans.Data[j]) + offset
			denoised := p.Round(noised)
			vals = append(vals, denoised)
		}

		ans.MatrixAdd(interm)
		results[i] = ReconstructElem(vals, indexes[i], info)

		results[i] = UintPToUintQTrunc(results[i], p.Logp, 64, (info.BitsPerEntry+63)/64)
	}

	return results
}

func (pi *SimplePIR) Reset(DB *InternalDB) {
	DB.Unsquish()
	DB.Data.Sub(pi.Params.P / 2)
}

func (pi *SimplePIR) LoadHint(id string, filepath string, internalDB *InternalDB) (*State, *State) {
	serverHintFilename := filepath + id + "server_state.gob"
	clientHintFilename := filepath + id + "client_state.gob"
	seedFilename := filepath + id + "seed.gob"

	if utils.FileExists(serverHintFilename) && utils.FileExists(clientHintFilename) && utils.FileExists(seedFilename) {
		fmt.Printf("Loading precomputed hints for ID: %s...\n", id)

		utils.LoadFromFile(&pi.seed, seedFilename)
		A := MatrixRand(pi.Params.M, pi.Params.N, pi.Params.Logq, 0, pi.seed)
		pi.A = MakeState(A)

		serverHint := &State{}
		clientHint := &State{}
		utils.LoadFromFile(serverHint, serverHintFilename)
		utils.LoadFromFile(clientHint, clientHintFilename)

		pi.FakeSetup(internalDB)

		return serverHint, clientHint
	}

	fmt.Println("Hints not found. Running full Setup (this may take a while)...")
	sHint, cHint := pi.Setup(internalDB)

	utils.SaveToFile(pi.seed, seedFilename)
	utils.SaveToFile(sHint, serverHintFilename)
	utils.SaveToFile(cHint, clientHintFilename)

	return &sHint, &cHint
}
