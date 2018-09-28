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
	"github.com/miekg/pkcs11"
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

//ContextHandle encapsulate basic pkcs11.Ctx operations and manages sessions
type ContextHandle struct {
	ctx                *pkcs11.Ctx
	slot               uint
	pin                string
	lib                string
	label              string
	sessions           chan pkcs11.SessionHandle
	opts               ctxOpts
	reloadNotification chan struct{}
	lock               sync.RWMutex
}

// NotifyCtxReload registers a channel to get notification when underlying pkcs11.Ctx is recreated
func (handle *ContextHandle) NotifyCtxReload(ch chan struct{}) {
	handle.reloadNotification = ch
}

//OpenSession opens a session between an application and a token.
func (handle *ContextHandle) OpenSession() (pkcs11.SessionHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	var session pkcs11.SessionHandle
	var err error
	for i := 0; i < handle.opts.openSessionRetry; i++ {
		session, err = handle.ctx.OpenSession(handle.slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
		if err != nil {
			logger.Warnf("OpenSession failed, retrying [%s]\n", err)
		} else {
			logger.Debug("OpenSession succeeded")
			break
		}
	}
	return session, err
}

// Login logs a user into a token
func (handle *ContextHandle) Login(session pkcs11.SessionHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	if handle.pin == "" {
		return errors.New("No PIN set")
	}
	err := handle.ctx.Login(session, pkcs11.CKU_USER, handle.pin)
	if err != nil && err != pkcs11.Error(pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
		return errors.Errorf("Login failed [%s]", err)
	}
	return nil
}

//ReturnSession returns session back into the session pool
//if pool is pull or session is invalid then discards session
func (handle *ContextHandle) ReturnSession(session pkcs11.SessionHandle) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	_, e := handle.ctx.GetSessionInfo(session)
	if e != nil {
		logger.Warnf("not returning session [%d], due to error [%s]. Discarding it", session, e)
		e = handle.ctx.CloseSession(session)
		if e != nil {
			logger.Warn("unable to close session:", e)
		}
		return
	}

	select {
	case handle.sessions <- session:
		// returned session back to session cache
	default:
		// have plenty of sessions in cache, dropping
		e = handle.ctx.CloseSession(session)
		if e != nil {
			logger.Warn("unable to close session: ", e)
		}
	}
}

//GetSession returns session from session pool
//if pool is empty or completely in use, creates new session
//if new session is invalid recreates one after reloading ctx and re-login
func (handle *ContextHandle) GetSession() (session pkcs11.SessionHandle) {
	handle.lock.RLock()
	select {
	case session = <-handle.sessions:
		logger.Debugf("Reusing existing pkcs11 session %+v on slot %d\n", session, handle.slot)

	default:

		// cache is empty (or completely in use), create a new session
		s, err := handle.OpenSession()
		if err != nil {
			handle.lock.RUnlock()
			panic(fmt.Errorf("OpenSession failed [%s]", err))
		}
		logger.Debugf("Created new pkcs11 session %+v on slot %d\n", s, handle.slot)
		session = s
		cachebridge.ClearSession(fmt.Sprintf("%d", session))
	}
	handle.lock.RUnlock()
	return handle.validateSession(session)
}

// GetAttributeValue obtains the value of one or more object attributes.
func (handle *ContextHandle) GetAttributeValue(session pkcs11.SessionHandle, objectHandle pkcs11.ObjectHandle, attrs []*pkcs11.Attribute) ([]*pkcs11.Attribute, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.GetAttributeValue(session, objectHandle, attrs)
}

// SetAttributeValue modifies the value of one or more object attributes
func (handle *ContextHandle) SetAttributeValue(session pkcs11.SessionHandle, objectHandle pkcs11.ObjectHandle, attrs []*pkcs11.Attribute) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.SetAttributeValue(session, objectHandle, attrs)
}

// GenerateKeyPair generates a public-key/private-key pair creating new key objects.
func (handle *ContextHandle) GenerateKeyPair(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, public, private []*pkcs11.Attribute) (pkcs11.ObjectHandle, pkcs11.ObjectHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.GenerateKeyPair(session, m, public, private)
}

//GenerateKey generates a secret key, creating a new key object.
func (handle *ContextHandle) GenerateKey(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, temp []*pkcs11.Attribute) (pkcs11.ObjectHandle, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.GenerateKey(session, m, temp)
}

// FindObjectsInit initializes a search for token and session objects that match a template.
func (handle *ContextHandle) FindObjectsInit(session pkcs11.SessionHandle, temp []*pkcs11.Attribute) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.FindObjectsInit(session, temp)
}

