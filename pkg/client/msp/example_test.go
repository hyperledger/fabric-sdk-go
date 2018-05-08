/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package msp

import (
	"fmt"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

func Example() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	username := randomUsername()

	enrollmentSecret, err := c.Register(&RegistrationRequest{Name: username})
	if err != nil {
		fmt.Printf("Register return error %v", err)
		return
	}

	err = c.Enroll(username, WithSecret(enrollmentSecret))
	if err != nil {
		fmt.Printf("failed to enroll user: %v", err)
		return
	}
	fmt.Println("enroll user is completed")

	// Output: enroll user is completed

}

func ExampleNew() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	if c != nil {
		fmt.Println("msp client created")
	}

	// Output: msp client created
}

func ExampleWithOrg() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx, WithOrg("org1"))
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	if c != nil {
		fmt.Println("msp client created with org")
	}

	// Output: msp client created with org
}

func ExampleWithSecret() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	err = c.Enroll(randomUsername(), WithSecret("enrollmentSecret"))
	if err != nil {
		fmt.Printf("failed to enroll user: %v", err)
		return
	}
	fmt.Println("enroll user is completed")

	// Output: enroll user is completed

}

func ExampleClient_Register() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	_, err = c.Register(&RegistrationRequest{Name: randomUsername()})
	if err != nil {
		fmt.Printf("Register return error %v", err)
		return
	}
	fmt.Println("register user is completed")

	// Output: register user is completed
}

func ExampleClient_Enroll() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	err = c.Enroll(randomUsername(), WithSecret("enrollmentSecret"))
	if err != nil {
		fmt.Printf("failed to enroll user: %v", err)
		return
	}
	fmt.Println("enroll user is completed")

	// Output: enroll user is completed
}

func ExampleClient_Reenroll() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	username := randomUsername()

	err = c.Enroll(username, WithSecret("enrollmentSecret"))
	if err != nil {
		fmt.Printf("failed to enroll user: %v", err)
		return
	}

	err = c.Reenroll(username)
	if err != nil {
		fmt.Printf("failed to reenroll user: %v", err)
		return
	}

	fmt.Println("reenroll user is completed")

	// Output: reenroll user is completed

}

func ExampleClient_GetSigningIdentity() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	username := randomUsername()

	err = c.Enroll(username, WithSecret("enrollmentSecret"))
	if err != nil {
		fmt.Printf("failed to enroll user: %v", err)
		return
	}
	enrolledUser, err := c.GetSigningIdentity(username)
	if err != nil {
		fmt.Printf("user not found %v", err)
		return
	}

	if enrolledUser.Identifier().ID != username {
		fmt.Printf("Enrolled user name doesn't match")
		return
	}

	fmt.Println("enroll user is completed")

	// Output: enroll user is completed
}

func ExampleClient_Revoke() {

	ctx := mockClientProvider()

	// Create msp client
	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create msp client")
		return
	}

	_, err = c.Revoke(&RevocationRequest{Name: "testuser"})
	if err != nil {
		fmt.Printf("revoke return error %v", err)
	}
	fmt.Println("revoke user is completed")

	// Output: revoke user is completed
}

func mockClientProvider() context.ClientProvider {
	log.SetLogger(nil)
	f := testFixture{}
	sdk := f.setup()
	logging.SetLevel("fabsdk/fab", logging.ERROR)
	return sdk.Context()
}
