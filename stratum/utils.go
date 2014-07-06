package stratum

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"github.com/conformal/btcwire"
	"github.com/conformal/fastsha256"
	"math"
	"math/big"
	"strconv"
	"time"
)

func randhash() string {
	randbytes := make([]byte, 4)
	rand.Read(randbytes)

	h := sha1.New()
	h.Write(randbytes)
	return hex.EncodeToString(h.Sum(nil))
}

func HexToInt64(ntime string) (n int64, err error) {
	return strconv.ParseInt(ntime, 16, 64)
}

func HexToInt32(version string) (n int32, err error) {
	v, err := strconv.ParseInt(version, 16, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func HexToUint32(version string) (n uint32, err error) {
	v, err := strconv.ParseUint(version, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

func DoubleSha256(b []byte) []byte {
	hasher := fastsha256.New()
	hasher.Write(b)
	sum := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write(sum)
	return hasher.Sum(nil)
}

func CoinbaseHash(coinbase1, nonce1, nonce2, coinbase2 string) []byte {
	coinbase := coinbase1 + nonce1 + nonce2 + coinbase2
	buf, _ := hex.DecodeString(coinbase)
	return DoubleSha256(buf)
}

func HexToString(raw []byte) string {
	return hex.EncodeToString(raw)
}

// -------------------------------------------------------------------
// The following methods are copied from btcchain with modification.
// We do not want to introduced too much dependency, eg: btcdb.
// -------------------------------------------------------------------

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in bitcoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with bitcoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// ShaHashToBig converts a btcwire.ShaHash into a big.Int that can be used to
// perform math comparisons.
func ShaHashToBig(hash *btcwire.ShaHash) *big.Int {
	// A ShaHash is in little-endian, but the big package wants the bytes
	// in big-endian.  Reverse them.  ShaHash.Bytes makes a copy, so it
	// is safe to modify the returned buffer.
	buf := hash.Bytes()
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

// HashMerkleBranches takes two hashes, treated as the left and right tree
// nodes, and returns the hash of their concatenation.  This is a helper
// function used to aid in the generation of a merkle tree.
func HashMerkleBranches(left *btcwire.ShaHash, right *btcwire.ShaHash) *btcwire.ShaHash {
	// Concatenate the left and right nodes.
	var sha [btcwire.HashSize * 2]byte
	copy(sha[:btcwire.HashSize], left.Bytes())
	copy(sha[btcwire.HashSize:], right.Bytes())

	// Create a new sha hash from the double sha 256.  Ignore the error
	// here since SetBytes can't fail here due to the fact DoubleSha256
	// always returns a []byte of the right size regardless of input.
	newSha, _ := btcwire.NewShaHash(btcwire.DoubleSha256(sha[:]))
	return newSha
}

func BuildMerkleTree(mkBranches []*btcwire.ShaHash) []*btcwire.ShaHash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(mkBranches))
	arraySize := nextPoT*2 - 1
	merkles := make([]*btcwire.ShaHash, arraySize)
	copy(merkles, mkBranches)

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newSha := HashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newSha

		// The normal case sets the parent node to the double sha256
		// of the concatentation of the left and right children.
		default:
			newSha := HashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newSha
		}
		offset++
	}

	return merkles
}

// NewShaHashFromStr converts a hash string in the standard bitcoin big-endian
// form to a ShaHash (which is little-endian).
func NewShaHashFromMerkleBranch(hash string) (*btcwire.ShaHash, error) {
	// Return error if hash string is too long.
	if len(hash) > btcwire.MaxHashStringSize {
		return nil, btcwire.ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.
	if len(hash)%2 != 0 {
		hash = "0" + hash
	}

	// Convert string hash to bytes.
	buf, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	// Make sure the byte slice is the right length by appending zeros to
	// pad it out.
	blen := len(buf)
	pbuf := buf
	if btcwire.HashSize-blen > 0 {
		pbuf = make([]byte, btcwire.HashSize)
		copy(pbuf, buf)
	}

	// Create the sha hash using the byte slice and return it.
	return btcwire.NewShaHash(pbuf)
}

// Merkle branches has at least one which is coinbase hash.
func BuildMerkleRoot(mkBranches []*btcwire.ShaHash) *btcwire.ShaHash {
	// merkles := BuildMerkleTree(mkBranches)
	// return merkles[len(merkles)-1]
  root := mkBranches[0]
	if len(mkBranches) == 0 {
		return root
	}
  for _, node := range mkBranches[1:] {
		root = HashMerkleBranches(root, node)
	}
  return root
}

func SerializeHeader(job *Job, merkleRoot *btcwire.ShaHash, ntime string, nonce string) (*btcwire.BlockHeader, error) {
	Version, err := HexToInt32(job.Version)
	if err != nil {
		return nil, err
	}
	Bits, err := HexToUint32(job.Nbits)
	if err != nil {
		return nil, err
	}
	Nonce, err := HexToUint32(nonce)
	if err != nil {
		return nil, err
	}
	reversedHash, err := ReversePrevHash(job.PrevHash)
	if err != nil {
		return nil, err
	}
	PrevHash, err := btcwire.NewShaHash(reversedHash)
	// PrevHash, err := btcwire.NewShaHashFromStr(job.PrevHash)
	if err != nil {
		return nil, err
	}
	tsUnix, err := HexToInt64(ntime)
	if err != nil {
		return nil, err
	}
	Timestamp := time.Unix(tsUnix, 0)

	// https://en.bitcoin.it/wiki/Protocol_specification#Block_Headers
	header := &btcwire.BlockHeader{
		Version:    Version,
		PrevBlock:  *PrevHash,
		MerkleRoot: *merkleRoot,
		Timestamp:  Timestamp,
		Bits:       Bits,
		Nonce:      Nonce,
	}
	return header, nil
}

func HeaderToBig(header *btcwire.BlockHeader) *big.Int {
	headerHash, _ := header.BlockSha()
	return ShaHashToBig(&headerHash)
}


// Maximum target?
//
// The maximum target used by SHA256 mining devices is:
// 0x00000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
//
// Because Bitcoin stores the target as a floating-point type, this is truncated:
// 0x00000000FFFF0000000000000000000000000000000000000000000000000000
//
// Since a lower target makes Bitcoin generation more difficult, the maximum
// target is the lowest possible difficulty.
func Target(diff1_target string, newDiff int) {
	
}

// Server as a single purpose, reverse prevhash in stratum job.
// The prevhash is the hash of the previous block. Apparently mixing
// big-ending and little-endian isn't confusing enough so this hash value
// also has every block of 4 bytes reversed.
// See: http://stackoverflow.com/questions/9245235/golang-midstate-sha-256-hash
func ReversePrevHash(hash string) ([]byte, error) {
	buf, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	length := len(buf)
	ret := make([]byte, length)
	times := length / 4 // divide by four bytes group
	for i := 0; i < times; i++ {
		index := i * 4
		copy(ret[index:index+4], ReverseBytes(buf[index:index+4]))
	}
	return ret, nil
}

func ReverseBytes(bytes []byte) []byte {
	length := len(bytes)
  ret := make([]byte, length)
  for i := 0; i < length; i++ {
    ret[i] = bytes[length-1-i]
  }
  return ret
}
