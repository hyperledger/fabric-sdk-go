// Code generated by counterfeiter. DO NOT EDIT.
package resmgmt

import (
	"context"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

type MockLifecycleResource struct {
	GetInstalledPackageStub        func(context.Context, string, fab.ProposalProcessor, ...resource.Opt) ([]byte, error)
	getInstalledPackageMutex       sync.RWMutex
	getInstalledPackageArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 fab.ProposalProcessor
		arg4 []resource.Opt
	}
	getInstalledPackageReturns struct {
		result1 []byte
		result2 error
	}
	getInstalledPackageReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	InstallStub        func(context.Context, []byte, []fab.ProposalProcessor, ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error)
	installMutex       sync.RWMutex
	installArgsForCall []struct {
		arg1 context.Context
		arg2 []byte
		arg3 []fab.ProposalProcessor
		arg4 []resource.Opt
	}
	installReturns struct {
		result1 []*resource.LifecycleInstallProposalResponse
		result2 error
	}
	installReturnsOnCall map[int]struct {
		result1 []*resource.LifecycleInstallProposalResponse
		result2 error
	}
	QueryCommittedStub        func(context.Context, string, string, fab.ProposalProcessor, ...resource.Opt) ([]*resource.LifecycleQueryCommittedResponse, error)
	queryCommittedMutex       sync.RWMutex
	queryCommittedArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 fab.ProposalProcessor
		arg5 []resource.Opt
	}
	queryCommittedReturns struct {
		result1 []*resource.LifecycleQueryCommittedResponse
		result2 error
	}
	queryCommittedReturnsOnCall map[int]struct {
		result1 []*resource.LifecycleQueryCommittedResponse
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *MockLifecycleResource) GetInstalledPackage(arg1 context.Context, arg2 string, arg3 fab.ProposalProcessor, arg4 ...resource.Opt) ([]byte, error) {
	fake.getInstalledPackageMutex.Lock()
	ret, specificReturn := fake.getInstalledPackageReturnsOnCall[len(fake.getInstalledPackageArgsForCall)]
	fake.getInstalledPackageArgsForCall = append(fake.getInstalledPackageArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 fab.ProposalProcessor
		arg4 []resource.Opt
	}{arg1, arg2, arg3, arg4})
	fake.recordInvocation("GetInstalledPackage", []interface{}{arg1, arg2, arg3, arg4})
	fake.getInstalledPackageMutex.Unlock()
	if fake.GetInstalledPackageStub != nil {
		return fake.GetInstalledPackageStub(arg1, arg2, arg3, arg4...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getInstalledPackageReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *MockLifecycleResource) GetInstalledPackageCallCount() int {
	fake.getInstalledPackageMutex.RLock()
	defer fake.getInstalledPackageMutex.RUnlock()
	return len(fake.getInstalledPackageArgsForCall)
}

func (fake *MockLifecycleResource) GetInstalledPackageCalls(stub func(context.Context, string, fab.ProposalProcessor, ...resource.Opt) ([]byte, error)) {
	fake.getInstalledPackageMutex.Lock()
	defer fake.getInstalledPackageMutex.Unlock()
	fake.GetInstalledPackageStub = stub
}

func (fake *MockLifecycleResource) GetInstalledPackageArgsForCall(i int) (context.Context, string, fab.ProposalProcessor, []resource.Opt) {
	fake.getInstalledPackageMutex.RLock()
	defer fake.getInstalledPackageMutex.RUnlock()
	argsForCall := fake.getInstalledPackageArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *MockLifecycleResource) GetInstalledPackageReturns(result1 []byte, result2 error) {
	fake.getInstalledPackageMutex.Lock()
	defer fake.getInstalledPackageMutex.Unlock()
	fake.GetInstalledPackageStub = nil
	fake.getInstalledPackageReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) GetInstalledPackageReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.getInstalledPackageMutex.Lock()
	defer fake.getInstalledPackageMutex.Unlock()
	fake.GetInstalledPackageStub = nil
	if fake.getInstalledPackageReturnsOnCall == nil {
		fake.getInstalledPackageReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.getInstalledPackageReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) Install(arg1 context.Context, arg2 []byte, arg3 []fab.ProposalProcessor, arg4 ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error) {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	var arg3Copy []fab.ProposalProcessor
	if arg3 != nil {
		arg3Copy = make([]fab.ProposalProcessor, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.installMutex.Lock()
	ret, specificReturn := fake.installReturnsOnCall[len(fake.installArgsForCall)]
	fake.installArgsForCall = append(fake.installArgsForCall, struct {
		arg1 context.Context
		arg2 []byte
		arg3 []fab.ProposalProcessor
		arg4 []resource.Opt
	}{arg1, arg2Copy, arg3Copy, arg4})
	fake.recordInvocation("Install", []interface{}{arg1, arg2Copy, arg3Copy, arg4})
	fake.installMutex.Unlock()
	if fake.InstallStub != nil {
		return fake.InstallStub(arg1, arg2, arg3, arg4...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.installReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *MockLifecycleResource) InstallCallCount() int {
	fake.installMutex.RLock()
	defer fake.installMutex.RUnlock()
	return len(fake.installArgsForCall)
}

func (fake *MockLifecycleResource) InstallCalls(stub func(context.Context, []byte, []fab.ProposalProcessor, ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error)) {
	fake.installMutex.Lock()
	defer fake.installMutex.Unlock()
	fake.InstallStub = stub
}

func (fake *MockLifecycleResource) InstallArgsForCall(i int) (context.Context, []byte, []fab.ProposalProcessor, []resource.Opt) {
	fake.installMutex.RLock()
	defer fake.installMutex.RUnlock()
	argsForCall := fake.installArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *MockLifecycleResource) InstallReturns(result1 []*resource.LifecycleInstallProposalResponse, result2 error) {
	fake.installMutex.Lock()
	defer fake.installMutex.Unlock()
	fake.InstallStub = nil
	fake.installReturns = struct {
		result1 []*resource.LifecycleInstallProposalResponse
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) InstallReturnsOnCall(i int, result1 []*resource.LifecycleInstallProposalResponse, result2 error) {
	fake.installMutex.Lock()
	defer fake.installMutex.Unlock()
	fake.InstallStub = nil
	if fake.installReturnsOnCall == nil {
		fake.installReturnsOnCall = make(map[int]struct {
			result1 []*resource.LifecycleInstallProposalResponse
			result2 error
		})
	}
	fake.installReturnsOnCall[i] = struct {
		result1 []*resource.LifecycleInstallProposalResponse
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) QueryCommitted(arg1 context.Context, arg2 string, arg3 string, arg4 fab.ProposalProcessor, arg5 ...resource.Opt) ([]*resource.LifecycleQueryCommittedResponse, error) {
	fake.queryCommittedMutex.Lock()
	ret, specificReturn := fake.queryCommittedReturnsOnCall[len(fake.queryCommittedArgsForCall)]
	fake.queryCommittedArgsForCall = append(fake.queryCommittedArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 fab.ProposalProcessor
		arg5 []resource.Opt
	}{arg1, arg2, arg3, arg4, arg5})
	fake.recordInvocation("QueryCommitted", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.queryCommittedMutex.Unlock()
	if fake.QueryCommittedStub != nil {
		return fake.QueryCommittedStub(arg1, arg2, arg3, arg4, arg5...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.queryCommittedReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *MockLifecycleResource) QueryCommittedCallCount() int {
	fake.queryCommittedMutex.RLock()
	defer fake.queryCommittedMutex.RUnlock()
	return len(fake.queryCommittedArgsForCall)
}

func (fake *MockLifecycleResource) QueryCommittedCalls(stub func(context.Context, string, string, fab.ProposalProcessor, ...resource.Opt) ([]*resource.LifecycleQueryCommittedResponse, error)) {
	fake.queryCommittedMutex.Lock()
	defer fake.queryCommittedMutex.Unlock()
	fake.QueryCommittedStub = stub
}

func (fake *MockLifecycleResource) QueryCommittedArgsForCall(i int) (context.Context, string, string, fab.ProposalProcessor, []resource.Opt) {
	fake.queryCommittedMutex.RLock()
	defer fake.queryCommittedMutex.RUnlock()
	argsForCall := fake.queryCommittedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *MockLifecycleResource) QueryCommittedReturns(result1 []*resource.LifecycleQueryCommittedResponse, result2 error) {
	fake.queryCommittedMutex.Lock()
	defer fake.queryCommittedMutex.Unlock()
	fake.QueryCommittedStub = nil
	fake.queryCommittedReturns = struct {
		result1 []*resource.LifecycleQueryCommittedResponse
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) QueryCommittedReturnsOnCall(i int, result1 []*resource.LifecycleQueryCommittedResponse, result2 error) {
	fake.queryCommittedMutex.Lock()
	defer fake.queryCommittedMutex.Unlock()
	fake.QueryCommittedStub = nil
	if fake.queryCommittedReturnsOnCall == nil {
		fake.queryCommittedReturnsOnCall = make(map[int]struct {
			result1 []*resource.LifecycleQueryCommittedResponse
			result2 error
		})
	}
	fake.queryCommittedReturnsOnCall[i] = struct {
		result1 []*resource.LifecycleQueryCommittedResponse
		result2 error
	}{result1, result2}
}

func (fake *MockLifecycleResource) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getInstalledPackageMutex.RLock()
	defer fake.getInstalledPackageMutex.RUnlock()
	fake.installMutex.RLock()
	defer fake.installMutex.RUnlock()
	fake.queryCommittedMutex.RLock()
	defer fake.queryCommittedMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *MockLifecycleResource) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
