# pkcs11 helper tool

Meant to help import keys for use in fabric.

```
# prepare slot
softhsm2-util --init-token --label ForFabric --pin 98765432 --free --so-pin 1234

# tool options
./pkcs11helper -help
Usage of ./pkcs11helper:
  -action string
    	list,import (default "list")
  -keyFile string
    	path to pem encoded EC private key you want to import (default "testdata/key.ec.pem")
  -lib string
    	Location of pkcs11 library (Defaults to a list of possible paths to libsofthsm2.so)
  -pin string
    	Slot PIN (default "98765432")
  -slot string
    	Slot Label (default "ForFabric")

# import ec key
./pkcs11helper -action import -keyFile testdata/key.ec.pem
PKCS11 provider found specified slot label: ForFabric (slot: 0, index: 0)
Object was imported with CKA_LABEL:BCPUB1 CKA_ID:018f389d200e48536367f05b99122f355ba33572009bd2b8b521cdbbb717a5b5
Object was imported with CKA_LABEL:BCPRV1 CKA_ID:018f389d200e48536367f05b99122f355ba33572009bd2b8b521cdbbb717a5b5

# list objects
./pkcs11helper -action list
PKCS11 provider found specified slot label: ForFabric (slot: 0, index: 0)
+-------+-----------------+-----------+------------------------------------------------------------------+
| COUNT |    CKA CLASS    | CKA LABEL |                              CKA ID                              |
+-------+-----------------+-----------+------------------------------------------------------------------+
|   001 | CKO_PUBLIC_KEY  | BCPUB1    | 018f389d200e48536367f05b99122f355ba33572009bd2b8b521cdbbb717a5b5 |
|   002 | CKO_PRIVATE_KEY | BCPRV1    | 018f389d200e48536367f05b99122f355ba33572009bd2b8b521cdbbb717a5b5 |
+-------+-----------------+-----------+------------------------------------------------------------------+
Total objects found (max 50): 2
```
