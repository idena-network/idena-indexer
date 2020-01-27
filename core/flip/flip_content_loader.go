package flip

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/ecies"
	"github.com/idena-network/idena-go/rlp"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"math/big"
	"time"
)

const (
	getCidsRetryInterval = time.Minute * 10
)

type ContentDbAccessor interface {
	GetFlipsToLoadContent(timestamp *big.Int, limit int) ([]*db.FlipToLoadContent, error)
	SaveFlipsContent(failedFlips []*db.FailedFlipContent, flipsContent []*db.FlipContent) error
}

type ContentLoader struct {
	db            ContentDbAccessor
	flipper       *flip.Flipper
	batchSize     int
	attemptsLimit int
	retryInterval time.Duration
	logger        log.Logger
}

func StartContentLoader(
	db ContentDbAccessor,
	batchSize int,
	attemptsLimit int,
	retryInterval time.Duration,
	flipper *flip.Flipper,
	logger log.Logger,
) {
	l := &ContentLoader{
		db:            db,
		batchSize:     batchSize,
		attemptsLimit: attemptsLimit,
		retryInterval: retryInterval,
		flipper:       flipper,
		logger:        logger,
	}
	l.initialize()
}

func (l *ContentLoader) initialize() {
	go l.loop()
}

func (l *ContentLoader) loop() {
	for {
		flips, err := l.db.GetFlipsToLoadContent(new(big.Int).SetInt64(time.Now().UTC().Unix()), l.batchSize)
		if err != nil {
			l.logger.Error(errors.Wrap(err, "Unable to get flip cids").Error())
			time.Sleep(getCidsRetryInterval)
			continue
		}
		if len(flips) == 0 {
			l.logger.Debug("No flips to load")
			time.Sleep(getCidsRetryInterval)
			continue
		}
		l.logger.Debug(fmt.Sprintf("%d flips to load", len(flips)))
		failedFlips, flipsContent := l.handleFlips(flips)
		if err = l.db.SaveFlipsContent(failedFlips, flipsContent); err != nil {
			l.logger.Error(errors.Wrap(err, "Unable to save flips content").Error())
			time.Sleep(getCidsRetryInterval)
			continue
		}
		l.logger.Debug("Flips content saved")
	}
}

func (l *ContentLoader) handleFlips(flips []*db.FlipToLoadContent) ([]*db.FailedFlipContent, []*db.FlipContent) {
	var failedFlips []*db.FailedFlipContent
	var flipsContent []*db.FlipContent
	for _, f := range flips {
		flipContent, err := l.getFlipContent(f)
		if err != nil {
			l.logger.Error(errors.
				Wrapf(err, "unable to get flip content (cid %s, attempt %d)", f.Cid, f.Attempts+1).Error())
			failedCid := &db.FailedFlipContent{
				Cid:                  f.Cid,
				AttemptsLimitReached: f.Attempts+1 == l.attemptsLimit,
			}
			if !failedCid.AttemptsLimitReached {
				failedCid.NextAttemptTimestamp = new(big.Int).SetInt64(time.Now().Add(l.retryInterval).UTC().Unix())
			}
			failedFlips = append(failedFlips, failedCid)
			continue
		}
		l.logger.Debug(fmt.Sprintf("Loaded flip %s", f.Cid))
		flipsContent = append(flipsContent, flipContent)
	}
	return failedFlips, flipsContent
}

func (l *ContentLoader) getFlipContent(flip *db.FlipToLoadContent) (*db.FlipContent, error) {
	encryptionKey, err := convertKey(flip.Key)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert key")
	}
	flipCid, _ := cid.Decode(flip.Cid)
	flipData, err := l.getFlipData(flipCid.Bytes(), encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get flip data")
	}
	parsedData, err := l.parseFlip(flip.Cid, flipData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse flip data")
	}
	return &parsedData, nil
}

func convertKey(keyHex string) (*ecies.PrivateKey, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, err
	}
	ecdsaKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, err
	}
	return ecies.ImportECDSA(ecdsaKey), nil
}

func (l *ContentLoader) getFlipData(cid []byte, encryptionKey *ecies.PrivateKey) ([]byte, error) {
	ipfsFlip, err := l.flipper.GetRawFlip(cid)
	if err != nil {
		return nil, err
	}
	if encryptionKey == nil {
		return nil, nil
	}
	decryptedFlip, err := encryptionKey.Decrypt(ipfsFlip.PublicPart, nil, nil)
	if err != nil {
		return nil, err
	}
	return decryptedFlip, nil
}

func (l *ContentLoader) parseFlip(flipCidStr string, data []byte) (db.FlipContent, error) {
	arr := make([]interface{}, 2)
	err := rlp.DecodeBytes(data, &arr)
	if err != nil || len(arr) == 0 {
		return db.FlipContent{}, err
	}
	var pics [][]byte
	for _, b := range arr[0].([]interface{}) {
		pics = append(pics, b.([]byte))
	}
	var allOrders [][]byte
	if len(arr) > 1 {
		for _, b := range arr[1].([]interface{}) {
			var orders []byte
			for _, bb := range b.([]interface{}) {
				var order byte
				if len(bb.([]byte)) > 0 {
					order = bb.([]byte)[0]
				}
				orders = append(orders, order)
			}
			allOrders = append(allOrders, orders)
		}
	}
	var icon []byte

	if len(pics) > 0 {
		icon, err = compressPic(pics[0])
		if err != nil {
			l.logger.Warn(errors.Wrapf(err, "Unable to create flip icon, cid %s", flipCidStr).Error())
		}
	}
	return db.FlipContent{
		Cid:    flipCidStr,
		Pics:   pics,
		Orders: allOrders,
		Icon:   icon,
	}, nil
}

func compressPic(src []byte) ([]byte, error) {
	srcImage, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	var x, y int
	if srcImage.Bounds().Max.X > srcImage.Bounds().Max.Y {
		x = 64
		y = int(float32(srcImage.Bounds().Max.Y) / float32(srcImage.Bounds().Max.X) * 64)
	} else {
		y = 64
		x = int(float32(srcImage.Bounds().Max.X) / float32(srcImage.Bounds().Max.Y) * 64)
	}

	dr := image.Rect(0, 0, x, y)
	dst := image.NewRGBA(dr)
	draw.CatmullRom.Scale(dst, dr, srcImage, srcImage.Bounds(), draw.Src, nil)

	var res bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&res), dst, nil)
	if err != nil {
		return nil, err
	}
	if len(res.Bytes()) == 0 {
		return nil, errors.New("empty converted pic")
	}
	return res.Bytes(), nil
}
