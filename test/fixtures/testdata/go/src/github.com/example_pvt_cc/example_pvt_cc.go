/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type invokeFunc func(stub shim.ChaincodeStubInterface, args []string) pb.Response
type funcMap map[string]invokeFunc

const (
	getFunc               = "get"
	putFunc               = "put"
	delFunc               = "del"
	putPrivateFunc        = "putprivate"
	getPrivateFunc        = "getprivate"
	delPrivateFunc        = "delprivate"
	putBothFunc           = "putboth"
	getAndPutBothFunc     = "getandputboth"
	invokeCCFunc          = "invokecc"
	addToIntFunc          = "addToInt"
	getPrivateByRangeFunc = "getprivatebyrange"
)

// ExampleCC example chaincode that puts and gets state and private data
type ExampleCC struct {
	funcRegistry funcMap
}

// Init ...
func (cc *ExampleCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke invoke the chaincode with a given function
func (cc *ExampleCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "" {
		return shim.Error("Expecting function")
	}

	f, ok := cc.funcRegistry[function]
	if !ok {
		return shim.Error(fmt.Sprintf("Unknown function [%s]. Expecting one of: %v", function, cc.functions()))
	}

	return f(stub, args)
}

func (cc *ExampleCC) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Invalid args. Expecting key and value")
	}

	key := args[0]
	value := args[1]

	existingValue, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting data for key [%s]: %s", key, err))
	}
	if existingValue != nil {
		value = string(existingValue) + "-" + value
	}

	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting data for key [%s]: %s", key, err))
	}

	return shim.Success([]byte(value))
}

func (cc *ExampleCC) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid args. Expecting key")
	}

	key := args[0]

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting data for key [%s]: %s", key, err))
	}

	return shim.Success([]byte(value))
}

func (cc *ExampleCC) del(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid args. Expecting key")
	}

	key := args[1]

	err := stub.DelState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to delete state for [%s]: %s", key, err))
	}

	return shim.Success(nil)
}

func (cc *ExampleCC) putPrivate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Invalid args. Expecting collection, key and value")
	}

	coll := args[0]
	key := args[1]
	value := args[2]

	if err := stub.PutPrivateData(coll, key, []byte(value)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	return shim.Success(nil)
}

func (cc *ExampleCC) getPrivate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Invalid args. Expecting collection and key")
	}

	coll := args[0]
	key := args[1]

	value, err := stub.GetPrivateData(coll, key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	return shim.Success([]byte(value))
}

func (cc *ExampleCC) getPrivateByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Invalid args. Expecting collection and keyFrom, keyTo")
	}

	coll := args[0]
	keyFrom := args[1]
	keyTo := args[2]

	it, err := stub.GetPrivateDataByRange(coll, keyFrom, keyTo)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting private data by range for collection [%s] and keys [%s to %s]: %s", coll, keyFrom, keyTo, err))
	}

	kvPair := ""
	for it.HasNext() {
		kv, err := it.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next value for private data collection [%s]: %s", coll, err))
		}
		kvPair += fmt.Sprintf("%s=%s ", kv.Key, kv.Value)
	}

	return shim.Success([]byte(kvPair))
}

func (cc *ExampleCC) delPrivate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Invalid args. Expecting collection and key")
	}

	coll := args[0]
	key := args[1]

	err := stub.DelPrivateData(coll, key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	return shim.Success(nil)
}

func (cc *ExampleCC) putBoth(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("Invalid args. Expecting key, value, collection, privkey and privvalue")
	}

	key := args[0]
	value := args[1]
	coll := args[2]
	privKey := args[3]
	privValue := args[4]

	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting state for key [%s]: %s", key, err))
	}
	if err := stub.PutPrivateData(coll, privKey, []byte(privValue)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	return shim.Success(nil)
}

