package chaincfg

import (
	"math/big"
	"time"

	"realy.lol/ec/chainhash"
	"realy.lol/ec/wire"
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowLimit is the highest proof of work value a Bitcoin block can
	// have for the main network.  It is the value 2^224 - 1.
	mainPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 224), bigOne)
)

// Constants that define the deployment offset in the deployments field of the
// parameters for each deployment.  This is useful to be able to get the details
// of a specific deployment by name.
const (
	// DeploymentTestDummy defines the rule change deployment ID for testing
	// purposes.
	DeploymentTestDummy = iota

	// DeploymentTestDummyMinActivation defines the rule change deployment
	// ID for testing purposes. This differs from the DeploymentTestDummy
	// in that it specifies the newer params the taproot fork used for
	// activation: a custom threshold and a min activation height.
	DeploymentTestDummyMinActivation

	// DeploymentCSV defines the rule change deployment ID for the CSV
	// soft-fork package. The CSV package includes the deployment of BIPS
	// 68, 112, and 113.
	DeploymentCSV

	// DeploymentSegwit defines the rule change deployment ID for the
	// Segregated Witness (segwit) soft-fork package. The segwit package
	// includes the deployment of BIPS 141, 142, 144, 145, 147 and 173.
	DeploymentSegwit

	// DeploymentTaproot defines the rule change deployment ID for the
	// Taproot (+Schnorr) soft-fork package. The taproot package includes
	// the deployment of BIPS 340, 341 and 342.
	DeploymentTaproot

	// NOTE: DefinedDeployments must always come last since it is used to
	// determine how many defined deployments there currently are.

	// DefinedDeployments is the number of currently defined deployments.
	DefinedDeployments
)

// ConsensusDeployment defines details related to a specific consensus rule
// change that is voted in.  This is part of BIP0009.
type ConsensusDeployment struct {
	// BitNumber defines the specific bit number within the block version
	// this particular soft-fork deployment refers to.
	BitNumber uint8

	// MinActivationHeight is an optional field that when set (default
	// value being zero), modifies the traditional BIP 9 state machine by
	// only transitioning from LockedIn to Active once the block height is
	// greater than (or equal to) thus specified height.
	MinActivationHeight uint32

	// CustomActivationThreshold if set (non-zero), will _override_ the
	// existing RuleChangeActivationThreshold value set at the
	// network/chain level. This value divided by the active
	// MinerConfirmationWindow denotes the threshold required for
	// activation. A value of 1815 block denotes a 90% threshold.
	CustomActivationThreshold uint32

	// DeploymentStarter is used to determine if the given
	// ConsensusDeployment has started or not.
	DeploymentStarter ConsensusDeploymentStarter

	// DeploymentEnder is used to determine if the given
	// ConsensusDeployment has ended or not.
	DeploymentEnder ConsensusDeploymentEnder
}

// Checkpoint identifies a known good point in the block chain.  Using
// checkpoints allows a few optimizations for old blocks during initial download
// and also prevents forks from old blocks.
//
// Each checkpoint is selected based upon several factors.  See the
// documentation for blockchain.IsCheckpointCandidate for details on the
// selection criteria.
type Checkpoint struct {
	Height int32
	Hash   *chainhash.Hash
}

// DNSSeed identifies a DNS seed.
type DNSSeed struct {
	// Host defines the hostname of the seed.
	Host string

	// HasFiltering defines whether the seed supports filtering
	// by service flags (wire.ServiceFlag).
	HasFiltering bool
}

