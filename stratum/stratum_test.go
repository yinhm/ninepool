package stratum_test

import (
	"bytes"
	"encoding/hex"
	"github.com/conformal/btcchain"
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

func TestMerkleRoot(t *testing.T) {
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

	txList := make([]*btcwire.ShaHash, 0, len(txHashes))
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

	txHashes, _ := stratum.MerkleHashesFromList(list[4])
	merkleRoot := stratum.BuildMerkleRoot(txHashes)
	expected, _ := btcwire.NewShaHashFromStr("023b0945b83c971237afb78b79fafe60a961a1833c431b2375576fe6fc80b63f")
	if !expected.IsEqual(merkleRoot) {
		t.Errorf("Merkle root hash does not match.")
	}

	ntime := "53058d7b"
	nonce := "e8832204"

	header, _ := stratum.SerializeHeader(job, merkleRoot, ntime, nonce)
	var buf bytes.Buffer
	_ = header.Serialize(&buf)
	blockHeaderLen := 80
	headerStr := hex.EncodeToString(buf.Bytes()[0:blockHeaderLen])
	expectedHeader := "0200000000000000010000007803c817330edab897599b55e255adf2c18ed1f717975b973fb680fce66f5775231b433c83a161a960fefa798bb7af3712973cb845093b027b8d0553535f0119042283e8"
	if headerStr != expectedHeader {
		t.Errorf("Not expected header: %s", headerStr)
	}

	headerHash, _ := header.BlockSha()
	// given little-endian hash string
	buf2, _ := hex.DecodeString("5d495a2f92a67ac6df4b2f84c7ee76df7bc0633d57394dd8b9c2253f420ddef6")
	expectedHash, _ := btcwire.NewShaHash(buf2)
	if !headerHash.IsEqual(expectedHash) {
		t.Errorf("wrong header hash %v", headerHash.String())
	}

	// diff1 := 0x00000000FFFF0000000000000000000000000000000000000000000000000000
	compact := uint32(0x1d00ffff)
	diff1 := stratum.CompactToBig(compact)
	hash, _ := btcwire.NewShaHashFromStr("00000000FFFF0000000000000000000000000000000000000000000000000000")
	diff2 := btcchain.ShaHashToBig(hash)

	if diff1.Cmp(diff2) != 0 {
		t.Errorf("d1 != d2: %v, %v", diff1, diff2)
	}

	shareDiff := new(big.Int).Div(diff1, stratum.ShaHashToBig(&headerHash))
	if shareDiff.Cmp(big.NewInt(int64(1))) > 0 {
		t.Errorf("share diff >1")
	}
}
