// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"database/sql"
	"time"
)

type Addr struct {
	ID               int64
	Version          int16
	AssetVersion     int16
	GenesisAssetID   int64
	GroupKey         []byte
	ScriptKeyID      int64
	TaprootKeyID     int64
	TapscriptSibling []byte
	TaprootOutputKey []byte
	Amount           int64
	AssetType        int16
	CreationTime     time.Time
	ManagedFrom      sql.NullTime
	ProofCourierAddr []byte
}

type AddrEvent struct {
	ID                  int64
	CreationTime        time.Time
	AddrID              int64
	Status              int16
	ChainTxnID          int64
	ChainTxnOutputIndex int32
	ManagedUtxoID       int64
	AssetProofID        sql.NullInt64
	AssetID             sql.NullInt64
}

type Asset struct {
	AssetID                  int64
	GenesisID                int64
	Version                  int32
	ScriptKeyID              int64
	AssetGroupWitnessID      sql.NullInt64
	ScriptVersion            int32
	Amount                   int64
	LockTime                 sql.NullInt32
	RelativeLockTime         sql.NullInt32
	SplitCommitmentRootHash  []byte
	SplitCommitmentRootValue sql.NullInt64
	AnchorUtxoID             sql.NullInt64
	Spent                    bool
}

type AssetGroup struct {
	GroupID         int64
	TweakedGroupKey []byte
	TapscriptRoot   []byte
	InternalKeyID   int64
	GenesisPointID  int64
}

type AssetGroupWitness struct {
	WitnessID    int64
	WitnessStack []byte
	GenAssetID   int64
	GroupKeyID   int64
}

type AssetMintingBatch struct {
	BatchID           int64
	BatchState        int16
	MintingTxPsbt     []byte
	ChangeOutputIndex sql.NullInt32
	GenesisID         sql.NullInt64
	HeightHint        int32
	CreationTimeUnix  time.Time
	TapscriptSibling  []byte
}

type AssetProof struct {
	ProofID   int64
	AssetID   int64
	ProofFile []byte
}

type AssetSeedling struct {
	SeedlingID         int64
	AssetName          string
	AssetVersion       int16
	AssetType          int16
	AssetSupply        int64
	AssetMetaID        int64
	EmissionEnabled    bool
	BatchID            int64
	GroupGenesisID     sql.NullInt64
	GroupAnchorID      sql.NullInt64
	ScriptKeyID        sql.NullInt64
	GroupInternalKeyID sql.NullInt64
	GroupTapscriptRoot []byte
}

type AssetTransfer struct {
	ID               int64
	HeightHint       int32
	AnchorTxnID      int64
	TransferTimeUnix time.Time
}

type AssetTransferInput struct {
	InputID     int64
	TransferID  int64
	AnchorPoint []byte
	AssetID     []byte
	ScriptKey   []byte
	Amount      int64
}

type AssetTransferOutput struct {
	OutputID                 int64
	TransferID               int64
	AnchorUtxo               int64
	ScriptKey                int64
	ScriptKeyLocal           bool
	Amount                   int64
	AssetVersion             int32
	SerializedWitnesses      []byte
	SplitCommitmentRootHash  []byte
	SplitCommitmentRootValue sql.NullInt64
	ProofSuffix              []byte
	NumPassiveAssets         int32
	OutputType               int16
	ProofCourierAddr         []byte
}

type AssetWitness struct {
	WitnessID            int64
	AssetID              int64
	PrevOutPoint         []byte
	PrevAssetID          []byte
	PrevScriptKey        []byte
	WitnessStack         []byte
	SplitCommitmentProof []byte
	WitnessIndex         int32
}

type AssetsMetum struct {
	MetaID       int64
	MetaDataHash []byte
	MetaDataBlob []byte
	MetaDataType sql.NullInt16
}

type ChainTxn struct {
	TxnID       int64
	Txid        []byte
	ChainFees   int64
	RawTx       []byte
	BlockHeight sql.NullInt32
	BlockHash   []byte
	TxIndex     sql.NullInt32
}

type FederationGlobalSyncConfig struct {
	ProofType       string
	AllowSyncInsert bool
	AllowSyncExport bool
}