// Params defines a Bitcoin network by its parameters.  These parameters may be
// used by Bitcoin applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net wire.BitcoinNet

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []DNSSeed

	// GenesisBlock defines the first block of the chain.
	GenesisBlock *wire.MsgBlock

	// GenesisHash is the starting block hash.
	GenesisHash *chainhash.Hash

	// PowLimit defines the highest allowed proof of work value for a block
	// as a uint256.
	PowLimit *big.Int

	// PowLimitBits defines the highest allowed proof of work value for a
	// block in compact form.
	PowLimitBits uint32

	// PoWNoRetargeting defines whether the network has difficulty
	// retargeting enabled or not. This should only be set to true for
	// regtest like networks.
	PoWNoRetargeting bool

	// These fields define the block heights at which the specified softfork
	// BIP became active.
	BIP0034Height int32
	BIP0065Height int32
	BIP0066Height int32

	// CoinbaseMaturity is the number of blocks required before newly mined
	// coins (coinbase transactions) can be spent.
	CoinbaseMaturity uint16

	// SubsidyReductionInterval is the interval of blocks before the subsidy
	// is reduced.
	SubsidyReductionInterval int32

	// TargetTimespan is the desired amount of time that should elapse
	// before the block difficulty requirement is examined to determine how
	// it should be changed in order to maintain the desired block
	// generation rate.
	TargetTimespan time.Duration

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// RetargetAdjustmentFactor is the adjustment factor used to limit
	// the minimum and maximum amount of adjustment that can occur between
	// difficulty retargets.
	RetargetAdjustmentFactor int64

	// ReduceMinDifficulty defines whether the network should reduce the
	// minimum required difficulty after a long enough period of time has
	// passed without finding a block.  This is really only useful for test
	// networks and should not be set on a main network.
	ReduceMinDifficulty bool

	// MinDiffReductionTime is the amount of time after which the minimum
	// required difficulty should be reduced when a block hasn't been found.
	//
	// NOTE: This only applies if ReduceMinDifficulty is true.
	MinDiffReductionTime time.Duration

	// GenerateSupported specifies whether or not CPU mining is allowed.
	GenerateSupported bool

	// Checkpoints ordered from oldest to newest.
	Checkpoints []Checkpoint

	// These fields are related to voting on consensus rule changes as
	// defined by BIP0009.
	//
	// RuleChangeActivationThreshold is the number of blocks in a threshold
	// state retarget window for which a positive vote for a rule change
	// must be cast in order to lock in a rule change. It should typically
	// be 95% for the main network and 75% for test networks.
	//
	// MinerConfirmationWindow is the number of blocks in each threshold
	// state retarget window.
	//
	// Deployments define the specific consensus rule changes to be voted
	// on.
	RuleChangeActivationThreshold uint32
	MinerConfirmationWindow       uint32
	Deployments                   [DefinedDeployments]ConsensusDeployment

	// Mempool parameters
	RelayNonStdTxs bool

	// Human-readable part for Bech32 encoded segwit addresses, as defined
	// in BIP 173.
	Bech32HRPSegwit B

	// Address encoding magics
	PubKeyHashAddrID        byte // First byte of a P2PKH address
	ScriptHashAddrID        byte // First byte of a P2SH address
	PrivateKeyID            byte // First byte of a WIF private key
	WitnessPubKeyHashAddrID byte // First byte of a P2WPKH address
	WitnessScriptHashAddrID byte // First byte of a P2WSH address

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID [4]byte
	HDPublicKeyID  [4]byte

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType uint32
}

