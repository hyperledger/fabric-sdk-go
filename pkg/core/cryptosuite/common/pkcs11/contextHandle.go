/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/cachebridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	mPkcs11 "github.com/miekg/pkcs11"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/core")
var ctxCache *lazycache.Cache
var once sync.Once
var errSlotIDChanged = fmt.Errorf("slot id changed")

//LoadPKCS11ContextHandle loads PKCS11 context handler instance from underlying cache
func LoadPKCS11ContextHandle(lib, label, pin string, opts ...Options) (*ContextHandle, error) {
	return getInstance(&pkcs11CtxCacheKey{lib: lib, label: label, pin: pin, opts: getCtxOpts(opts...)}, false)
}

//ReloadPKCS11ContextHandle deletes PKCS11 instance from underlying cache and loads new PKCS11 context  handler in cache
func ReloadPKCS11ContextHandle(lib, label, pin string, opts ...Options) (*ContextHandle, error) {
	return getInstance(&pkcs11CtxCacheKey{lib: lib, label: label, pin: pin, opts: getCtxOpts(opts...)}, true)
}

//LoadContextAndLogin loads Context handle and performs login
func LoadContextAndLogin(lib, pin, label string) (*ContextHandle, error) {
	logger.Debugf("Loading context and performing login for [%s-%s]", lib, label)
	pkcs11Context, err := LoadPKCS11ContextHandle(lib, label, pin)
	if err != nil {
		return nil, err
	}

	session, err := pkcs11Context.OpenSession()
	if err != nil {
		return nil, err
	}

	err = pkcs11Context.Login(session)
	if err != nil {
		return nil, err
	}

	pkcs11Context.ReturnSession(session)
	cachebridge.ClearAllSession()

	return pkcs11Context, err
}

//ContextHandle encapsulate basic mPkcs11.Ctx operations and manages sessions
type ContextHandle struct {
	ctx                *mPkcs11.Ctx
	slot               uint
	pin                string
	lib                string
	label              string
	sessions           chan mPkcs11.SessionHandle
	opts               ctxOpts
	reloadNotification chan struct{}
	lock               sync.RWMutex
	recovery           bool
}

// NotifyCtxReload registers a channel to get notification when underlying mPkcs11.Ctx is recreated
func (handle *ContextHandle) NotifyCtxReload(ch chan struct{}) {
	handle.reloadNotification = ch
}

//OpenSession opens a session between an application and a token.
func (handle *ContextHandle) OpenSession() (mPkcs11.SessionHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	err := handle.isInRecovery()
	if err != nil {
		return 0, err
	}

	return handle.ctx.OpenSession(handle.slot, mPkcs11.CKF_SERIAL_SESSION|mPkcs11.CKF_RW_SESSION)
}

// Login logs a user into a token
func (handle *ContextHandle) Login(session mPkcs11.SessionHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	if handle.pin == "" {
		return errors.New("No PIN set")
	}

	err := handle.isInRecovery()
	if err != nil {
		return err
	}

	err = handle.ctx.Login(session, mPkcs11.CKU_USER, handle.pin)
	if err != nil && err != mPkcs11.Error(mPkcs11.CKR_USER_ALREADY_LOGGED_IN) {
		return errors.Errorf("Login failed [%s]", err)
	}
	return nil
}

//ReturnSession returns session back into the session pool
//if pool is pull or session is invalid then discards session
func (handle *ContextHandle) ReturnSession(session mPkcs11.SessionHandle) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := isEmpty(session)
	if e != nil {
		logger.Warnf("not returning session [%d], due to error [%s]. Discarding it", session, e)
		return
	}

	e = handle.isInRecovery()
	if e != nil {
		logger.Warnf("not returning session [%d], due to error [%s]. Discarding it", session, e)
		return
	}

	_, e = handle.ctx.GetSessionInfo(session)
	if e != nil {
		logger.Debugf("not returning session [%d], due to error [%s]. Discarding it", session, e)
		e = handle.ctx.CloseSession(session)
		if e != nil {
			logger.Warn("unable to close session:", e)
		}
		cachebridge.ClearSession(fmt.Sprintf("%d", session))
		return
	}

	logger.Debugf("Returning session : %d", session)

	select {
	case handle.sessions <- session:
		// returned session back to session cache
	default:
		// have plenty of sessions in cache, dropping
		e = handle.ctx.CloseSession(session)
		cachebridge.ClearSession(fmt.Sprintf("%d", session))
		if e != nil {
			logger.Warn("unable to close session: ", e)
		}
	}
}

