// build unittest

package skyr

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"reflect"
	"testing"
)

type testBlock struct {
	size int
	time int64
	txs  []string
}

var testParseBlockTxs = map[int]testBlock{
	1224551: {
		size: 452,
		time: 1723452660,
		txs: []string{
			"e1c275c20cff9f34a39e9c61fa89b1b3de351f60e7b89e1bf844095077f73527",
                        "9a4f9c488f21dbf9fc6d2032588987392370e34075409e3a080394f562335fd3",
		},
	},
	// block with cold staking
	1405103: {
		size: 481,
		time: 1734608640,
		txs: []string{
			"027904d694898c430324b56845cfcdaa529a717316eeefca3d489787cf115ba4",
                        "7a4e30bd02e763230015ef14fa3d86981d4754dabd792da45767a4981f00d7ee",
		},
	},
}

func helperLoadBlock(t *testing.T, height int) []byte {
	name := fmt.Sprintf("block_dump.%d", height)
	path := filepath.Join("testdata", name)

	d, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	d = bytes.TrimSpace(d)

	b := make([]byte, hex.DecodedLen(len(d)))
	_, err = hex.Decode(b, d)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestParseBlock(t *testing.T) {
	p := NewSkyrParser(GetChainParams("main"), &btc.Configuration{})

	for height, tb := range testParseBlockTxs {
		b := helperLoadBlock(t, height)

		blk, err := p.ParseBlock(b)
		if err != nil {
			t.Errorf("ParseBlock() error %v", err)
		}

		if blk.Size != tb.size {
			t.Errorf("ParseBlock() block size: got %d, want %d", blk.Size, tb.size)
		}

		if blk.Time != tb.time {
			t.Errorf("ParseBlock() block time: got %d, want %d", blk.Time, tb.time)
		}

		if len(blk.Txs) != len(tb.txs) {
			t.Errorf("ParseBlock() number of transactions: got %d, want %d", len(blk.Txs), len(tb.txs))
		}

		for ti, tx := range tb.txs {
			if blk.Txs[ti].Txid != tx {
				t.Errorf("ParseBlock() transaction %d: got %s, want %s", ti, blk.Txs[ti].Txid, tx)
			}
		}
	}
}

func Test_GetAddrDescFromAddress_Mainnet(t *testing.T) {
    type args struct {
        address string
    }
    tests := []struct {
        name    string
        args    args
        want    string
        wantErr bool
    }{
        {
            name:    "P2PKH1",
            args:    args{address: "BQf7Y3E6VEXECRzCro9YNc6vWHJTnhGciX"},
            want:    "76a914dda91c0396050d660f9c0e38f78064486bbfcb2c88ac",
            wantErr: false,
        },
    }
    parser := NewSkyrParser(GetChainParams("main"), &btc.Configuration{})

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.GetAddrDescFromAddress(tt.args.address)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetAddrDescFromAddress() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            h := hex.EncodeToString(got)
            if !reflect.DeepEqual(h, tt.want) {
                t.Errorf("GetAddrDescFromAddress() = %v, want %v", h, tt.want)
            }
        })
    }
}

