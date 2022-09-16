package tarodb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/lightninglabs/taro/address"
	"github.com/lightninglabs/taro/asset"
	"github.com/lightninglabs/taro/tarodb/sqlite"
	"github.com/lightningnetwork/lnd/keychain"
)

type (
	// AddrQuery as a type alias for a query into the set of known
	// addresses.
	AddrQuery = sqlite.FetchAddrsParams

	// NewAddr is a type alias for the params to create a new address.
	NewAddr = sqlite.InsertAddrParams

	// Addresses is a type alias for the full address row with key locator
	// information.
	Addresses = sqlite.FetchAddrsRow

	// AddrByTaprootOutput is a type alias for returning an address by its
	// Taproot output key.
	AddrByTaprootOutput = sqlite.FetchAddrByTaprootOutputKeyRow
)

// AddrBook is an interface that represents the storage backed needed to create
// the TaroAddressBook book. We need to be able to insert/fetch addresses, and
// also make internal keys since each address has an internal key and a script
// key (tho they can be the same).
type AddrBook interface {
	// FetchAddrs returns all the addresses based on the constraints of the
	// passed AddrQuery.
	FetchAddrs(ctx context.Context, arg AddrQuery) ([]Addresses, error)

	// FetchAddrByTaprootOutputKey returns a single address based on its
	// Taproot output key or a sql.ErrNoRows error if no such address
	// exists.
	FetchAddrByTaprootOutputKey(ctx context.Context,
		arg []byte) (AddrByTaprootOutput, error)

	// InsertAddr inserts a new address into the database.
	InsertAddr(ctx context.Context, arg NewAddr) (int32, error)

	// UpsertInternalKey inserts a new or updates an existing internal key
	// into the database and returns the primary key.
	UpsertInternalKey(ctx context.Context, arg InternalKey) (int32, error)

	// UpsertScriptKey inserts a new script key on disk into the DB.
	UpsertScriptKey(context.Context, NewScriptKey) (int32, error)
}

// AddrBookTxOptions defines the set of db txn options the AddrBook
// understands.
type AddrBookTxOptions struct {
	// readOnly governs if a read only transaction is needed or not.
	readOnly bool
}

// ReadOnly returns true if the transaction should be read only.
//
// NOTE: This implements the TxOptions
func (a *AddrBookTxOptions) ReadOnly() bool {
	return a.readOnly
}

// NewAddrBookReadTx creates a new read transaction option set.
func NewAddrBookReadTx() AssetStoreTxOptions {
	return AssetStoreTxOptions{
		readOnly: true,
	}
}

// BatchedAddrBook is a version of the AddrBook that's capable of batched
// database operations.
type BatchedAddrBook interface {
	AddrBook

	BatchedTx[AddrBook, TxOptions]
}

// TaroAddressBook represents a storage backend for all the Taro addresses a
// daemon has created.
type TaroAddressBook struct {
	db BatchedAddrBook
}

// NewTaroAddressBook creates a new TaroAddressBook instance given a open
// BatchedAddrBook storage backend.
func NewTaroAddressBook(db BatchedAddrBook) *TaroAddressBook {
	return &TaroAddressBook{
		db: db,
	}
}

// insertInternalKey inserts a new internal key into the DB and returns the
// primary key of the internal key.
func insertInternalKey(ctx context.Context, a AddrBook,
	desc keychain.KeyDescriptor) (int32, error) {

	return a.UpsertInternalKey(ctx, InternalKey{
		RawKey:    desc.PubKey.SerializeCompressed(),
		KeyFamily: int32(desc.Family),
		KeyIndex:  int32(desc.Index),
	})
}

