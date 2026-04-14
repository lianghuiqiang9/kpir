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

type DoublePIR struct {
	Params Params
	seed   [16]byte
	A      State
	DBInfo DBinfo
}

// Offline download: matrix H2
// Online query: matrices q1, q2
// Online download: matrices h1, a2, h2

// Server state: matrix H1
// Client state: matrices secret1, secret2
// Shared state: matrices A1, A2

// Ratio between first-level DB and second-level DB
const COMP_RATIO = uint64(64)

func (pi *DoublePIR) Name() string {
	return "DoublePIR"
}

func (pi *DoublePIR) InitParams(numEntries, bitsPerEntry uint64) Params {
	n := uint64(1 << 10)
	logq := uint64(32)

	good_p := Params{}
	found := false

	// Iteratively refine p and DB dims, until find tight values
	for mod_p := uint64(2); ; mod_p += 1 {
		l, m := ApproxDatabaseDims(numEntries, bitsPerEntry, mod_p, COMP_RATIO*n)

		p := Params{
			N:    n,
			Logq: logq,
			L:    l,
			M:    m,
		}
		p.PickParams(true, l, m)

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
func (pi *DoublePIR) MakeRandomInteranlDB() *InternalDB {
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

func (pi *DoublePIR) MakeInternalDB(db *utils.EncodedDB) *InternalDB {
	D := pi.SetupDB()
	p := pi.Params
	D.Data = MatrixZeros(p.L, p.M)

	for i := uint64(0); i < pi.DBInfo.NumEntries; i++ {
		temp := db.GetEntry(i)
		temp2 := UintPToUintQ(temp, uint64(64), p.Logp)

		for j := uint64(0); j < D.Info.Ne; j++ {
			D.Data.Set(temp2[j], (uint64(i)/p.M)*D.Info.Ne+j, uint64(i)%p.M)
		}
	}

	D.Data.Sub(p.P / 2)

	return D
}

func (pi *DoublePIR) SetupDB() *InternalDB {
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

func (pi *DoublePIR) PickParamsGivenDimensions(l, m, n, logq uint64) Params {
	p := Params{
		N:    n,
		Logq: logq,
		L:    l,
		M:    m,
	}
	p.PickParams(true, l, m)
	return p
}

func (pi *DoublePIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.delta()*info.X*p.N*p.N*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.M*p.Logq+info.Ne/info.X*p.L/info.X*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.delta()*info.X*p.N*p.Logq+p.delta()*p.N*info.Ne*p.Logq+p.delta()*info.Ne*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
}

func (pi *DoublePIR) Init() State {
	var seed [16]byte
	if _, err := rand.Read(seed[:]); err != nil {
		panic(err)
	}
	pi.seed = seed
	A1 := MatrixRand(pi.Params.M, pi.Params.N, pi.Params.Logq, 0, pi.seed)
	nextSeed := NextSeed(pi.seed)
	A2 := MatrixRand(pi.Params.L/pi.DBInfo.X, pi.Params.N, pi.Params.Logq, 0, nextSeed)
	pi.A = MakeState(A1, A2)
	return pi.A
}

func (pi *DoublePIR) Setup(DB *InternalDB) (State, State) {
	pi.Init()
	A1 := pi.A.Data[0]
	A2 := pi.A.Data[1]
	H1 := MatrixMul(DB.Data, A1)
	H1.Transpose()
	H1.Expand(pi.Params.P, pi.Params.delta())
	H1.ConcatCols(DB.Info.X)
	H2 := MatrixMul(H1, A2)

	// pack the database more tightly, because the online computation is memory-bound
	DB.Data.Add(pi.Params.P / 2)
	DB.Squish()
	pi.DBInfo = DB.Info

	H1.Add(pi.Params.P / 2)
	H1.Squish(10, 3)

	A2_copy := A2.RowsDeepCopy(0, A2.Rows) // deep copy whole matrix
	if A2_copy.Rows%3 != 0 {
		A2_copy.Concat(MatrixZeros(3-(A2_copy.Rows%3), A2_copy.Cols))
	}
	A2_copy.Transpose()

	return MakeState(H1, A2_copy), MakeState(H2)
}

func (pi *DoublePIR) FakeSetup(DB *InternalDB) (State, float64) {
	var seed128 [16]byte
	if _, err := rand.Read(seed128[:]); err != nil {
		panic(err)
	}

	info := pi.DBInfo
	H1 := MatrixRand(pi.Params.N*pi.Params.delta()*info.X, pi.Params.L/info.X, 0, pi.Params.P, seed128)
	offline_download := float64(pi.Params.N*pi.Params.delta()*info.X*pi.Params.N*uint64(pi.Params.Logq)) / (8.0 * 1024.0)
	//fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	// pack the database more tightly, because the online computation is memory-bound
	DB.Data.Add(pi.Params.P / 2)
	DB.Squish()
	pi.DBInfo = DB.Info

	H1.Add(pi.Params.P / 2)
	H1.Squish(10, 3)

	A2_rows := pi.Params.L / info.X
	if A2_rows%3 != 0 {
		A2_rows += (3 - (A2_rows % 3))
	}
	if _, err := rand.Read(seed128[:]); err != nil {
		panic(err)
	}
	A2_copy := MatrixRand(pi.Params.N, A2_rows, pi.Params.Logq, 0, seed128)

	return MakeState(H1, A2_copy), offline_download
}

func (pi *DoublePIR) Query(indexes []uint64) (Msg, State) {
	//batchSize := len(indexes)
	queries := Msg{}  //make([]Msg, batchSize)
	states := State{} //make([]State, batchSize)

	info := pi.DBInfo
	p := pi.Params
	A1 := pi.A.Data[0]
	A2 := pi.A.Data[1]
	delta := p.Delta()
	ne_x := info.Ne / info.X

	for _, idx := range indexes {
		var seed128 [16]byte
		if _, err := rand.Read(seed128[:]); err != nil {
			panic(err)
		}

		secret1 := MatrixRand(p.N, 1, p.Logq, 0, seed128)
		err1 := MatrixGaussian(p.M, 1)
		query1 := MatrixMul(A1, secret1)
		query1.MatrixAdd(err1)

		i2 := idx % p.M
		query1.Data[i2] += C.Elem(delta)

		if p.M%info.Squishing != 0 {
			query1.AppendZeros(info.Squishing - (p.M % info.Squishing))
		}

		//currState := MakeState(secret1)
		//currMsg := MakeMsg(query1)
		queries.Data = append(queries.Data, query1)
		states.Data = append(states.Data, secret1)

		i1_base := (idx / p.M) * ne_x

		for j := uint64(0); j < ne_x; j++ {
			if _, err := rand.Read(seed128[:]); err != nil {
				panic(err)
			}

			secret2 := MatrixRand(p.N, 1, p.Logq, 0, seed128)
			err2 := MatrixGaussian(p.L/info.X, 1)
			query2 := MatrixMul(A2, secret2)
			query2.MatrixAdd(err2)

			query2.Data[i1_base+j] += C.Elem(delta)

			l_x := p.L / info.X
			if l_x%info.Squishing != 0 {
				query2.AppendZeros(info.Squishing - (l_x % info.Squishing))
			}

			queries.Data = append(queries.Data, query2)
			states.Data = append(states.Data, secret2)
		}

		//queries.Data = append(queries.Data, currMsg)
		//states.Data = append(states.Data, currState)
	}

	return queries, states
}

func (pi *DoublePIR) QueryOffline(batchSize uint64) (Msg, State) {
	states := State{}
	msgs := Msg{}

	info := pi.DBInfo
	p := pi.Params
	A1 := pi.A.Data[0]
	A2 := pi.A.Data[1]
	ne_x := info.Ne / info.X

	for b := uint64(0); b < batchSize; b++ {
		var seed128 [16]byte

		if _, err := rand.Read(seed128[:]); err != nil {
			panic(err)
		}
		secret1 := MatrixRand(p.N, 1, p.Logq, 0, seed128)
		err1 := MatrixGaussian(p.M, 1)
		query1 := MatrixMul(A1, secret1)
		query1.MatrixAdd(err1)

		if p.M%info.Squishing != 0 {
			query1.AppendZeros(info.Squishing - (p.M % info.Squishing))
		}

		//currState := MakeState(secret1)
		//currMsg := MakeMsg(query1)
		// 1 + ne_x
		states.Data = append(states.Data, secret1)
		msgs.Data = append(msgs.Data, query1)

		for j := uint64(0); j < ne_x; j++ {
			if _, err := rand.Read(seed128[:]); err != nil {
				panic(err)
			}
			secret2 := MatrixRand(p.N, 1, p.Logq, 0, seed128)
			err2 := MatrixGaussian(p.L/info.X, 1)
			query2 := MatrixMul(A2, secret2)
			query2.MatrixAdd(err2)

			if (p.L/info.X)%info.Squishing != 0 {
				query2.AppendZeros(info.Squishing - ((p.L / info.X) % info.Squishing))
			}

			states.Data = append(states.Data, secret2)
			msgs.Data = append(msgs.Data, query2)
		}

		//states[b] = currState
		//msgs[b] = currMsg
	}

	return msgs, states
}

func (pi *DoublePIR) QueryOnline(indexes []uint64, offlineMsgs Msg) Msg {
	//batchSize := len(indexes)
	//if batchSize != len(offlineMsgs) {
	//	panic("DoublePIR Online: Batch size mismatch")
	//}

	onlineMsgs := Msg{}
	info := pi.DBInfo
	p := pi.Params
	delta := p.Delta()
	ne_x := info.Ne / info.X

	for i, idx := range indexes {
		offset0 := (1 + ne_x) * uint64(i)
		query1 := offlineMsgs.Data[offset0]
		i2 := idx % p.M
		query1.Data[i2] += C.Elem(delta)

		//currMsg := MakeMsg(query1)
		onlineMsgs.Data = append(onlineMsgs.Data, query1)

		i1_base := (idx / p.M) * ne_x
		for j := uint64(0); j < ne_x; j++ {
			query2 := offlineMsgs.Data[offset0+j+1]
			query2.Data[i1_base+j] += C.Elem(delta)

			onlineMsgs.Data = append(onlineMsgs.Data, query2)
		}

		//onlineMsgs.Data = append(onlineMsgs.Data, currMsg)
	}

	return onlineMsgs
}

func (pi *DoublePIR) Answer(DB *InternalDB, queries Msg, serverHint State) Msg {

	answers := Msg{}
	H1 := serverHint.Data[0]
	A2_transpose := serverHint.Data[1]

	len := len(queries.Data)

	ne_x := int(DB.Info.Ne / DB.Info.X)

	batchSize := len / (ne_x + 1)

	for i := 0; i < batchSize; i++ {
		offset := (ne_x + 1) * i

		a1 := new(Matrix)
		batch_sz := DB.Data.Rows

		q1 := queries.Data[offset]
		batch_sz = DB.Data.Rows

		a := MatrixMulVecPacked(DB.Data.SelectRows(0, batch_sz),
			q1, DB.Info.Basis, DB.Info.Squishing)
		a1.Concat(a)

		a1.TransposeAndExpandAndConcatColsAndSquish(pi.Params.P, pi.Params.delta(), DB.Info.X, 10, 3)
		h1 := MatrixMulTransposedPacked(a1, A2_transpose, 10, 3)
		//fmt.Println("h1.Rows : ", h1.Rows, " h1.Cols : ", h1.Cols)

		//msg := MakeMsg(h1)
		answers.Data = append(answers.Data, h1)

		for j := 0; j < ne_x; j++ {
			q2 := queries.Data[offset+1+j]
			a2 := MatrixMulVecPacked(H1, q2, 10, 3)
			h2 := MatrixMulVecPacked(a1, q2, 10, 3)

			//msg.Data = append(msg.Data, a2)
			answers.Data = append(answers.Data, a2)

			//msg.Data = append(msg.Data, h2)
			answers.Data = append(answers.Data, h2)
		}

	}

	return answers
}

func (pi *DoublePIR) Reconstruct(indexes []uint64, clientHint State, query Msg, answer Msg, clientState State) [][]uint64 {
	batchSize := len(indexes)
	results := make([][]uint64, batchSize)
	info := pi.DBInfo
	p := pi.Params
	H2 := clientHint.Data[0]

	ne_x := int(info.Ne / info.X)
	for i := 0; i < batchSize; i++ {
		offset0 := (1 + ne_x) * i   // query, state
		offset1 := offset0 + ne_x*i // (1 + 2*ne_x) * i // answer

		h1 := answer.Data[offset1].RowsDeepCopy(0, answer.Data[offset1].Rows) // deep copy whole matrix
		secret1 := clientState.Data[offset0]

		ratio := p.P / 2
		val1 := uint64(0)
		for j := uint64(0); j < p.M; j++ {
			val1 += ratio * query.Data[offset0].Get(j, 0)
		}
		val1 %= (1 << p.Logq)
		val1 = (1 << p.Logq) - val1

		val2 := uint64(0)
		for j := uint64(0); j < p.L/info.X; j++ {
			val2 += ratio * query.Data[offset0+1].Get(j, 0)
		}
		val2 %= (1 << p.Logq)
		val2 = (1 << p.Logq) - val2

		A2 := pi.A.Data[1]
		if (A2.Cols != p.N) || (h1.Cols != p.N) {
			panic("Should not happen!")
		}
		for j1 := uint64(0); j1 < p.N; j1++ {
			val3 := uint64(0)
			for j2 := uint64(0); j2 < A2.Rows; j2++ {
				val3 += ratio * A2.Get(j2, j1)
			}
			val3 %= (1 << p.Logq)
			val3 = (1 << p.Logq) - val3
			v := C.Elem(val3)
			for k := uint64(0); k < h1.Rows; k++ {
				h1.Data[k*h1.Cols+j1] += v
			}
		}

		//offset := int(0) //(info.Ne / info.X * 2) // for batching
		var vals []uint64
		for j := int(0); j < int(info.Ne/info.X); j++ {
			offset2 := offset1 + 1 + 2*j

			a2 := answer.Data[offset2]
			h2 := answer.Data[offset2+1]
			secret2 := clientState.Data[offset0+1+j]
			h2.Add(val2)

			for k := uint64(0); k < info.X; k++ {
				state := a2.RowsDeepCopy(k*p.N*p.delta(), p.N*p.delta())
				state.Add(val2)
				state.Concat(h2.SelectRows(k*p.delta(), p.delta()))

				hint := H2.RowsDeepCopy(k*p.N*p.delta(), p.N*p.delta())
				hint.Concat(h1.SelectRows(k*p.delta(), p.delta()))

				interm := MatrixMul(hint, secret2)
				state.MatrixSub(interm)
				state.Round(p)
				state.Contract(p.P, p.delta())

				noised := uint64(state.Data[p.N]) + val1
				for l := uint64(0); l < p.N; l++ {
					noised -= uint64(secret1.Data[l] * state.Data[l])
					noised = noised % (1 << p.Logq)
				}
				vals = append(vals, p.Round(noised))
			}
		}

		results[i] = ReconstructElem(vals, indexes[i], info)
		results[i] = UintPToUintQTrunc(results[i], p.Logp, uint64(64), (pi.DBInfo.BitsPerEntry+63)/64)
	}

	return results
}

func (pi *DoublePIR) Reset(DB *InternalDB) {
	DB.Unsquish()
	DB.Data.Sub(pi.Params.P / 2)
}
func (pi *DoublePIR) LoadHint(id string, filepath string, internalDB *InternalDB) (*State, *State) {

	serverHintFilename := filepath + id + "server_state" + ".gob"
	clientHintFilename := filepath + id + "client_state" + ".gob"
	seedFilename := filepath + id + "seed" + ".gob"

	if utils.FileExists(serverHintFilename) && utils.FileExists(clientHintFilename) && utils.FileExists(seedFilename) {
		fmt.Printf("Loading precomputed DoublePIR hints for ID: %s...\n", id)

		utils.LoadFromFile(&pi.seed, seedFilename)

		A1 := MatrixRand(pi.Params.M, pi.Params.N, pi.Params.Logq, 0, pi.seed)

		nextSeed := NextSeed(pi.seed)
		A2 := MatrixRand(pi.Params.L/pi.DBInfo.X, pi.Params.N, pi.Params.Logq, 0, nextSeed)

		pi.A = MakeState(A1, A2)

		serverHint := &State{}
		utils.LoadFromFile(&serverHint, serverHintFilename)

		clientHint := &State{}
		utils.LoadFromFile(&clientHint, clientHintFilename)

		pi.FakeSetup(internalDB)

		return serverHint, clientHint
	} else {
		fmt.Println("DoublePIR Hints not found. Computing Setup...")

		serverHint, clientHint := pi.Setup(internalDB)

		utils.SaveToFile(pi.seed, seedFilename)
		utils.SaveToFile(serverHint, serverHintFilename)
		utils.SaveToFile(clientHint, clientHintFilename)

		return &serverHint, &clientHint
	}
}
