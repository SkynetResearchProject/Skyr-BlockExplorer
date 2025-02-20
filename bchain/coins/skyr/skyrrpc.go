// The package is forked from PIVX
package skyr

import (
    "blockbook/bchain"
    "blockbook/bchain/coins/btc"
    "encoding/json"
    "net"
    "net/http"
    "strings"
    "fmt"
    "io/ioutil"
    "time"

    "github.com/golang/glog"
    "github.com/juju/errors"
)

// SkyrRPC is an interface to JSON-RPC bitcoind service.
type SkyrRPC struct {
    *btc.BitcoinRPC
    BitcoinGetChainInfo func() (*bchain.ChainInfo, error)
}

// NewSkyrRPC returns new SkyrRPC instance.
func NewSkyrRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
    b, err := btc.NewBitcoinRPC(config, pushHandler)
    if err != nil {
        return nil, err
    }

    s := &SkyrRPC{
        b.(*btc.BitcoinRPC),
        b.GetChainInfo,
    }
    s.RPCMarshaler = btc.JSONMarshalerV1{}
    s.ChainConfig.SupportsEstimateFee = true
    s.ChainConfig.SupportsEstimateSmartFee = false

    return s, nil
}

// Initialize initializes SkyrRPC instance.
func (b *SkyrRPC) Initialize() error {
    ci, err := b.GetChainInfo()
    if err != nil {
        return err
    }
    chainName := ci.Chain

    glog.Info("Chain name ", chainName)
    params := GetChainParams(chainName)

    // always create parser
    b.Parser = NewSkyrParser(params, b.ChainConfig)

    // parameters for getInfo request
    if params.Net == MainnetMagic {
        b.Testnet = false
        b.Network = "livenet"
    } else {
        b.Testnet = true
        b.Network = "testnet"
    }

    glog.Info("rpc: block chain ", params.Name)

    return nil
}

// GetBlock returns block with given hash.
func (z *SkyrRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
    var err error
    if hash == "" && height > 0 {
        hash, err = z.GetBlockHash(height)
        if err != nil {
            return nil, err
        }
    }

    glog.V(1).Info("rpc: getblock (verbosity=1) ", hash)

    res := btc.ResGetBlockThin{}
    req := btc.CmdGetBlock{Method: "getblock"}
    req.Params.BlockHash = hash
    req.Params.Verbosity = 1
    err = z.Call(&req, &res)

    if err != nil {
        return nil, errors.Annotatef(err, "hash %v", hash)
    }
    if res.Error != nil {
        return nil, errors.Annotatef(res.Error, "hash %v", hash)
    }

    txs := make([]bchain.Tx, 0, len(res.Result.Txids))
    for _, txid := range res.Result.Txids {
        tx, err := z.GetTransaction(txid)
        if err != nil {
            if err == bchain.ErrTxNotFound {
                glog.Errorf("rpc: getblock: skipping transanction in block %s due error: %s", hash, err)
                continue
            }
            return nil, err
        }
        txs = append(txs, *tx)
    }
    block := &bchain.Block{
        BlockHeader: res.Result.BlockHeader,
        Txs:         txs,
    }
    return block, nil
}


// getinfo

type CmdGetInfo struct {
    Method string `json:"method"`
}

type ResGetInfo struct {
    Error  *bchain.RPCError `json:"error"`
    Result struct {
        TransparentSupply   json.Number `json:"transparentsupply"`
        ShieldSupply   json.Number `json:"shieldsupply"`
        MoneySupply   json.Number `json:"moneysupply"`
    } `json:"result"`
}

// getmasternodecount

type CmdGetMasternodeCount struct {
    Method string `json:"method"`
}

type ResGetMasternodeCount struct {
    Error  *bchain.RPCError `json:"error"`
    Result struct {
        Total   int    `json:"total"`
        Stable   int    `json:"stable"`
        Enabled   int    `json:"enabled"`
        InQueue   int    `json:"inqueue"`
    } `json:"result"`
}

// getconnectioncount

type CmdGetConnectionCount struct {
    Method string `json:"method"`
}

type ResGetConnectionCount struct {
    Error  *bchain.RPCError `json:"error"`
    Result  int  `json:"result"`
}

// listmasternodes

type CmdListMasternodes struct {
	Method string `json:"method"`
}

type ResListMasternodes struct {
	Error  *bchain.RPCError `json:"error"`
	Result *bchain.RPCMasternodes `json:"result"`
}

//getpeerinfo
type CmdGetPeerInfo struct {
        Method string `json:"method"`
}

type ResGetPeerInfo struct {
        Error  *bchain.RPCError `json:"error"`
        Result *bchain.RPCPeers `json:"result"`
}

// GetNextSuperBlock returns the next superblock height after nHeight
func (b *SkyrRPC) GetNextSuperBlock(nHeight int) int {
    nBlocksPerPeriod := 43200
    if b.Testnet {
        nBlocksPerPeriod = 144
    }
    return nHeight - nHeight % nBlocksPerPeriod + nBlocksPerPeriod
}

