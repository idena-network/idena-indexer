package mempool

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"time"
)

type Indexer struct {
	db               db.Accessor
	flipKeyChan      chan *flipKeyWrapper
	answerHashTxChan chan *txWrapper
	cache            *indexerCache
	log              log.Logger
	mutex            sync.Mutex
}

type indexerCache struct {
	fktMutex                sync.Mutex
	flipKeyTimestamps       []*db.MemPoolActionTimestamp
	ahttMutex               sync.Mutex
	answersHashTxTimestamps []*db.MemPoolActionTimestamp
}

type flipKeyWrapper struct {
	key  *types.FlipKey
	time *big.Int
}

type txWrapper struct {
	tx   *types.Transaction
	time *big.Int
}

func NewIndexer(db db.Accessor, log log.Logger) *Indexer {
	return &Indexer{
		db:               db,
		flipKeyChan:      make(chan *flipKeyWrapper, 1000),
		answerHashTxChan: make(chan *txWrapper, 1000),
		log:              log,
		cache:            &indexerCache{},
	}
}

func (indexer *Indexer) Initialize(bus eventbus.Bus) {
	indexer.subscribe(bus)
	go indexer.saveDataLoop()
	go indexer.listenFlipKeys()
	go indexer.listenAnswersHashTxs()
}

func (indexer *Indexer) subscribe(bus eventbus.Bus) {
	bus.Subscribe(events.NewFlipKeyID, func(e eventbus.Event) {
		newFlipKeyEvent := e.(*events.NewFlipKeyEvent)
		w := &flipKeyWrapper{
			key:  newFlipKeyEvent.Key,
			time: new(big.Int).SetInt64(time.Now().Unix()),
		}
		indexer.flipKeyChan <- w
	})

	bus.Subscribe(events.NewTxEventID, func(e eventbus.Event) {
		newTxEvent := e.(*events.NewTxEvent)
		if newTxEvent.Tx.Type != types.SubmitAnswersHashTx {
			return
		}
		w := &txWrapper{
			tx:   newTxEvent.Tx,
			time: new(big.Int).SetInt64(time.Now().Unix()),
		}
		indexer.answerHashTxChan <- w
	})
}

func (indexer *Indexer) listenFlipKeys() {
	for {
		flipKey := <-indexer.flipKeyChan
		sender, err := types.SenderFlipKey(flipKey.key)
		if err != nil {
			indexer.log.Error(errors.Wrapf(err, "Unable to define flip key (%v) sender", flipKey.key).Error())
			continue
		}
		flipKeyTimestamp := &db.MemPoolActionTimestamp{
			Address: conversion.ConvertAddress(sender),
			Epoch:   uint64(flipKey.key.Epoch),
			Time:    flipKey.time,
		}
		indexer.addFlipKeyTimestamp(flipKeyTimestamp)
	}
}

func (indexer *Indexer) listenAnswersHashTxs() {
	for {
		tx := <-indexer.answerHashTxChan

		sender, err := types.Sender(tx.tx)
		if err != nil {
			indexer.log.Error(errors.Wrapf(err, "Unable to define tx (%v) sender", tx.tx.Hash().Hex()).Error())
			continue
		}
		txTimestamp := &db.MemPoolActionTimestamp{
			Address: conversion.ConvertAddress(sender),
			Epoch:   uint64(tx.tx.Epoch),
			Time:    tx.time,
		}
		indexer.addAnswersHashTxTimestamp(txTimestamp)
	}
}

func (indexer *Indexer) addFlipKeyTimestamp(timestamp *db.MemPoolActionTimestamp) {
	indexer.cache.fktMutex.Lock()
	defer indexer.cache.fktMutex.Unlock()

	indexer.cache.flipKeyTimestamps = append(indexer.cache.flipKeyTimestamps, timestamp)
	indexer.log.Trace(fmt.Sprintf("Got mem pool flip key timestamp: %v", timestamp))
}

func (indexer *Indexer) addFlipKeyTimestamps(timestamps []*db.MemPoolActionTimestamp) {
	indexer.cache.fktMutex.Lock()
	defer indexer.cache.fktMutex.Unlock()

	indexer.cache.flipKeyTimestamps = append(indexer.cache.flipKeyTimestamps, timestamps...)
}

func (indexer *Indexer) getFlipKeyTimestampsToSave() []*db.MemPoolActionTimestamp {
	indexer.cache.fktMutex.Lock()
	defer indexer.cache.fktMutex.Unlock()

	result := indexer.cache.flipKeyTimestamps
	indexer.cache.flipKeyTimestamps = nil
	return result
}

func (indexer *Indexer) addAnswersHashTxTimestamp(timestamp *db.MemPoolActionTimestamp) {
	indexer.cache.ahttMutex.Lock()
	defer indexer.cache.ahttMutex.Unlock()

	indexer.cache.answersHashTxTimestamps = append(indexer.cache.answersHashTxTimestamps, timestamp)
	indexer.log.Trace(fmt.Sprintf("Got mem pool answer hash tx timestamp: %v", timestamp))
}

func (indexer *Indexer) addAnswersHashTxTimestamps(timestamps []*db.MemPoolActionTimestamp) {
	indexer.cache.ahttMutex.Lock()
	defer indexer.cache.ahttMutex.Unlock()

	indexer.cache.answersHashTxTimestamps = append(indexer.cache.answersHashTxTimestamps, timestamps...)
}

func (indexer *Indexer) getAnswersHashTxTimestampsToSave() []*db.MemPoolActionTimestamp {
	indexer.cache.ahttMutex.Lock()
	defer indexer.cache.ahttMutex.Unlock()

	result := indexer.cache.answersHashTxTimestamps
	indexer.cache.answersHashTxTimestamps = nil
	return result
}

func (indexer *Indexer) saveDataLoop() {
	for {
		time.Sleep(time.Second * 10)
		indexer.saveData()
	}
}

func (indexer *Indexer) saveData() {
	indexer.mutex.Lock()
	defer indexer.mutex.Unlock()

	data := &db.MemPoolData{
		FlipKeyTimestamps:       indexer.getFlipKeyTimestampsToSave(),
		AnswersHashTxTimestamps: indexer.getAnswersHashTxTimestampsToSave(),
	}
	if len(data.FlipKeyTimestamps)+len(data.AnswersHashTxTimestamps) == 0 {
		return
	}
	err := indexer.db.SaveMemPoolData(data)
	if err != nil {
		indexer.addFlipKeyTimestamps(data.FlipKeyTimestamps)
		indexer.addAnswersHashTxTimestamps(data.AnswersHashTxTimestamps)
		indexer.log.Error(errors.Wrapf(err,
			"Unable to save %d answers hash timestamps (current cache length: %d) "+
				"and %d flip key timestamps (current cache length: %d)",
			len(data.AnswersHashTxTimestamps), len(indexer.cache.answersHashTxTimestamps),
			len(data.FlipKeyTimestamps), len(indexer.cache.flipKeyTimestamps)).Error())
		return
	}
	indexer.log.Info(fmt.Sprintf("Saved %d flip key timestamps and %d answers hash timestamps",
		len(data.FlipKeyTimestamps),
		len(data.AnswersHashTxTimestamps)))
}

func (indexer *Indexer) Destroy() {
	indexer.saveData()
}
