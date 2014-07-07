package stratum_test

import (
	"bytes"
	"encoding/hex"
	"github.com/conformal/btcwire"
	"github.com/yinhm/ninepool/birpc"
	"github.com/yinhm/ninepool/stratum"
	"math/big"
	"testing"
)

func TestExtraNonceCounter(t *testing.T) {
	counter := stratum.NewExtraNonceCounter()
	if counter.Size != 4 {
		t.Errorf("incorrect counter size %d != 4", counter.Size)
	}

	if counter.Next() != "08000001" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Next() != "08000002" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Nonce2Size() != 4 {
		t.Errorf("incorrect Nonce2Size")
	}
}

func TestProxyExtraNonceCounter(t *testing.T) {
	counter := stratum.NewProxyExtraNonceCounter("08000001", 2, 2)

	if next := counter.Next(); next != "080000010001" {
		t.Errorf("incorrect next nonce1: %v", next)
	}

	if counter.Next() != "080000010002" {
		t.Errorf("incorrect next nonce1")
	}

	if counter.Nonce2Size() != 2 {
		t.Errorf("incorrect Nonce2Size")
	}
}

func TestHexToInt64(t *testing.T) {
	ntime, err := stratum.HexToInt64("504e86ed")
	if err != nil || ntime != int64(1347323629) {
		t.Errorf("failed on parse ntime")
	}
}

func TestParseInt32(t *testing.T) {
	version, err := stratum.HexToInt32("00000002")
	if err != nil || version != int32(2) {
		t.Errorf("failed on parse version")
	}
}

func TestCoinbaseHash(t *testing.T) {
	coinbase := stratum.HexToString(
		stratum.CoinbaseHash(
			"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
			"08000001",
			"0001",
			"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000",
		))
	if coinbase != "94f317184323c9965abd532450519e6db6859b53b0551c6b8702c1f300ec9b51" {
		t.Errorf("failed to build coinbase %s", coinbase)
	}
}

func TestBuildMerkleRoot(t *testing.T) {
	// https://blockexplorer.com/rawblock/0000000000000000151f00e7b882b15f1523587f4c97c8f16cac185946039ba1
	txHashes := []string{
		"e57c35461e4be6b197b22f126d43561022d4107cc1326a9cb1e892b43e4d48db",
		"e0909212d97ec600196fadc209b2d7981a89a7cd903e5c971341980f828db7a4",
		"50e94516832160237f8b391510636ed9a31ebe75296a5f193755f802b40e2e8d",
		"c79660e2eedd16bfd64c908995c469a844c723c270290afb138edca2203f332b",
		"1841d48c65df428aef3ce5e1a74fe1d8e87cee3b46e40be0aa773fa0efac1f9f",
		"3d0502aebdfac4f5b7595bac307e81c8fcb0ab96fadffb38f350029beba705db",
		"c4b035e3d51318eed15361f306cce27123ca7a3c8e3ce565c68a649faf3d5338",
	}

	txList := make([]*btcwire.ShaHash, 0, len(txHashes)+1)
	for _, hash := range txHashes {
		txHash, _ := btcwire.NewShaHashFromStr(hash)
		txList = append(txList, txHash)
	}

	mkRoot := stratum.BuildMerkleRoot(txList)
	expected, _ := btcwire.NewShaHashFromStr("1844f9fbe5ca95527b7413484f3bcdd7a247df3a7c7d5dee2e16330996fa1e77")
	if !expected.IsEqual(mkRoot) {
		t.Errorf("Merkle root hash not match:\n%s\n%s", mkRoot, expected)
	}
}

