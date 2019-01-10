// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
)

func callQuery(cc *Client, request Request, options ...RequestOption) (Response, error) {
	meterLabels := []string{
		"chaincode", request.ChaincodeID,
		"Fcn", request.Fcn,
	}
	cc.metrics.QueriesReceived.With(meterLabels...).Add(1)
	startTime := time.Now()
	r, err := cc.InvokeHandler(invoke.NewQueryHandler(), request, options...)
	if err != nil {
		if s, ok := err.(*status.Status); ok {
			if s.Code == status.Timeout.ToInt32() {
				meterLabels = append(meterLabels, "fail", "timeout")
				cc.metrics.QueryTimeouts.With(meterLabels...).Add(1)
				return r, err
			}
			meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Group:%s - Code:%d", s.Group.String(), s.Code))
			cc.metrics.QueriesFailed.With(meterLabels...).Add(1)
			return r, err
		}
		meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Generic: %s", err))
		cc.metrics.QueriesFailed.With(meterLabels...).Add(1)
		return r, err
	}
	cc.metrics.QueryDuration.With(meterLabels...).Observe(time.Since(startTime).Seconds())
	return r, err
}

func callExecute(cc *Client, request Request, options ...RequestOption) (Response, error) {
	meterLabels := []string{
		"chaincode", request.ChaincodeID,
		"Fcn", request.Fcn,
	}
	cc.metrics.ExecutionsReceived.With(meterLabels...).Add(1)
	startTime := time.Now()
	r, err := cc.InvokeHandler(invoke.NewExecuteHandler(), request, options...)
	if err != nil {
		if s, ok := err.(*status.Status); ok {
			if s.Code == status.Timeout.ToInt32() {
				meterLabels = append(meterLabels, "fail", "timeout")
				cc.metrics.ExecutionTimeouts.With(meterLabels...).Add(1)
				return r, err
			}
			meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Group:%s - Code:%d", s.Group.String(), s.Code))
			cc.metrics.ExecutionsFailed.With(meterLabels...).Add(1)
			return r, err
		}
		meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Generic: %s", err))
		cc.metrics.ExecutionsFailed.With(meterLabels...).Add(1)
		return r, err
	}

	cc.metrics.ExecutionDuration.With(meterLabels...).Observe(time.Since(startTime).Seconds())
	return r, err
}