func (cc *ExampleCC) getAndPutBoth(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("Invalid args. Expecting key, value, collection, privkey and privvalue")
	}

	key := args[0]
	value := args[1]
	coll := args[2]
	privKey := args[3]
	privValue := args[4]

	oldValue, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state for key [%s]: %s", key, err))
	}
	if oldValue != nil {
		value = value + "_" + string(oldValue)
	}
	oldPrivValue, err := stub.GetPrivateData(coll, privKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting private data for collection [%s] and key [%s]: %s", coll, privKey, err))
	}
	if oldPrivValue != nil {
		privValue = privValue + "_" + string(oldPrivValue)
	}

	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting state for key [%s]: %s", key, err))
	}
	if err := stub.PutPrivateData(coll, privKey, []byte(privValue)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	return shim.Success(nil)
}

// Adds a given int amount to the value stored in the given private collection's key, storing the result using the same key.
// If the given key does not already exist then it will be added and stored with the given amount.
func (cc *ExampleCC) addToInt(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Invalid args. Expecting collection, key, amountToAdd")
	}

	coll := args[0]
	key := args[1]
	amountToAdd, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid arg: amountToAdd is not an int")
	}

	oldValue, err := stub.GetPrivateData(coll, key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting private data for collection [%s] and key [%s]: %s", coll, key, err))
	}

	var oldValueInt int
	if oldValue != nil {
		oldValueInt, err = strconv.Atoi(string(oldValue))
		if err != nil {
			return shim.Error(fmt.Sprintf("Error parsing existing amount [%s]: %s", string(oldValue), err))
		}
	} else {
		oldValueInt = 0
	}

	newValueInt := oldValueInt + amountToAdd
	if err := stub.PutPrivateData(coll, key, []byte(strconv.Itoa(newValueInt))); err != nil {
		return shim.Error(fmt.Sprintf("Error storing new sum [%s] to key [%s] in private collection [%s]: %s", newValueInt, key, coll, err))
	}

	return shim.Success(nil)
}

type argStruct struct {
	Args []string `json:"Args"`
}

func asBytes(args []string) [][]byte {
	bytes := make([][]byte, len(args))
	for i, arg := range args {
		bytes[i] = []byte(arg)
	}
	return bytes
}

func (cc *ExampleCC) invokeCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error(`Invalid args. Expecting name of chaincode to invoke and chaincode args in the format {"Args":["arg1","arg2",...]}`)
	}

	ccName := args[0]
	invokeArgsJSON := strings.Replace(args[1], "`", `"`, -1)

	argStruct := argStruct{}
	if err := json.Unmarshal([]byte(invokeArgsJSON), &argStruct); err != nil {
		return shim.Error(fmt.Sprintf("Invalid invoke args: %s", err))
	}

	if err := stub.PutState(stub.GetTxID()+"_invokedcc", []byte(ccName)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting state: %s", err))
	}

	return stub.InvokeChaincode(ccName, asBytes(argStruct.Args), "")
}

func (cc *ExampleCC) initRegistry() {
	cc.funcRegistry = make(map[string]invokeFunc)
	cc.funcRegistry[getFunc] = cc.get
	cc.funcRegistry[putFunc] = cc.put
	cc.funcRegistry[delFunc] = cc.del
	cc.funcRegistry[getPrivateFunc] = cc.getPrivate
	cc.funcRegistry[putPrivateFunc] = cc.putPrivate
	cc.funcRegistry[delPrivateFunc] = cc.delPrivate
	cc.funcRegistry[putBothFunc] = cc.putBoth
	cc.funcRegistry[getAndPutBothFunc] = cc.getAndPutBoth
	cc.funcRegistry[invokeCCFunc] = cc.invokeCC
	cc.funcRegistry[addToIntFunc] = cc.addToInt
	cc.funcRegistry[getPrivateByRangeFunc] = cc.getPrivateByRange
}

func (cc *ExampleCC) functions() []string {
	var funcs []string
	for key := range cc.funcRegistry {
		funcs = append(funcs, key)
	}
	return funcs
}

func main() {
	cc := new(ExampleCC)
	cc.initRegistry()
	err := shim.Start(cc)
	if err != nil {
		fmt.Printf("Error starting example chaincode: %s", err)
	}
}
