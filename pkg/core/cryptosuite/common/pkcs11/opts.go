/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

const (
	defaultSessionCacheSize = 10
	defaultOpenSessionRetry = 10
)

//ctxOpts options for conext handler
type ctxOpts struct {
	//sessionCacheSize size of session cache pool
	sessionCacheSize int
	//openSessionRetry number of retry for open session logic
	openSessionRetry int
}

//Options for PKCS11 ContextHandle
type Options func(opts *ctxOpts)

func getCtxOpts(opts ...Options) ctxOpts {
	ctxOpts := ctxOpts{}
	for _, option := range opts {
		option(&ctxOpts)
	}

	if ctxOpts.sessionCacheSize == 0 {
		ctxOpts.sessionCacheSize = defaultSessionCacheSize
	}

	if ctxOpts.openSessionRetry == 0 {
		ctxOpts.openSessionRetry = defaultOpenSessionRetry
	}

	return ctxOpts
}

//WithSessionCacheSize size of session cache pool
func WithSessionCacheSize(size int) Options {
	return func(o *ctxOpts) {
		o.sessionCacheSize = size
	}
}

//WithOpenSessionRetry number of retry for open session logic
func WithOpenSessionRetry(count int) Options {
	return func(o *ctxOpts) {
		o.openSessionRetry = count
	}
}
