/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"testing"

	"time"

	"os"
	"strings"

	mPkcs11 "github.com/miekg/pkcs11"
	"github.com/stretchr/testify/assert"
)

const (
	pin              = "98765432"
	label            = "ForFabric"
	label1           = "ForFabric1"
	label2           = "ForFabric2"
	allLibs          = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so,/usr/lib/softhsm/libsofthsm2.so,/usr/lib/s390x-linux-gnu/softhsm/libsofthsm2.so,/usr/lib/powerpc64le-linux-gnu/softhsm/libsofthsm2.so, /usr/local/Cellar/softhsm/2.1.0/lib/softhsm/libsofthsm2.so"
	ctxReloadTimeout = 2 * time.Second
)

var lib string

func TestContextHandleFeatures(t *testing.T) {

	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	//Test session
	session, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	//Test login
	err = handle.Login(session)
	assert.NoError(t, err)

	//test return/get session
	assert.Equal(t, 0, len(handle.sessions))
	handle.ReturnSession(session)
	assert.Equal(t, 1, len(handle.sessions))
	session = handle.GetSession()
	assert.Equal(t, 0, len(handle.sessions))
	handle.ReturnSession(session)
	assert.Equal(t, 1, len(handle.sessions))

	//add new 2 session to pool, externally
	session1, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	session2, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	handle.ReturnSession(session1)
	handle.ReturnSession(session2)

	assert.Equal(t, 3, len(handle.sessions))

	//empty pool
	session1 = handle.GetSession()
	session2 = handle.GetSession()
	session3 := handle.GetSession()

	assert.Equal(t, 0, len(handle.sessions))

	//even if pool is empty should be able to get session
	session4 := handle.GetSession()
	assert.Equal(t, 0, len(handle.sessions))

	//return all sessions to pool
	handle.ReturnSession(session1)
	handle.ReturnSession(session2)
	handle.ReturnSession(session3)
	handle.ReturnSession(session4)
	assert.Equal(t, 4, len(handle.sessions))

	//reset session pool after test
	handle.sessions = make(chan mPkcs11.SessionHandle, handle.opts.sessionCacheSize)

	//force reload
	handle, err = ReloadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

}

func TestMultipleContextHandleInstances(t *testing.T) {

	testSessions := func(handle *ContextHandle) {
		//Test session
		session, err := handle.OpenSession()
		assert.NoError(t, err)
		assert.True(t, session > 0)

		//Test login
		err = handle.Login(session)
		assert.NoError(t, err)

		//test return/get session
		assert.Equal(t, 0, len(handle.sessions))
		handle.ReturnSession(session)
		assert.Equal(t, 1, len(handle.sessions))
		session = handle.GetSession()
		assert.Equal(t, 0, len(handle.sessions))
		handle.ReturnSession(session)
		assert.Equal(t, 1, len(handle.sessions))

		//add new 2 session to pool, externally
		session1, err := handle.OpenSession()
		assert.NoError(t, err)
		assert.True(t, session > 0)

		session2, err := handle.OpenSession()
		assert.NoError(t, err)
		assert.True(t, session > 0)

		handle.ReturnSession(session1)
		handle.ReturnSession(session2)

		assert.Equal(t, 3, len(handle.sessions))

		//empty pool
		session1 = handle.GetSession()
		session2 = handle.GetSession()
		session3 := handle.GetSession()

		assert.Equal(t, 0, len(handle.sessions))

		//even if pool is empty should be able to get session
		session4 := handle.GetSession()
		assert.Equal(t, 0, len(handle.sessions))

		//return all sessions to pool
		handle.ReturnSession(session1)
		handle.ReturnSession(session2)
		handle.ReturnSession(session3)
		handle.ReturnSession(session4)
		assert.Equal(t, 4, len(handle.sessions))

		//reset session pool after test
		handle.sessions = make(chan mPkcs11.SessionHandle, handle.opts.sessionCacheSize)
	}

	handle1, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle1)
	assert.NotNil(t, handle1.ctx)
	assert.Equal(t, handle1.lib, lib)
	assert.Equal(t, handle1.label, label)
	assert.Equal(t, handle1.pin, pin)
	testSessions(handle1)

	handle2, err := LoadPKCS11ContextHandle(lib, label1, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle2)
	assert.NotNil(t, handle2.ctx)
	assert.Equal(t, handle2.lib, lib)
	assert.Equal(t, handle2.label, label1)
	assert.Equal(t, handle2.pin, pin)
	testSessions(handle2)

	//different label means different slot
	assert.NotEqual(t, handle1.slot, handle2.slot)

	//get session each from handle1 & 2
	session1 := handle1.GetSession()
	session2 := handle2.GetSession()

	//return them back to opposite handlers
	handle1.ReturnSession(session2)
	handle2.ReturnSession(session1)

	//Test if sessions are still valid(since lib/pin are same)
	assert.Equal(t, session2, handle1.GetSession())
	assert.Equal(t, session1, handle2.GetSession())
}