func TestJobMerkleRoot(t *testing.T) {
	list := birpc.List{
		"1",
		"16fec96ac8501b7178c41590c7b378b940120cfd3c869b2c0000d25100000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703f81104062f503253482f04041db65308",
		"0d2f6e6f64655374726174756d2f000000000240eda87e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988acc00b5a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",
		birpc.List{},
		"00000002",
		"1b013164",
		"53b61d05",
		false,
	}
	job, _ := stratum.NewJob(list)

	extraNonce1 := "38000000"
	extraNonce2 := "00000000"

	merkleRoot := job.MerkleRoot(extraNonce1, extraNonce2)
	expected, _ := btcwire.NewShaHashFromStr("e69718c9c04d411af24522315652306479d70b2e6bef31cc202fa13b58125bba")
	if !expected.IsEqual(merkleRoot) {
		t.Errorf("merkleRoot: %v", merkleRoot)
	}
}

func TestReversePrevHash(t *testing.T) {
	actual, _ := stratum.ReversePrevHash("69fc72e76db0e764615a858f483e3566e42d56b2bc7a03adce9492887010eda8")
	expected, _ := hex.DecodeString("e772fc6964e7b06d8f855a6166353e48b2562de4ad037abc889294cea8ed1070")
	if string(actual) != string(expected) {
		t.Errorf("prevhash reverse failed,\n%s\n%s", string(actual), string(expected))
	}
}

func TestDifficulity(t *testing.T) {
	// sha256d
	// upstream <=> cpuminer
	//{"id":1,"result":[[["mining.set_difficulty","deadbeefcafebabe0100000000000000"],["mining.notify","deadbeefcafebabe0100000000000000"]],"38000000",4],"error":null}
	//{"id":null,"method":"mining.set_difficulty","params":[1]}
	//{"id":null,"method":"mining.notify","params":["4","16fec96ac8501b7178c41590c7b378b940120cfd3c869b2c0000d25100000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703f81104062f503253482f04041db65308","0d2f6e6f64655374726174756d2f000000000240eda87e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988acc00b5a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",[],"00000002","1b013164","53b61d05",false]}
	//{"method": "mining.submit", "params": ["n4p4cLr6mfp1obJAda8jt1gJjuXyjQ8GTk", "4", "00000000", "53b61d05", "0ae44d20"], "id":4}

	list := birpc.List{
		"4",
		"16fec96ac8501b7178c41590c7b378b940120cfd3c869b2c0000d25100000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703f81104062f503253482f04041db65308",
		"0d2f6e6f64655374726174756d2f000000000240eda87e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988acc00b5a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",
		birpc.List{},
		"00000002",
		"1b013164",
		"53b61d05",
		false,
	}
	job, _ := stratum.NewJob(list)

	extraNonce1 := "38000000"
	extraNonce2 := "00000000"
	ntime := "53b61d05"
	nonce := "0ae44d20"

	merkleRoot := job.MerkleRoot(extraNonce1, extraNonce2)
	header, err := stratum.SerializeHeader(job, merkleRoot, ntime, nonce)
	if err != nil {
		t.Errorf("unexpected: %v", err)
		return
	}

	var buf bytes.Buffer
	_ = header.Serialize(&buf)
	headerStr := hex.EncodeToString(buf.Bytes()[0:80])

	if headerStr != "020000006ac9fe16711b50c89015c478b978b3c7fd0c12402c9b863c51d2000000000000ba5b12583ba12f20cc31ef6b2e0bd77964305256312245f21a414dc0c91897e6051db6536431011b204de40a" {
		t.Errorf("header buffer not expected: %v", headerStr)
	}

	headerHash, _ := header.BlockSha()
	// hash in standard bitcoin big-endian form.
	expHeaderHash := "00000000d3d7347bfebb9587d01ebbdd4840579ebb6f6bae0c190bf363d0cd3d"
	if headerHash.String() != expHeaderHash {
		t.Errorf("wrong header hash %v", headerHash.String())
	}
	shareDiff := stratum.ShaHashToBig(&headerHash)

	// diff1 := 0x00000000FFFF0000000000000000000000000000000000000000000000000000
	compact := uint32(0x1d00ffff)
	diff1 := stratum.CompactToBig(compact)

	target := new(big.Int).Div(diff1, big.NewInt(int64(1)))
	if shareDiff.Cmp(target) > 0 {
		t.Errorf("share difficulty not meet the target.")
		t.Errorf("header big: %v", shareDiff)
		t.Errorf("job target: %v", target)
	}
}