//GetSession returns session from session pool
//if pool is empty or completely in use, creates new session
//if new session is invalid recreates one after reloading ctx and re-login
func (handle *ContextHandle) GetSession() (session mPkcs11.SessionHandle) {
	handle.lock.RLock()
	logger.Debugf("Total number of sessions currently in pool is %d\n", len(handle.sessions))
	select {
	case session = <-handle.sessions:
		logger.Debugf("Reusing existing pkcs11 session %+v on slot %d\n", session, handle.slot)
		handle.lock.RUnlock()
	default:
		handle.lock.RUnlock()
		logger.Debug("Opening a new session since cache is empty (or completely in use)")
		// cache is empty (or completely in use), create a new session
		s, err := handle.OpenSession()
		if err != nil {
			logger.Debugf("Opening a new session failed [%v], will retry %d times", err, handle.opts.openSessionRetry)
			handle.lock.Lock()
			defer handle.lock.Unlock()
			for i := 0; i < handle.opts.openSessionRetry; i++ {
				logger.Debugf("Trying re-login and open session attempt[%v]", i+1)
				s, err = handle.reLogin()
				if err != nil {
					logger.Debugf("Failed to re-login, attempt[%d], error[%s], trying again now", i+1, err)
					continue
				} else {
					logger.Debugf("Successfully able to re-login and open session[%d], attempt[%d], clearing cache now for new session", s, i+1)
					cachebridge.ClearSession(fmt.Sprintf("%d", s))
					return s
				}
			}
			logger.Debugf("Exhausted all attempts to recover session, failed with error [%s], returning 0 session", err)
			return s
		}
		logger.Debugf("Created new pkcs11 session %+v on slot %d", s, handle.slot)
		cachebridge.ClearSession(fmt.Sprintf("%d", s))
		return s
	}
	return handle.validateSession(session)
}

// GetAttributeValue obtains the value of one or more object attributes.
func (handle *ContextHandle) GetAttributeValue(session mPkcs11.SessionHandle, objectHandle mPkcs11.ObjectHandle, attrs []*mPkcs11.Attribute) ([]*mPkcs11.Attribute, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return nil, e
	}

	return handle.ctx.GetAttributeValue(session, objectHandle, attrs)
}

// SetAttributeValue modifies the value of one or more object attributes
func (handle *ContextHandle) SetAttributeValue(session mPkcs11.SessionHandle, objectHandle mPkcs11.ObjectHandle, attrs []*mPkcs11.Attribute) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.SetAttributeValue(session, objectHandle, attrs)
}

// GenerateKeyPair generates a public-key/private-key pair creating new key objects.
func (handle *ContextHandle) GenerateKeyPair(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, public, private []*mPkcs11.Attribute) (mPkcs11.ObjectHandle, mPkcs11.ObjectHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	err := isEmpty(session)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to generate key pair")
	}

	e := handle.isInRecovery()
	if e != nil {
		return 0, 0, e
	}

	return handle.ctx.GenerateKeyPair(session, m, public, private)
}

//GenerateKey generates a secret key, creating a new key object.
func (handle *ContextHandle) GenerateKey(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, temp []*mPkcs11.Attribute) (mPkcs11.ObjectHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return 0, e
	}

	return handle.ctx.GenerateKey(session, m, temp)
}

// FindObjectsInit initializes a search for token and session objects that match a template.
func (handle *ContextHandle) FindObjectsInit(session mPkcs11.SessionHandle, temp []*mPkcs11.Attribute) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.FindObjectsInit(session, temp)
}

// FindObjects continues a search for token and session objects that match a template, obtaining additional object
// handles. The returned boolean indicates if the list would have been larger than max.
func (handle *ContextHandle) FindObjects(session mPkcs11.SessionHandle, max int) ([]mPkcs11.ObjectHandle, bool, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return nil, false, e
	}

	return handle.ctx.FindObjects(session, max)
}

