package stratum

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"github.com/conformal/btcwire"
	"github.com/conformal/fastsha256"
	"math"
	"strconv"
)

func randhash() string {
	randbytes := make([]byte, 4)
	rand.Read(randbytes)

	h := sha1.New()
	h.Write(randbytes)
	return hex.EncodeToString(h.Sum(nil))
}

func DecodeNtime(ntime string) (n int64, err error) {
	return strconv.ParseInt(ntime, 16, 64)
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

// The following methods are copied from btcchain with modification.
// We do not want to introduced too much dependency, eg: btcdb.

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

func BuildMerkleRoot(mkBranches []*btcwire.ShaHash) *btcwire.ShaHash {
	merkles := BuildMerkleTree(mkBranches)
	return merkles[len(merkles)-1]
}
