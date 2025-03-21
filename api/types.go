package api

import (
    "blockbook/bchain"
    "blockbook/common"
    "blockbook/db"
    "encoding/json"
    "errors"
    "math/big"
    "time"
)

const maxUint32 = ^uint32(0)
const maxInt = int(^uint(0) >> 1)
const maxInt64 = int64(^uint64(0) >> 1)

// AccountDetails specifies what data returns GetAddress and GetXpub calls
type AccountDetails int

const (
    // AccountDetailsBasic - only that address is indexed and some basic info
    AccountDetailsBasic AccountDetails = iota
    // AccountDetailsTokens - basic info + tokens
    AccountDetailsTokens
    // AccountDetailsTokenBalances - basic info + token with balance
    AccountDetailsTokenBalances
    // AccountDetailsTxidHistory - basic + token balances + txids, subject to paging
    AccountDetailsTxidHistory
    // AccountDetailsTxHistoryLight - basic + tokens + easily obtained tx data (not requiring requests to backend), subject to paging
    AccountDetailsTxHistoryLight
    // AccountDetailsTxHistory - basic + tokens + full tx data, subject to paging
    AccountDetailsTxHistory
)

// ErrUnsupportedXpub is returned when coin type does not support xpub address derivation or provided string is not an xpub
var ErrUnsupportedXpub = errors.New("XPUB not supported")

// APIError extends error by information if the error details should be returned to the end user
type APIError struct {
    Text   string
    Public bool
}

func (e *APIError) Error() string {
    return e.Text
}

// NewAPIError creates ApiError
func NewAPIError(s string, public bool) error {
    return &APIError{
        Text:   s,
        Public: public,
    }
}

// Amount is datatype holding amounts
type Amount big.Int

// IsZeroBigInt if big int has zero value
func IsZeroBigInt(b *big.Int) bool {
    return len(b.Bits()) == 0
}

// MarshalJSON Amount serialization
func (a *Amount) MarshalJSON() (out []byte, err error) {
    if a == nil {
        return []byte(`"0"`), nil
    }
    return []byte(`"` + (*big.Int)(a).String() + `"`), nil
}

func (a *Amount) String() string {
    if a == nil {
        return ""
    }
    return (*big.Int)(a).String()
}

// DecimalString returns amount with decimal point placed according to parameter d
func (a *Amount) DecimalString(d int) string {
    return bchain.AmountToDecimalString((*big.Int)(a), d)
}

// AsBigInt returns big.Int type for the Amount (empty if Amount is nil)
func (a *Amount) AsBigInt() big.Int {
    if a == nil {
        return *new(big.Int)
    }
    return big.Int(*a)
}

// AsInt64 returns Amount as int64 (0 if Amount is nil).
// It is used only for legacy interfaces (socket.io)
// and generally not recommended to use for possible loss of precision.
func (a *Amount) AsInt64() int64 {
    if a == nil {
        return 0
    }
    return (*big.Int)(a).Int64()
}

// Vin contains information about single transaction input
type Vin struct {
    Txid       string                   `json:"txid,omitempty"`
    Vout       uint32                   `json:"vout,omitempty"`
    Sequence   int64                    `json:"sequence,omitempty"`
    N          int                      `json:"n"`
    AddrDesc   bchain.AddressDescriptor `json:"-"`
    Addresses  []string                 `json:"addresses,omitempty"`
    Searchable bool                     `json:"-"`
    ValueSat   *Amount                  `json:"value,omitempty"`
    Hex        string                   `json:"hex,omitempty"`
    Asm        string                   `json:"asm,omitempty"`
    Coinbase   string                   `json:"coinbase,omitempty"`
}