// GetMasternodesInfo
func (b *SkyrRPC) GetMasternodesInfo() (*bchain.RPCMasternodes, error){
    rv, err := b.BitcoinGetChainInfo()
    if err != nil {
        return nil, err
    }

    var bestHeight = rv.Blocks

    glog.V(1).Info("rpc: listmasternodes")

    resMns := ResListMasternodes{}
    err = b.Call(&CmdListMasternodes{Method: "listmasternodes"}, &resMns)
    if err != nil {
        return nil, err
    }
    if resMns.Error != nil {
        return nil, resMns.Error
    }

    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err == nil {
        defer conn.Close()
        //localip_ := "193.233.165.116:16888"  //for debug in docker
        localip_ := conn.LocalAddr().(*net.UDPAddr).String()
        localip := strings.Split(localip_, ":")
        var Mn = *resMns.Result
        for i:=0; i<len(Mn); i++{
            Ip_ := Mn[i].Ip
            Ip := strings.Split(Ip_, ":")
            if localip[0] == Ip[0] || len(Ip[0]) > 12 { //it's fake, will be new rpc-command 'getlastblock "ip"'
                Mn[i].Lastblock = int(bestHeight)
            } else{
                Mn[i].Lastblock = -1
            }
         }
    }
    return resMns.Result, nil
}

// The structures are partially implementeds ---
type Name struct {
    En  string    `json:"en"`
}

type Country struct {
    Geoname_id              int         `json:"geoname_id"`
    Is_in_european_union    bool        `json:"geoname_igeoname_id"`
    Iso_code                string      `json:"iso_code"`
    Names                   Name        `json:"names"`
}

type Location struct {
    Country     Country `json:"country"`
}
// ---

// GetPeersInfo
func (b *SkyrRPC) GetPeersInfo() (*bchain.RPCPeers, error){
    var err error
    glog.V(1).Info("rpc: getpeerinfo")

    resPeers := ResGetPeerInfo{}
    err = b.Call(&CmdGetPeerInfo{Method: "getpeerinfo"}, &resPeers)
    if err != nil {
        return nil, err
    }
    if resPeers.Error != nil {
        return nil, resPeers.Error
    }

    cfg := *b.ChainConfig
    if len(cfg.Geolocation_url) > 0 {
        var Peers = *resPeers.Result
        for i:=0; i<len(Peers); i++ {
           ip := strings.Split(Peers[i].Addr, ":")
           client := http.Client{Timeout: 5 * time.Second,}
           url := fmt.Sprintf(cfg.Geolocation_url, ip[0])
           resp, err := client.Get(url)
           defer resp.Body.Close()
           if err == nil {
               var loc Location
               // read json http response
               jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
               if err == nil {
                    err = json.Unmarshal([]byte(jsonDataFromHttp), &loc)
                    if err == nil {
                         Peers[i].Location = loc.Country.Names.En
                    }
               }
           }
        }
    }
    return resPeers.Result, nil
}

// GetChainInfo returns information about the connected backend
// SKYR adds Money Supply to btc implementation
func (b *SkyrRPC) GetChainInfo() (*bchain.ChainInfo, error) {
    rv, err := b.BitcoinGetChainInfo()
    if err != nil {
        return nil, err
    }
    glog.V(1).Info("rpc: getinfo")

    resGi := ResGetInfo{}
    err = b.Call(&CmdGetInfo{Method: "getinfo"}, &resGi)
    if err != nil {
        return nil, err
    }
    if resGi.Error != nil {
        return nil, resGi.Error
    }
    rv.TransparentSupply = resGi.Result.TransparentSupply
        rv.ShieldSupply = resGi.Result.ShieldSupply
        rv.MoneySupply = resGi.Result.MoneySupply

    glog.V(1).Info("rpc: getmasternodecount")

    resMc := ResGetMasternodeCount{}
    err = b.Call(&CmdGetMasternodeCount{Method: "getmasternodecount"}, &resMc)
    if err != nil {
        return nil, err
    }
    if resMc.Error != nil {
        return nil, resMc.Error
    }
    rv.MasternodeCount = resMc.Result.Enabled

    glog.V(1).Info("rpc: getconnectioncount")

    resCc := ResGetConnectionCount{}
    err = b.Call(&CmdGetConnectionCount{Method: "getconnectioncount"}, &resCc)
    if err != nil {
        return nil, err
    }
    if resCc.Error != nil {
        return nil, resCc.Error
    }
    rv.ConnectionCount = resCc.Result

    rv.NextSuperBlock = b.GetNextSuperBlock(rv.Headers)

    return rv, nil
}

// findserial
type CmdFindSerial struct {
    Method string   `json:"method"`
    Params []string `json:"params"`
}

type ResFindSerial struct {
    Error  *bchain.RPCError `json:"error"`
    Result struct {
        Success bool      `json:"success"`
        Txid    string    `json:"txid"`
    } `json:"result"`
}

func (b *SkyrRPC) Findzcserial(serialHex string) (string, error) {
    glog.V(1).Info("rpc: findserial")

    res := ResFindSerial{}
    req := CmdFindSerial{Method: "findserial"}
    req.Params = []string{serialHex}
    err := b.Call(&req, &res)

    if err != nil {
        return "", err
    }
    if res.Error != nil {
        return "", res.Error
    }
    if !res.Result.Success {
        return "Serial not found in blockchain", nil
    }
    return res.Result.Txid, nil
}
