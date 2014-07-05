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

func TestSerializeHeader(t *testing.T) {
	// http://www.righto.com/2014/02/bitcoin-mining-hard-way-algorithms.html
	//{"id":1,"result":[[["mining.set_difficulty","b4b6693b72a50c7116db18d6497cac52"],["mining.notify","ae6812eb4cd7735a302a8a9dd95cf71f"]],"4bc6af58",4],"error":null}
	list := birpc.List{
		"58af8d8c",
		"975b9717f7d18ec1f2ad55e2559b5997b8da0e3317c803780000000100000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4803636004062f503253482f04428b055308",
		"2e522cfabe6d6da0bd01f57abe963d25879583eea5ea6f08f83e3327eba9806b14119718cbb1cf04000000000000000000000001fb673495000000001976a91480ad90d403581fa3bf46086a91b2d9d4125db6c188ac00000000",
		birpc.List{"ea9da84d55ebf07f47def6b9b35ab30fc18b6e980fc618f262724388f2e9c591", "f8578e6b5900de614aabe563c9622a8f514e11d368caa78890ac2ed615a2300c", "1632f2b53febb0a999784c4feb1655144793c4e662226aff64b71c6837430791", "ad4328979dba3e30f11c2d94445731f461a25842523fcbfa53cd42b585e63fcd", "a904a9a41d1c8f9e860ba2b07ba13187b41aa7246f341489a730c6dc6fb42701", "dd7e026ac1fff0feac6bed6872b6964f5ea00bd8913a956e6b2eb7e22363dc5c", "2c3b18d8edff29c013394c28888c6b50ed8733760a3d4d9082c3f1f5a43afa64"},
		"00000002",
		"19015f53",
		"53058b41",
		false,
	}
	job, _ := stratum.NewJob(list)


  extraNonce1 := "4bc6af58"
  extraNonce2 := "00000000"
	ntime := "53058d7b"
	nonce := "e8832204"

	merkleRoot := job.MerkleRoot(extraNonce1, extraNonce2)
	header, _ := stratum.SerializeHeader(job, merkleRoot, ntime, nonce)
	headerHash, _ := header.BlockSha()
	// given little-endian hash string
	expected := "4a77d5d2e3f51ecc8aec8a75d8f157ec2637d44450e3f0949db78dbb21f7ed5a"
	if headerHash.String() != expected {
		t.Errorf("wrong header hash %v", headerHash.String())
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

	list := birpc.List{
		"31",
		"5da479a459762b2082abb7496cea48694c50b7a2bc070457bb7fd0f400000000",
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2703851204062f503253482f0446ecb65308",
		"0d2f6e6f64655374726174756d2f00000000029d6bab7e000000001976a914efc72872187fbb5688001065c5df01ed84e6f25988ac677c5a16000000001976a914aa9eded884f09c5d5844df00093453dda8881b5b88ac00000000",
		birpc.List{"2595fd2d192731df10d131eb77e108449c189e88a6797d7b6027be328476433e","01cd5c94bb486111524040ef65fb23855fd524ea27f508581f521994ac72c44c","6be8f181cce5d752f9b994c1e2dfe371bcce0d602d4a8cd6da9e84c5e87b3db2","2cfa58a93fa1e1c75eda9fedfd49eec22a2af91f6de192e253b86c7df7a5aebf","692a1ab4bba06b055c5e5ec519e7800665b2ab6df86d05ffe36dbf769ada9d68"},
		"00000002",
		"1b013164",
		"53b6ec46",
		false,
	}
	job, _ := stratum.NewJob(list)

	extraNonce1 := "580000020001"
	extraNonce2 := "0000"
	ntime := "53b6ec46"
	nonce := "5ea00f07"

	merkleRoot := job.MerkleRoot(extraNonce1, extraNonce2)
	header, err := stratum.SerializeHeader(job, merkleRoot, ntime, nonce)
	if err != nil {
		t.Errorf("unexpected: %v", err)
		return
	}
	headerHash, _ := header.BlockSha()
	// hash in standard bitcoin big-endian form.
	expHeaderHash := "00000000421a0986cafd57f748e8d33469eb2d0583243b4a299060ac0ebc0b3c"
	// NOMP header hash: 3c0bbc0eac6090294a3b2483052deb6934d3e848f757fdca86091a4200000000
	if headerHash.String() != expHeaderHash {
		t.Errorf("wrong header hash %v", headerHash.String())
	}
}