// InsertAddrs inserts a new address into the database.
func (t *TaroAddressBook) InsertAddrs(ctx context.Context,
	addrs ...address.AddrWithKeyInfo) error {

	var writeTxOpts AddrBookTxOptions
	return t.db.ExecTx(ctx, &writeTxOpts, func(db AddrBook) error {
		// For each of the addresses listed, we'll insert the two new
		// internal keys, then use those returned primary key IDs to
		// returned to insert the address itself.
		for _, addr := range addrs {
			rawScriptKeyID, err := insertInternalKey(
				ctx, db, addr.ScriptKeyTweak.RawKey,
			)
			if err != nil {
				return fmt.Errorf("unable to insert internal "+
					"script key: %w", err)
			}
			scriptKeyID, err := db.UpsertScriptKey(ctx, NewScriptKey{
				InternalKeyID:    rawScriptKeyID,
				TweakedScriptKey: addr.ScriptKey.SerializeCompressed(),
				Tweak:            addr.ScriptKeyTweak.Tweak,
			})
			if err != nil {
				return fmt.Errorf("unable to insert script "+
					"key: %w", err)
			}

			taprootKeyID, err := insertInternalKey(
				ctx, db, addr.InternalKeyDesc,
			)
			if err != nil {
				return fmt.Errorf("unable to insert internal "+
					"taproot key: %w", err)
			}

			var famKeyBytes []byte
			if addr.FamilyKey != nil {
				famKeyBytes = addr.FamilyKey.SerializeCompressed()
			}
			_, err = db.InsertAddr(ctx, NewAddr{
				Version:      int16(addr.Version),
				AssetID:      addr.ID[:],
				FamKey:       famKeyBytes,
				ScriptKeyID:  scriptKeyID,
				TaprootKeyID: taprootKeyID,
				TaprootOutputKey: schnorr.SerializePubKey(
					&addr.TaprootOutputKey,
				),
				Amount:       int64(addr.Amount),
				AssetType:    int16(addr.Type),
				CreationTime: addr.CreationTime,
			})
			if err != nil {
				return fmt.Errorf("unable to insert addr: %w",
					err)
			}
		}

		return nil
	})
}