// Vout contains information about single transaction output
type Vout struct {
    ValueSat    *Amount                  `json:"value,omitempty"`
    N           int                      `json:"n"`
    Spent       bool                     `json:"spent,omitempty"`
    SpentTxID   string                   `json:"spentTxId,omitempty"`
    SpentIndex  int                      `json:"spentIndex,omitempty"`
    SpentHeight int                      `json:"spentHeight,omitempty"`
    Hex         string                   `json:"hex,omitempty"`
    Asm         string                   `json:"asm,omitempty"`
    AddrDesc    bchain.AddressDescriptor `json:"-"`
    Addresses   []string                 `json:"addresses"`
    Searchable  bool                     `json:"-"`
    Type        string                   `json:"type,omitempty"`
}

// TokenType specifies type of token
type TokenType string

// ERC20TokenType is Ethereum ERC20 token
const ERC20TokenType TokenType = "ERC20"

// XPUBAddressTokenType is address derived from xpub
const XPUBAddressTokenType TokenType = "XPUBAddress"

// Token contains info about tokens held by an address
type Token struct {
    Type             TokenType `json:"type"`
    Name             string    `json:"name"`
    Path             string    `json:"path,omitempty"`
    Contract         string    `json:"contract,omitempty"`
    Transfers        int       `json:"transfers"`
    Symbol           string    `json:"symbol,omitempty"`
    Decimals         int       `json:"decimals,omitempty"`
    BalanceSat       *Amount   `json:"balance,omitempty"`
    TotalReceivedSat *Amount   `json:"totalReceived,omitempty"`
    TotalSentSat     *Amount   `json:"totalSent,omitempty"`
    ContractIndex    string    `json:"-"`
}

// TokenTransfer contains info about a token transfer done in a transaction
type TokenTransfer struct {
    Type     TokenType `json:"type"`
    From     string    `json:"from"`
    To       string    `json:"to"`
    Token    string    `json:"token"`
    Name     string    `json:"name"`
    Symbol   string    `json:"symbol"`
    Decimals int       `json:"decimals"`
    Value    *Amount   `json:"value"`
}

// EthereumSpecific contains ethereum specific transaction data
type EthereumSpecific struct {
    Status   int      `json:"status"` // 1 OK, 0 Fail, -1 pending
    Nonce    uint64   `json:"nonce"`
    GasLimit *big.Int `json:"gasLimit"`
    GasUsed  *big.Int `json:"gasUsed"`
    GasPrice *Amount  `json:"gasPrice"`
}

// Tx holds information about a transaction
type Tx struct {
    Txid             string            `json:"txid"`
    Version          int32             `json:"version,omitempty"`
    Locktime         uint32            `json:"lockTime,omitempty"`
    Vin              []Vin             `json:"vin"`
    Vout             []Vout            `json:"vout"`
    Blockhash        string            `json:"blockHash,omitempty"`
    Blockheight      int               `json:"blockHeight"`
    Confirmations    uint32            `json:"confirmations"`
    Blocktime        int64             `json:"blockTime"`
    Size             int               `json:"size,omitempty"`
    ValueOutSat      *Amount           `json:"value"`
    ValueInSat       *Amount           `json:"valueIn,omitempty"`
    ShieldIns        int               `json:"shieldIns,omitempty"`
    ShieldOuts       int               `json:"shielOuts,omitempty"`
    ShieldValBal     *Amount           `json:"valueBalanceSat,omitempty"`
    FeesSat          *Amount           `json:"fees,omitempty"`
    Hex              string            `json:"hex,omitempty"`
    CoinSpecificData interface{}       `json:"-"`
    CoinSpecificJSON json.RawMessage   `json:"-"`
    TokenTransfers   []TokenTransfer   `json:"tokenTransfers,omitempty"`
    EthereumSpecific *EthereumSpecific `json:"ethereumSpecific,omitempty"`
}

// Paging contains information about paging for address, blocks and block
type Paging struct {
    Page        int `json:"page,omitempty"`
    TotalPages  int `json:"totalPages,omitempty"`
    ItemsOnPage int `json:"itemsOnPage,omitempty"`
}