// FindObjects continues a search for token and session objects that match a template, obtaining additional object
// handles. The returned boolean indicates if the list would have been larger than max.
func (handle *ContextHandle) FindObjects(session pkcs11.SessionHandle, max int) ([]pkcs11.ObjectHandle, bool, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.FindObjects(session, max)
}

//FindObjectsFinal finishes a search for token and session objects.
func (handle *ContextHandle) FindObjectsFinal(session pkcs11.SessionHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.FindObjectsFinal(session)
}

//Encrypt encrypts single-part data.
func (handle *ContextHandle) Encrypt(session pkcs11.SessionHandle, message []byte) ([]byte, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.Encrypt(session, message)
}

//EncryptInit initializes an encryption operation.
func (handle *ContextHandle) EncryptInit(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, o pkcs11.ObjectHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.EncryptInit(session, m, o)
}

//DecryptInit initializes a decryption operation.
func (handle *ContextHandle) DecryptInit(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, o pkcs11.ObjectHandle) error {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.DecryptInit(session, m, o)
}

//Decrypt decrypts encrypted data in a single part.
func (handle *ContextHandle) Decrypt(session pkcs11.SessionHandle, cypher []byte) ([]byte, error) {

	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.Decrypt(session, cypher)
}

// SignInit initializes a signature (private key encryption)
// operation, where the signature is (will be) an appendix to
// the data, and plaintext cannot be recovered from the signature.
func (handle *ContextHandle) SignInit(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, o pkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.SignInit(session, m, o)
}

// Sign signs (encrypts with private key) data in a single part, where the signature
// is (will be) an appendix to the data, and plaintext cannot be recovered from the signature.
func (handle *ContextHandle) Sign(session pkcs11.SessionHandle, message []byte) ([]byte, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.Sign(session, message)
}

// VerifyInit initializes a verification operation, where the
// signature is an appendix to the data, and plaintext cannot
// be recovered from the signature (e.g. DSA).
func (handle *ContextHandle) VerifyInit(session pkcs11.SessionHandle, m []*pkcs11.Mechanism, key pkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.VerifyInit(session, m, key)
}

// Verify verifies a signature in a single-part operation,
// where the signature is an appendix to the data, and plaintext
// cannot be recovered from the signature.
func (handle *ContextHandle) Verify(session pkcs11.SessionHandle, data []byte, signature []byte) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.Verify(session, data, signature)
}

// CreateObject creates a new object.
func (handle *ContextHandle) CreateObject(session pkcs11.SessionHandle, temp []*pkcs11.Attribute) (pkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.CreateObject(session, temp)
}

// CopyObject creates a copy of an object.
func (handle *ContextHandle) CopyObject(sh pkcs11.SessionHandle, o pkcs11.ObjectHandle, temp []*pkcs11.Attribute) (pkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.CopyObject(sh, o, temp)
}

// DestroyObject destroys an object.
func (handle *ContextHandle) DestroyObject(sh pkcs11.SessionHandle, oh pkcs11.ObjectHandle) error {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return handle.ctx.DestroyObject(sh, oh)
}

//FindKeyPairFromSKI finds key pair by SKI
func (handle *ContextHandle) FindKeyPairFromSKI(session pkcs11.SessionHandle, ski []byte, keyType bool) (*pkcs11.ObjectHandle, error) {
	handle.lock.RLock()
	defer handle.lock.RUnlock()

	return cachebridge.GetKeyPairFromSessionSKI(&cachebridge.KeyPairCacheKey{Mod: handle.ctx, Session: session, SKI: ski, KeyType: keyType})
}