type FederationProofSyncLog struct {
	ID             int64
	Status         string
	Timestamp      time.Time
	AttemptCounter int64
	SyncDirection  string
	ProofLeafID    int64
	UniverseRootID int64
	ServersID      int64
}

type FederationUniSyncConfig struct {
	Namespace       string
	AssetID         []byte
	GroupKey        []byte
	ProofType       string
	AllowSyncInsert bool
	AllowSyncExport bool
}

type GenesisAsset struct {
	GenAssetID     int64
	AssetID        []byte
	AssetTag       string
	MetaDataID     sql.NullInt64
	OutputIndex    int32
	AssetType      int16
	GenesisPointID int64
}

type GenesisInfoView struct {
	GenAssetID  int64
	AssetID     []byte
	AssetTag    string
	MetaHash    []byte
	OutputIndex int32
	AssetType   int16
	PrevOut     []byte
	AnchorTxid  []byte
	BlockHeight sql.NullInt32
}

type GenesisPoint struct {
	GenesisID  int64
	PrevOut    []byte
	AnchorTxID sql.NullInt64
}

type InternalKey struct {
	KeyID     int64
	RawKey    []byte
	KeyFamily int32
	KeyIndex  int32
}

type KeyGroupInfoView struct {
	WitnessID       int64
	GenAssetID      int64
	WitnessStack    []byte
	TapscriptRoot   []byte
	TweakedGroupKey []byte
	RawKey          []byte
	KeyIndex        int32
	KeyFamily       int32
	XOnlyGroupKey   []byte
}

type Macaroon struct {
	ID      []byte
	RootKey []byte
}

type ManagedUtxo struct {
	UtxoID           int64
	Outpoint         []byte
	AmtSats          int64
	InternalKeyID    int64
	TaprootAssetRoot []byte
	TapscriptSibling []byte
	MerkleRoot       []byte
	TxnID            int64
	LeaseOwner       []byte
	LeaseExpiry      sql.NullTime
	RootVersion      sql.NullInt16
}

type MssmtNode struct {
	HashKey   []byte
	LHashKey  []byte
	RHashKey  []byte
	Key       []byte
	Value     []byte
	Sum       int64
	Namespace string
}

type MssmtRoot struct {
	Namespace string
	RootHash  []byte
}

type MultiverseLeafe struct {
	ID                int64
	MultiverseRootID  int64
	AssetID           []byte
	GroupKey          []byte
	LeafNodeKey       []byte
	LeafNodeNamespace string
}

type MultiverseRoot struct {
	ID            int64
	NamespaceRoot string
	ProofType     string
}

type PassiveAsset struct {
	PassiveID       int64
	TransferID      int64
	AssetID         int64
	NewAnchorUtxo   int64
	ScriptKey       []byte
	AssetVersion    int32
	NewWitnessStack []byte
	NewProof        []byte
}

type ProofTransferLog struct {
	TransferType     string
	ProofLocatorHash []byte
	TimeUnix         time.Time
}

type ScriptKey struct {
	ScriptKeyID      int64
	InternalKeyID    int64
	TweakedScriptKey []byte
	Tweak            []byte
	DeclaredKnown    sql.NullBool
}

type TapscriptEdge struct {
	EdgeID     int64
	RootHashID int64
	NodeIndex  int64
	RawNodeID  int64
}

type TapscriptNode struct {
	NodeID  int64
	RawNode []byte
}

type TapscriptRoot struct {
	RootID     int64
	RootHash   []byte
	BranchOnly bool
}

type UniverseEvent struct {
	EventID        int64
	EventType      string
	UniverseRootID int64
	EventTime      time.Time
	EventTimestamp int64
}

type UniverseLeafe struct {
	ID                int64
	AssetGenesisID    int64
	MintingPoint      []byte
	ScriptKeyBytes    []byte
	UniverseRootID    int64
	LeafNodeKey       []byte
	LeafNodeNamespace string
}

type UniverseRoot struct {
	ID            int64
	NamespaceRoot string
	AssetID       []byte
	GroupKey      []byte
	ProofType     string
}

type UniverseServer struct {
	ID           int64
	ServerHost   string
	LastSyncTime time.Time
}

type UniverseStat struct {
	TotalAssetSyncs  int64
	TotalAssetProofs int64
	AssetID          []byte
	GroupKey         []byte
	ProofType        string
}
