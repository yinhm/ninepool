package stratum_test

import (
	"github.com/conformal/btcwire"
	"github.com/yinhm/ninepool/stratum"
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

func TestDecodeNtime(t *testing.T) {
	ntime, err := stratum.DecodeNtime("504e86ed")
	if err != nil || ntime != int64(1347323629) {
		t.Errorf("failed on parse ntime")
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