//validateSession validates given session
//if session is invalid recreates one after reloading ctx and re-login
//care should be taken since handle.lock should be read locked before calling this function
func (handle *ContextHandle) validateSession(currentSession pkcs11.SessionHandle) pkcs11.SessionHandle {

	handle.lock.RLock()

	e := handle.detectErrorCondition(currentSession)

	switch e {
	case errSlotIDChanged,
		pkcs11.Error(pkcs11.CKR_OBJECT_HANDLE_INVALID),
		pkcs11.Error(pkcs11.CKR_SESSION_HANDLE_INVALID),
		pkcs11.Error(pkcs11.CKR_SESSION_CLOSED),
		pkcs11.Error(pkcs11.CKR_TOKEN_NOT_PRESENT),
		pkcs11.Error(pkcs11.CKR_DEVICE_ERROR),
		pkcs11.Error(pkcs11.CKR_GENERAL_ERROR),
		pkcs11.Error(pkcs11.CKR_USER_NOT_LOGGED_IN):

		logger.Warnf("Found error condition [%s], attempting to recreate pkcs11 context and re-login....", e)

		handle.lock.RUnlock()
		handle.lock.Lock()
		defer handle.lock.Unlock()

		handle.disposePKCS11Ctx()

		//create new context
		newCtx := handle.createNewPKCS11Ctx()
		if newCtx == nil {
			logger.Warn("Failed to recreate new pkcs11 context for given library")
			return 0
		}

		//find slot
		slot, found := handle.findSlot(newCtx)
		if !found {
			logger.Warnf("Unable to find slot for label :%s", handle.label)
			return 0
		}
		logger.Debug("got the slot ", slot)

		//open new session for given slot
		newSession, err := createNewSession(newCtx, slot)
		if err != nil {
			logger.Fatalf("OpenSession [%s]\n", err)
			return 0
		}
		logger.Debugf("Recreated new pkcs11 session %+v on slot %d\n", newSession, slot)

		//login with new session
		err = newCtx.Login(newSession, pkcs11.CKU_USER, handle.pin)
		if err != nil && err != pkcs11.Error(pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
			logger.Warnf("Unable to login with new session :%s", newSession)
			return 0
		}

		handle.sendNotification()

		handle.ctx = newCtx
		handle.slot = slot
		handle.sessions = make(chan pkcs11.SessionHandle, handle.opts.sessionCacheSize)

		logger.Infof("Able to login with recreated session successfully")
		return newSession

	case pkcs11.Error(pkcs11.CKR_DEVICE_MEMORY),
		pkcs11.Error(pkcs11.CKR_DEVICE_REMOVED):
		handle.lock.RUnlock()
		panic(fmt.Sprintf("PKCS11 Session failure: [%s]", e))

	default:
		handle.lock.RUnlock()
		// default should be a valid session or valid error, return session as it is
		return currentSession
	}
}

//detectErrorCondition checks if given session handle has errors
func (handle *ContextHandle) detectErrorCondition(currentSession pkcs11.SessionHandle) error {
	var e error
	slot, ok := handle.findSlot(handle.ctx)
	if !ok || slot != handle.slot {
		e = errSlotIDChanged
	}

	if e == nil {
		_, e = handle.ctx.GetSessionInfo(currentSession)
		if e == nil {
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

//disposePKCS11Ctx disposes pkcs11.Ctx object
func (handle *ContextHandle) disposePKCS11Ctx() {
	//ignore error on close all sessions
	err := handle.ctx.CloseAllSessions(handle.slot)
	if err != nil {
		logger.Warnf("Unable to close session", err)
	}

	//clear cache
	cachebridge.ClearAllSession()

	//Initialize context
	err = handle.ctx.Finalize()
	if err != nil {
		logger.Warnf("unable to finalize pkcs11 ctx for [%s, %s] : %s", handle.lib, handle.label, err)
	}

	//Destroy context
	handle.ctx.Destroy()
}

//createNewPKCS11Ctx creates new pkcs11.Ctx
func (handle *ContextHandle) createNewPKCS11Ctx() *pkcs11.Ctx {
	newCtx := pkcs11.New(handle.lib)
	if newCtx == nil {
		logger.Warn("Failed to recreate new context for given library")
		return nil
	}

	//initialize new context
	err := newCtx.Initialize()
	if err != nil {
		if err != pkcs11.Error(pkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED) {
			logger.Warn("Failed to initialize context:", err)
			return nil
		}
	}

	return newCtx
}

//findSlot finds slot for given pkcs11 ctx and label
func (handle *ContextHandle) findSlot(ctx *pkcs11.Ctx) (uint, bool) {

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

func createNewSession(ctx *pkcs11.Ctx, slot uint) (pkcs11.SessionHandle, error) {
	var newSession pkcs11.SessionHandle
	var err error
	for i := 0; i < 10; i++ {
		newSession, err = ctx.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
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
	return fmt.Sprintf("%x_%s_%s_%d_%d", key.lib, key.label, key.pin, key.opts.sessionCacheSize, key.opts.openSessionRetry)
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

		ctx := pkcs11.New(ctxKey.lib)
		if ctx == nil {
			return &ContextHandle{}, errors.Errorf("Instantiate failed [%s]", ctxKey.lib)
		}

		err := ctx.Initialize()
		if err != nil && err != pkcs11.Error(pkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED) {
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
		sessions := make(chan pkcs11.SessionHandle, ctxKey.opts.sessionCacheSize)
		return &ContextHandle{ctx: ctx, slot: slot, pin: ctxKey.pin, lib: ctxKey.lib, label: ctxKey.label, sessions: sessions, opts: ctxKey.opts}, nil
	}
}