// TokensToReturn specifies what tokens are returned by GetAddress and GetXpubAddress
type TokensToReturn int

const (
    // AddressFilterVoutOff disables filtering of transactions by vout
    AddressFilterVoutOff = -1
    // AddressFilterVoutInputs specifies that only txs where the address is as input are returned
    AddressFilterVoutInputs = -2
    // AddressFilterVoutOutputs specifies that only txs where the address is as output are returned
    AddressFilterVoutOutputs = -3

    // TokensToReturnNonzeroBalance - return only tokens with nonzero balance
    TokensToReturnNonzeroBalance TokensToReturn = 0
    // TokensToReturnUsed - return tokens with some transfers (even if they have zero balance now)
    TokensToReturnUsed TokensToReturn = 1
    // TokensToReturnDerived - return all derived tokens
    TokensToReturnDerived TokensToReturn = 2
)

// AddressFilter is used to filter data returned from GetAddress api method
type AddressFilter struct {
    Vout           int
    Contract       string
    FromHeight     uint32
    ToHeight       uint32
    TokensToReturn TokensToReturn
    // OnlyConfirmed set to true will ignore mempool transactions; mempool is also ignored if FromHeight/ToHeight filter is specified
    OnlyConfirmed bool
}

// Address holds information about address and its transactions
type Address struct {
    Paging
    AddrStr               string                `json:"address"`
    BalanceSat            *Amount               `json:"balance"`
    TotalReceivedSat      *Amount               `json:"totalReceived,omitempty"`
    TotalSentSat          *Amount               `json:"totalSent,omitempty"`
    UnconfirmedBalanceSat *Amount               `json:"unconfirmedBalance"`
    UnconfirmedTxs        int                   `json:"unconfirmedTxs"`
    Txs                   int                   `json:"txs"`
    NonTokenTxs           int                   `json:"nonTokenTxs,omitempty"`
    Transactions          []*Tx                 `json:"transactions,omitempty"`
    Txids                 []string              `json:"txids,omitempty"`
    Nonce                 string                `json:"nonce,omitempty"`
    UsedTokens            int                   `json:"usedTokens,omitempty"`
    Tokens                []Token               `json:"tokens,omitempty"`
    Erc20Contract         *bchain.Erc20Contract `json:"erc20Contract,omitempty"`
    // helpers for explorer
    Filter        string              `json:"-"`
    XPubAddresses map[string]struct{} `json:"-"`
}

// Utxo is one unspent transaction output
type Utxo struct {
    Txid          string  `json:"txid"`
    Vout          int32   `json:"vout"`
    AmountSat     *Amount `json:"value"`
    Height        int     `json:"height,omitempty"`
    Confirmations int     `json:"confirmations"`
    Address       string  `json:"address,omitempty"`
    Path          string  `json:"path,omitempty"`
    Locktime      uint32  `json:"lockTime,omitempty"`
}

// Utxos is array of Utxo
type Utxos []Utxo

func (a Utxos) Len() int      { return len(a) }
func (a Utxos) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Utxos) Less(i, j int) bool {
    // sort in reverse order, unconfirmed (height==0) utxos on top
    hi := a[i].Height
    hj := a[j].Height
    if hi == 0 {
        hi = maxInt
    }
    if hj == 0 {
        hj = maxInt
    }
    return hi >= hj
}

// Blocks is list of blocks with paging information
type Blocks struct {
    Paging
    Blocks []db.BlockInfo `json:"blocks"`
}

// BlockInfo contains extended block header data and a list of block txids
type BlockInfo struct {
    Hash          string      `json:"hash"`
    Prev          string      `json:"previousBlockHash,omitempty"`
    Next          string      `json:"nextBlockHash,omitempty"`
    Height        uint32      `json:"height"`
    Confirmations int         `json:"confirmations"`
    Size          int         `json:"size"`
    Time          int64       `json:"time,omitempty"`
    Version       json.Number `json:"version"`
    MerkleRoot    string      `json:"merkleRoot"`
    Nonce         string      `json:"nonce"`
    Bits          string      `json:"bits"`
    Difficulty    string      `json:"difficulty"`
    Txids         []string    `json:"tx,omitempty"`
    SaplingRoot   string      `json:"finalsaplingroot,omitempty"`
}

