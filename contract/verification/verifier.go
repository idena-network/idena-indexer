package verification

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"time"
)

type Verifier interface {
	Submit(contractAddress common.Address, code []byte, fileName string) (usrErr, err error)
}

type verifierImpl struct {
	db     VerifierDb
	info   WasmInfo
	logger log.Logger
}

func NewVerifier(db VerifierDb, info WasmInfo, logger log.Logger) Verifier {
	res := &verifierImpl{
		db:     db,
		info:   info,
		logger: logger,
	}
	go res.loop()
	return res
}

func (v *verifierImpl) Submit(contractAddress common.Address, code []byte, fileName string) (usrErr, err error) {
	return v.db.SavePendingVerification(contractAddress, code, fileName)
}

func (v *verifierImpl) loop() {
	for {
		time.Sleep(time.Second * 5)
		if err := v.verifyPendingContract(); err != nil {
			v.logger.Warn(fmt.Sprintf("failed to verify contract: %v", err))
		}
	}
}

func (v *verifierImpl) verifyPendingContract() error {
	verification, err := v.db.GetPendingVerification()
	if err != nil {
		return errors.Wrap(err, "failed to get pending verification")
	}

	if verification == nil {
		return nil
	}

	start := time.Now()
	v.logger.Info(fmt.Sprintf("start verifying contract %v, contract src len: %v", verification.Address.Hex(), len(verification.Data)))

	codeHash, err := v.getCodeHash(verification.Code)
	if err != nil {
		return errors.Wrap(err, "failed to get code hash")
	}

	dataHash, verificationErr := v.getDataInfo(verification.Data)
	if verificationErr != nil {
		v.logger.Warn(fmt.Sprintf("failed to verify contract: %v", err))
		verificationErr = errors.New("failed to compile")
	} else if bytes.Compare(codeHash, dataHash) != 0 {
		v.logger.Warn(fmt.Sprintf("different hashes, actual: %v, provided: %v", hexutil.Encode(codeHash), hexutil.Encode(dataHash)))
		verificationErr = errors.New("wrong compiled contract hash")
	}

	verified := verificationErr == nil

	var state State
	var errorMessage *string
	if verified {
		state = StateVerified
	} else {
		state = StateFailed
		v := verificationErr.Error()
		errorMessage = &v
	}

	if err := v.db.UpdateVerificationState(verification.Address, state, verification.Data, errorMessage); err != nil {
		return errors.Wrap(err, "failed to update verification state")
	}

	v.logger.Info(fmt.Sprintf("contract %v verification result: %v, err: %v", verification.Address.Hex(), verified, verificationErr), "d", time.Since(start))

	return nil
}

func (v *verifierImpl) getDataInfo(data []byte) ([]byte, error) {
	return v.info.Hash(data)
}

func (v *verifierImpl) getCodeHash(code []byte) ([]byte, error) {
	hash := sha256.Sum256(code)
	return hash[:], nil
}