func TestDifficulityWithTnx(t *testing.T) {
	// sha256d
	// < {"id":1,"result":[[["mining.set_difficulty","4d7c80542434bde760ef7182e665895bbc4d67f0"],["mining.notify","4d7c80542434bde760ef7182e665895bbc4d67f0"]],"580000020001",2]}
	// < {"method":"mining.notify","params":["30","5da479a459762b2082abb7496cea48694c50b7a2bc070457bb7fd0f400000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703851204062f503253482f0446ecb65308","0d2f6e6f64655374726174756d2f00000000029d6bab7e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988ac677c5a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",["2595fd2d192731df10d131eb77e108449c189e88a6797d7b6027be328476433e","01cd5c94bb486111524040ef65fb23855fd524ea27f508581f521994ac72c44c","6be8f181cce5d752f9b994c1e2dfe371bcce0d602d4a8cd6da9e84c5e87b3db2","2cfa58a93fa1e1c75eda9fedfd49eec22a2af91f6de192e253b86c7df7a5aebf","692a1ab4bba06b055c5e5ec519e7800665b2ab6df86d05ffe36dbf769ada9d68"],"00000002","1b013164","53b6ec46",false]}
	// hash <= target
	// Hash:   00000000421a0986cafd57f748e8d33469eb2d0583243b4a299060ac0ebc0b3c
	// Target: 00000000ffff000000000000000000000000000000000000000000000000000
	// > {"method": "mining.submit", "params": ["1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "30", "0000", "53b6ec46", "5ea00f07"], "id":4}

	//[2014-07-07 01:44:41] < {"id":1,"result":[[["mining.set_difficulty","9571966ef4f5f61b07ff237d2a09611ca6493aa9"],["mining.notify","9571966ef4f5f61b07ff237d2a09611ca6493aa9"]],"580000000002",2]}
	// [2014-07-07 01:48:49] < {"method":"mining.notify","params":["a","e4454cd2476e40492ca994025bf87527b1ed28818c66c17a0000cd0800000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff27031d1404062f503253482f04018cb95308","0d2f6e6f64655374726174756d2f00000000021072a97e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988ac30235a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",["036a7c012247b88a4b8943d81f54e89eae705f55de5cac61d3080388238ab32d","26acbcd2f3e76dc0827517fb4ef83ad2f8261f2d78269f5657707b33ce1f1dfd","14e3485ccca39e8c17fb669a72249ee8c3dc6dc23815ea4101f944bf46bdc773"],"00000002","1b013164","53b98c13",true]}
	// Hash:   0000000007a4a0e8212730fa0a832e89cb3d571445ca9f52b8eed811108cf3a4
	// Target: 00000000ffff0000000000000000000000000000000000000000000000000000
	// [2014-07-07 01:48:57] thread 1: 15252110 hashes, 1905 khash/s
	// [2014-07-07 01:48:57] > {"method": "mining.submit", "params": ["1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda", "a", "0000", "53b98c13", "e20f3e56"], "id":4}
	// [2014-07-07 01:48:57] < {"id":4,"result":true}

	list := birpc.List{
		"31",
		"e4454cd2476e40492ca994025bf87527b1ed28818c66c17a0000cd0800000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff27031d1404062f503253482f04018cb95308",
		"0d2f6e6f64655374726174756d2f00000000021072a97e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988ac30235a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",
		birpc.List{"036a7c012247b88a4b8943d81f54e89eae705f55de5cac61d3080388238ab32d", "26acbcd2f3e76dc0827517fb4ef83ad2f8261f2d78269f5657707b33ce1f1dfd", "14e3485ccca39e8c17fb669a72249ee8c3dc6dc23815ea4101f944bf46bdc773"},
		"00000002",
		"1b013164",
		"53b98c13",
		false,
	}
	job, _ := stratum.NewJob(list)

	extraNonce1 := "580000000002"
	extraNonce2 := "0000"
	ntime := "53b98c13"
	nonce := "e20f3e56"

	merkleRoot := job.MerkleRoot(extraNonce1, extraNonce2)
	expectedRoot, _ := btcwire.NewShaHashFromStr("ec4ccb35edd4930c86d35a24fa911b8840a8394228de534ed7a65ba9f4968f86")
	if !merkleRoot.IsEqual(expectedRoot) {
		t.Errorf("unexpected merkle root: %v", merkleRoot.String())
	}
	header, err := stratum.SerializeHeader(job, expectedRoot, ntime, nonce)
	if err != nil {
		t.Errorf("unexpected: %v", err)
		return
	}
	headerHash, _ := header.BlockSha()
	// hash in standard bitcoin big-endian form.
	// 26e3f8fefd1467c01bb0e0b7842e49d20cc42c81b17dfdfbbaffa71ea332307d
	expHeaderHash := "0000000007a4a0e8212730fa0a832e89cb3d571445ca9f52b8eed811108cf3a4"
	// NOMP header hash: 3c0bbc0eac6090294a3b2483052deb6934d3e848f757fdca86091a4200000000
	if headerHash.String() != expHeaderHash {
		t.Errorf("wrong header hash %v", headerHash.String())
	}
}

