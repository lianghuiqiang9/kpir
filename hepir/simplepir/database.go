package simplepir

import (
	"math"
)

type DBinfo struct {
	NumEntries   uint64 // number of DB entries.
	BitsPerEntry uint64 // number of bits per DB entry.

	Packing uint64 // number of DB entries per Z_p elem, if log(p) > DB entry size.
	Ne      uint64 // number of Z_p elems per DB entry, if DB entry size > log(p).

	X uint64 // tunable param that governs communication,
	// must be in range [1, ne] and must be a divisor of ne;
	// represents the number of times the scheme is repeated.
	P    uint64 // plaintext modulus.
	Logq uint64 // (logarithm of) ciphertext modulus.

	// For in-memory DB compression
	Basis     uint64
	Squishing uint64
	Cols      uint64
}

type InternalDB struct {
	Info DBinfo
	Data *Matrix
}

func (DB *InternalDB) Squish() {

	DB.Info.Basis = 10
	DB.Info.Squishing = 3
	DB.Info.Cols = DB.Data.Cols
	DB.Data.Squish(DB.Info.Basis, DB.Info.Squishing)

	if (DB.Info.P > (1 << DB.Info.Basis)) || (DB.Info.Logq < DB.Info.Basis*DB.Info.Squishing) {
		panic("Bad params")
	}
}

func (DB *InternalDB) Unsquish() {
	DB.Data.Unsquish(DB.Info.Basis, DB.Info.Squishing, DB.Info.Cols)
}

/*
	func ReconstructElem(vals []uint64, index uint64, info DBinfo) uint64 {
		q := uint64(1 << info.Logq)

		for i, _ := range vals {
			vals[i] = (vals[i] + info.P/2) % q
			vals[i] = vals[i] % info.P
		}

		val := Reconstruct_from_base_p(info.P, vals)
		fmt.Println("info.Packing: ", info.Packing)
		if info.Packing > 0 {
			val = Base_p((1 << info.BitsPerEntry), val, index%info.Packing)
		}

		return val
	}
*/
func ReconstructElem(vals []uint64, index uint64, info DBinfo) []uint64 {
	q := uint64(1 << info.Logq)

	for i, _ := range vals {
		vals[i] = (vals[i] + info.P/2) % q
		vals[i] = vals[i] % info.P
	}

	val := vals

	return val
}

/*
func (DB *Database) GetElem(i uint64) uint64 {
	if i >= DB.Info.NumEntries {
		panic("Index out of range")
	}

	col := i % DB.Data.Cols
	row := i / DB.Data.Cols

	if DB.Info.Packing > 0 {
		new_i := i / DB.Info.Packing
		col = new_i % DB.Data.Cols
		row = new_i / DB.Data.Cols
	}

	var vals []uint64
	for j := row * DB.Info.Ne; j < (row+1)*DB.Info.Ne; j++ {
		vals = append(vals, DB.Data.Get(j, col))
	}
	fmt.Println("vals: ", vals)
	return ReconstructElem(vals, i, DB.Info)
}
*/
/*
	func (DB *Database) GetElemVec(i uint64) []uint64 {
		if i >= DB.Info.NumEntries {
			panic("Index out of range")
		}

		col := i % DB.Data.Cols
		row := i / DB.Data.Cols

		if DB.Info.Packing > 0 {
			new_i := i / DB.Info.Packing
			col = new_i % DB.Data.Cols
			row = new_i / DB.Data.Cols
		}

		var vals []uint64
		for j := row * DB.Info.Ne; j < (row+1)*DB.Info.Ne; j++ {
			vals = append(vals, DB.Data.Get(j, col))
		}
		return ReconstructElemVec(vals, i, DB.Info)
	}
*/
func ApproxSquareDatabaseDims(N, row_length, p uint64) (uint64, uint64) {
	db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
	l := uint64(math.Floor(math.Sqrt(float64(db_elems))))
	rem := l % elems_per_entry
	if rem != 0 {
		l += elems_per_entry - rem
	}

	m := uint64(math.Ceil(float64(db_elems) / float64(l)))
	return l, m
}

func ApproxDatabaseDims(N, row_length, p, lower_bound_m uint64) (uint64, uint64) {
	l, m := ApproxSquareDatabaseDims(N, row_length, p)
	if m >= lower_bound_m {
		return l, m
	}

	m = lower_bound_m
	db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
	l = uint64(math.Ceil(float64(db_elems) / float64(m)))

	rem := l % elems_per_entry
	if rem != 0 {
		l += elems_per_entry - rem
	}

	return l, m
}