// Block contains information about block
type Block struct {
    Paging
    BlockInfo
    TxCount      int   `json:"txCount"`
    Transactions []*Tx `json:"txs,omitempty"`
}

// BlockbookInfo contains information about the running blockbook instance
type BlockbookInfo struct {
    Coin              string                       `json:"coin"`
    Host              string                       `json:"host"`
    Version           string                       `json:"version"`
    GitCommit         string                       `json:"gitCommit"`
    BuildTime         string                       `json:"buildTime"`
    SyncMode          bool                         `json:"syncMode"`
    InitialSync       bool                         `json:"initialSync"`
    InSync            bool                         `json:"inSync"`
    BestHeight        uint32                       `json:"bestHeight"`
    LastBlockTime     time.Time                    `json:"lastBlockTime"`
    InSyncMempool     bool                         `json:"inSyncMempool"`
    LastMempoolTime   time.Time                    `json:"lastMempoolTime"`
    MempoolSize       int                          `json:"mempoolSize"`
    Decimals          int                          `json:"decimals"`
    DbSize            int64                        `json:"dbSize"`
    DbSizeFromColumns int64                        `json:"dbSizeFromColumns,omitempty"`
    DbColumns         []common.InternalStateColumn `json:"dbColumns,omitempty"`
    About             string                       `json:"about"`
}

// BackendInfo is used to get information about blockchain
type BackendInfo struct {
    Chain              string       `json:"chain"`
    Blocks             int          `json:"blocks"`
    Headers            int          `json:"headers"`
    Bestblockhash      string       `json:"bestBlockHash"`
    Difficulty         string       `json:"difficulty"`
    SizeOnDisk         int64        `json:"sizeOnDisk"`
    Version            string       `json:"version"`
    Subversion         string       `json:"subversion"`
    ProtocolVersion    string       `json:"protocolVersion"`
    Timeoffset         float64      `json:"timeOffset"`
    Warnings           string       `json:"warnings"`
    TransparentSupply  json.Number  `json:"transparentsupply"`
    ShieldSupply       json.Number  `json:"shieldsupply"`
    MoneySupply        json.Number  `json:"moneysupply"`
    MasternodeCount    int          `json:"masternodecount"`
    ConnectionCount    int          `json:"connectioncount"`
    NextSuperBlock     int          `json:"masternodecount"`
}

// SystemInfo contains information about the running blockbook and backend instance
type SystemInfo struct {
    Blockbook *BlockbookInfo `json:"blockbook"`
    Backend   *BackendInfo   `json:"backend"`
}

type MasternodesInfo struct{
   Masternodes        *bchain.RPCMasternodes  `json:"masternodes"`
}

type Version struct{
   Subver string
   Cnt      int
}

type Service struct{
   Service string
   Cnt      int
}

type PeersInfo struct{
   Peers              *bchain.RPCPeers   `json:"peers"`
   Version            *[]Version           `json:"resv"`
   Services           *[]Service           `json:"ress"`
}

type TopInfo struct {
    Tops              *[]db.Top             `json:"tops"`
    WealthDists       *[]db.WealthDist      `json:"wealthdist"`
}

// MempoolTxid contains information about a transaction in mempool
type MempoolTxid struct {
    Time int64  `json:"time"`
    Txid string `json:"txid"`
}

// MempoolTxids contains a list of mempool txids with paging information
type MempoolTxids struct {
    Paging
    Mempool     []MempoolTxid `json:"mempool"`
    MempoolSize int           `json:"mempoolSize"`
}
