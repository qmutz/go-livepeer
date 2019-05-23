package pm

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoundsManager struct {
	mock.Mock
}

func (m *mockRoundsManager) LastInitializedRound() (*big.Int, error) {
	args := m.Called()
	round := args.Get(0)
	err := args.Error(1)
	if round == nil {
		return nil, err
	}

	return round.(*big.Int), err
}

func (m *mockRoundsManager) BlockHashForRound(round *big.Int) ([32]byte, error) {
	args := m.Called(round)
	blkHash := args.Get(0)
	err := args.Error(1)
	if blkHash == nil {
		return [32]byte{}, err
	}

	return blkHash.([32]byte), err
}

func TestCreate(t *testing.T) {
	rm := &mockRoundsManager{}
	c := &RoundAuxDataCreator{rm}

	round := big.NewInt(5)
	var blkHash [32]byte
	copy(blkHash[:], ethcommon.FromHex("7624778dedc75f8b322b9fa1632a610d40b85e106c7d9bf0e743a9ce291b9c6f"))

	assert := assert.New(t)

	// Test LastInitializedRound error

	expErr := errors.New("LastInitializedRound error")
	rm.On("LastInitializedRound").Return(nil, expErr).Once()
	auxData, err := c.Create()
	assert.NotNil(err)
	assert.EqualError(err, expErr.Error())

	// Test BlockHashForRound error

	expErr = errors.New("BlockHashForRound error")
	rm.On("LastInitializedRound").Return(big.NewInt(5), nil).Once()
	rm.On("BlockHashForRound", big.NewInt(5)).Return(nil, expErr).Once()
	auxData, err = c.Create()
	assert.NotNil(err)
	assert.EqualError(err, expErr.Error())

	// Test round = 0

	rm.On("LastInitializedRound").Return(big.NewInt(0), nil).Once()
	rm.On("BlockHashForRound", big.NewInt(0)).Return(blkHash, nil).Once()
	auxData, err = c.Create()
	assert.Nil(err)
	assert.Equal(64, len(auxData))
	assert.Equal(
		ethcommon.FromHex("00000000000000000000000000000000000000000000000000000000000000007624778dedc75f8b322b9fa1632a610d40b85e106c7d9bf0e743a9ce291b9c6f"),
		auxData,
	)

	// Test empty block hash

	var emptyBlkHash [32]byte
	copy(emptyBlkHash[:], ethcommon.LeftPadBytes([]byte{}, 32))
	rm.On("LastInitializedRound").Return(round, nil).Once()
	rm.On("BlockHashForRound", round).Return(emptyBlkHash, nil).Once()
	auxData, err = c.Create()
	assert.Nil(err)
	assert.Equal(64, len(auxData))
	assert.Equal(
		ethcommon.FromHex("00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000000"),
		auxData,
	)

	// Test round = 0 and empty block hash

	rm.On("LastInitializedRound").Return(big.NewInt(0), nil).Once()
	rm.On("BlockHashForRound", big.NewInt(0)).Return(emptyBlkHash, nil).Once()
	auxData, err = c.Create()
	assert.Nil(err)
	assert.Equal(64, len(auxData))
	assert.Equal(
		ethcommon.FromHex("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		auxData,
	)

	// Test normal case

	rm.On("LastInitializedRound").Return(round, nil).Once()
	rm.On("BlockHashForRound", round).Return(blkHash, nil).Once()
	auxData, err = c.Create()
	assert.Nil(err)
	assert.Equal(64, len(auxData))
	assert.Equal(
		ethcommon.FromHex("00000000000000000000000000000000000000000000000000000000000000057624778dedc75f8b322b9fa1632a610d40b85e106c7d9bf0e743a9ce291b9c6f"),
		auxData,
	)
}

func TestValidate(t *testing.T) {
	rm := &mockRoundsManager{}
	v := &RoundAuxDataValidator{rm}

	round := big.NewInt(5)
	var blkHash [32]byte
	copy(blkHash[:], ethcommon.FromHex("7624778dedc75f8b322b9fa1632a610d40b85e106c7d9bf0e743a9ce291b9c6f"))

	assert := assert.New(t)

	// Test invalid length

	auxData := []byte{}
	assert.EqualError(v.Validate(auxData), errInvalidAuxDataLength.Error())

	// Test LastInitializedRound error

	expErr := errors.New("LastInitializedRound error")
	auxData = ethcommon.FromHex("00000000000000000000000000000000000000000000000000000000000000057624778dedc75f8b322b9fa1632a610d40b85e106c7d9bf0e743a9ce291b9c6f")
	rm.On("LastInitializedRound").Return(nil, expErr).Once()
	assert.EqualError(v.Validate(auxData), expErr.Error())

	// Test BlockHashForRound error

	expErr = errors.New("BlockHashForRound error")
	rm.On("LastInitializedRound").Return(round, nil).Once()
	rm.On("BlockHashForRound", round).Return(nil, expErr).Once()
	assert.EqualError(v.Validate(auxData), expErr.Error())

	// Test invalid creation round

	rm.On("LastInitializedRound").Return(big.NewInt(2), nil).Once()
	rm.On("BlockHashForRound", big.NewInt(2)).Return(blkHash, nil).Once()
	assert.EqualError(v.Validate(auxData), errInvalidCreationRound.Error())

	// Test invalid creation round block hash

	rm.On("LastInitializedRound").Return(round, nil).Once()
	rm.On("BlockHashForRound", round).Return([32]byte{}, nil).Once()
	assert.EqualError(v.Validate(auxData), errInvalidCreationRoundBlockHash.Error())

	// Test valid creation round and block hash

	rm.On("LastInitializedRound").Return(round, nil).Once()
	rm.On("BlockHashForRound", round).Return(blkHash, nil).Once()
	assert.Nil(v.Validate(auxData))
}
