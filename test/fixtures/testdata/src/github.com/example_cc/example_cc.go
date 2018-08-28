/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("examplecc")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init ...
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	txID := stub.GetTxID()
	logger.Debugf("[txID %s] ########### example_cc Init ###########\n", txID)
	_, args := stub.GetFunctionAndParameters()

	err := t.reset(stub, txID, args)
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			logger.Debugf("[txID %s] Transient data in 'init' : %s\n", txID, transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)

}

func (t *SimpleChaincode) resetCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Deletes an entity from its state
	if err := t.reset(stub, stub.GetTxID(), args); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) reset(stub shim.ChaincodeStubInterface, txID string, args []string) error {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return errors.New("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return errors.New("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return errors.New("Expecting integer value for asset holding")
	}
	logger.Debugf("[txID %s] Aval = %d, Bval = %d\n", txID, Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return err
	}

	return nil
}

// Query ...
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call")
}

//set sets given key-value in state
func (t *SimpleChaincode) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting a key and a value")
	}

	// Initialize the chaincode
	key := args[1]
	value := args[2]
	eventID := "testEvent"
	if len(args) >= 4 {
		eventID = args[3]
	}

	logger.Debugf("Setting value for key[%s]", key)

	// Write the state to the ledger
	err = stub.PutState(key, []byte(value))
	if err != nil {
		logger.Errorf("Failed to set value for key[%s] : ", key, err)
		return shim.Error(err.Error())
	}

	err = stub.SetEvent(eventID, []byte("Test Payload"))
	if err != nil {
		logger.Errorf("Failed to set event for key[%s] : ", key, err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invoke ...
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debugf("[txID %s] ########### example_cc Invoke ###########\n", stub.GetTxID())
	function, args := stub.GetFunctionAndParameters()

	if function == "invokecc" {
		return t.invokeCC(stub, args)
	}

	if function == "reset" {
		return t.resetCC(stub, args)
	}

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}

	if args[0] == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}

	if args[0] == "set" {
		// setting an entity state
		return t.set(stub, args)
	}

	if args[0] == "move" {
		eventID := "testEvent"
		if len(args) >= 5 {
			eventID = args[4]
		}
		if err := stub.SetEvent(eventID, []byte("Test Payload")); err != nil {
			return shim.Error("Unable to set CC event: testEvent. Aborting transaction ...")
		}
		return t.move(stub, args)
	}
	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'")
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	txID := stub.GetTxID()
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error
	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[1]
	B = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	logger.Debugf("[txID %s] Aval = %d, Bval = %d\n", txID, Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			logger.Debugf("[txID %s] Transient data in 'move' : %s\n", txID, transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[1]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[1]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Debugf("[txID %s] Query Response:%s\n", stub.GetTxID(), jsonResp)
	return shim.Success(Avalbytes)
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

// invokeCC invokes another chaincode
// arg0: ID of chaincode to invoke
// arg1: Chaincode arguments in the form: {"Args": ["arg0", "arg1",...]}
func (t *SimpleChaincode) invokeCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting ID of chaincode to invoke and args")
	}

	ccID := args[0]
	invokeArgsJSON := args[1]

	argStruct := argStruct{}
	if err := json.Unmarshal([]byte(invokeArgsJSON), &argStruct); err != nil {
		return shim.Error(fmt.Sprintf("Invalid invoke args: %s", err))
	}

	if err := stub.PutState(stub.GetTxID()+"_invokedcc", []byte(ccID)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting state: %s", err))
	}

	return stub.InvokeChaincode(ccID, asBytes(argStruct.Args), "")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