func TestContextHandleInstance(t *testing.T) {

	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	defer func() {
		//reload pkcs11 context for other tests to succeed
		handle, err := ReloadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		assert.NotNil(t, handle.ctx)
		assert.Equal(t, handle.lib, lib)
		assert.Equal(t, handle.label, label)
		assert.Equal(t, handle.pin, pin)
	}()

	//destroy pkcs11 ctx inside
	handle.ctx.Destroy()

	//test if this impacted other 'LoadPKCS11ContextHandle' calls
	t.Run("test corrupted context handler instance", func(t *testing.T) {

		//get it again
		handle1, err := LoadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handle1)

		//Open session should fail it is destroyed by previous instance
		err = handle1.ctx.CloseAllSessions(handle.slot)
		assert.Error(t, err, mPkcs11.CKR_CRYPTOKI_NOT_INITIALIZED)
	})

}

func TestContextHandleOpts(t *testing.T) {

	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin, WithOpenSessionRetry(10), WithSessionCacheSize(2))
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	//get 4 sessions
	session1 := handle.GetSession()
	session2 := handle.GetSession()
	session3 := handle.GetSession()
	session4 := handle.GetSession()

	//return all 4, but pool size is 2, so last 2 will sessions will be closed
	handle.ReturnSession(session1)
	handle.ReturnSession(session2)
	handle.ReturnSession(session3)
	handle.ReturnSession(session4)

	//session1 should be valid
	_, e := handle.ctx.GetSessionInfo(session1)
	assert.NoError(t, e)

	//session2 should be valid
	_, e = handle.ctx.GetSessionInfo(session2)
	assert.NoError(t, e)

	//session3 should be closed
	_, e = handle.ctx.GetSessionInfo(session3)
	assert.Equal(t, mPkcs11.Error(mPkcs11.CKR_SESSION_HANDLE_INVALID), e)

	//session4 should be closed
	_, e = handle.ctx.GetSessionInfo(session4)
	assert.Equal(t, mPkcs11.Error(mPkcs11.CKR_SESSION_HANDLE_INVALID), e)

}

func TestContextHandleCacheCollision(t *testing.T) {

	const pin1 = "22334455"

	//get context handler-1
	handle1, err := LoadPKCS11ContextHandle(lib, label2, pin, WithOpenSessionRetry(10), WithSessionCacheSize(2))
	assert.NoError(t, err)
	assert.NotNil(t, handle1)
	assert.NotNil(t, handle1.ctx)
	assert.Equal(t, handle1.lib, lib)
	assert.Equal(t, handle1.label, label2)
	assert.Equal(t, handle1.pin, pin)

	//get context handler-2
	handle2, err := LoadPKCS11ContextHandle(lib, label2, pin1, WithOpenSessionRetry(10), WithSessionCacheSize(2))
	assert.NoError(t, err)
	assert.NotNil(t, handle2)
	assert.NotNil(t, handle2.ctx)
	assert.Equal(t, handle2.lib, lib)
	assert.Equal(t, handle2.label, label2)

	//collision happens when different PINs used under same label and lib
	assert.Equal(t, handle2.pin, handle1.pin)

	//To fix, use connection name to distinguish instances in cache
	handle2, err = LoadPKCS11ContextHandle(lib, label2, pin1, WithOpenSessionRetry(10), WithSessionCacheSize(2), WithConnectionName("connection-2"))
	assert.NoError(t, err)
	assert.NotNil(t, handle2)
	assert.NotNil(t, handle2.ctx)
	assert.Equal(t, handle2.lib, lib)
	assert.Equal(t, handle2.label, label2)
	assert.Equal(t, handle2.pin, pin1)
}

func TestContextHandleCommonInstance(t *testing.T) {
	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	oldCtx := handle.ctx
	for i := 0; i < 20; i++ {
		handleX, err := LoadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handleX)
		//Should be same instance, for same set of lib, label, pin
		assert.Equal(t, oldCtx, handleX.ctx)
	}
}