//FindObjectsFinal finishes a search for token and session objects.
func (handle *ContextHandle) FindObjectsFinal(session mPkcs11.SessionHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.FindObjectsFinal(session)
}

//Encrypt encrypts single-part data.
func (handle *ContextHandle) Encrypt(session mPkcs11.SessionHandle, message []byte) ([]byte, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return nil, e
	}

	return handle.ctx.Encrypt(session, message)
}

//EncryptInit initializes an encryption operation.
func (handle *ContextHandle) EncryptInit(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, o mPkcs11.ObjectHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.EncryptInit(session, m, o)
}

//DecryptInit initializes a decryption operation.
func (handle *ContextHandle) DecryptInit(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, o mPkcs11.ObjectHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.DecryptInit(session, m, o)
}

//Decrypt decrypts encrypted data in a single part.
func (handle *ContextHandle) Decrypt(session mPkcs11.SessionHandle, cypher []byte) ([]byte, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return nil, e
	}

	return handle.ctx.Decrypt(session, cypher)
}

// SignInit initializes a signature (private key encryption)
// operation, where the signature is (will be) an appendix to
// the data, and plaintext cannot be recovered from the signature.
func (handle *ContextHandle) SignInit(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, o mPkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.SignInit(session, m, o)
}

// Sign signs (encrypts with private key) data in a single part, where the signature
// is (will be) an appendix to the data, and plaintext cannot be recovered from the signature.
func (handle *ContextHandle) Sign(session mPkcs11.SessionHandle, message []byte) ([]byte, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return nil, e
	}

	return handle.ctx.Sign(session, message)
}

// VerifyInit initializes a verification operation, where the
// signature is an appendix to the data, and plaintext cannot
// be recovered from the signature (e.g. DSA).
func (handle *ContextHandle) VerifyInit(session mPkcs11.SessionHandle, m []*mPkcs11.Mechanism, key mPkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.VerifyInit(session, m, key)
}

// Verify verifies a signature in a single-part operation,
// where the signature is an appendix to the data, and plaintext
// cannot be recovered from the signature.
func (handle *ContextHandle) Verify(session mPkcs11.SessionHandle, data []byte, signature []byte) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.Verify(session, data, signature)
}

// CreateObject creates a new object.
func (handle *ContextHandle) CreateObject(session mPkcs11.SessionHandle, temp []*mPkcs11.Attribute) (mPkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return 0, e
	}

	return handle.ctx.CreateObject(session, temp)
}

// CopyObject creates a copy of an object.
func (handle *ContextHandle) CopyObject(sh mPkcs11.SessionHandle, o mPkcs11.ObjectHandle, temp []*mPkcs11.Attribute) (mPkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return 0, e
	}

	return handle.ctx.CopyObject(sh, o, temp)
}

// DestroyObject destroys an object.
func (handle *ContextHandle) DestroyObject(sh mPkcs11.SessionHandle, oh mPkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	e := handle.isInRecovery()
	if e != nil {
		return e
	}

	return handle.ctx.DestroyObject(sh, oh)
}

//FindKeyPairFromSKI finds key pair by SKI
func (handle *ContextHandle) FindKeyPairFromSKI(session mPkcs11.SessionHandle, ski []byte, keyType bool) (*mPkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	err := isEmpty(session)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find key pair from SKI")
	}

	e := handle.isInRecovery()
	if e != nil {
		return nil, e
	}

	return cachebridge.GetKeyPairFromSessionSKI(&cachebridge.KeyPairCacheKey{Mod: handle.ctx, Session: session, SKI: ski, KeyType: keyType})
}

