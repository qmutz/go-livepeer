package pm

import (
	"bytes"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	errInvalidAuxDataLength          = errors.New("invalid ticket aux data length")
	errInvalidCreationRound          = errors.New("invalid ticket creation round")
	errInvalidCreationRoundBlockHash = errors.New("invalid ticket creation round block hash")
)

// AuxDataCreator defines the methods for creating ticket aux data
type AuxDataCreator interface {
	Create() ([]byte, error)
}

// AuxDataValidator defines the methods for validating ticket aux data
type AuxDataValidator interface {
	Validate([]byte) error
}

// RoundsManager defines the methods for fetching the last
// initialized round and associated block hash of the Livepeer protocol
type RoundsManager interface {
	LastInitializedRound() (*big.Int, error)
	BlockHashForRound(round *big.Int) ([32]byte, error)
}

// RoundAuxDataCreator is an AuxDataCreator that generates ticket aux data using
// the last initialized round and associated block hash of the Livepeer protocol
type RoundAuxDataCreator struct {
	roundsManager RoundsManager
}

// RoundAuxDataValidator is an AuxDataValidator that validates ticket aux data
// based on the last initialized round and associated block hash of the Livepeer protocol
type RoundAuxDataValidator struct {
	roundsManager RoundsManager
}

// NewRoundAuxDataCreator returns a RoundAuxDataCreator
func NewRoundAuxDataCreator(roundsManager RoundsManager) *RoundAuxDataCreator {
	return &RoundAuxDataCreator{
		roundsManager: roundsManager,
	}
}

// Create returns the last initialized round and its block hash as a byte slice
func (c *RoundAuxDataCreator) Create() ([]byte, error) {
	round, err := c.roundsManager.LastInitializedRound()
	if err != nil {
		return nil, err
	}

	blkHash, err := c.roundsManager.BlockHashForRound(round)
	if err != nil {
		return nil, err
	}

	return append(
		ethcommon.LeftPadBytes(round.Bytes(), 32),
		blkHash[:]...,
	), nil
}

// NewRoundAuxDataValidator returns a RoundAuxDataValidator
func NewRoundAuxDataValidator(roundsManager RoundsManager) *RoundAuxDataValidator {
	return &RoundAuxDataValidator{
		roundsManager: roundsManager,
	}
}

// Validate returns a boolean indicating whether the provided ticket aux data
// is valid given the last initialized round and its block hash
func (v *RoundAuxDataValidator) Validate(auxData []byte) error {
	// auxData = creation round (32 bytes) + creation round block hash (32 bytes)
	if len(auxData) != 64 {
		return errInvalidAuxDataLength
	}

	creationRound := new(big.Int).SetBytes(auxData[:32])
	creationRoundBlkHash := auxData[32:]

	round, err := v.roundsManager.LastInitializedRound()
	if err != nil {
		return err
	}

	blkHash, err := v.roundsManager.BlockHashForRound(round)
	if err != nil {
		return err
	}

	if creationRound.Cmp(round) != 0 {
		return errInvalidCreationRound
	}

	if !bytes.Equal(creationRoundBlkHash, blkHash[:]) {
		return errInvalidCreationRoundBlockHash
	}

	return nil
}