func Test_GetAddressesFromAddrDesc(t *testing.T) {
    type args struct {
        script string
    }

    tnet_tests := []struct {
        name    string
        args    args
        want    []string
        want2   bool
        wantErr bool
    }{
        {
            name:    "P2CS",
            args:    args{script: "76a97b63d11488d337f2229fb5397180e50fd002c5572479fda56714630044871a3cec9cf469e7ccc7585736fbb70a5c6888ac"},
            want:    []string{"Wb9VojknMG5FStXrcaHuBhufxUZDPThmkH", "mpYRWeL569Fm2r5F1WMgAc6VQ1EJoKT4Ds"},
            want2:   true,
            wantErr: false,
        },
        {
            name:    "P2PKH1",
            args:    args{script: "76a914dda91c0396050d660f9c0e38f78064486bbfcb2c88ac"},
            want:    []string{"n1izDNrsYkNaXhyexsnx5PWadfTavxXTVR"},
            want2:   true,
            wantErr: false,
        },
        {
            name:    "pubkey",
            args:    args{script: "210251c5555ff3c684aebfca92f5329e2f660da54856299da067060a1bcf5e8fae73ac"},
            want:    []string{"muhuAnLvpTPG4XcBGuDf9xpbCZcoykEvnW"},
            want2:   false,
            wantErr: false,
        },
    }

    tests := []struct {
        name    string
        args    args
        want    []string
        want2   bool
        wantErr bool
    }{
        {
            name:    "P2CS",
            args:    args{script: "76a97b63d11488d337f2229fb5397180e50fd002c5572479fda56714630044871a3cec9cf469e7ccc7585736fbb70a5c6888ac"},
            want:    []string{"SZmTxemuFTSVGZ8zNNxiLTBofRynDcH6Bj", "BDUYqJhJ2dQQha5nuRiGTpgqGd5BfTJNah"},
            want2:   true,
            wantErr: false,
        },
        {
            name:    "P2PKH1",
            args:    args{script: "76a914dda91c0396050d660f9c0e38f78064486bbfcb2c88ac"},
            want:    []string{"BQf7Y3E6VEXECRzCro9YNc6vWHJTnhGciX"},
            want2:   true,
            wantErr: false,
        },
        {
            name:    "pubkey",
            args:    args{script: "210251c5555ff3c684aebfca92f5329e2f660da54856299da067060a1bcf5e8fae73ac"},
            want:    []string{"BJe2VSi9kwXujFcjApaFTBQw5BTgnfF4C5"},
            want2:   false,
            wantErr: false,
        },
    }

    parser := NewSkyrParser(GetChainParams("main"), &btc.Configuration{})

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b, _ := hex.DecodeString(tt.args.script)
            got, got2, err := parser.GetAddressesFromAddrDesc(b)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetAddressesFromAddrDesc() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got, tt.want)
            }
            if !reflect.DeepEqual(got2, tt.want2) {
                t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got2, tt.want2)
            }
        })
    }

    tnet_parser := NewSkyrParser(GetChainParams("test"), &btc.Configuration{})

    for _, tt := range tnet_tests {
        t.Run(tt.name, func(t *testing.T) {
            b, _ := hex.DecodeString(tt.args.script)
            got, got2, err := tnet_parser.GetAddressesFromAddrDesc(b)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetAddressesFromAddrDesc() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got, tt.want)
            }
            if !reflect.DeepEqual(got2, tt.want2) {
                t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got2, tt.want2)
            }
        })
    }
}