//validateSession validates given session
//if session is invalid recreates one after reloading ctx and re-login
//care should be taken since handle.lock should be read locked before calling this function
func (handle *ContextHandle) validateSession(currentSession mPkcs11.SessionHandle) mPkcs11.SessionHandle {

	handle.lock.RLock()

	e := isEmpty(currentSession)
	if e != nil {
		logger.Debugf("Not validating session[%d] due to [%d], ", currentSession, e)
		return currentSession
	}

	e = handle.isInRecovery()
	if e != nil {
		return 0
	}

	logger.Debugf("Validating session[%+v], for any error condition....", currentSession)
	e = handle.detectErrorCondition(currentSession)
	if e != nil {
		logger.Debugf("Found error condition, while validating session [%+v], error:[%v]", currentSession, e)
	}

	switch e {
	case errSlotIDChanged,
		mPkcs11.Error(mPkcs11.CKR_OBJECT_HANDLE_INVALID),
		mPkcs11.Error(mPkcs11.CKR_SESSION_HANDLE_INVALID),
		mPkcs11.Error(mPkcs11.CKR_SESSION_CLOSED),
		mPkcs11.Error(mPkcs11.CKR_TOKEN_NOT_PRESENT),
		mPkcs11.Error(mPkcs11.CKR_DEVICE_ERROR),
		mPkcs11.Error(mPkcs11.CKR_GENERAL_ERROR),
		mPkcs11.Error(mPkcs11.CKR_USER_NOT_LOGGED_IN):

		logger.Debugf("Found error condition [%s] for session[%+v], attempting to recreate pkcs11 context and re-login....", e, currentSession)

		handle.lock.RUnlock()
		handle.lock.Lock()
		defer handle.lock.Unlock()

		newSession, err := handle.reLogin()
		if err != nil {
			logger.Warnf("Failed to recover session[%v], cause: %s,", currentSession, err)
			return 0
		}
		return newSession

	case mPkcs11.Error(mPkcs11.CKR_DEVICE_MEMORY),
		mPkcs11.Error(mPkcs11.CKR_DEVICE_REMOVED):
		handle.lock.RUnlock()
		panic(fmt.Sprintf("PKCS11 Session failure: [%s]", e))

	default:
		logger.Debugf("Not an Error condition [%+v], didn't match any condition....", e)
		handle.lock.RUnlock()
		// default should be a valid session or valid error, return session as it is
		return currentSession
	}
}

// isInRecovery returns if current context handle is in recovery mode
func (handle *ContextHandle) isInRecovery() error {

	if handle.recovery {
		logger.Debugf("Attempt to access ctx which is in recovery mode, returning error")
		return errors.New("pkcs11 ctx is under recovery, try again later")
	}

	return nil
}

// reLogin destroys pkcs11 context and tries to re-login and returns new session
// Note: this function isn't thread safe, recommended to use write lock for calling this function
func (handle *ContextHandle) reLogin() (mPkcs11.SessionHandle, error) {

	// dispose existing pkcs11 ctx (closing sessions)
	handle.disposePKCS11Ctx()
	logger.Debugf("Disposed ctx, Number of sessions left in pool %d\n", len(handle.sessions))

	// create new context
	var err error
	handle.ctx, err = handle.createNewPKCS11Ctx()
	if err != nil {
		logger.Warn("Failed to recreate new pkcs11 context for given library", err)
		return 0, errors.WithMessage(err, "failed to recreate new pkcs11 context for given library")
	}

	// find slot
	slot, found := handle.findSlot(handle.ctx)
	if !found {
		logger.Warnf("Unable to find slot for label :%s", handle.label)
		return 0, errors.Errorf("unable to find slot for label :%s", handle.label)
	}
	logger.Debugf("Able to find slot : %d ", slot)

	// open new session for given slot
	newSession, err := createNewSession(handle.ctx, slot)
	if err != nil {
		logger.Errorf("Failed to open session with given slot [%s]\n", err)
		return 0, errors.Errorf("failed to open session with given slot :%s", err)
	}
	logger.Debugf("Recreated new pkcs11 session %+v on slot %d\n", newSession, slot)

	// login with new session
	err = handle.ctx.Login(newSession, mPkcs11.CKU_USER, handle.pin)
	if err != nil && err != mPkcs11.Error(mPkcs11.CKR_USER_ALREADY_LOGGED_IN) {
		logger.Errorf("Unable to login with new session :%d, error:%v", newSession, err)
		return 0, errors.Errorf("unable to login with new session :%d", newSession)
	}
	handle.slot = slot

	logger.Infof("Able to re-login with recreated session[%+v] successfully", newSession)
	return newSession, nil
}

