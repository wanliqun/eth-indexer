// Copyright 2018 The eth-indexer Authors
// This file is part of the eth-indexer library.
//
// The eth-indexer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The eth-indexer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the eth-indexer library. If not, see <http://www.gnu.org/licenses/>.

package subscription

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/getamis/sirius/log"
	"github.com/jinzhu/gorm"

	idxCommon "github.com/getamis/eth-indexer/common"
	"github.com/getamis/eth-indexer/model"
)

const (
	ErrCodeDuplicateKey uint16 = 1062
)

//go:generate mockery -name Store
type Store interface {
	BatchInsert(subs []*model.Subscription) ([]common.Address, error)
	BatchUpdateBlockNumber(blockNumber int64, addrs [][]byte) error
	Find(blockNumber int64) (result []*model.Subscription, err error)
	// FindOldSubscriptions find old subscriptions by addresses
	FindOldSubscriptions(addrs [][]byte) (result []*model.Subscription, err error)
	FindByGroup(groupID int64, query *model.QueryParameters) (result []*model.Subscription, total uint64, err error)
	ListOldSubscriptions(query *model.QueryParameters) (result []*model.Subscription, total uint64, err error)

	// Total balance
	InsertTotalBalance(data *model.TotalBalance) error
	FindTotalBalance(blockNumber int64, token common.Address, group int64) (result *model.TotalBalance, err error)

	Reset(from, to int64) error
}

type store struct {
	db *gorm.DB
}

func NewWithDB(db *gorm.DB) Store {
	return &store{
		db: db,
	}
}

func (t *store) BatchInsert(subs []*model.Subscription) (duplicated []common.Address, err error) {
	dbTx := t.db.Begin()
	defer func() {
		if err != nil {
			dbTx.Rollback()
			return
		}
		err = dbTx.Commit().Error
		if err != nil {
			log.Error("Failed to commit db", "err", err)
		}
	}()
	for _, sub := range subs {
		createErr := dbTx.Create(sub).Error
		if createErr != nil {
			if idxCommon.DuplicateError(createErr) {
				duplicated = append(duplicated, common.BytesToAddress(sub.Address))
			} else {
				return nil, createErr
			}
		}
	}
	return duplicated, nil
}

func (t *store) BatchUpdateBlockNumber(blockNumber int64, addrs [][]byte) error {
	if len(addrs) == 0 {
		return nil
	}
	return t.db.Model(model.Subscription{}).Where("address in (?)", addrs).Updates(map[string]interface{}{"block_number": blockNumber}).Error
}

func (t *store) Find(blockNumber int64) (result []*model.Subscription, err error) {
	err = t.db.Where("block_number = ?", blockNumber).Find(&result).Error
	return
}

func (t *store) FindOldSubscriptions(addrs [][]byte) (result []*model.Subscription, err error) {
	if len(addrs) == 0 {
		return []*model.Subscription{}, nil
	}

	err = t.db.Where("address in (?) AND block_number > 0", addrs).Find(&result).Error
	if err != nil {
		return
	}
	return
}

func (t *store) InsertTotalBalance(data *model.TotalBalance) error {
	return t.db.Create(data).Error
}

func (t *store) FindTotalBalance(blockNumber int64, token common.Address, group int64) (*model.TotalBalance, error) {
	result := &model.TotalBalance{}
	err := t.db.Where("block_number <= ? AND token = ? AND `group` = ?", blockNumber, token.Bytes(), group).Order("block_number DESC").Limit(1).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *store) Reset(from, to int64) error {
	// Set the block number of subscription to 0
	err := t.db.Table(model.Subscription{}.TableName()).Where("block_number >= ? AND block_number <= ?", from, to).UpdateColumn("block_number", 0).Error
	if err != nil {
		return err
	}
	// Delete total balances
	return t.db.Delete(model.TotalBalance{}, "block_number >= ? AND block_number <= ?", from, to).Error
}

func (t *store) FindByGroup(groupID int64, params *model.QueryParameters) ([]*model.Subscription, uint64, error) {
	return t.find(t.db.Model(&model.Subscription{}).Where(&model.Subscription{
		Group: groupID,
	}), params)
}

func (t *store) ListOldSubscriptions(params *model.QueryParameters) ([]*model.Subscription, uint64, error) {
	return t.find(t.db.Model(&model.Subscription{}).Where("block_number > 0"), params)
}

func (t *store) find(db *gorm.DB, params *model.QueryParameters) ([]*model.Subscription, uint64, error) {
	var total uint64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	start := (params.Page - 1) * params.Limit
	var result []*model.Subscription
	err = db.Offset(start).Limit(params.Limit).Order(params.OrderString()).Find(&result).Error
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}