var (
	testTx1 = bchain.Tx{
		Blocktime:     1738312335,
		//Confirmations: 26292, // not present in gettransaction output
		Hex:           "010000000137791a8820d3eecbc911fcdd832529b3c6959d5e65dfd8590c3302c572dc82b7000000006b483045022100fac5aa73f2a6c2e50e51341054310fe59401db2422d30fabdbbe3259bfe40f8002207eacb731c14687685a62718c214147634ee711d739600683f6ab2a7972e524b8012103eb9cbe09f80ec5789a0c61f37cc3c6c5cf7b41aa1197aaa4552106310a28b4d4ffffffff022cd5024cf60e00001976a914ea8c651f229a18c13ec9e9f74aaedfcf986497fd88ac00d4b2b49c2700001976a9147756e29ac9c9c631e459b3d60796934e59b8e2aa88ac00000000",
		LockTime:      0,
		Time:          1738312335,
		Txid:          "adb5cbc5bfd7d662544e7ef9f2f1a1adab06551df524209288dea72b2022fb04",
		Version:       1,
		Vin: []bchain.Vin{
			{
				Txid: "b782dc72c502330c59d8df655e9d95c6b3292583ddfc11c9cbeed320881a7937",
				Vout: 0,
				ScriptSig: bchain.ScriptSig{
					Hex: "483045022100fac5aa73f2a6c2e50e51341054310fe59401db2422d30fabdbbe3259bfe40f8002207eacb731c14687685a62718c214147634ee711d739600683f6ab2a7972e524b8012103eb9cbe09f80ec5789a0c61f37cc3c6c5cf7b41aa1197aaa4552106310a28b4d4",
				},
				Sequence: 4294967295,
			},
		},
		Vout: []bchain.Vout{
			{
				N: 0,
				ScriptPubKey: bchain.ScriptPubKey{
					Addresses: []string{"BRqFvRJi8g8e4VFmAa7motGmF7wJjaZF8c"},
					Hex:       "76a914ea8c651f229a18c13ec9e9f74aaedfcf986497fd88ac",
				},
				ValueSat: *big.NewInt(1645099999774),
			},
			{
				N: 1,
				ScriptPubKey: bchain.ScriptPubKey{
					Addresses: []string{"BFL67RxNyxMXhPB8WXiWM4Y6CEZrmEJEjr"},
					Hex:       "76a9147756e29ac9c9c631e459b3d60796934e59b8e2aa88ac",
				},
				ValueSat: *big.NewInt(435540),
			},
		},
	}
	testTxPacked1 = "0a20adb5cbc5bfd7d662544e7ef9f2f1a1adab06551df524209288dea72b2022fb0412e201010000000137791a8820d3eecbc911fcdd832529b3c6959d5e65dfd8590c3302c572dc82b7000000006b483045022100fac5aa73f2a6c2e50e51341054310fe59401db2422d30fabdbbe3259bfe40f8002207eacb731c14687685a62718c214147634ee711d739600683f6ab2a7972e524b8012103eb9cbe09f80ec5789a0c61f37cc3c6c5cf7b41aa1197aaa4552106310a28b4d4ffffffff022cd5024cf60e00001976a914ea8c651f229a18c13ec9e9f74aaedfcf986497fd88ac00d4b2b49c2700001976a9147756e29ac9c9c631e459b3d60796934e59b8e2aa88ac00000000188f95f2bc06200028d3b7593299010a001220b782dc72c502330c59d8df655e9d95c6b3292583ddfc11c9cbeed320881a79371800226b483045022100fac5aa73f2a6c2e50e51341054310fe59401db2422d30fabdbbe3259bfe40f8002207eacb731c14687685a62718c214147634ee711d739600683f6ab2a7972e524b8012103eb9cbe09f80ec5789a0c61f37cc3c6c5cf7b41aa1197aaa4552106310a28b4d428ffffffff0f3a490a06017f0799e21e10001a1976a914ea8c651f229a18c13ec9e9f74aaedfcf986497fd88ac22224252714676524a69386738653456466d4161376d6f74476d4637774a6a615a4638633a460a0306a55410011a1976a9147756e29ac9c9c631e459b3d60796934e59b8e2aa88ac222242464c363752784e79784d5868504238575869574d34593643455a726d454a456a724001"
)

func Test_PackTx(t *testing.T) {
	type args struct {
		tx        bchain.Tx
		height    uint32
		blockTime int64
		parser    *SkyrParser
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "skyr-1",
			args: args{
				tx:        testTx1,
				height:    1465299,
				blockTime: 1738312335,
				parser:    NewSkyrParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    testTxPacked1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.parser.PackTx(&tt.args.tx, tt.args.height, tt.args.blockTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("packTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			h := hex.EncodeToString(got)
			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("packTx() = %v, want %v", h, tt.want)
			}
		})
	}
}

func Test_UnpackTx(t *testing.T) {
	type args struct {
		packedTx string
		parser   *SkyrParser
	}
	tests := []struct {
		name    string
		args    args
		want    *bchain.Tx
		want1   uint32
		wantErr bool
	}{
		{
			name: "skyr-1",
			args: args{
				packedTx: testTxPacked1,
				parser:   NewSkyrParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    &testTx1,
			want1:   1465299,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := hex.DecodeString(tt.args.packedTx)
			got, got1, err := tt.args.parser.UnpackTx(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unpackTx() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("unpackTx() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