//detectErrorCondition checks if given session handle has errors
func (handle *ContextHandle) detectErrorCondition(currentSession mPkcs11.SessionHandle) error {
	var e error
	slot, ok := handle.findSlot(handle.ctx)
	if !ok || slot != handle.slot {
		e = errSlotIDChanged
	}

	if e == nil {
		_, e = handle.ctx.GetSessionInfo(currentSession)
		if e == nil {
			logger.Debugf("Validating operation state for session[%+v]", currentSession)
			_, e = handle.ctx.GetOperationState(currentSession)
		}
	}

	return e
}

//sendNotification sends ctx reload notificatin if channel available
func (handle *ContextHandle) sendNotification() {
	if handle.reloadNotification != nil {
		select {
		case handle.reloadNotification <- struct{}{}:
			logger.Info("Notification sent for recreated pkcs11 ctx")
		default:
			logger.Warn("Unable to send notification for recreated pkcs11 ctx")
		}
	}
}

//disposePKCS11Ctx disposes mPkcs11.Ctx object
func (handle *ContextHandle) disposePKCS11Ctx() {

	logger.Debugf("Disposing pkcs11 ctx for [%s, %s]", handle.lib, handle.label)

	e := handle.isInRecovery()
	if e != nil {
		//already disposed
		return
	}

	// switch on recovery mode
	handle.recovery = true
	// flush all sessions from pool
	handle.sessions = make(chan mPkcs11.SessionHandle, handle.opts.sessionCacheSize)

	// ignore error on close all sessions
	err := handle.ctx.CloseAllSessions(handle.slot)
	if err != nil {
		logger.Warn("Unable to close session", err)
	}

	// clear cache
	cachebridge.ClearAllSession()

	// Finalize context
	err = handle.ctx.Finalize()
	if err != nil {
		logger.Warnf("unable to finalize pkcs11 ctx for [%s, %s] : %s", handle.lib, handle.label, err)
	}

	// Destroy context
	handle.ctx.Destroy()
}

//createNewPKCS11Ctx creates new mPkcs11.Ctx
func (handle *ContextHandle) createNewPKCS11Ctx() (*mPkcs11.Ctx, error) {
	newCtx := mPkcs11.New(handle.lib)
	if newCtx == nil {
		logger.Warn("Failed to recreate new context for given library")
		return nil, errors.New("Failed to recreate new context for given library")
	}

	//initialize new context
	err := newCtx.Initialize()
	if err != nil {
		if err != mPkcs11.Error(mPkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED) {
			logger.Warn("Failed to initialize context:", err)
			return nil, err
		}
	}

	// ctx recovered
	handle.recovery = false
	//send notification about ctx update
	handle.sendNotification()
	return newCtx, nil
}

//findSlot finds slot for given pkcs11 ctx and label
func (handle *ContextHandle) findSlot(ctx *mPkcs11.Ctx) (uint, bool) {

	var found bool
	var slot uint

	//get all slots
	slots, err := ctx.GetSlotList(true)
	if err != nil {
		logger.Warn("Failed to get slot list for recreated context:", err)
		return slot, found
	}

	//find slot matching label
	for _, s := range slots {
		info, err := ctx.GetTokenInfo(s)
		if err != nil {
			continue
		}
		logger.Debugf("Looking for %s, found label %s\n", handle.label, info.Label)
		if handle.label == info.Label {
			found = true
			slot = s
			break
		}
	}

	return slot, found
}

func createNewSession(ctx *mPkcs11.Ctx, slot uint) (mPkcs11.SessionHandle, error) {
	var newSession mPkcs11.SessionHandle
	var err error
	for i := 0; i < 10; i++ {
		newSession, err = ctx.OpenSession(slot, mPkcs11.CKF_SERIAL_SESSION|mPkcs11.CKF_RW_SESSION)
		if err != nil {
			logger.Warnf("OpenSession failed, retrying [%s]\n", err)
		} else {
			return newSession, nil
		}
	}
	return newSession, err
}

