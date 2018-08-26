/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

type params struct {
	eventConsumerBufferSize           uint
	eventConsumerTimeout              time.Duration
	initialLastBlockNum               uint64
	initialBlockRegistrations         []*BlockReg
	initialFilteredBlockRegistrations []*FilteredBlockReg
	initialCCRegistrations            []*ChaincodeReg
	initialTxStatusRegistrations      []*TxStatusReg
}

func defaultParams() *params {
	return &params{
		eventConsumerBufferSize: 100,
		eventConsumerTimeout:    500 * time.Millisecond,
	}
}

// WithEventConsumerBufferSize sets the size of the registered consumer's event channel.
func WithEventConsumerBufferSize(value uint) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(eventConsumerBufferSizeSetter); ok {
			setter.SetEventConsumerBufferSize(value)
		}
	}
}

// WithEventConsumerTimeout is the timeout when sending events to a registered consumer.
// If < 0, if buffer full, unblocks immediately and does not send.
// If 0, if buffer full, will block and guarantee the event will be sent out.
// If > 0, if buffer full, blocks util timeout.
func WithEventConsumerTimeout(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(eventEventConsumerTimeoutSetter); ok {
			setter.SetEventConsumerTimeout(value)
		}
	}
}

// WithSnapshot sets the given TxStatus registrations.
func WithSnapshot(value fab.EventSnapshot) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(snapshotSetter); ok {
			err := setter.SetSnapshot(value)
			if err != nil {
				logger.Errorf("Unable to set snapshot: %s", err)
			}
		}
	}
}

type eventConsumerBufferSizeSetter interface {
	SetEventConsumerBufferSize(value uint)
}

type eventEventConsumerTimeoutSetter interface {
	SetEventConsumerTimeout(value time.Duration)
}

func (p *params) SetEventConsumerBufferSize(value uint) {
	logger.Debugf("EventConsumerBufferSize: %d", value)
	p.eventConsumerBufferSize = value
}

func (p *params) SetEventConsumerTimeout(value time.Duration) {
	logger.Debugf("EventConsumerTimeout: %s", value)
	p.eventConsumerTimeout = value
}

type snapshotSetter interface {
	SetSnapshot(value fab.EventSnapshot) error
}

func (p *params) SetSnapshot(value fab.EventSnapshot) error {
	logger.Debugf("Snapshot: %#v", value)

	bRegistrations, err := asBlockRegistrations(value.BlockRegistrations())
	if err != nil {
		return err
	}
	fbRegistrations, err := asFBlockRegistrations(value.FilteredBlockRegistrations())
	if err != nil {
		return err
	}
	ccRegistrations, err := asCCRegistrations(value.CCRegistrations())
	if err != nil {
		return err
	}
	txRegistrations, err := asTxRegistrations(value.TxStatusRegistrations())
	if err != nil {
		return err
	}

	p.initialLastBlockNum = value.LastBlockReceived()
	p.initialBlockRegistrations = bRegistrations
	p.initialFilteredBlockRegistrations = fbRegistrations
	p.initialCCRegistrations = ccRegistrations
	p.initialTxStatusRegistrations = txRegistrations

	return nil
}

func asBlockRegistrations(registrations []fab.Registration) ([]*BlockReg, error) {
	var bRegistrations []*BlockReg
	for _, reg := range registrations {
		breg, ok := reg.(*BlockReg)
		if !ok {
			return nil, errors.New("invalid block registration")
		}
		bRegistrations = append(bRegistrations, breg)
	}
	return bRegistrations, nil
}

func asFBlockRegistrations(registrations []fab.Registration) ([]*FilteredBlockReg, error) {
	var fbRegistrations []*FilteredBlockReg
	for _, reg := range registrations {
		fbreg, ok := reg.(*FilteredBlockReg)
		if !ok {
			return nil, errors.New("invalid filtered block registration")
		}
		fbRegistrations = append(fbRegistrations, fbreg)
	}
	return fbRegistrations, nil
}

func asCCRegistrations(registrations []fab.Registration) ([]*ChaincodeReg, error) {
	var ccRegistrations []*ChaincodeReg
	for _, reg := range registrations {
		ccreg, ok := reg.(*ChaincodeReg)
		if !ok {
			return nil, errors.New("invalid chaincode registration")
		}
		ccRegistrations = append(ccRegistrations, ccreg)
	}
	return ccRegistrations, nil
}

func asTxRegistrations(registrations []fab.Registration) ([]*TxStatusReg, error) {
	var txRegistrations []*TxStatusReg
	for _, reg := range registrations {
		txreg, ok := reg.(*TxStatusReg)
		if !ok {
			return nil, errors.New("invalid TxStatus registration")
		}
		txRegistrations = append(txRegistrations, txreg)
	}
	return txRegistrations, nil
}
