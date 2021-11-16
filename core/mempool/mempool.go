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
	"sync"
	"time"
)

const queueSize = 20000

type Indexer struct {
	db                         db.Accessor
	flipPrivateKeysPackageChan chan *flipPrivateKeysPackageWrapper
	flipKeyChan                chan *flipKeyWrapper
	answerHashTxChan           chan *txWrapper
	shortAnswersTxChan         chan *txWrapper
	cache                      *indexerCache
	log                        log.Logger
	mutex                      sync.Mutex
}

type indexerCache struct {
	flipPrivateKeysPackageTimestampsMutex sync.Mutex
	flipPrivateKeysPackageTimestamps      []*db.MemPoolActionTimestamp
	fktMutex                              sync.Mutex
	flipKeyTimestamps                     []*db.MemPoolActionTimestamp
	ahttMutex                             sync.Mutex
	answersHashTxTimestamps               []*db.MemPoolActionTimestamp
	shortAnswersTxTimestampsMutex         sync.Mutex
	shortAnswersTxTimestamps              []*db.MemPoolActionTimestamp
}

type flipPrivateKeysPackageWrapper struct {
	key  *types.PrivateFlipKeysPackage
	time int64
}

type flipKeyWrapper struct {
	key  *types.PublicFlipKey
	time int64
}

type txWrapper struct {
	tx   *types.Transaction
	time int64
}

func NewIndexer(db db.Accessor, log log.Logger) *Indexer {
	return &Indexer{
		db:                         db,
		flipPrivateKeysPackageChan: make(chan *flipPrivateKeysPackageWrapper, queueSize),
		flipKeyChan:                make(chan *flipKeyWrapper, queueSize),
		answerHashTxChan:           make(chan *txWrapper, queueSize),
		shortAnswersTxChan:         make(chan *txWrapper, queueSize),
		log:                        log,
		cache:                      &indexerCache{},
	}
}

func (indexer *Indexer) Initialize(bus eventbus.Bus) {
	indexer.subscribe(bus)
	go indexer.saveDataLoop()
	go indexer.listenFlipPrivateKeysPackages()
	go indexer.listenFlipKeys()
	go indexer.listenAnswersHashTxs()
	go indexer.listenShortAnswersTxs()
}

func (indexer *Indexer) subscribe(bus eventbus.Bus) {
	bus.Subscribe(events.NewFlipKeysPackageID, func(e eventbus.Event) {
		newFlipKeysPackageEvent := e.(*events.NewFlipKeysPackageEvent)
		w := &flipPrivateKeysPackageWrapper{
			key:  newFlipKeysPackageEvent.Key,
			time: time.Now().UTC().Unix(),
		}
		indexer.flipPrivateKeysPackageChan <- w
	})

	bus.Subscribe(events.NewFlipKeyID, func(e eventbus.Event) {
		newFlipKeyEvent := e.(*events.NewFlipKeyEvent)
		w := &flipKeyWrapper{
			key:  newFlipKeyEvent.Key,
			time: time.Now().UTC().Unix(),
		}
		indexer.flipKeyChan <- w
	})

	bus.Subscribe(events.NewTxEventID, func(e eventbus.Event) {
		newTxEvent := e.(*events.NewTxEvent)
		if newTxEvent.Tx.Type == types.SubmitAnswersHashTx {
			w := &txWrapper{
				tx:   newTxEvent.Tx,
				time: time.Now().UTC().Unix(),
			}
			indexer.answerHashTxChan <- w
			return
		}
		if newTxEvent.Tx.Type == types.SubmitShortAnswersTx {
			w := &txWrapper{
				tx:   newTxEvent.Tx,
				time: time.Now().UTC().Unix(),
			}
			indexer.shortAnswersTxChan <- w
			return
		}
	})
}

