package simplepir

import (
	"math"
)

type State struct {
	Data []*Matrix
}

type Msg struct {
	Data []*Matrix
}

// return all Elem nums in Data
func (m *State) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

// return all Elem nums in Data
func (m *Msg) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

func MakeState(elems ...*Matrix) State {
	st := State{}
	for _, elem := range elems {
		st.Data = append(st.Data, elem)
	}
	return st
}

func MakeMsg(elems ...*Matrix) Msg {
	msg := Msg{}
	for _, elem := range elems {
		//fmt.Println("k")
		msg.Data = append(msg.Data, elem)
	}
	return msg
}

// Returns the i-th elem in the representation of m in base p.
func Base_p(p, m, i uint64) uint64 {
	for j := uint64(0); j < i; j++ {
		m = m / p
	}
	return (m % p)
}

// Returns the element whose base-p decomposition is given by the values in vals
func Reconstruct_from_base_p(p uint64, vals []uint64) uint64 {
	res := uint64(0)
	coeff := uint64(1)
	for _, v := range vals {
		res += coeff * v
		coeff *= p
	}
	return res
}

// Returns how many entries in Z_p are needed to represent an element in Z_q
func Compute_num_entries_base_p(p, log_q uint64) uint64 {
	log_p := uint64(math.Log2(float64(p)))
	return uint64(math.Ceil(float64(log_q) / float64(log_p)))
}

// Returns how many Z_p elements are needed to represent a database of N entries,
// each consisting of row_length bits.
func Num_DB_entries(N, row_length, p uint64) (uint64, uint64, uint64) {
	// use multiple Z_p elems to represent a single DB entry
	ne := Compute_num_entries_base_p(p, row_length)
	return N * ne, ne, 0
}

func avg(data []float64) float64 {
	sum := 0.0
	num := 0.0
	for _, elem := range data {
		sum += elem
		num += 1.0
	}
	return sum / num
}

func stddev(data []float64) float64 {
	avg := avg(data)
	sum := 0.0
	num := 0.0
	for _, elem := range data {
		sum += math.Pow(elem-avg, 2)
		num += 1.0
	}
	variance := sum / num // not -1!
	return math.Sqrt(variance)
}

func UintPToUintQTrunc(data []uint64, p, q, t uint64) []uint64 {
	temp := UintPToUintQ(data, p, q)
	return temp[:t]
}

func UintPToUintQ(data []uint64, p, q uint64) []uint64 {
	if len(data) == 0 {
		return nil
	}

	totalBits := uint64(len(data)) * p
	numWords := (totalBits + q - 1) / q
	result := make([]uint64, numWords)

	var (
		bitBuf   uint128_manual
		bitCount uint64
		outIdx   uint64
		maskP    uint64 = (1 << p) - 1
		maskQ    uint64 = (1 << q) - 1
	)

	if p == 64 {
		maskP = 0xFFFFFFFFFFFFFFFF
	}
	if q == 64 {
		maskQ = 0xFFFFFFFFFFFFFFFF
	}

	for _, word := range data {
		cleanWord := word & maskP
		bitBuf.low |= (cleanWord << bitCount)
		if bitCount+p > 64 {
			bitBuf.high |= (cleanWord >> (64 - bitCount))
		}
		bitCount += p

		for bitCount >= q {
			result[outIdx] = bitBuf.low & maskQ
			outIdx++

			bitBuf.low = (bitBuf.low >> q) | (bitBuf.high << (64 - q))
			bitBuf.high >>= q
			bitCount -= q
		}
	}

	if bitCount > 0 && outIdx < numWords {
		result[outIdx] = bitBuf.low & maskQ
	}

	return result
}

type uint128_manual struct {
	low  uint64
	high uint64
}