func TestContextRefreshOnInvalidSession(t *testing.T) {

	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	//get session
	session := handle.GetSession()

	//close this session and return it, validation on return session should stop it
	handle.ctx.CloseSession(session)
	handle.ReturnSession(session)
	//session pool unchanged, since returned session was invalid
	assert.Equal(t, 0, len(handle.sessions))

	//just for test manually add it into pool
	handle.sessions <- session
	assert.Equal(t, 1, len(handle.sessions))

	oldCtx := handle.ctx
	assert.Equal(t, oldCtx, handle.ctx)

	//get session again, now ctx should be refreshed
	ch := make(chan struct{}, 1)
	handle.NotifyCtxReload(ch)
	session = handle.GetSession()
	assert.NotEqual(t, oldCtx, handle.ctx)
	assert.NotNil(t, session)

	var receivedNotification bool
	select {
	case <-ch:
		receivedNotification = true
	case <-time.After(ctxReloadTimeout):
		t.Fatal("couldn't get notification on ctx update")
	}

	assert.True(t, receivedNotification)
	//reset session pool after test
	handle.sessions = make(chan mPkcs11.SessionHandle, handle.opts.sessionCacheSize)
}

func TestSessionsFromDifferentPKCS11Ctx(t *testing.T) {

	//Testing if session created by a ctx can be validated by of some other ctx created using same lib/label/pin
	ctxAndSession := func(label string) (*mPkcs11.Ctx, mPkcs11.SessionHandle) {
		ctx := mPkcs11.New(lib)
		assert.NotNil(t, ctx)
		err := ctx.Initialize()
		assert.False(t, err != nil && err != mPkcs11.Error(mPkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED))

		var found bool
		var slot uint
		//get all slots
		slots, err := ctx.GetSlotList(true)
		if err != nil {
			t.Fatal("Failed to get slot list for recreated context:", err)
		}

		//find slot matching label
		for _, s := range slots {
			info, err := ctx.GetTokenInfo(s)
			if err != nil {
				continue
			}
			if label == info.Label {
				found = true
				slot = s
				break
			}
		}

		assert.True(t, found)

		session, err := createNewSession(ctx, slot)
		assert.NoError(t, err)
		return ctx, session
	}

	ctx1, session1 := ctxAndSession(label)
	ctx2, session2 := ctxAndSession(label)

	//ctx2 validating session1 from ctx1
	sessionInfo, err := ctx2.GetSessionInfo(session1)
	assert.NoError(t, err)
	assert.NotNil(t, sessionInfo)

	//ctx1 validating session2 from ctx2
	sessionInfo, err = ctx1.GetSessionInfo(session2)
	assert.NoError(t, err)
	assert.NotNil(t, sessionInfo)

	//test between different slot/label
	ctx3, session3 := ctxAndSession(label1)

	sessionInfo, err = ctx1.GetSessionInfo(session3)
	assert.NoError(t, err)
	assert.NotNil(t, sessionInfo)

	sessionInfo, err = ctx3.GetSessionInfo(session1)
	assert.NoError(t, err)
	assert.NotNil(t, sessionInfo)
}

func TestContextHandlerConcurrency(t *testing.T) {

	handlersCount := 5
	concurrency := 5000

	var err error
	handlers := make([]*ContextHandle, handlersCount)
	for i := 0; i < handlersCount; i++ {
		handlers[i], err = LoadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
	}

	testDone := make(chan bool)

	runTest := func(handle *ContextHandle) {
		session1 := handle.GetSession()
		assert.True(t, session1 > 0)

		session2 := handle.GetSession()
		assert.True(t, session2 > 0)

		handle.ReturnSession(session1)
		handle.ReturnSession(session2)

		session1 = handle.GetSession()
		assert.True(t, session1 > 0)

		session2 = handle.GetSession()
		assert.True(t, session2 > 0)

		handle.ReturnSession(session1)
		handle.ReturnSession(session2)

		session3, err := handle.OpenSession()
		assert.NoError(t, err)
		assert.True(t, session3 > 0)

		err = handle.Login(session3)
		assert.NoError(t, err)

		testDone <- true
	}

	for i := 0; i < concurrency; i++ {
		go runTest(handlers[i%handlersCount])
	}

	testsReturned := 0
	for i := 0; i < concurrency; i++ {
		select {
		case b := <-testDone:
			assert.True(t, b)
			testsReturned++
		case <-time.After(time.Second * 10):
			t.Fatalf("Timed out waiting for test %d", i)
		}
	}

	assert.Equal(t, concurrency, testsReturned)
}

