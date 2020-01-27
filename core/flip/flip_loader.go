package flip

import (
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"time"
)

const (
	requestRetryInterval = time.Minute * 60
	flipsToRetry         = 50
)

type Loader interface {
	SubmitToLoad(cidBytes []byte, txHash common.Hash)
}

func NewLoader(
	getEpoch func() uint64,
	enabledRetry func() bool,
	db DbAccessor,
	flipper *flip.Flipper,
	logger log.Logger,
) Loader {
	l := &loaderImpl{
		db:             db,
		flipper:        flipper,
		logger:         logger,
		headers:        make(chan *flipHeader, 10),
		headersToRetry: make(chan *flipHeader),
		getEpoch:       getEpoch,
		enabledRetry:   enabledRetry,
	}
	l.initialize()
	return l
}

type loaderImpl struct {
	db             DbAccessor
	flipper        *flip.Flipper
	headers        chan *flipHeader
	headersToRetry chan *flipHeader
	logger         log.Logger
	getEpoch       func() uint64
	enabledRetry   func() bool
}

type DbAccessor interface {
	GetEpochFlipsWithoutSize(epoch uint64, limit int) (cids []string, err error)
	SaveFlipSize(flipCid string, size int) error
}

type flipHeader struct {
	cidBytes []byte
	txHash   common.Hash
}

type flipBody struct {
	cidStr   string
	ipfsFlip *flip.IpfsFlip
}

func (l *loaderImpl) initialize() {
	go l.loadLoop()
	go l.processLoop()
}

func (l *loaderImpl) SubmitToLoad(cidBytes []byte, txHash common.Hash) {
	l.headers <- &flipHeader{
		cidBytes: cidBytes,
		txHash:   txHash,
	}
}

func (l *loaderImpl) load(header *flipHeader) (*flipBody, error) {
	flipCid, err := cid.Parse(header.cidBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse flip cid, tx %s", header.txHash.Hex())
	}
	ipfsFlip, err := l.flipper.GetRawFlip(header.cidBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load flip, cid %s", flipCid)
	}
	return &flipBody{
		cidStr:   flipCid.String(),
		ipfsFlip: ipfsFlip,
	}, nil
}

func (l *loaderImpl) processLoop() {
	for {
		select {
		case header := <-l.headers:
			l.processHeader(header)
			continue
		default:
		}

		select {
		case header := <-l.headers:
			l.processHeader(header)
			continue
		case header := <-l.headersToRetry:
			l.processHeader(header)
			continue
		}
	}
}

func (l *loaderImpl) processHeader(header *flipHeader) {
	l.logger.Debug("Start processing flip")
	body, err := l.load(header)
	if err != nil {
		l.logger.Error(errors.Wrap(err, "Unable to load flip").Error())
		return
	}
	if err := l.db.SaveFlipSize(body.cidStr, len(body.ipfsFlip.PublicPart)); err != nil {
		l.logger.Error(errors.Wrap(err, "Unable to save flip size").Error())
		return
	}
	l.logger.Debug(fmt.Sprintf("Processed flip %s", body.cidStr))
}

func (l *loaderImpl) loadLoop() {
	for {
		time.Sleep(requestRetryInterval)
		if !l.enabledRetry() {
			l.logger.Debug("Retry is disabled")
			continue
		}
		l.logger.Debug("Getting flips to retry loading")
		cids, err := l.db.GetEpochFlipsWithoutSize(l.getEpoch(), flipsToRetry)
		if err != nil {
			l.logger.Error(errors.Wrap(err, "Unable to get flip from db to retry loading").Error())
			continue
		}
		if len(cids) == 0 {
			l.logger.Debug("No flips to retry loading")
			continue
		}
		l.logger.Debug(fmt.Sprintf("%d flips to retry loading: %v", len(cids), cids))
		for _, flipCidStr := range cids {
			flipCid, _ := cid.Parse(flipCidStr)
			l.headersToRetry <- &flipHeader{
				cidBytes: flipCid.Bytes(),
			}
		}
	}
}