// pkcs11CtxCacheKey for context handler
type pkcs11CtxCacheKey struct {
	lib   string
	label string
	pin   string
	opts  ctxOpts
}

//String return string value for pkcs11CtxCacheKey
func (key *pkcs11CtxCacheKey) String() string {
	return fmt.Sprintf("%x_%s_%s_%d_%d", key.lib, key.label, key.opts.connectionName, key.opts.sessionCacheSize, key.opts.openSessionRetry)
}

//getInstance loads ContextHandle instance from cache
//key - cache key
//reload - if true then deletes the existing cache instance and recreates one
func getInstance(key lazycache.Key, reload bool) (*ContextHandle, error) {

	once.Do(func() {
		ctxCache = newCtxCache()
		//anyway, loading first time, no need to reload
		reload = false
	})

	if reload {
		ctxCache.Delete(key)
	}

	ref, err := ctxCache.Get(key)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pkcs11 ctx cache for given key")
	}

	return ref.(*ContextHandle), nil
}

//newCtxCache creates new lazycache instance of context handle cache
func newCtxCache() *lazycache.Cache {
	return lazycache.New(
		"PKCS11_Context_Cache",
		loadLibInitializer(),
		lazyref.WithFinalizer(finalizer()),
	)
}

//finalizer finalizer for context handler cache
func finalizer() lazyref.Finalizer {
	return func(v interface{}) {
		if handle, ok := v.(*ContextHandle); ok {
			logger.Debugf("Finalizing pkcs11 ctx for [%s, %s]", handle.lib, handle.label)
			err := handle.ctx.CloseAllSessions(handle.slot)
			if err != nil {
				logger.Warnf("unable to close all sessions in finalizer for [%s, %s] : %s", handle.lib, handle.label, err)
			}
			err = handle.ctx.Finalize()
			if err != nil {
				logger.Warnf("unable to finalize pkcs11 ctx in finalizer for [%s, %s] : %s", handle.lib, handle.label, err)
			}
			handle.ctx.Destroy()
			cachebridge.ClearAllSession()
		}
	}
}

//loadLibInitializer initializer for context handler cache
func loadLibInitializer() lazycache.EntryInitializer {
	return func(key lazycache.Key) (interface{}, error) {

		ctxKey := key.(*pkcs11CtxCacheKey)
		var slot uint
		logger.Debugf("Loading pkcs11 library [%s]\n", ctxKey.lib)
		if ctxKey.lib == "" {
			return &ContextHandle{}, errors.New("No PKCS11 library default")
		}

		ctx := mPkcs11.New(ctxKey.lib)
		if ctx == nil {
			return &ContextHandle{}, errors.Errorf("Instantiate failed [%s]", ctxKey.lib)
		}

		err := ctx.Initialize()
		if err != nil && err != mPkcs11.Error(mPkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED) {
			logger.Warn("Failed to initialize context:", err)
			return &ContextHandle{}, errors.WithMessage(err, "Failed to initialize pkcs11 ctx")
		}

		slots, err := ctx.GetSlotList(true)
		if err != nil {
			return &ContextHandle{}, errors.WithMessage(err, "Could not get Slot List")
		}
		found := false
		for _, s := range slots {
			info, errToken := ctx.GetTokenInfo(s)
			if errToken != nil {
				continue
			}
			logger.Debugf("Looking for %s, found label %s\n", ctxKey.label, info.Label)
			if ctxKey.label == info.Label {
				found = true
				slot = s
				break
			}
		}
		if !found {
			return &ContextHandle{}, errors.Errorf("Could not find token with label %s", ctxKey.label)
		}
		sessions := make(chan mPkcs11.SessionHandle, ctxKey.opts.sessionCacheSize)
		return &ContextHandle{ctx: ctx, slot: slot, pin: ctxKey.pin, lib: ctxKey.lib, label: ctxKey.label, sessions: sessions, opts: ctxKey.opts}, nil
	}
}

// isEmpty validates if session is valid (not default zero handle)
func isEmpty(session mPkcs11.SessionHandle) error {

	if session > 0 {
		return nil
	}
	return errors.New("invalid session detected")
}