func TestSessionHandle(t *testing.T) {
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	//make sure session pool is empty
	for len(handle.sessions) > 0 {
		<-handle.sessions
	}

	//get session
	session := handle.GetSession()
	err = isEmpty(session)
	assert.NoError(t, err)

	//tamper pin, so that get session should fail
	pinBackup := handle.pin
	slotBackup := handle.slot

	handle.pin = "9999"
	handle.slot = 8888

	//get session should fail
	session = handle.GetSession()
	err = isEmpty(session)
	assert.Error(t, err)

	//try again
	session = handle.GetSession()
	err = isEmpty(session)
	assert.Error(t, err)

	//recover tampered pin and slot
	handle.pin = pinBackup
	handle.slot = slotBackup

	//try again
	session = handle.GetSession()
	err = isEmpty(session)
	assert.NoError(t, err)
}

func TestGetSessionResilience(t *testing.T) {

	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	//make sure session pool is empty
	for len(handle.sessions) > 0 {
		<-handle.sessions
	}

	//get session
	session := handle.GetSession()
	err = isEmpty(session)
	assert.NoError(t, err)

	//tamper pin, so that get session should fail
	pinBackup := handle.pin
	slotBackup := handle.slot

	resetPinAndSlot := func() {
		handle.lock.Lock()
		defer handle.lock.Unlock()
		handle.pin = pinBackup
		handle.slot = slotBackup
	}

	handle.pin = "1111"
	handle.slot = 8888

	//make sure get session should fail
	session = handle.GetSession()
	err = isEmpty(session)
	assert.Error(t, err)

	const retry = 5
	interval := 200 * time.Millisecond
	done := make(chan bool)

	// launch get session with retry
	go func() {
		for i := 0; i < retry; i++ {
			session = handle.GetSession()
			if err := isEmpty(session); err == nil {
				done <- true
				break
			}
			time.Sleep(interval)
			continue
		}
	}()

	time.Sleep(500 * time.Millisecond)

	go resetPinAndSlot()

	select {
	case <-done:
		t.Log("session recovered")
		handle.pin = pinBackup
		handle.slot = slotBackup
	case <-time.After(ctxReloadTimeout):
		t.Fatal("couldn't recover session")
	}

}

func TestContextHandleInRecovery(t *testing.T) {
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	// make sure session pool is empty
	for len(handle.sessions) > 0 {
		<-handle.sessions
	}

	// get session
	session := handle.GetSession()
	err = isEmpty(session)
	assert.NoError(t, err)

	// tamper pin, so that get session should fail
	libBackup := handle.lib
	slotBackup := handle.slot

	resetLib := func() {
		handle.lock.Lock()
		defer handle.lock.Unlock()
		handle.lib = libBackup
		handle.slot = slotBackup
	}
	defer resetLib()

	// recreate failure scenario
	handle.lib = handle.lib + "_invalid"
	handle.slot = 9999

	// get session, should fail
	session = handle.GetSession()
	err = isEmpty(session)
	assert.Error(t, err)

	assert.True(t, handle.recovery)

	verifyRecoveryError := func(err error) {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pkcs11 ctx is under recovery, try again later")
	}

	_, err = handle.OpenSession()
	verifyRecoveryError(err)

	err = handle.Login(4)
	verifyRecoveryError(err)

	lenBefore := len(handle.sessions)
	handle.ReturnSession(4)
	assert.Equal(t, lenBefore, len(handle.sessions))

	_, err = handle.GetAttributeValue(4, 4, nil)
	verifyRecoveryError(err)

	err = handle.SetAttributeValue(4, 4, nil)
	verifyRecoveryError(err)

	_, _, err = handle.GenerateKeyPair(4, nil, nil, nil)
	verifyRecoveryError(err)
}

func TestMain(m *testing.M) {

	possibilities := strings.Split(allLibs, ",")
	for _, path := range possibilities {
		trimpath := strings.TrimSpace(path)
		if _, err := os.Stat(trimpath); !os.IsNotExist(err) {
			lib = trimpath
			break
		}
	}

	os.Exit(m.Run())
}