func (indexer *Indexer) listenFlipPrivateKeysPackages() {
	for {
		flipPrivateKeysPackage := <-indexer.flipPrivateKeysPackageChan
		sender, err := types.SenderFlipKeysPackage(flipPrivateKeysPackage.key)
		if err != nil {
			indexer.log.Error(errors.Wrapf(err, "Unable to define flip keys package (%v) sender", flipPrivateKeysPackage.key).Error())
			continue
		}
		flipKeyTimestamp := &db.MemPoolActionTimestamp{
			Address: conversion.ConvertAddress(sender),
			Epoch:   uint64(flipPrivateKeysPackage.key.Epoch),
			Time:    flipPrivateKeysPackage.time,
		}
		indexer.addFlipPrivateKeysPackageTimestamp(flipKeyTimestamp)
	}
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

func (indexer *Indexer) listenShortAnswersTxs() {
	for {
		tx := <-indexer.shortAnswersTxChan

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
		indexer.addShortAnswersTxTimestamp(txTimestamp)
	}
}

func (indexer *Indexer) addFlipPrivateKeysPackageTimestamp(timestamp *db.MemPoolActionTimestamp) {
	indexer.cache.flipPrivateKeysPackageTimestampsMutex.Lock()
	defer indexer.cache.flipPrivateKeysPackageTimestampsMutex.Unlock()

	indexer.cache.flipPrivateKeysPackageTimestamps = append(indexer.cache.flipPrivateKeysPackageTimestamps, timestamp)
	indexer.log.Trace(fmt.Sprintf("Got mem pool flip keys package timestamp: %v", timestamp))
}

func (indexer *Indexer) addFlipPrivateKeysPackageTimestamps(timestamps []*db.MemPoolActionTimestamp) {
	indexer.cache.flipPrivateKeysPackageTimestampsMutex.Lock()
	defer indexer.cache.flipPrivateKeysPackageTimestampsMutex.Unlock()

	indexer.cache.flipPrivateKeysPackageTimestamps = append(indexer.cache.flipPrivateKeysPackageTimestamps, timestamps...)
}

func (indexer *Indexer) getFlipKeysPackageTimestampsToSave() []*db.MemPoolActionTimestamp {
	indexer.cache.flipPrivateKeysPackageTimestampsMutex.Lock()
	defer indexer.cache.flipPrivateKeysPackageTimestampsMutex.Unlock()

	result := indexer.cache.flipPrivateKeysPackageTimestamps
	indexer.cache.flipPrivateKeysPackageTimestamps = nil
	return result
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

func (indexer *Indexer) addShortAnswersTxTimestamp(timestamp *db.MemPoolActionTimestamp) {
	indexer.cache.shortAnswersTxTimestampsMutex.Lock()
	defer indexer.cache.shortAnswersTxTimestampsMutex.Unlock()

	indexer.cache.shortAnswersTxTimestamps = append(indexer.cache.shortAnswersTxTimestamps, timestamp)
	indexer.log.Trace(fmt.Sprintf("Got mem pool short answers tx timestamp: %v", timestamp))
}

func (indexer *Indexer) addAnswersHashTxTimestamps(timestamps []*db.MemPoolActionTimestamp) {
	indexer.cache.ahttMutex.Lock()
	defer indexer.cache.ahttMutex.Unlock()

	indexer.cache.answersHashTxTimestamps = append(indexer.cache.answersHashTxTimestamps, timestamps...)
}

func (indexer *Indexer) addShortAnswersTxTimestamps(timestamps []*db.MemPoolActionTimestamp) {
	indexer.cache.shortAnswersTxTimestampsMutex.Lock()
	defer indexer.cache.shortAnswersTxTimestampsMutex.Unlock()

	indexer.cache.shortAnswersTxTimestamps = append(indexer.cache.shortAnswersTxTimestamps, timestamps...)
}

func (indexer *Indexer) getAnswersHashTxTimestampsToSave() []*db.MemPoolActionTimestamp {
	indexer.cache.ahttMutex.Lock()
	defer indexer.cache.ahttMutex.Unlock()

	result := indexer.cache.answersHashTxTimestamps
	indexer.cache.answersHashTxTimestamps = nil
	return result
}

func (indexer *Indexer) getShortAnswersTxTimestampsToSave() []*db.MemPoolActionTimestamp {
	indexer.cache.shortAnswersTxTimestampsMutex.Lock()
	defer indexer.cache.shortAnswersTxTimestampsMutex.Unlock()

	result := indexer.cache.shortAnswersTxTimestamps
	indexer.cache.shortAnswersTxTimestamps = nil
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
		FlipPrivateKeysPackageTimestamps: indexer.getFlipKeysPackageTimestampsToSave(),
		FlipKeyTimestamps:                indexer.getFlipKeyTimestampsToSave(),
		AnswersHashTxTimestamps:          indexer.getAnswersHashTxTimestampsToSave(),
		ShortAnswersTxTimestamps:         indexer.getShortAnswersTxTimestampsToSave(),
	}
	if len(data.FlipPrivateKeysPackageTimestamps)+
		len(data.FlipKeyTimestamps)+
		len(data.AnswersHashTxTimestamps)+
		len(data.ShortAnswersTxTimestamps) == 0 {
		return
	}
	start := time.Now()
	err := indexer.db.SaveMemPoolData(data)
	duration := time.Since(start)
	if err != nil {
		indexer.addFlipPrivateKeysPackageTimestamps(data.FlipPrivateKeysPackageTimestamps)
		indexer.addFlipKeyTimestamps(data.FlipKeyTimestamps)
		indexer.addAnswersHashTxTimestamps(data.AnswersHashTxTimestamps)
		indexer.addShortAnswersTxTimestamps(data.ShortAnswersTxTimestamps)
		indexer.log.Error(errors.Wrapf(err,
			"Unable to save %d flip keys package timestamps (current cache length: %d) "+
				", %d flip key timestamps (current cache length: %d)"+
				", %d answers hash timestamps (current cache length: %d)"+
				", %d short answers timestamps (current cache length: %d)",
			len(data.FlipPrivateKeysPackageTimestamps), len(indexer.cache.flipPrivateKeysPackageTimestamps),
			len(data.FlipKeyTimestamps), len(indexer.cache.flipKeyTimestamps),
			len(data.AnswersHashTxTimestamps), len(indexer.cache.answersHashTxTimestamps),
			len(data.ShortAnswersTxTimestamps), len(indexer.cache.shortAnswersTxTimestamps),
		).Error(), "d", duration)
		return
	}
	indexer.log.Info(fmt.Sprintf("Saved timestamps: %d flip keys packages, %d flip keys, %d answers hashes, %d short answers",
		len(data.FlipPrivateKeysPackageTimestamps),
		len(data.FlipKeyTimestamps),
		len(data.AnswersHashTxTimestamps),
		len(data.ShortAnswersTxTimestamps),
	), "d", duration)
}

func (indexer *Indexer) Destroy() {
	indexer.saveData()
}