func TestBuildMerkleRootFromBranches(t *testing.T) {
	coinbase := stratum.HexToString(
		stratum.CoinbaseHash(
			"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703291404062f503253482f04c59eb95308",
			"68000000",
			"00020000",
			"0d2f6e6f64655374726174756d2f00000000024cdfaa7e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988aca4635a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",
		))
	if coinbase != "f72a5e25e7f311b1d131a73aabd5296fd4af0d17c16dd1ee544108bac0fb674e" {
		t.Errorf("failed to build coinbase %s", coinbase)
	}

	// [2014-07-07 03:08:53] < {"method":"mining.notify","params":["12","5de35fcda61f9c20bc2e045662123bf02e5c037f642cc2f400003c7700000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703291404062f503253482f04c59eb95308","0d2f6e6f64655374726174756d2f00000000024cdfaa7e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988aca4635a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",["99d31d7ae854079485d41aa5bf2c44b5a187a697d641d0472affd1ce8b524079","3b556c1f5442ede1df8be8063632948b00fd26b279a0ed566bf129e72945ae54","c31bf4de370220db4b211186fb5a4db95b6fdafa27a8c2fe22deb1b6e18d02db","b02c2eab55dda7d3b83a48405d1b939e0371cc6418a12a6861372ae16b21be26"],"00000002","1b013164","53b99ec5",false]}
	txHashes := []string{
		"f72a5e25e7f311b1d131a73aabd5296fd4af0d17c16dd1ee544108bac0fb674e",
		"99d31d7ae854079485d41aa5bf2c44b5a187a697d641d0472affd1ce8b524079",
		"3b556c1f5442ede1df8be8063632948b00fd26b279a0ed566bf129e72945ae54",
		"c31bf4de370220db4b211186fb5a4db95b6fdafa27a8c2fe22deb1b6e18d02db",
		"b02c2eab55dda7d3b83a48405d1b939e0371cc6418a12a6861372ae16b21be26",
	}

	txList := make([]*btcwire.ShaHash, 0, len(txHashes)+1)
	for _, hash := range txHashes {
		txHash, _ := stratum.NewShaHashFromMerkleBranch(hash)
		txList = append(txList, txHash)
	}

	merkleRoot := stratum.MerkleRootFromBranches(txList)
	expectedRoot, _ := btcwire.NewShaHashFromStr("57db1d776125ff2d4eab6a3feec14fed78ec0ffa2099f492949caf5c4589019e")
	if !merkleRoot.IsEqual(expectedRoot) {
		t.Errorf("unexpected merkle root: %v", merkleRoot.String())
	}
}