// MainNetParams defines the network parameters for the main Bitcoin network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         wire.MainNet,
	DefaultPort: "8333",
	DNSSeeds: []DNSSeed{
		{"seed.bitcoin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.bitcoin.dashjr.org", false},
		{"seed.bitcoinstats.com", true},
		{"seed.bitnodes.io", false},
		{"seed.bitcoin.jonasschnelli.ch", true},
	},

	// Chain parameters
	GenesisBlock:             &genesisBlock,
	GenesisHash:              &genesisHash,
	PowLimit:                 mainPowLimit,
	PowLimitBits:             0x1d00ffff,
	BIP0034Height:            227931, // 000000000000024b89b42a942fe0d9fea3bb44ab7bd1b19115dd6a759c0808b8
	BIP0065Height:            388381, // 000000000000000004c2b624ed5d7756c508d90fd0da2c7c679febfa6c4735f0
	BIP0066Height:            363725, // 00000000000000000379eaa19dce8c9b722d46ae6a57c2f1a988119488b50931
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
		{11111,
			newHashFromStr("0000000069e244f73d78e8fd29ba2fd2ed618bd6fa2ee92559f542fdb26e7c1d")},
		{33333,
			newHashFromStr("000000002dd5588a74784eaa7ab0507a18ad16a236e7b1ce69f00d7ddfb5d0a6")},
		{74000,
			newHashFromStr("0000000000573993a3c9e41ce34471c079dcf5f52a0e824a81e7f953b8661a20")},
		{105000,
			newHashFromStr("00000000000291ce28027faea320c8d2b054b2e0fe44a773f3eefb151d6bdc97")},
		{134444,
			newHashFromStr("00000000000005b12ffd4cd315cd34ffd4a594f430ac814c91184a0d42d2b0fe")},
		{168000,
			newHashFromStr("000000000000099e61ea72015e79632f216fe6cb33d7899acb35b75c8303b763")},
		{193000,
			newHashFromStr("000000000000059f452a5f7340de6682a977387c17010ff6e6c3bd83ca8b1317")},
		{210000,
			newHashFromStr("000000000000048b95347e83192f69cf0366076336c639f9b7228e9ba171342e")},
		{216116,
			newHashFromStr("00000000000001b4f4b433e81ee46494af945cf96014816a4e2370f11b23df4e")},
		{225430,
			newHashFromStr("00000000000001c108384350f74090433e7fcf79a606b8e797f065b130575932")},
		{250000,
			newHashFromStr("000000000000003887df1f29024b06fc2200b55f8af8f35453d7be294df2d214")},
		{267300,
			newHashFromStr("000000000000000a83fbd660e918f218bf37edd92b748ad940483c7c116179ac")},
		{279000,
			newHashFromStr("0000000000000001ae8c72a0b0c301f67e3afca10e819efa9041e458e9bd7e40")},
		{300255,
			newHashFromStr("0000000000000000162804527c6e9b9f0563a280525f9d08c12041def0a0f3b2")},
		{319400,
			newHashFromStr("000000000000000021c6052e9becade189495d1c539aa37c58917305fd15f13b")},
		{343185,
			newHashFromStr("0000000000000000072b8bf361d01a6ba7d445dd024203fafc78768ed4368554")},
		{352940,
			newHashFromStr("000000000000000010755df42dba556bb72be6a32f3ce0b6941ce4430152c9ff")},
		{382320,
			newHashFromStr("00000000000000000a8dc6ed5b133d0eb2fd6af56203e4159789b092defd8ab2")},
		{400000,
			newHashFromStr("000000000000000004ec466ce4732fe6f1ed1cddc2ed4b328fff5224276e3f6f")},
		{430000,
			newHashFromStr("000000000000000001868b2bb3a285f3cc6b33ea234eb70facf4dcdf22186b87")},
		{460000,
			newHashFromStr("000000000000000000ef751bbce8e744ad303c47ece06c8d863e4d417efc258c")},
		{490000,
			newHashFromStr("000000000000000000de069137b17b8d5a3dfbd5b145b2dcfb203f15d0c4de90")},
		{520000,
			newHashFromStr("0000000000000000000d26984c0229c9f6962dc74db0a6d525f2f1640396f69c")},
		{550000,
			newHashFromStr("000000000000000000223b7a2298fb1c6c75fb0efc28a4c56853ff4112ec6bc9")},
		{560000,
			newHashFromStr("0000000000000000002c7b276daf6efb2b6aa68e2ce3be67ef925b3264ae7122")},
		{563378,
			newHashFromStr("0000000000000000000f1c54590ee18d15ec70e68c8cd4cfbadb1b4f11697eee")},
		{597379,
			newHashFromStr("00000000000000000005f8920febd3925f8272a6a71237563d78c2edfdd09ddf")},
		{623950,
			newHashFromStr("0000000000000000000f2adce67e49b0b6bdeb9de8b7c3d7e93b21e7fc1e819d")},
		{654683,
			newHashFromStr("0000000000000000000b9d2ec5a352ecba0592946514a92f14319dc2b367fc72")},
		{691719,
			newHashFromStr("00000000000000000008a89e854d57e5667df88f1cdef6fde2fbca1de5b639ad")},
		{724466,
			newHashFromStr("000000000000000000052d314a259755ca65944e68df6b12a067ea8f1f5a7091")},
		{751565,
			newHashFromStr("00000000000000000009c97098b5295f7e5f183ac811fb5d1534040adb93cabd")},
	},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1916, // 95% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016, //
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber: 28,
			DeploymentStarter: NewMedianTimeDeploymentStarter(
				time.Unix(11991456010, 0), // January 1, 2008 UTC
			),
			DeploymentEnder: NewMedianTimeDeploymentEnder(
				time.Unix(1230767999, 0), // December 31, 2008 UTC
			),
		},
		DeploymentTestDummyMinActivation: {
			BitNumber:                 22,
			CustomActivationThreshold: 1815,    // Only needs 90% hash rate.
			MinActivationHeight:       10_0000, // Can only activate after height 10k.
			DeploymentStarter: NewMedianTimeDeploymentStarter(
				time.Time{}, // Always available for vote
			),
			DeploymentEnder: NewMedianTimeDeploymentEnder(
				time.Time{}, // Never expires
			),
		},
		DeploymentCSV: {
			BitNumber: 0,
			DeploymentStarter: NewMedianTimeDeploymentStarter(
				time.Unix(1462060800, 0), // May 1st, 2016
			),
			DeploymentEnder: NewMedianTimeDeploymentEnder(
				time.Unix(1493596800, 0), // May 1st, 2017
			),
		},
		DeploymentSegwit: {
			BitNumber: 1,
			DeploymentStarter: NewMedianTimeDeploymentStarter(
				time.Unix(1479168000, 0), // November 15, 2016 UTC
			),
			DeploymentEnder: NewMedianTimeDeploymentEnder(
				time.Unix(1510704000, 0), // November 15, 2017 UTC.
			),
		},
		DeploymentTaproot: {
			BitNumber: 2,
			DeploymentStarter: NewMedianTimeDeploymentStarter(
				time.Unix(1619222400, 0), // April 24th, 2021 UTC.
			),
			DeploymentEnder: NewMedianTimeDeploymentEnder(
				time.Unix(1628640000, 0), // August 11th, 2021 UTC.
			),
			CustomActivationThreshold: 1815, // 90%
			MinActivationHeight:       709_632,
		},
	},

	// Mempool parameters
	RelayNonStdTxs: false,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: B("bc"), // always bc for main net

	// Address encoding magics
	PubKeyHashAddrID:        0x00, // starts with 1
	ScriptHashAddrID:        0x05, // starts with 3
	PrivateKeyID:            0x80, // starts with 5 (uncompressed) or K (compressed)
	WitnessPubKeyHashAddrID: 0x06, // starts with p2
	WitnessScriptHashAddrID: 0x0A, // starts with 7Xh

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 0,
}

// newHashFromStr converts the passed big-endian hex string into a
// chainhash.Hash.  It only differs from the one available in chainhash in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(hexStr)
	if err != nil {
		// Ordinarily I don't like panics in library code since it
		// can take applications down without them having a chance to
		// recover which is extremely annoying, however an exception is
		// being made in this case because the only way this can panic
		// is if there is an error in the hard-coded hashes.  Thus it
		// will only ever potentially panic on init and therefore is
		// 100% predictable.
		panic(err)
	}
	return hash
}