// QueryAddrs attempts to query for the set of addresses on disk given the
// passed set of query params.
func (t *TaroAddressBook) QueryAddrs(ctx context.Context,
	params address.QueryParams) ([]address.AddrWithKeyInfo, error) {

	var addrs []address.AddrWithKeyInfo

	// If the created before time is zero, then we'll use a very large date
	// to ensure that we don't restrict based on this field.
	if params.CreatedBefore.IsZero() {
		params.CreatedBefore = time.Unix(math.MaxInt64, 0)
	}

	// Similarly, for sqlite using LIMIT with a value of -1 means no rows
	// should be limited.
	//
	// TODO(roasbeef): needs to be more portable
	limit := int32(-1)
	if params.Limit != 0 {
		limit = params.Limit
	}

	readOpts := NewAddrBookReadTx()
	err := t.db.ExecTx(ctx, &readOpts, func(db AddrBook) error {
		// First, fetch the set of addresses based on the set of query
		// parameters.
		dbAddrs, err := db.FetchAddrs(ctx, AddrQuery{
			CreatedAfter:  params.CreatedAfter,
			CreatedBefore: params.CreatedBefore,
			NumOffset:     int32(params.Offset),
			NumLimit:      limit,
		})
		if err != nil {
			return err
		}

		// Next, we'll need to map each of the addresses into an
		// AddrWithKeyInfo struct that can be used in a general
		// context.
		for _, addr := range dbAddrs {
			var assetID asset.ID
			copy(assetID[:], addr.AssetID)

			var famKey *btcec.PublicKey
			if addr.FamKey != nil {
				famKey, err = btcec.ParsePubKey(addr.FamKey)
				if err != nil {
					return fmt.Errorf("unable to decode "+
						"fam key: %w", err)
				}
			}

			rawScriptKey, err := btcec.ParsePubKey(
				addr.RawScriptKey,
			)
			if err != nil {
				return fmt.Errorf("unable to decode "+
					"script key: %w", err)
			}
			rawScriptKeyDesc := keychain.KeyDescriptor{
				KeyLocator: keychain.KeyLocator{
					Family: keychain.KeyFamily(
						addr.ScriptKeyFamily,
					),
					Index: uint32(addr.ScriptKeyIndex),
				},
				PubKey: rawScriptKey,
			}

			internalKey, err := btcec.ParsePubKey(addr.RawTaprootKey)
			if err != nil {
				return fmt.Errorf("unable to decode "+
					"taproot key: %w", err)
			}
			internalKeyDesc := keychain.KeyDescriptor{
				KeyLocator: keychain.KeyLocator{
					Family: keychain.KeyFamily(
						addr.TaprootKeyFamily,
					),
					Index: uint32(addr.TaprootKeyIndex),
				},
				PubKey: internalKey,
			}

			scriptKey, err := btcec.ParsePubKey(addr.TweakedScriptKey)
			if err != nil {
				return err
			}

			taprootOutputKey, err := schnorr.ParsePubKey(
				addr.TaprootOutputKey,
			)
			if err != nil {
				return fmt.Errorf("unable to parse taproot "+
					"output key: %w", err)
			}

			addrs = append(addrs, address.AddrWithKeyInfo{
				Taro: &address.Taro{
					Version:     asset.Version(addr.Version),
					ID:          assetID,
					FamilyKey:   famKey,
					ScriptKey:   *scriptKey,
					InternalKey: *internalKey,
					Amount:      uint64(addr.Amount),
					Type:        asset.Type(addr.AssetType),
				},
				ScriptKeyTweak: asset.TweakedScriptKey{
					RawKey: rawScriptKeyDesc,
					Tweak:  addr.ScriptKeyTweak,
				},
				InternalKeyDesc:  internalKeyDesc,
				TaprootOutputKey: *taprootOutputKey,
				CreationTime:     addr.CreationTime,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

// AddrByTaprootOutput returns a single address based on its Taproot output
// key or a sql.ErrNoRows error if no such address exists.
func (t *TaroAddressBook) AddrByTaprootOutput(ctx context.Context,
	key *btcec.PublicKey) (*address.AddrWithKeyInfo, error) {

	var (
		addr     *address.AddrWithKeyInfo
		readOpts = NewAddrBookReadTx()
	)
	err := t.db.ExecTx(ctx, &readOpts, func(db AddrBook) error {
		var err error
		addr, err = fetchAddr(ctx, db, key)
		return err
	})
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// fetchAddr fetches a single address identified by its taproot output key from
// the database and populates all its fields.
func fetchAddr(ctx context.Context, db AddrBook,
	taprootOutputKey *btcec.PublicKey) (*address.AddrWithKeyInfo, error) {

	dbAddr, err := db.FetchAddrByTaprootOutputKey(
		ctx, schnorr.SerializePubKey(taprootOutputKey),
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, address.ErrNoAddr

	case err != nil:
		return nil, err
	}

	var assetID asset.ID
	copy(assetID[:], dbAddr.AssetID)

	var famKey *btcec.PublicKey
	if dbAddr.FamKey != nil {
		famKey, err = btcec.ParsePubKey(dbAddr.FamKey)
		if err != nil {
			return nil, fmt.Errorf("unable to decode fam key: %w",
				err)
		}
	}

	rawScriptKey, err := btcec.ParsePubKey(dbAddr.RawScriptKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode script key: %w", err)
	}
	scriptKeyDesc := keychain.KeyDescriptor{
		KeyLocator: keychain.KeyLocator{
			Family: keychain.KeyFamily(
				dbAddr.ScriptKeyFamily,
			),
			Index: uint32(dbAddr.ScriptKeyIndex),
		},
		PubKey: rawScriptKey,
	}

	scriptKey, err := btcec.ParsePubKey(dbAddr.TweakedScriptKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode script key: %w", err)
	}

	internalKey, err := btcec.ParsePubKey(dbAddr.RawTaprootKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode taproot key: %w", err)
	}
	internalKeyDesc := keychain.KeyDescriptor{
		KeyLocator: keychain.KeyLocator{
			Family: keychain.KeyFamily(
				dbAddr.TaprootKeyFamily,
			),
			Index: uint32(dbAddr.TaprootKeyIndex),
		},
		PubKey: internalKey,
	}

	return &address.AddrWithKeyInfo{
		Taro: &address.Taro{
			Version:     asset.Version(dbAddr.Version),
			ID:          assetID,
			FamilyKey:   famKey,
			ScriptKey:   *scriptKey,
			InternalKey: *internalKey,
			Amount:      uint64(dbAddr.Amount),
			Type:        asset.Type(dbAddr.AssetType),
		},
		ScriptKeyTweak: asset.TweakedScriptKey{
			RawKey: scriptKeyDesc,
			Tweak:  dbAddr.ScriptKeyTweak,
		},
		InternalKeyDesc:  internalKeyDesc,
		TaprootOutputKey: *taprootOutputKey,
		CreationTime:     dbAddr.CreationTime,
	}, nil
}

// A compile-time assertion to ensure that TaroAddressBook meets the
// address.Storage interface.
var _ address.Storage = (*TaroAddressBook)(nil)
