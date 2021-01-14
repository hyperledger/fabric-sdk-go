## v1.0.0 
Thu 14 Jan 2021 12:46:07 EST

* [080cc92e](https://github.com/hyperledger/fabric-sdk-go/commit/080cc92e) FABG-1026 Fix broken link in gateway godoc (#159)
* [88e5e474](https://github.com/hyperledger/fabric-sdk-go/commit/88e5e474) [FABG-1024] Put valid PublicKey in client PrivateKey (#157)

## v1.0.0-rc1
Thu  3 Dec 2020 08:17:58 GMT

* [2dcfaa90](https://github.com/hyperledger/fabric-sdk-go/commit/2dcfaa90) [[FAB-1022](https://jira.hyperledger.org/browse/FAB-1022)] Release v1.0.0-rc1 (#156)
* [e84f33e9](https://github.com/hyperledger/fabric-sdk-go/commit/e84f33e9) [FABG-1021] Add properties to Peer (#155)
* [5b6912cc](https://github.com/hyperledger/fabric-sdk-go/commit/5b6912cc) [FABG-1020] Changed lifecycle payloads to use camel-case in JSON fields (#154)
* [2f93a320](https://github.com/hyperledger/fabric-sdk-go/commit/2f93a320) Fix Gateway Evaluate ignoring endpoint option (#153)
* [a6ea771b](https://github.com/hyperledger/fabric-sdk-go/commit/a6ea771b) Fix typing error (#152)
* [91355b02](https://github.com/hyperledger/fabric-sdk-go/commit/91355b02) Fix see redundant connector caching. (#151)
* [9c426dcc](https://github.com/hyperledger/fabric-sdk-go/commit/9c426dcc) FABG-1018 TRANSIENT_FAILURE in SubmitTransaction (#150)
* [03d41dd2](https://github.com/hyperledger/fabric-sdk-go/commit/03d41dd2) [FABG-1017] pkcs11 resilience - handling CKR_OPERATION_NOT_INITIALIZED (#149)
* [a64e1ef9](https://github.com/hyperledger/fabric-sdk-go/commit/a64e1ef9) [FABG-1016] Remove PackageID field from check commit readiness request (#148)
* [45ff9b13](https://github.com/hyperledger/fabric-sdk-go/commit/45ff9b13) [FABG-1015] Added JSON tags to lifecycle structs (#147)
* [a18228fa](https://github.com/hyperledger/fabric-sdk-go/commit/a18228fa) [FABG-1008] Allows ecdsakey to return private key bytes (#145)
* [842c4b3e](https://github.com/hyperledger/fabric-sdk-go/commit/842c4b3e) [FABG-1012] Retry on private data dissemination error (#144)
* [0c18d1aa](https://github.com/hyperledger/fabric-sdk-go/commit/0c18d1aa) [FABG-1010] Added ConnectionFailed to ResMgmtDefaultRetryableCodes (#142)
* [166507a9](https://github.com/hyperledger/fabric-sdk-go/commit/166507a9) [FABG-1009] Wait for responses from other Discovery targets (#141)

## v1.0.0-beta3 
Wed 26 Aug 2020 08:14:23 EDT

* [f152f4b9](https://github.com/hyperledger/fabric-sdk-go/commit/f152f4b9) [FABG-1007] Release v1.0.0-beta3 (#140)
* [92e563b5](https://github.com/hyperledger/fabric-sdk-go/commit/92e563b5) fix lint-submodules errors (#137)
* [f34cc66b](https://github.com/hyperledger/fabric-sdk-go/commit/f34cc66b) [FABG 998]Integration tests (#136)
* [bdbc9bd3](https://github.com/hyperledger/fabric-sdk-go/commit/bdbc9bd3) [FABG-1006] Fix for query committed "responses do not match" (#129)
* [470fd07b](https://github.com/hyperledger/fabric-sdk-go/commit/470fd07b) [FABG-1005] chaincoded should use distinct separator (#125)
* [7b97fdf0](https://github.com/hyperledger/fabric-sdk-go/commit/7b97fdf0) [FABG-997] Lifecycle query committed chaincode (#116)
* [58655d5a](https://github.com/hyperledger/fabric-sdk-go/commit/58655d5a) [FABG-996] Lifecycle commit chaincode (#115)
* [1c1fe186](https://github.com/hyperledger/fabric-sdk-go/commit/1c1fe186) [FABG-995] Implement lifecycle check commit readiness (#114)
* [9e1c2cff](https://github.com/hyperledger/fabric-sdk-go/commit/9e1c2cff) [FABG-994] Implement lifecycle query approved chaincode (#113)
* [ec040555](https://github.com/hyperledger/fabric-sdk-go/commit/ec040555) [FABG-993] Implemented lifecycle approve chaincode (#112)
* [c0312bed](https://github.com/hyperledger/fabric-sdk-go/commit/c0312bed) [FABG-992] Lifecycle query installed chaincodes (#110)
* [fd2560a0](https://github.com/hyperledger/fabric-sdk-go/commit/fd2560a0) [FABG-991] Implemented Lifecycle Install Chaincode (#109)
* [7ced0c31](https://github.com/hyperledger/fabric-sdk-go/commit/7ced0c31) [FABG-990] Implemented lifecycle chaincode packager (#108)
* [a7b00f76](https://github.com/hyperledger/fabric-sdk-go/commit/a7b00f76) [FABG-977] Compatibility with cryptogen v2 (#107)
* [37764bc9](https://github.com/hyperledger/fabric-sdk-go/commit/37764bc9) [FABG-1002] Update dependencies for v2.2 (#106)
* [61187183](https://github.com/hyperledger/fabric-sdk-go/commit/61187183) [FABG-1001] Update to Go 1.14 (#105)
* [c97fe34d](https://github.com/hyperledger/fabric-sdk-go/commit/c97fe34d) [FABG-1000] Update pinning for fabric-ca (#104)
* [a336df62](https://github.com/hyperledger/fabric-sdk-go/commit/a336df62) [FABG-999] Update pinning to Fabric v2.2.0 (#102)
* [e11d47b9](https://github.com/hyperledger/fabric-sdk-go/commit/e11d47b9) [FABG-988] Rename cauthdsl to policydsl (#101)
* [c39fc7e8](https://github.com/hyperledger/fabric-sdk-go/commit/c39fc7e8) [FABG-989] Import protolator from fabric-config (#100)
* [a4fd674b](https://github.com/hyperledger/fabric-sdk-go/commit/a4fd674b) [FABG-988] Update to Policy DSL v2.2 (#99)
* [87f5eb8a](https://github.com/hyperledger/fabric-sdk-go/commit/87f5eb8a) [FABG-979] Update stable target to Fabric 2.2.0 (#98)
* [b9dcc017](https://github.com/hyperledger/fabric-sdk-go/commit/b9dcc017) [FABG-956] system cert pool in identity config (#97)
* [01fc6e28](https://github.com/hyperledger/fabric-sdk-go/commit/01fc6e28) [FABG-975] Compatability with v2 chaincode builder (#94)
* [47216cd5](https://github.com/hyperledger/fabric-sdk-go/commit/47216cd5) [FABG-982] Remove legacy status codes (#91)
* [cef0d995](https://github.com/hyperledger/fabric-sdk-go/commit/cef0d995) [FABG-973] Remove legacy CA auth token (#92)
* [d716237d](https://github.com/hyperledger/fabric-sdk-go/commit/d716237d) Fix the Ability to Override CredentialStorePath (#85)
* [d44e5d63](https://github.com/hyperledger/fabric-sdk-go/commit/d44e5d63) [FABG-982] Remove legacy endorser status parsing (#89)
* [57e60e60](https://github.com/hyperledger/fabric-sdk-go/commit/57e60e60) [FABG-981] Process endorser response for chaincode not found (#88)
* [21648a00](https://github.com/hyperledger/fabric-sdk-go/commit/21648a00) [FABG-981] Fix negative integration tests retries (#86)
* [ebe8cd97](https://github.com/hyperledger/fabric-sdk-go/commit/ebe8cd97) [FABG-972] Update stable target to Fabric 2.1.1 (#84)
* [ac702767](https://github.com/hyperledger/fabric-sdk-go/commit/ac702767) [FABG-971] Update stable target to Fabric 1.4.7 (#83)
* [4881b0df](https://github.com/hyperledger/fabric-sdk-go/commit/4881b0df) [FABG-970] Add CODE_OF_CONDUCT and CONTRIBUTING (#82)
* [72d7c7ab](https://github.com/hyperledger/fabric-sdk-go/commit/72d7c7ab) Add Andrew Coleman as a maintainer (#81)
* [163bbe66](https://github.com/hyperledger/fabric-sdk-go/commit/163bbe66) FABG-966 Add 2-arg form of getContract (#79)
* [9338f544](https://github.com/hyperledger/fabric-sdk-go/commit/9338f544) FABG-965: Change MSP base option to v1.4.3 (#78)

## v1.0.0-beta2 
Mon  1 Jun 2020 15:47:02 EDT

* [fdb95089](https://github.com/hyperledger/fabric-sdk-go/commit/fdb95089) [FABG-964] Release v1.0.0-beta2 (#77)
* [e937e5c3](https://github.com/hyperledger/fabric-sdk-go/commit/e937e5c3) FABG-943 Remove unused code (#76)
* [ad7b0438](https://github.com/hyperledger/fabric-sdk-go/commit/ad7b0438) FABG-943 Enhance documentation and add examples (#75)
* [219a09aa](https://github.com/hyperledger/fabric-sdk-go/commit/219a09aa) FABG-936 Support minimal ccp for fabric samples (#74)
* [5fe41b9c](https://github.com/hyperledger/fabric-sdk-go/commit/5fe41b9c) [FABG-962] Fix deadlock in event service (#72)
* [c3b3ab65](https://github.com/hyperledger/fabric-sdk-go/commit/c3b3ab65) [FAB-17777](https://jira.hyperledger.org/browse/FAB-17777) Create basic settings.yaml (#71)
* [2d73fad4](https://github.com/hyperledger/fabric-sdk-go/commit/2d73fad4) FABG-940 Register for block/contract/commit events (#63)
* [f7729f18](https://github.com/hyperledger/fabric-sdk-go/commit/f7729f18) Removed cancellation of all outstanding requests in discovery service (#62)
* [594a8dce](https://github.com/hyperledger/fabric-sdk-go/commit/594a8dce) [FABG-959] Unable to specify empty affiliation (#70)
* [60484ddc](https://github.com/hyperledger/fabric-sdk-go/commit/60484ddc) [FABG-961] Use system cert pool when connecting to fabric-ca (#69)
* [0e550fc1](https://github.com/hyperledger/fabric-sdk-go/commit/0e550fc1) Simplified upstream patching (#67)
* [20b887dd](https://github.com/hyperledger/fabric-sdk-go/commit/20b887dd) Update to golangci-lint v1.23.8 (#66)
* [d7264b5e](https://github.com/hyperledger/fabric-sdk-go/commit/d7264b5e) Fix empty msp, capabilities and anchor peer in ConfigValue (#64)
* [e2b6c739](https://github.com/hyperledger/fabric-sdk-go/commit/e2b6c739) [[FAB-17632](https://jira.hyperledger.org/browse/FAB-17632)] Exclude unnecessary code from ecdsa.go  (#60)
* [e71412ff](https://github.com/hyperledger/fabric-sdk-go/commit/e71412ff) FABG-938 Transient Data (#58)
* [3e8998ec](https://github.com/hyperledger/fabric-sdk-go/commit/3e8998ec) ci: update ci versions and code coverage (#59)
* [c47f8b4a](https://github.com/hyperledger/fabric-sdk-go/commit/c47f8b4a) FABG-937 FileSystemWallet (#54)
* [ab4de7e1](https://github.com/hyperledger/fabric-sdk-go/commit/ab4de7e1) [FABG-948] Add support for extra hosts at enrollment (#46)
* [8bcbcfb4](https://github.com/hyperledger/fabric-sdk-go/commit/8bcbcfb4) Fixed capability version matcher for cases with 3 or more levels (#57)
* [4919c923](https://github.com/hyperledger/fabric-sdk-go/commit/4919c923) [FABG-949] Add support for generating CRL with Revoke (#49)
* [ff3bdd73](https://github.com/hyperledger/fabric-sdk-go/commit/ff3bdd73) [FABG-954] Cancel stream context when event client closes (#55)
* [ae164c98](https://github.com/hyperledger/fabric-sdk-go/commit/ae164c98) fix multi.Errors panic on nil append (#53)
* [be7b275c](https://github.com/hyperledger/fabric-sdk-go/commit/be7b275c) FABG-933 Gateway package for Go SDK (#51)
* [745ca5b8](https://github.com/hyperledger/fabric-sdk-go/commit/745ca5b8) FABG-952 codecov ignore integration test paths
* [8b9e0757](https://github.com/hyperledger/fabric-sdk-go/commit/8b9e0757) [FABG-950] Stablize integrations tests
* [92bda8e1](https://github.com/hyperledger/fabric-sdk-go/commit/92bda8e1) [FABG-946] Honor Excluded Orderer in SDK Config (#43)
* [b5424d1f](https://github.com/hyperledger/fabric-sdk-go/commit/b5424d1f) Copying Policy and ModPolicy when fetching config from orderer. (#48)
* [5f7f0b02](https://github.com/hyperledger/fabric-sdk-go/commit/5f7f0b02) [FABG-947] Terminate event reconnect routine after close (#45)
* [51ecb63a](https://github.com/hyperledger/fabric-sdk-go/commit/51ecb63a) Correcting debug message (#42)
* [53615676](https://github.com/hyperledger/fabric-sdk-go/commit/53615676) chore: [FABG-785] removed unnecessary warning from PKCS11 context handle
* [8f3d32c9](https://github.com/hyperledger/fabric-sdk-go/commit/8f3d32c9) [FABG-785] Supported node chaincode (#40)
* [700785af](https://github.com/hyperledger/fabric-sdk-go/commit/700785af) Update maintainers list
* [e7b9b0dc](https://github.com/hyperledger/fabric-sdk-go/commit/e7b9b0dc) [FABG-932] Add java chaincode integration test (#38)
* [4f8dd0dc](https://github.com/hyperledger/fabric-sdk-go/commit/4f8dd0dc) [FABG-736] Supported java chaincode (#37)
* [e1055f39](https://github.com/hyperledger/fabric-sdk-go/commit/e1055f39) [FABG-930] Ignore _lifecycle namespace in endorsement handler (#35)
* [f5bfb4c3](https://github.com/hyperledger/fabric-sdk-go/commit/f5bfb4c3) [FABG-929] Use retry options from channel policy for selection (#36)
* [6f74e787](https://github.com/hyperledger/fabric-sdk-go/commit/6f74e787) [FABG-927] Fixed third party pin README
* [17677af8](https://github.com/hyperledger/fabric-sdk-go/commit/17677af8) [FABG-927] Patch upstream
* [73d44b63](https://github.com/hyperledger/fabric-sdk-go/commit/73d44b63) [FABG-927] Apply upstream
* [629d3b78](https://github.com/hyperledger/fabric-sdk-go/commit/629d3b78) [FABG-927] Fixed pinning scripts
* [468cb900](https://github.com/hyperledger/fabric-sdk-go/commit/468cb900) [FABG-925] update codecov project target
* [22206ad3](https://github.com/hyperledger/fabric-sdk-go/commit/22206ad3) [FABG-925] add codecov
* [f6335e79](https://github.com/hyperledger/fabric-sdk-go/commit/f6335e79) [FABG-924] Patch upstream.
* [ba124715](https://github.com/hyperledger/fabric-sdk-go/commit/ba124715) [FABG-924] Apply upstream
* [fe2f1556](https://github.com/hyperledger/fabric-sdk-go/commit/fe2f1556) [FABG-924] Simplify upstream patching
* [2d4e7a10](https://github.com/hyperledger/fabric-sdk-go/commit/2d4e7a10) [FABG-921] Update collection.proto message references
* [e9ab174a](https://github.com/hyperledger/fabric-sdk-go/commit/e9ab174a) [FABG-918] Update stale bot configuration
* [22b56d7d](https://github.com/hyperledger/fabric-sdk-go/commit/22b56d7d) [FABG-911] PKCS11 context resilience for errors
* [6fa500f4](https://github.com/hyperledger/fabric-sdk-go/commit/6fa500f4) [FABG-916] Evict connection in TRANSIENT_FAILURE state
* [d424433b](https://github.com/hyperledger/fabric-sdk-go/commit/d424433b) [FABG-915] Update README with CI badge and remove gerrit
* [c365e23e](https://github.com/hyperledger/fabric-sdk-go/commit/c365e23e) [FABG-914] Use parent context in Discovery Dial
* [4dd1938a](https://github.com/hyperledger/fabric-sdk-go/commit/4dd1938a) [FABG-913] Fail-fast for Discovery Service
* [bbaa6377](https://github.com/hyperledger/fabric-sdk-go/commit/bbaa6377) Add default SECURITY policy
* [65095eb3](https://github.com/hyperledger/fabric-sdk-go/commit/65095eb3) Defang stalebot
* [acd13024](https://github.com/hyperledger/fabric-sdk-go/commit/acd13024) [FABG-912] Azure pipeline configuration
* [c13e0cd0](https://github.com/hyperledger/fabric-sdk-go/commit/c13e0cd0) [FABG-910] Update to golangci-lint v1.19.1
* [4a6db410](https://github.com/hyperledger/fabric-sdk-go/commit/4a6db410) [FABG-911] PKCS11 context resilience for errors

## v1.0.0-beta1
Wed Sep 18 11:37:33 EDT 2019

* [5d7ae7a5](https://github.com/hyperledger/fabric-sdk-go/commit/5d7ae7a5) [FABG-907] Release v1.0.0-beta1
* [fc08b96a](https://github.com/hyperledger/fabric-sdk-go/commit/fc08b96a) FABG-909 - Multi-org channel config update
* [ba370d29](https://github.com/hyperledger/fabric-sdk-go/commit/ba370d29) [FABG-908] Update to Go 1.13
* [d578af09](https://github.com/hyperledger/fabric-sdk-go/commit/d578af09) [FABG-905] Fix the warn
* [3b103fa1](https://github.com/hyperledger/fabric-sdk-go/commit/3b103fa1) FABG-904 Ability to update channel configuration
* [1fab3508](https://github.com/hyperledger/fabric-sdk-go/commit/1fab3508) [[FAB-16489](https://jira.hyperledger.org/browse/FAB-16489)] Add CODEOWNERS
* [6c3f788a](https://github.com/hyperledger/fabric-sdk-go/commit/6c3f788a) [FABG-900] Remove third_party fabric module
* [7dc798e6](https://github.com/hyperledger/fabric-sdk-go/commit/7dc798e6) [FABG-902] Update to latest fabric third_party pins
* [7483e052](https://github.com/hyperledger/fabric-sdk-go/commit/7483e052) [FABG-901] Remove third_party protos
* [d2b42602](https://github.com/hyperledger/fabric-sdk-go/commit/d2b42602) [FABG-899] Use fabric-protos-go module
* [e1fa7b9b](https://github.com/hyperledger/fabric-sdk-go/commit/e1fa7b9b) FABG-898 Update to latest fabric pins
* [7eeb1434](https://github.com/hyperledger/fabric-sdk-go/commit/7eeb1434) FABG-897 Convert protobuf to/from JSON
* [bdd5eed6](https://github.com/hyperledger/fabric-sdk-go/commit/bdd5eed6) [FABG-896] Remove GoDep definitions
* [54aba206](https://github.com/hyperledger/fabric-sdk-go/commit/54aba206) [FABG-895] Add more prominent link to GoDocs in README
* [ae6252cc](https://github.com/hyperledger/fabric-sdk-go/commit/ae6252cc) [FABG-893] Update to latest protos mod
* [ab2f64f5](https://github.com/hyperledger/fabric-sdk-go/commit/ab2f64f5) [FABG-893] Update to latest fabric pins
* [267a1810](https://github.com/hyperledger/fabric-sdk-go/commit/267a1810) FABG-891 - Retrieve channel config block
* [bf5e77cb](https://github.com/hyperledger/fabric-sdk-go/commit/bf5e77cb) [FABG-889] add CreateConfigSignatureFromReader method
* [ba9dfce3](https://github.com/hyperledger/fabric-sdk-go/commit/ba9dfce3) FABG-890 Fixed argument names
* [56799614](https://github.com/hyperledger/fabric-sdk-go/commit/56799614) FABG-888 - Add ID property to CA definition
* [8e3a9008](https://github.com/hyperledger/fabric-sdk-go/commit/8e3a9008) [FABG-886] wrap internal discovery.Request
* [1ad92455](https://github.com/hyperledger/fabric-sdk-go/commit/1ad92455) [FABG-887] ci: fix path when finding changed packages
* [fb049b2a](https://github.com/hyperledger/fabric-sdk-go/commit/fb049b2a) [FABG-884] copy more fields when extract chConf
* [e0aa15e4](https://github.com/hyperledger/fabric-sdk-go/commit/e0aa15e4) [FABG-883] Multiple CAs per organization
* [4db5c822](https://github.com/hyperledger/fabric-sdk-go/commit/4db5c822) [FABG-881] Update to test target to 1.4.2 and Go 1.12
* [f264087c](https://github.com/hyperledger/fabric-sdk-go/commit/f264087c) [FABG-875] Fix round-robin balancer in selection
* [48690b25](https://github.com/hyperledger/fabric-sdk-go/commit/48690b25) [FABG-879] Fix panic when logging disconnect event
* [33876cd6](https://github.com/hyperledger/fabric-sdk-go/commit/33876cd6) [FABG-871] Generate the genesis block MSP directory
* [6160a00c](https://github.com/hyperledger/fabric-sdk-go/commit/6160a00c) [FABG-867] Create anchor peer transaction
* [94a6ed7c](https://github.com/hyperledger/fabric-sdk-go/commit/94a6ed7c) [[FAB-15553](https://jira.hyperledger.org/browse/FAB-15553)] Remove Fabric baseimage
* [b6fe995b](https://github.com/hyperledger/fabric-sdk-go/commit/b6fe995b) [FABG-870] Improve logging in PKCS11 ContextHandle
* [54fa6147](https://github.com/hyperledger/fabric-sdk-go/commit/54fa6147) [FABG-863] Create a channel creation tx
* [e3ec4247](https://github.com/hyperledger/fabric-sdk-go/commit/e3ec4247) [FABG-869] Fixed go.mod dependenies
* [7a8c0e0e](https://github.com/hyperledger/fabric-sdk-go/commit/7a8c0e0e) [FABG-862] Create genesis block
* [4def0f92](https://github.com/hyperledger/fabric-sdk-go/commit/4def0f92) [FABG-860] fix dial ctx error using same context
* [48bb0d19](https://github.com/hyperledger/fabric-sdk-go/commit/48bb0d19) [FABG-859] Fabric Logging Variable Update
* [75ddf2f7](https://github.com/hyperledger/fabric-sdk-go/commit/75ddf2f7) [FABG-857] Add QueryConfigBlock in ledger client
* [745c9b98](https://github.com/hyperledger/fabric-sdk-go/commit/745c9b98) [[FAB-15242](https://jira.hyperledger.org/browse/FAB-15242)] To remove newline character.
* [65e82965](https://github.com/hyperledger/fabric-sdk-go/commit/65e82965) [[FAB-15170](https://jira.hyperledger.org/browse/FAB-15170)] add Channel Client Query Benchmark
* [8e846971](https://github.com/hyperledger/fabric-sdk-go/commit/8e846971) [FABG-855] Update stable test target to 1.4.1
* [5a9a0e74](https://github.com/hyperledger/fabric-sdk-go/commit/5a9a0e74) [FABG-849] Cleanup scripts
* [3e09cdf4](https://github.com/hyperledger/fabric-sdk-go/commit/3e09cdf4) [FABG-851] Allow tests to run outside of GOPATH
* [33f4504f](https://github.com/hyperledger/fabric-sdk-go/commit/33f4504f) [FABG-852] Enable 'default-exclude' lint checks
* [8bb794de](https://github.com/hyperledger/fabric-sdk-go/commit/8bb794de) [FABG-849] Update fabric pins to v2.0.0-alpha
* [f1fd02ac](https://github.com/hyperledger/fabric-sdk-go/commit/f1fd02ac) [FABG-850] Remove vendor populate during tests
* [5b5ef5e2](https://github.com/hyperledger/fabric-sdk-go/commit/5b5ef5e2) [FABG-847] use third_party fabric module
* [0ef97e4f](https://github.com/hyperledger/fabric-sdk-go/commit/0ef97e4f) [FABG-848] Use errors.WithMessagef
* [8d90000c](https://github.com/hyperledger/fabric-sdk-go/commit/8d90000c) [FABG-847] Combine protos into third_party module
* [ff5642a5](https://github.com/hyperledger/fabric-sdk-go/commit/ff5642a5) [FABG-846] Decouple SDK usage of third_party utils
* [118e73cf](https://github.com/hyperledger/fabric-sdk-go/commit/118e73cf) [FABG-845] Enable modules during tests
* [b1f8907c](https://github.com/hyperledger/fabric-sdk-go/commit/b1f8907c) [FABG-844] use third_party fabric module
* [06a75db4](https://github.com/hyperledger/fabric-sdk-go/commit/06a75db4) [FABG-844] third_party fabric pins as a module
* [ca2a7a33](https://github.com/hyperledger/fabric-sdk-go/commit/ca2a7a33) [FABG-843] update to golangci-lint v1.16.0
* [37a26c64](https://github.com/hyperledger/fabric-sdk-go/commit/37a26c64) [FABG-842] Use gobin to invoke dev tools
* [f328fa6f](https://github.com/hyperledger/fabric-sdk-go/commit/f328fa6f) [FABG-841] Fix go.mod file
* [f44a4617](https://github.com/hyperledger/fabric-sdk-go/commit/f44a4617) [FABG-840] Allow Go 1.12
* [a7a340b8](https://github.com/hyperledger/fabric-sdk-go/commit/a7a340b8) [FABG-839] Relax Go patch version check
* [b60a9ff1](https://github.com/hyperledger/fabric-sdk-go/commit/b60a9ff1) [FABG-838] Switch populate script to go mod
* [e7f22b3d](https://github.com/hyperledger/fabric-sdk-go/commit/e7f22b3d) [FABG-834] Rebase Fabric pins to 2.0 master
* [6ec70c17](https://github.com/hyperledger/fabric-sdk-go/commit/6ec70c17) [FABG-837] Update miekg/pkcs11 dependency
* [93c3fcb2](https://github.com/hyperledger/fabric-sdk-go/commit/93c3fcb2) [FABG-836] Prepare for Go modules
* [0e710ceb](https://github.com/hyperledger/fabric-sdk-go/commit/0e710ceb) [FABG-833] Release v1.0.0-alpha5

## v1.0.0-alpha5
Wed Mar 27 16:38:23 EDT 2019

* [aa1c8816](https://github.com/hyperledger/fabric-sdk-go/commit/aa1c8816) [FABG-835] Detect softhsm2 image existence
* [658cc4a5](https://github.com/hyperledger/fabric-sdk-go/commit/658cc4a5) [FABG-832] Switch linter : gometalinter->golangci-lint
* [46e05989](https://github.com/hyperledger/fabric-sdk-go/commit/46e05989) [FABG-831] Update Go Max Supported Version to 1.11.5
* [11dc83c4](https://github.com/hyperledger/fabric-sdk-go/commit/11dc83c4) [FABG-830] Add wait method for chaincoded
* [00a50ffa](https://github.com/hyperledger/fabric-sdk-go/commit/00a50ffa) [FABG-828] do not include pin in ctx handle key
* [f198238e](https://github.com/hyperledger/fabric-sdk-go/commit/f198238e) [FABG-825] Add error handler option
* [458319b4](https://github.com/hyperledger/fabric-sdk-go/commit/458319b4) [FABG-826] Close client context
* [5fd3e988](https://github.com/hyperledger/fabric-sdk-go/commit/5fd3e988) [FABG-824] Allow for provider-specific options
* [bda01c94](https://github.com/hyperledger/fabric-sdk-go/commit/bda01c94) [FABG-822] test pvt data reconciliation
* [7df511b8](https://github.com/hyperledger/fabric-sdk-go/commit/7df511b8) [FABG-813] Ch Client Metrics-Switch to Fabric impl
* [ce2814e7](https://github.com/hyperledger/fabric-sdk-go/commit/ce2814e7) [FABG-814] Derive URLs from hostname if omitted
* [fc535d14](https://github.com/hyperledger/fabric-sdk-go/commit/fc535d14) [FABG-787] Document BCCSP section as optional
* [8dfefe44](https://github.com/hyperledger/fabric-sdk-go/commit/8dfefe44) [FABG-766] Auth token support for Fab V1.3
* [3e710760](https://github.com/hyperledger/fabric-sdk-go/commit/3e710760) [FABG-819] Fix typo in loadAllConfigs() Debugf()
* [752dda15](https://github.com/hyperledger/fabric-sdk-go/commit/752dda15) [FABG-816] Fix typo in Debugf()
* [27a1c09d](https://github.com/hyperledger/fabric-sdk-go/commit/27a1c09d) Fix intergration-test-local target with fabric1.4
* [56ebf9ad](https://github.com/hyperledger/fabric-sdk-go/commit/56ebf9ad) [FABG-815] make multi-errors on a single line
* [f50af625](https://github.com/hyperledger/fabric-sdk-go/commit/f50af625) Updated the Fabric Fixture version to 1.4
* [aabee4d1](https://github.com/hyperledger/fabric-sdk-go/commit/aabee4d1) [FABG-812] Selection handling of 'access denied' error
* [a19e0b06](https://github.com/hyperledger/fabric-sdk-go/commit/a19e0b06) [FABG-811] Update versions in README
* [5e291d3a](https://github.com/hyperledger/fabric-sdk-go/commit/5e291d3a) [FABG-810] Discovery handling of 'access denied' error
* [295d4cbe](https://github.com/hyperledger/fabric-sdk-go/commit/295d4cbe) [FABG-811] Update to Fabric 1.4.0
* [3b2b8762](https://github.com/hyperledger/fabric-sdk-go/commit/3b2b8762) [FABG-809] Proper handling of 403 code
* [95e07420](https://github.com/hyperledger/fabric-sdk-go/commit/95e07420) [FABG-808] unique protos update in thirdparty
* [e23671d4](https://github.com/hyperledger/fabric-sdk-go/commit/e23671d4) [FABG-807] Update to Go 1.11
* [c410a99e](https://github.com/hyperledger/fabric-sdk-go/commit/c410a99e) [FABG-806] Rebase pinning to Fabric 1.4 RC2
* [f53e21bd](https://github.com/hyperledger/fabric-sdk-go/commit/f53e21bd) [FABG-804] Update prerelease target to 1.4 RC2
* [cb6e8f94](https://github.com/hyperledger/fabric-sdk-go/commit/cb6e8f94) [FABG-805] MSP tests failing against Fabric 1.4 RC
* [c17b014d](https://github.com/hyperledger/fabric-sdk-go/commit/c17b014d) Configure Stale ProBot
* [371a591d](https://github.com/hyperledger/fabric-sdk-go/commit/371a591d) [FABG-803] Sync the latest files of fabric.
* [fa4badf7](https://github.com/hyperledger/fabric-sdk-go/commit/fa4badf7) [FABG-788] Set PeerChannelConfig defaults properly
* [05ffa0a5](https://github.com/hyperledger/fabric-sdk-go/commit/05ffa0a5) [FABG-800] fallback to global orderers section
* [ea8cf696](https://github.com/hyperledger/fabric-sdk-go/commit/ea8cf696) [FABG-798] remove cache package
* [3081d701](https://github.com/hyperledger/fabric-sdk-go/commit/3081d701) [FABG-801] fix stale function comment
* [de52b8aa](https://github.com/hyperledger/fabric-sdk-go/commit/de52b8aa) [FABG-779]Add an function to modify enroll request
* [740d7e29](https://github.com/hyperledger/fabric-sdk-go/commit/740d7e29) [FABG-795] Ensure updateLastBlockNum is 64-bit aligned
* [946ab57c](https://github.com/hyperledger/fabric-sdk-go/commit/946ab57c) [FABG-769] query chaincode's collection configuration
* [11a8a7ee](https://github.com/hyperledger/fabric-sdk-go/commit/11a8a7ee) [FABG-787] Add defaults for cryptoconfig
* [d10be338](https://github.com/hyperledger/fabric-sdk-go/commit/d10be338) [FABG-793] Add note about libtool requirement
* [2c76aff6](https://github.com/hyperledger/fabric-sdk-go/commit/2c76aff6) [FABG-793] Non-conflicting namespace for pkcs11 import
* [1dcd2338](https://github.com/hyperledger/fabric-sdk-go/commit/1dcd2338) [[FAB-12464](https://jira.hyperledger.org/browse/FAB-12464)] add explicit retry for CC2CC test
* [303f01c9](https://github.com/hyperledger/fabric-sdk-go/commit/303f01c9) [[FAB-12464](https://jira.hyperledger.org/browse/FAB-12464)] add explicit retries in integration tests
* [d8a85e8d](https://github.com/hyperledger/fabric-sdk-go/commit/d8a85e8d) [FABG-789] Make depend should work when GOBIN is set
* [9efe90fc](https://github.com/hyperledger/fabric-sdk-go/commit/9efe90fc) [FABG-778] Corrections, Win10-MINGW64 Docker call
* [34bbf266](https://github.com/hyperledger/fabric-sdk-go/commit/34bbf266) [FABG-681] MSP Client: CAInfo
* [85e62441](https://github.com/hyperledger/fabric-sdk-go/commit/85e62441) [FABG-781] make fails on Windows using Git Bash
* [15f675d3](https://github.com/hyperledger/fabric-sdk-go/commit/15f675d3) [FABG-783] Invalid last block received in event client
* [74d819f0](https://github.com/hyperledger/fabric-sdk-go/commit/74d819f0) [[FAB-12464](https://jira.hyperledger.org/browse/FAB-12464)] fix query retries for MVCC Conflicts
* [a43e084a](https://github.com/hyperledger/fabric-sdk-go/commit/a43e084a) [FABG-682] MSP Client: Affiliation Service
* [ca9ef66b](https://github.com/hyperledger/fabric-sdk-go/commit/ca9ef66b) [FABG-782] Fix default event service policy
* [0ff6dad6](https://github.com/hyperledger/fabric-sdk-go/commit/0ff6dad6) [DEV-10050] fix local tests for Dist Signatures
* [1610d6a0](https://github.com/hyperledger/fabric-sdk-go/commit/1610d6a0) [[FAB-12464](https://jira.hyperledger.org/browse/FAB-12464)] fix additional z build errors
* [6a3c34e1](https://github.com/hyperledger/fabric-sdk-go/commit/6a3c34e1) [FABG-780] Update regex, dependencies.sh
* [db2b8eb1](https://github.com/hyperledger/fabric-sdk-go/commit/db2b8eb1) [[FAB-12464](https://jira.hyperledger.org/browse/FAB-12464)] z build failing due to Integration error
* [e0602c4f](https://github.com/hyperledger/fabric-sdk-go/commit/e0602c4f) [FABG-777] matchers substitution expression fix
* [7e1ea6e0](https://github.com/hyperledger/fabric-sdk-go/commit/7e1ea6e0) [FABG-776] Added private data range query test
* [de259d3c](https://github.com/hyperledger/fabric-sdk-go/commit/de259d3c) [FABG-773] CI test failure fixes
* [a3cc1c33](https://github.com/hyperledger/fabric-sdk-go/commit/a3cc1c33) [FABG-723] switched temp dir to default temp dir
* [8b2e7a26](https://github.com/hyperledger/fabric-sdk-go/commit/8b2e7a26) [FABG-774] Updated regex to string, populate-vendor.sh
* [b474c35c](https://github.com/hyperledger/fabric-sdk-go/commit/b474c35c) Updated regex to string, populate-fixtures.sh
* [aa0f268f](https://github.com/hyperledger/fabric-sdk-go/commit/aa0f268f) [FABG-772] Import Fabric 1.3.0
* [8e44ac8c](https://github.com/hyperledger/fabric-sdk-go/commit/8e44ac8c) [FABG-773] Update to fabric 1.3.0
* [f7ae259c](https://github.com/hyperledger/fabric-sdk-go/commit/f7ae259c) [FABG-768] unique protos update in thirdparty
* [76876694](https://github.com/hyperledger/fabric-sdk-go/commit/76876694) [FABG-723] add int tests for dist signatures
* [0809869e](https://github.com/hyperledger/fabric-sdk-go/commit/0809869e) [FABG-771] Add test case for put and get private data
* [f93fe443](https://github.com/hyperledger/fabric-sdk-go/commit/f93fe443) [FABG-770] Expose TargetSorter in channel client
* [9549fe54](https://github.com/hyperledger/fabric-sdk-go/commit/9549fe54) [FABG-768] Import latest fabric code
* [0c73d46b](https://github.com/hyperledger/fabric-sdk-go/commit/0c73d46b) [FABG-765] Fail the build if Go version is not correct
* [f10bc6b4](https://github.com/hyperledger/fabric-sdk-go/commit/f10bc6b4) [FABG-767] Update to Fabric 1.2.1
* [ea10f6c5](https://github.com/hyperledger/fabric-sdk-go/commit/ea10f6c5) [FABG-764] query on revoked peer test
* [adcd251d](https://github.com/hyperledger/fabric-sdk-go/commit/adcd251d) [FABG-764] revoke user integration test
* [b69bfc21](https://github.com/hyperledger/fabric-sdk-go/commit/b69bfc21) [FABG-764] peer revoke integration test
* [0ac402c8](https://github.com/hyperledger/fabric-sdk-go/commit/0ac402c8) [FABG-762] Make event service config channel-specific
* [cf7dc625](https://github.com/hyperledger/fabric-sdk-go/commit/cf7dc625) [FABG-763] Use hard-coded channel config if no _default
* [1eb1d287](https://github.com/hyperledger/fabric-sdk-go/commit/1eb1d287) [FABG-721] notification on pkcs11 ctx reload
* [a739834c](https://github.com/hyperledger/fabric-sdk-go/commit/a739834c) [FABG-761] Config for event client peer resolvers
* [b9db2f44](https://github.com/hyperledger/fabric-sdk-go/commit/b9db2f44) [FABG-761] Added prefer-org and prefer-peer resolvers
* [c3753acd](https://github.com/hyperledger/fabric-sdk-go/commit/c3753acd) [FABG-761] Pluggable peer resolvers for event client
* [c8fd21dc](https://github.com/hyperledger/fabric-sdk-go/commit/c8fd21dc) [DEV-9154] Private data rollback MVCC_READ_CONFLICT
* [dd9c90e6](https://github.com/hyperledger/fabric-sdk-go/commit/dd9c90e6) [FABG-756] logs for SW cryptosuite in corefactory
* [b3680321](https://github.com/hyperledger/fabric-sdk-go/commit/b3680321) [FABG-755] Round-robin balancer using block height
* [915a9324](https://github.com/hyperledger/fabric-sdk-go/commit/915a9324) [FABG-755] Round-robin balancer using block height
* [ff64cd61](https://github.com/hyperledger/fabric-sdk-go/commit/ff64cd61) [FABG-721] additional functions in pkcs11 wrapper
* [4fc4b170](https://github.com/hyperledger/fabric-sdk-go/commit/4fc4b170) [FABG-721] additional functions in pkcs11 wrapper
* [6721a4c5](https://github.com/hyperledger/fabric-sdk-go/commit/6721a4c5) [FABG-723] Support distributed signing identities
* [8113decd](https://github.com/hyperledger/fabric-sdk-go/commit/8113decd) [FABG-751] SDK PKCS11 tests are broken with Mutual TLS
* [33034df7](https://github.com/hyperledger/fabric-sdk-go/commit/33034df7) [FABG-750] Make UserStore persistence optional
* [dcb26451](https://github.com/hyperledger/fabric-sdk-go/commit/dcb26451) [FABG-748] example cc unique keys in int-tests
* [de7bf5f8](https://github.com/hyperledger/fabric-sdk-go/commit/de7bf5f8) [FABG-748] fix for intermittent SeekTypes test failure
* [c8e1c6cc](https://github.com/hyperledger/fabric-sdk-go/commit/c8e1c6cc) [FABG-749] Channel Client: MVCC Read Conflict Test
* [b2585b78](https://github.com/hyperledger/fabric-sdk-go/commit/b2585b78) [FABG-748] fix for MVCC errors in channel tests
* [02c1c6f9](https://github.com/hyperledger/fabric-sdk-go/commit/02c1c6f9) [FABG-748] fix for lost events in CI
* [3fbef40a](https://github.com/hyperledger/fabric-sdk-go/commit/3fbef40a) [FABG-747] testSeekOpts test should fail in timeout
* [cdd0b4d2](https://github.com/hyperledger/fabric-sdk-go/commit/cdd0b4d2) [FABG-746] Allow transfer of event registrations
* [7c0d57df](https://github.com/hyperledger/fabric-sdk-go/commit/7c0d57df) [FABG-742] Fix docker-compose for standalone mode
* [4d38dbdd](https://github.com/hyperledger/fabric-sdk-go/commit/4d38dbdd) [FABG-728] Fix Org2 local test configuration
* [bafad14e](https://github.com/hyperledger/fabric-sdk-go/commit/bafad14e) [FABG-744] Default Channel: Default Peers and Orderers
* [43512154](https://github.com/hyperledger/fabric-sdk-go/commit/43512154) [FABG-721] pkcs11.ContextHandle - pinning scripts
* [1bb924a2](https://github.com/hyperledger/fabric-sdk-go/commit/1bb924a2) [FABG-745] Parameterize internal changed package calc
* [31c2fa4b](https://github.com/hyperledger/fabric-sdk-go/commit/31c2fa4b) [FABG-737] Default Channel: Channel Policies
* [9af76a3f](https://github.com/hyperledger/fabric-sdk-go/commit/9af76a3f) [FABG-563] CreateSigningIdentity with cert and key
* [0d25e261](https://github.com/hyperledger/fabric-sdk-go/commit/0d25e261) [FABG-712] Import latest fabric code
* [d45bf86f](https://github.com/hyperledger/fabric-sdk-go/commit/d45bf86f) [FABG-743] Update prev to Fabric 1.1.1
* [427c3294](https://github.com/hyperledger/fabric-sdk-go/commit/427c3294) [FABG-739] Use 1.2.1 images from dev registry
* [63f1c2ba](https://github.com/hyperledger/fabric-sdk-go/commit/63f1c2ba) [FABG-742] Add docker-compose pull to devstable
* [0cd98933](https://github.com/hyperledger/fabric-sdk-go/commit/0cd98933) [FABG-727] Remove socat dependency from Makefile
* [28c91d2f](https://github.com/hyperledger/fabric-sdk-go/commit/28c91d2f) [FABG-721] PKCS11 resilience
* [ee69064b](https://github.com/hyperledger/fabric-sdk-go/commit/ee69064b) [FABG-728] Fix Org2 test configuration
* [72066f71](https://github.com/hyperledger/fabric-sdk-go/commit/72066f71) [FABG-735] Allow client to provide nonce/creator
* [8d8c7b8b](https://github.com/hyperledger/fabric-sdk-go/commit/8d8c7b8b) [[FAB-11694](https://jira.hyperledger.org/browse/FAB-11694)] Improve user management test coverage
* [c8ac55a7](https://github.com/hyperledger/fabric-sdk-go/commit/c8ac55a7) [FABG-731] Some int tests missing sdk.Close call
* [e9474b49](https://github.com/hyperledger/fabric-sdk-go/commit/e9474b49) FABG-733] Negative revoked test not loading config
* [4834515c](https://github.com/hyperledger/fabric-sdk-go/commit/4834515c) [FABG-724] Don't Query all the peers for selection
* [22fbd8bb](https://github.com/hyperledger/fabric-sdk-go/commit/22fbd8bb) [FABG-726] monitorBlockHeight ticker chan not closed
* [7c970d52](https://github.com/hyperledger/fabric-sdk-go/commit/7c970d52) [FABG-727] Chaincoded no longer needs docker
* [be3009e2](https://github.com/hyperledger/fabric-sdk-go/commit/be3009e2) [FABG-718] integration test seektype default vs newest
* [523c5b27](https://github.com/hyperledger/fabric-sdk-go/commit/523c5b27) [FABG-725] Ticker should be stopped in conn janitor
* [16651c0d](https://github.com/hyperledger/fabric-sdk-go/commit/16651c0d) Updated list of maintainers
* [f6c4a63d](https://github.com/hyperledger/fabric-sdk-go/commit/f6c4a63d) [FABG-718] increase timeout for event client test
* [2b73bc23](https://github.com/hyperledger/fabric-sdk-go/commit/2b73bc23) [FABG-721] HSM Resilience in internal/fabric/bccsp
* [beccd9cb](https://github.com/hyperledger/fabric-sdk-go/commit/beccd9cb) [FABG-718] fix for ChainCodeEvent getting older events
* [b25b359f](https://github.com/hyperledger/fabric-sdk-go/commit/b25b359f) [FABG-712] Import latest fabric CA code
* [36698e6d](https://github.com/hyperledger/fabric-sdk-go/commit/36698e6d) [FABG-717] fix - Unhandled Panic in TLSCertHash
* [ed12209d](https://github.com/hyperledger/fabric-sdk-go/commit/ed12209d) [FABG-712] Import latest fabric code
* [28ed3d22](https://github.com/hyperledger/fabric-sdk-go/commit/28ed3d22) [FABG-714] fix typoes
* [9f88b347](https://github.com/hyperledger/fabric-sdk-go/commit/9f88b347) [FABG-703] SSL target override for CA config
* [5754d267](https://github.com/hyperledger/fabric-sdk-go/commit/5754d267) [FABG-716] Remove Event Hub
* [6c18d00e](https://github.com/hyperledger/fabric-sdk-go/commit/6c18d00e) [FABG-713] fix: remove duplicated line and a typo
* [708bb639](https://github.com/hyperledger/fabric-sdk-go/commit/708bb639) [FABG-711] Update to dep 0.5.0
* [4e3fe268](https://github.com/hyperledger/fabric-sdk-go/commit/4e3fe268) [FABG-710] Panic in local integration tests
* [d542077c](https://github.com/hyperledger/fabric-sdk-go/commit/d542077c) [FABG-709] Use full SourceURL in E2E test
* [779a8c11](https://github.com/hyperledger/fabric-sdk-go/commit/779a8c11) [FABG-708] No orderer e2e: verify from same peer
* [cacc1a71](https://github.com/hyperledger/fabric-sdk-go/commit/cacc1a71) [FABG-707] Increase max time window on events test
* [52884c8d](https://github.com/hyperledger/fabric-sdk-go/commit/52884c8d) [FABG-706] Fix bug in check for idle event client
* [d2e3b479](https://github.com/hyperledger/fabric-sdk-go/commit/d2e3b479) [FABG-705] fix: override clientTLS in backend
* [4b2d30f1](https://github.com/hyperledger/fabric-sdk-go/commit/4b2d30f1) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] tls certpool lock fix
* [c0144e2f](https://github.com/hyperledger/fabric-sdk-go/commit/c0144e2f) [FABG-692] fix make dockerenv-stable-up fail
* [972a009d](https://github.com/hyperledger/fabric-sdk-go/commit/972a009d) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] tls certpool lock fix
* [04245219](https://github.com/hyperledger/fabric-sdk-go/commit/04245219) [FABG-34] discovery:skip new peer till config refresh
* [b72f9cfc](https://github.com/hyperledger/fabric-sdk-go/commit/b72f9cfc) [FABG-700] fix typoes and imports
* [530eb847](https://github.com/hyperledger/fabric-sdk-go/commit/530eb847) [FABG-701] fix: use variable defaultVal, not const true
* [fe01e6e8](https://github.com/hyperledger/fabric-sdk-go/commit/fe01e6e8) [FABG-698] Event client reconn bug fix and enhancements
* [fa73e440](https://github.com/hyperledger/fabric-sdk-go/commit/fa73e440) [FABG-699] Add SDK config for reconnecting event client
* [bf173b43](https://github.com/hyperledger/fabric-sdk-go/commit/bf173b43) [FABG-698] Event client reconnect peer at higher block
* [233827d6](https://github.com/hyperledger/fabric-sdk-go/commit/233827d6) [FABG-697] don't query all the peers for discovery
* [a93c67eb](https://github.com/hyperledger/fabric-sdk-go/commit/a93c67eb) [[FAB-11399](https://jira.hyperledger.org/browse/FAB-11399)] Remove duplicate targets for proposal
* [4b6b3f86](https://github.com/hyperledger/fabric-sdk-go/commit/4b6b3f86) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] improving certpool peformance
* [0d623776](https://github.com/hyperledger/fabric-sdk-go/commit/0d623776) [FABG-693] Add Block Height to Dynamic Discovery Peers
* [a68049c9](https://github.com/hyperledger/fabric-sdk-go/commit/a68049c9) [FABG-691] Discard invalid cert
* [72bc6c64](https://github.com/hyperledger/fabric-sdk-go/commit/72bc6c64) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] int tests - discovery peer msp id
* [2fd5e4f8](https://github.com/hyperledger/fabric-sdk-go/commit/2fd5e4f8) [FABG-686] add Tally performance support to SDK
* [5809787e](https://github.com/hyperledger/fabric-sdk-go/commit/5809787e) [FABG-125] Add failfast test option to int tests
* [6eee5caa](https://github.com/hyperledger/fabric-sdk-go/commit/6eee5caa) [FABG-687] Ensure CC installed within int tests
* [0cb141ee](https://github.com/hyperledger/fabric-sdk-go/commit/0cb141ee) [FABG-688] Cleanup gitignore
* [ae6e1de1](https://github.com/hyperledger/fabric-sdk-go/commit/ae6e1de1) [FABG-688] Dynamically create crypto fixtures
* [37a93d37](https://github.com/hyperledger/fabric-sdk-go/commit/37a93d37) [[FAB-11128](https://jira.hyperledger.org/browse/FAB-11128)] Fix devstable pull of chaincoded base
* [1ed507fc](https://github.com/hyperledger/fabric-sdk-go/commit/1ed507fc) [[FAB-11229](https://jira.hyperledger.org/browse/FAB-11229)] Check for chaincode installed in orgs test
* [eebf8510](https://github.com/hyperledger/fabric-sdk-go/commit/eebf8510) [[FAB-11207](https://jira.hyperledger.org/browse/FAB-11207)] Fix local integration tests
* [4fadae50](https://github.com/hyperledger/fabric-sdk-go/commit/4fadae50) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] pinning scripts for pkcs11 object handle
* [b564dd5e](https://github.com/hyperledger/fabric-sdk-go/commit/b564dd5e) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] minor fix in pkcs11 object handle
* [73eef907](https://github.com/hyperledger/fabric-sdk-go/commit/73eef907) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] default config 'default' to '_default'
* [868c695f](https://github.com/hyperledger/fabric-sdk-go/commit/868c695f) [[FAB-11187](https://jira.hyperledger.org/browse/FAB-11187)] Extend event client test timing window
* [4f98d626](https://github.com/hyperledger/fabric-sdk-go/commit/4f98d626) [[FAB-11195](https://jira.hyperledger.org/browse/FAB-11195)] Events integration test retry txn errors
* [11958632](https://github.com/hyperledger/fabric-sdk-go/commit/11958632) [[FAB-11194](https://jira.hyperledger.org/browse/FAB-11194)] Fix human readable status code mapping
* [d9ca5ec8](https://github.com/hyperledger/fabric-sdk-go/commit/d9ca5ec8) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] default peer/orderer for url search
* [227d4ed4](https://github.com/hyperledger/fabric-sdk-go/commit/227d4ed4) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] IgnoreEndpoint minor fixes
* [37014538](https://github.com/hyperledger/fabric-sdk-go/commit/37014538) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] IgnoreEndpoint in peer/orderer/ca config
* [03823667](https://github.com/hyperledger/fabric-sdk-go/commit/03823667) [[FAB-11184](https://jira.hyperledger.org/browse/FAB-11184)] Set batch timeout to 500ms
* [d489eba9](https://github.com/hyperledger/fabric-sdk-go/commit/d489eba9) [[FAB-11156](https://jira.hyperledger.org/browse/FAB-11156)] update probuf+grpc package versions
* [51402604](https://github.com/hyperledger/fabric-sdk-go/commit/51402604) [[FAB-11154](https://jira.hyperledger.org/browse/FAB-11154)] make concurrent ChClient Reqs
* [27e0fd88](https://github.com/hyperledger/fabric-sdk-go/commit/27e0fd88) [[FAB-11040](https://jira.hyperledger.org/browse/FAB-11040)] Organize integration tests folder
* [4a8d4ae5](https://github.com/hyperledger/fabric-sdk-go/commit/4a8d4ae5) [[FAB-11134](https://jira.hyperledger.org/browse/FAB-11134)] Add reset to example_cc
* [78f2bba3](https://github.com/hyperledger/fabric-sdk-go/commit/78f2bba3) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] minor fix in pkcs11 object handle
* [fce3f3b4](https://github.com/hyperledger/fabric-sdk-go/commit/fce3f3b4) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] pinning scripts for pkcs11 object handle
* [b8a0344f](https://github.com/hyperledger/fabric-sdk-go/commit/b8a0344f) [[FAB-11128](https://jira.hyperledger.org/browse/FAB-11128)] Add retries to docker interceptor
* [5cecd167](https://github.com/hyperledger/fabric-sdk-go/commit/5cecd167) [[FAB-11135](https://jira.hyperledger.org/browse/FAB-11135)] Add cache for pkcs11 object handle
* [93590f03](https://github.com/hyperledger/fabric-sdk-go/commit/93590f03) [[FAB-11137](https://jira.hyperledger.org/browse/FAB-11137)] Include dependencies of test imports
* [a5e3ef8a](https://github.com/hyperledger/fabric-sdk-go/commit/a5e3ef8a) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] default peer and orderer config
* [e230c04e](https://github.com/hyperledger/fabric-sdk-go/commit/e230c04e) [[FAB-11128](https://jira.hyperledger.org/browse/FAB-11128)] Mock chaincode daemon
* [f771159d](https://github.com/hyperledger/fabric-sdk-go/commit/f771159d) [[FAB-11040](https://jira.hyperledger.org/browse/FAB-11040)] Fix typo in clean-tests target
* [333c1b92](https://github.com/hyperledger/fabric-sdk-go/commit/333c1b92) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] regex replace for entitymatcher mappedhost
* [02ba89a7](https://github.com/hyperledger/fabric-sdk-go/commit/02ba89a7) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] entity matchers refactoring
* [b5453c1d](https://github.com/hyperledger/fabric-sdk-go/commit/b5453c1d) [[FAB-10552](https://jira.hyperledger.org/browse/FAB-10552)] removed '_' from folder name
* [32bd5784](https://github.com/hyperledger/fabric-sdk-go/commit/32bd5784) [[FAB-11040](https://jira.hyperledger.org/browse/FAB-11040)] Organize PKCS11 integration tests
* [22115516](https://github.com/hyperledger/fabric-sdk-go/commit/22115516) [[FAB-11093](https://jira.hyperledger.org/browse/FAB-11093)] Cache vendor populate step
* [47c8f805](https://github.com/hyperledger/fabric-sdk-go/commit/47c8f805) [[FAB-11068](https://jira.hyperledger.org/browse/FAB-11068)] Resolve additional endorsers for pvt data
* [cb683524](https://github.com/hyperledger/fabric-sdk-go/commit/cb683524) [[FAB-11090](https://jira.hyperledger.org/browse/FAB-11090)] Lock dependency version and reduce freq
* [c8d677da](https://github.com/hyperledger/fabric-sdk-go/commit/c8d677da) [[FAB-11078](https://jira.hyperledger.org/browse/FAB-11078)] Increase cc execute timeout for tests
* [e1b9f492](https://github.com/hyperledger/fabric-sdk-go/commit/e1b9f492) [[FAB-11075](https://jira.hyperledger.org/browse/FAB-11075)] Change default arch to amd64
* [7f256568](https://github.com/hyperledger/fabric-sdk-go/commit/7f256568) [[FAB-10625](https://jira.hyperledger.org/browse/FAB-10625)] Local tests in CI
* [945c198d](https://github.com/hyperledger/fabric-sdk-go/commit/945c198d) [[FAB-11070](https://jira.hyperledger.org/browse/FAB-11070)] Update devstable to 1.3
* [c61ab2b2](https://github.com/hyperledger/fabric-sdk-go/commit/c61ab2b2) [[FAB-11064](https://jira.hyperledger.org/browse/FAB-11064)] Account for deleted files in tests
* [22a0f775](https://github.com/hyperledger/fabric-sdk-go/commit/22a0f775) [[FAB-11066](https://jira.hyperledger.org/browse/FAB-11066)] Remove v1.0 fixtures
* [12fab60d](https://github.com/hyperledger/fabric-sdk-go/commit/12fab60d) [[FAB-11066](https://jira.hyperledger.org/browse/FAB-11066)] Update stable version to 1.2.0
* [5af68299](https://github.com/hyperledger/fabric-sdk-go/commit/5af68299) [[FAB-11065](https://jira.hyperledger.org/browse/FAB-11065)] Fix panic in integration test
* [ea5e32e8](https://github.com/hyperledger/fabric-sdk-go/commit/ea5e32e8) [[FAB-11063](https://jira.hyperledger.org/browse/FAB-11063)] entity matchers refactoring
* [bed6fde5](https://github.com/hyperledger/fabric-sdk-go/commit/bed6fde5) [[FAB-11064](https://jira.hyperledger.org/browse/FAB-11064)] Run metalinters once
* [c5fce8dd](https://github.com/hyperledger/fabric-sdk-go/commit/c5fce8dd) [[FAB-11064](https://jira.hyperledger.org/browse/FAB-11064)] calc package dependencies in parallel
* [a2264e7d](https://github.com/hyperledger/fabric-sdk-go/commit/a2264e7d) [[FAB-10552](https://jira.hyperledger.org/browse/FAB-10552)] Benchmarks (Channel Client)
* [9241e0e5](https://github.com/hyperledger/fabric-sdk-go/commit/9241e0e5) [[FAB-11056](https://jira.hyperledger.org/browse/FAB-11056)] Close lazyref on error
* [0c6b7f19](https://github.com/hyperledger/fabric-sdk-go/commit/0c6b7f19) [[FAB-11056](https://jira.hyperledger.org/browse/FAB-11056)] Fix caching in Fabric & Dynamic Selection
* [2ad5906c](https://github.com/hyperledger/fabric-sdk-go/commit/2ad5906c) [FAB-10195](https://jira.hyperledger.org/browse/FAB-10195)] Bump to fabric 1.2.0
* [7a428410](https://github.com/hyperledger/fabric-sdk-go/commit/7a428410) [[FAB-11048](https://jira.hyperledger.org/browse/FAB-11048)] Fix metalinter issues
* [139bfdf9](https://github.com/hyperledger/fabric-sdk-go/commit/139bfdf9) [[FAB-11032](https://jira.hyperledger.org/browse/FAB-11032)] Handle no package changes on CI
* [c1130653](https://github.com/hyperledger/fabric-sdk-go/commit/c1130653) [[FAB-11027](https://jira.hyperledger.org/browse/FAB-11027)] Change errors pkg depend to 0.8.0
* [7a9d5ba4](https://github.com/hyperledger/fabric-sdk-go/commit/7a9d5ba4) [[FAB-10947](https://jira.hyperledger.org/browse/FAB-10947)] chaincode error code handling for v1.2
* [62295202](https://github.com/hyperledger/fabric-sdk-go/commit/62295202) [[FAB-10881](https://jira.hyperledger.org/browse/FAB-10881)] Update base image to 0.4.10
* [632e832b](https://github.com/hyperledger/fabric-sdk-go/commit/632e832b) [[FAB-10944](https://jira.hyperledger.org/browse/FAB-10944)] Use Invoker for int test retries
* [9a5b4fd6](https://github.com/hyperledger/fabric-sdk-go/commit/9a5b4fd6) [[FAB-10941](https://jira.hyperledger.org/browse/FAB-10941)] Fix transient error in selection
* [5f82d04a](https://github.com/hyperledger/fabric-sdk-go/commit/5f82d04a) [[FAB-10895](https://jira.hyperledger.org/browse/FAB-10895)] Channel client test wait for propagation
* [1d60584d](https://github.com/hyperledger/fabric-sdk-go/commit/1d60584d) [[FAB-10918](https://jira.hyperledger.org/browse/FAB-10918)] Increase event buffer size for test
* [6e07791d](https://github.com/hyperledger/fabric-sdk-go/commit/6e07791d) [[FAB-10895](https://jira.hyperledger.org/browse/FAB-10895)] Ledger change block test check all peers
* [142d0f13](https://github.com/hyperledger/fabric-sdk-go/commit/142d0f13) [[FAB-10918](https://jira.hyperledger.org/browse/FAB-10918)] Run integration tests individually
* [f429640f](https://github.com/hyperledger/fabric-sdk-go/commit/f429640f) [[FAB-10884](https://jira.hyperledger.org/browse/FAB-10884)] Fix CC success status >200 treated as error
* [a7dfa535](https://github.com/hyperledger/fabric-sdk-go/commit/a7dfa535) [[FAB-10898](https://jira.hyperledger.org/browse/FAB-10898)] Fix integration tests by adding retries
* [6decf5f3](https://github.com/hyperledger/fabric-sdk-go/commit/6decf5f3) [[FAB-10839](https://jira.hyperledger.org/browse/FAB-10839)] Fix PKCS11 unit test package array
* [3c9aad43](https://github.com/hyperledger/fabric-sdk-go/commit/3c9aad43) [[FAB-10893](https://jira.hyperledger.org/browse/FAB-10893)] Remove trailing spaces in Makefile
* [98164255](https://github.com/hyperledger/fabric-sdk-go/commit/98164255) [[FAB-10894](https://jira.hyperledger.org/browse/FAB-10894)] Fix S390x softhsm image build
* [2e13b4d5](https://github.com/hyperledger/fabric-sdk-go/commit/2e13b4d5) [[FAB-10893](https://jira.hyperledger.org/browse/FAB-10893)] Use canonical blank dockerhub namespace
* [c3f53d84](https://github.com/hyperledger/fabric-sdk-go/commit/c3f53d84) [[FAB-10880](https://jira.hyperledger.org/browse/FAB-10880)] Disable test cache by default
* [2a53231c](https://github.com/hyperledger/fabric-sdk-go/commit/2a53231c) [[FAB-10849](https://jira.hyperledger.org/browse/FAB-10849)] Improve package dependency checks
* [7b9e156b](https://github.com/hyperledger/fabric-sdk-go/commit/7b9e156b) [[FAB-10878](https://jira.hyperledger.org/browse/FAB-10878)] Change WARN to DEBUG
* [d1ad22a7](https://github.com/hyperledger/fabric-sdk-go/commit/d1ad22a7) [[FAB-10848](https://jira.hyperledger.org/browse/FAB-10848)] Invalid seek type in Event Client
* [a6217111](https://github.com/hyperledger/fabric-sdk-go/commit/a6217111) [[FAB-10858](https://jira.hyperledger.org/browse/FAB-10858)] Fix channel artifacts for fabric 1.2
* [cb4ad138](https://github.com/hyperledger/fabric-sdk-go/commit/cb4ad138) [[FAB-10849](https://jira.hyperledger.org/browse/FAB-10849)] Run int tests on changed packages
* [e4496994](https://github.com/hyperledger/fabric-sdk-go/commit/e4496994) [[FAB-10839](https://jira.hyperledger.org/browse/FAB-10839)] Run unit tests on changed packages
* [ed642008](https://github.com/hyperledger/fabric-sdk-go/commit/ed642008) [[FAB-10838](https://jira.hyperledger.org/browse/FAB-10838)] CI lints only changed packages
* [37201e91](https://github.com/hyperledger/fabric-sdk-go/commit/37201e91) [[FAB-10814](https://jira.hyperledger.org/browse/FAB-10814)] use fabric 1.2.0-rc1 docker images
* [43ff3e86](https://github.com/hyperledger/fabric-sdk-go/commit/43ff3e86) [[FAB-10195](https://jira.hyperledger.org/browse/FAB-10195)] Bump to fabric 1.2 RC1
* [e529e89b](https://github.com/hyperledger/fabric-sdk-go/commit/e529e89b) [[FAB-10667](https://jira.hyperledger.org/browse/FAB-10667)] Resolve additional endorsers for CC-to-CC
* [327a946d](https://github.com/hyperledger/fabric-sdk-go/commit/327a946d) [[FAB-10792](https://jira.hyperledger.org/browse/FAB-10792)] Only run e2e test for prev
* [79af81a8](https://github.com/hyperledger/fabric-sdk-go/commit/79af81a8) [[FAB-10694](https://jira.hyperledger.org/browse/FAB-10694)] InitializerWithData in lazyref
* [884d2cdb](https://github.com/hyperledger/fabric-sdk-go/commit/884d2cdb) [[FAB-10694](https://jira.hyperledger.org/browse/FAB-10694)] support for cauthdsl.AcceptAllPolicy
* [1dce66bf](https://github.com/hyperledger/fabric-sdk-go/commit/1dce66bf) [[FAB-10712](https://jira.hyperledger.org/browse/FAB-10712)] Add BeforeRetry callback to Channel Client
* [88455b48](https://github.com/hyperledger/fabric-sdk-go/commit/88455b48) [[FAB-10649](https://jira.hyperledger.org/browse/FAB-10649)] Fix local integration tests for v1.2

## v1.0.0-alpha4
Wed 13 Jun 2018 16:08:12 EDT

* [e1d3ddbf](https://github.com/hyperledger/fabric-sdk-go/commit/e1d3ddbf) [[FAB-10625](https://jira.hyperledger.org/browse/FAB-10625)] fixing local integration test
* [6c6462dd](https://github.com/hyperledger/fabric-sdk-go/commit/6c6462dd) [[FAB-10622](https://jira.hyperledger.org/browse/FAB-10622)] Add Invocation Chain to Channel Client
* [7dca4ca5](https://github.com/hyperledger/fabric-sdk-go/commit/7dca4ca5) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] refactoring endpointConfig.TLSCACertPool
* [c235146d](https://github.com/hyperledger/fabric-sdk-go/commit/c235146d) [[FAB-10605](https://jira.hyperledger.org/browse/FAB-10605)] Choose Selection Service using capabilities
* [74c04eaa](https://github.com/hyperledger/fabric-sdk-go/commit/74c04eaa) [[FAB-10610](https://jira.hyperledger.org/browse/FAB-10610)] Include prior version capabilities
* [68e40669](https://github.com/hyperledger/fabric-sdk-go/commit/68e40669) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] endpoint config refactoring
* [65f5a15d](https://github.com/hyperledger/fabric-sdk-go/commit/65f5a15d) [[FAB-10595](https://jira.hyperledger.org/browse/FAB-10595)] Fabric Selection Integration Tests
* [68e40669](https://github.com/hyperledger/fabric-sdk-go/commit/68e40669) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] endpoint config refactoring
* [83bfc050](https://github.com/hyperledger/fabric-sdk-go/commit/83bfc050) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] endpoint config refactoring
* [1fed3206](https://github.com/hyperledger/fabric-sdk-go/commit/1fed3206) [[FAB-9661](https://jira.hyperledger.org/browse/FAB-9661)] Selection based on Fabric's Discovery
* [b99774ac](https://github.com/hyperledger/fabric-sdk-go/commit/b99774ac) [[FAB-10591](https://jira.hyperledger.org/browse/FAB-10591)] Discovery Client Endorser Tests
* [aedc2268](https://github.com/hyperledger/fabric-sdk-go/commit/aedc2268) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] identity config follow up fix
* [6599f556](https://github.com/hyperledger/fabric-sdk-go/commit/6599f556) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] identity config refactoring
* [6b548678](https://github.com/hyperledger/fabric-sdk-go/commit/6b548678) [[FAB-10543](https://jira.hyperledger.org/browse/FAB-10543)] Fix SendDeliver tests to check closed chan
* [49f88a3b](https://github.com/hyperledger/fabric-sdk-go/commit/49f88a3b) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] remove client and CA from NetworkConfig
* [6314b19a](https://github.com/hyperledger/fabric-sdk-go/commit/6314b19a) [[FAB-10575](https://jira.hyperledger.org/browse/FAB-10575)] update to release version of cfssl
* [9cbdc1bb](https://github.com/hyperledger/fabric-sdk-go/commit/9cbdc1bb) [[FAB-10568](https://jira.hyperledger.org/browse/FAB-10568)] Identity Config refactoring
* [ec9053cc](https://github.com/hyperledger/fabric-sdk-go/commit/ec9053cc) [[FAB-10547](https://jira.hyperledger.org/browse/FAB-10547)] Pvt Collections in Selection API
* [7d9ebd2e](https://github.com/hyperledger/fabric-sdk-go/commit/7d9ebd2e) [[FAB-10556](https://jira.hyperledger.org/browse/FAB-10556)] CI: orderer panic on multi config update
* [63d2604b](https://github.com/hyperledger/fabric-sdk-go/commit/63d2604b) [[FAB-10551](https://jira.hyperledger.org/browse/FAB-10551)] Fix CI error: could not find CC
* [37c1b5a7](https://github.com/hyperledger/fabric-sdk-go/commit/37c1b5a7) [[FAB-10543](https://jira.hyperledger.org/browse/FAB-10543)] gracefully complete orderer calls
* [478d0974](https://github.com/hyperledger/fabric-sdk-go/commit/478d0974) [[FAB-10549](https://jira.hyperledger.org/browse/FAB-10549)] gracefully stop mock servers
* [cc5f23cd](https://github.com/hyperledger/fabric-sdk-go/commit/cc5f23cd) [[FAB-10518](https://jira.hyperledger.org/browse/FAB-10518)] Ensure testing prints to stdout is flushed
* [9ef508fa](https://github.com/hyperledger/fabric-sdk-go/commit/9ef508fa) [[FAB-10417](https://jira.hyperledger.org/browse/FAB-10417)] endpointconfig TLS client certs preload
* [df181da0](https://github.com/hyperledger/fabric-sdk-go/commit/df181da0) [[FAB-10518](https://jira.hyperledger.org/browse/FAB-10518)] Printf missing new lines
* [81694dd6](https://github.com/hyperledger/fabric-sdk-go/commit/81694dd6) [[FAB-10511](https://jira.hyperledger.org/browse/FAB-10511)] clean up sdk log calls
* [38d3ff88](https://github.com/hyperledger/fabric-sdk-go/commit/38d3ff88) [[FAB-10195](https://jira.hyperledger.org/browse/FAB-10195)] Bump to latest fabric
* [f58e994d](https://github.com/hyperledger/fabric-sdk-go/commit/f58e994d) [[FAB-10456](https://jira.hyperledger.org/browse/FAB-10456)] Use dynamic selection by default
* [428a1e62](https://github.com/hyperledger/fabric-sdk-go/commit/428a1e62) [[FAB-10417](https://jira.hyperledger.org/browse/FAB-10417)] Remove nil checks on NetworkConfig
* [d8cbd7f1](https://github.com/hyperledger/fabric-sdk-go/commit/d8cbd7f1) [[FAB-10495](https://jira.hyperledger.org/browse/FAB-10495)] Make channel config parsing less verbose
* [68c1ad06](https://github.com/hyperledger/fabric-sdk-go/commit/68c1ad06) [[FAB-10180](https://jira.hyperledger.org/browse/FAB-10180)] Discovery service based on capabilities
* [e0670202](https://github.com/hyperledger/fabric-sdk-go/commit/e0670202) [[FAB-10417](https://jira.hyperledger.org/browse/FAB-10417)] endpointconfig return type refactoring
* [a6ae0710](https://github.com/hyperledger/fabric-sdk-go/commit/a6ae0710) [[FAB-10491](https://jira.hyperledger.org/browse/FAB-10491)] Fix go routine leak in channel client
* [9bae2501](https://github.com/hyperledger/fabric-sdk-go/commit/9bae2501) [[FAB-10453](https://jira.hyperledger.org/browse/FAB-10453)] Discovery, Selection moved to chprovider
* [ced92a7e](https://github.com/hyperledger/fabric-sdk-go/commit/ced92a7e) [[FAB-10481](https://jira.hyperledger.org/browse/FAB-10481)] Fix data race in configless test
* [58ce93d3](https://github.com/hyperledger/fabric-sdk-go/commit/58ce93d3) [[FAB-10472](https://jira.hyperledger.org/browse/FAB-10472)] Fix configless e2e test
* [48f4a75d](https://github.com/hyperledger/fabric-sdk-go/commit/48f4a75d) [[FAB-10279](https://jira.hyperledger.org/browse/FAB-10279)] endpointConfig TLS/network refactoring
* [c83d774e](https://github.com/hyperledger/fabric-sdk-go/commit/c83d774e) [[FAB-10455](https://jira.hyperledger.org/browse/FAB-10455)] Clean up user store after test completion
* [1a4824e4](https://github.com/hyperledger/fabric-sdk-go/commit/1a4824e4) [[FAB-10325](https://jira.hyperledger.org/browse/FAB-10325)] Moved channel funcs out of InfraProvider
* [f3528ca3](https://github.com/hyperledger/fabric-sdk-go/commit/f3528ca3) [[FAB-10279](https://jira.hyperledger.org/browse/FAB-10279)] refactoring endpointconfig
* [9eaeaecb](https://github.com/hyperledger/fabric-sdk-go/commit/9eaeaecb) [[FAB-10422](https://jira.hyperledger.org/browse/FAB-10422)] Hide print of private keys
* [1663815e](https://github.com/hyperledger/fabric-sdk-go/commit/1663815e) [[FAB-9574](https://jira.hyperledger.org/browse/FAB-9574)] overridable Identity & Crypto Configs
* [e8566fe4](https://github.com/hyperledger/fabric-sdk-go/commit/e8566fe4) [[FAB-10279](https://jira.hyperledger.org/browse/FAB-10279)] pinning script updates for fabric-ca
* [e255416d](https://github.com/hyperledger/fabric-sdk-go/commit/e255416d) [[FAB-10196](https://jira.hyperledger.org/browse/FAB-10196)] Bootstrap test
* [300b3e74](https://github.com/hyperledger/fabric-sdk-go/commit/300b3e74) [[FAB-10279](https://jira.hyperledger.org/browse/FAB-10279)] fabric-ca client updates
* [3cc5ea3f](https://github.com/hyperledger/fabric-sdk-go/commit/3cc5ea3f) [[FAB-10195](https://jira.hyperledger.org/browse/FAB-10195)] Bump to latest fabric
* [81cbc396](https://github.com/hyperledger/fabric-sdk-go/commit/81cbc396) [[FAB-10179](https://jira.hyperledger.org/browse/FAB-10179)] Pick event service using capabilities
* [c776993b](https://github.com/hyperledger/fabric-sdk-go/commit/c776993b) [[FAB-10178](https://jira.hyperledger.org/browse/FAB-10178)] Add Capabilities to Channel Config
* [f3fee0ed](https://github.com/hyperledger/fabric-sdk-go/commit/f3fee0ed) [[FAB-10197](https://jira.hyperledger.org/browse/FAB-10197)] Fix CachingConnector deadlock
* [e3515e52](https://github.com/hyperledger/fabric-sdk-go/commit/e3515e52) [[FAB-10186](https://jira.hyperledger.org/browse/FAB-10186)] suppress repetitive config WARNs
* [2933804b](https://github.com/hyperledger/fabric-sdk-go/commit/2933804b) [[FAB-10079](https://jira.hyperledger.org/browse/FAB-10079)] Use successful response from discovery
* [89d1ed6f](https://github.com/hyperledger/fabric-sdk-go/commit/89d1ed6f) [[FAB-10063](https://jira.hyperledger.org/browse/FAB-10063)] Fix Access Denied error in Local Discovery
* [dd2868aa](https://github.com/hyperledger/fabric-sdk-go/commit/dd2868aa) [[FAB-9983](https://jira.hyperledger.org/browse/FAB-9983)] fixing endpointconfig channels
* [5ac2d979](https://github.com/hyperledger/fabric-sdk-go/commit/5ac2d979) [[FAB-10016](https://jira.hyperledger.org/browse/FAB-10016)] Handle a non-existent organization
* [0afd9938](https://github.com/hyperledger/fabric-sdk-go/commit/0afd9938) [[FAB-10009](https://jira.hyperledger.org/browse/FAB-10009)] Integration tests cleanup
* [ad035410](https://github.com/hyperledger/fabric-sdk-go/commit/ad035410) [[FAB-9896](https://jira.hyperledger.org/browse/FAB-9896)] MSP Client: Identity Management
* [5ea7cde7](https://github.com/hyperledger/fabric-sdk-go/commit/5ea7cde7) [[FAB-9983](https://jira.hyperledger.org/browse/FAB-9983)] fixing endpointconfig channels
* [156c189d](https://github.com/hyperledger/fabric-sdk-go/commit/156c189d) [[FAB-9975](https://jira.hyperledger.org/browse/FAB-9975)] logging: Docs and Examples
* [eadf6688](https://github.com/hyperledger/fabric-sdk-go/commit/eadf6688) [[FAB-9965](https://jira.hyperledger.org/browse/FAB-9965)] unit-test: Remove duplicated code
* [3086d546](https://github.com/hyperledger/fabric-sdk-go/commit/3086d546) [[FAB-9954](https://jira.hyperledger.org/browse/FAB-9954)] added certpool instance unit test
* [62edb3e1](https://github.com/hyperledger/fabric-sdk-go/commit/62edb3e1) [[FAB-9940](https://jira.hyperledger.org/browse/FAB-9940)] add "seek from" support for deliveryevent
* [7ecb50df](https://github.com/hyperledger/fabric-sdk-go/commit/7ecb50df) [[FAB-9954](https://jira.hyperledger.org/browse/FAB-9954)] sdk.New() updating identityconfig
* [491240c0](https://github.com/hyperledger/fabric-sdk-go/commit/491240c0) [[FAB-9952](https://jira.hyperledger.org/browse/FAB-9952)] Msp client: Docs and Examples
* [f06da85b](https://github.com/hyperledger/fabric-sdk-go/commit/f06da85b) [[FAB-9935](https://jira.hyperledger.org/browse/FAB-9935)] fixing typo in peerendorser
* [da948b4c](https://github.com/hyperledger/fabric-sdk-go/commit/da948b4c) [[FAB-9900](https://jira.hyperledger.org/browse/FAB-9900)] WithOrdererURL & WithTargetURLs renaming
* [90930224](https://github.com/hyperledger/fabric-sdk-go/commit/90930224) [[FAB-9935](https://jira.hyperledger.org/browse/FAB-9935)] retry on 'chaincode is already launching'
* [db6ac7bf](https://github.com/hyperledger/fabric-sdk-go/commit/db6ac7bf) [[FAB-9929](https://jira.hyperledger.org/browse/FAB-9929)] refactor pkg/fab/resource/api
* [914555ea](https://github.com/hyperledger/fabric-sdk-go/commit/914555ea) [[FAB-9871](https://jira.hyperledger.org/browse/FAB-9871)] reverting devstable int-test workarounds
* [56ed3fee](https://github.com/hyperledger/fabric-sdk-go/commit/56ed3fee) [[FAB-9850](https://jira.hyperledger.org/browse/FAB-9850)] Remove identity from channel cfg cache key
* [ae5783a1](https://github.com/hyperledger/fabric-sdk-go/commit/ae5783a1) [[FAB-9068](https://jira.hyperledger.org/browse/FAB-9068)] META-INF directory in root of ccpackage
* [93339443](https://github.com/hyperledger/fabric-sdk-go/commit/93339443) [[FAB-9808](https://jira.hyperledger.org/browse/FAB-9808)] test config file cleanup
* [c13c937c](https://github.com/hyperledger/fabric-sdk-go/commit/c13c937c) [[FAB-9849](https://jira.hyperledger.org/browse/FAB-9849)] Document status, multi, retry packages
* [0404e72e](https://github.com/hyperledger/fabric-sdk-go/commit/0404e72e) [[FAB-9822](https://jira.hyperledger.org/browse/FAB-9822)] Resource Mgmt: Docs and Examples
* [ee50b2f7](https://github.com/hyperledger/fabric-sdk-go/commit/ee50b2f7) [[FAB-9574](https://jira.hyperledger.org/browse/FAB-9574)] Sub interfaces Integration Test
* [72a5fdb4](https://github.com/hyperledger/fabric-sdk-go/commit/72a5fdb4) [[FAB-9808](https://jira.hyperledger.org/browse/FAB-9808)] override orderers configuration
* [fc7ae840](https://github.com/hyperledger/fabric-sdk-go/commit/fc7ae840) [[FAB-9602](https://jira.hyperledger.org/browse/FAB-9602)] Config overrides inplace of custom backends
* [08038bf8](https://github.com/hyperledger/fabric-sdk-go/commit/08038bf8) [[FAB-9602](https://jira.hyperledger.org/browse/FAB-9602)] refactoring network peers logic
* [bd1e7b9f](https://github.com/hyperledger/fabric-sdk-go/commit/bd1e7b9f) [[FAB-9767](https://jira.hyperledger.org/browse/FAB-9767)] Ledger Client: Docs and Examples
* [63b9e487](https://github.com/hyperledger/fabric-sdk-go/commit/63b9e487) [[FAB-9717](https://jira.hyperledger.org/browse/FAB-9717)] stabilising integration-tests-local
* [0c8195c1](https://github.com/hyperledger/fabric-sdk-go/commit/0c8195c1) [[FAB-9736](https://jira.hyperledger.org/browse/FAB-9736)] TLS config should not be required
* [4e624542](https://github.com/hyperledger/fabric-sdk-go/commit/4e624542) [[FAB-9717](https://jira.hyperledger.org/browse/FAB-9717)] support for multiple config backends
* [a9d0d918](https://github.com/hyperledger/fabric-sdk-go/commit/a9d0d918) [[FAB-9710](https://jira.hyperledger.org/browse/FAB-9710)] Channel Client: Documentation
* [29957cae](https://github.com/hyperledger/fabric-sdk-go/commit/29957cae) [[FAB-9678](https://jira.hyperledger.org/browse/FAB-9678)] Channel Client: Examples
* [ab35fb89](https://github.com/hyperledger/fabric-sdk-go/commit/ab35fb89) [[FAB-9601](https://jira.hyperledger.org/browse/FAB-9601)] Move cert pool wrapper into its own package
* [830bdeaa](https://github.com/hyperledger/fabric-sdk-go/commit/830bdeaa) [[FAB-9601](https://jira.hyperledger.org/browse/FAB-9601)] Make system cert pool access thread safe
* [e328e9d6](https://github.com/hyperledger/fabric-sdk-go/commit/e328e9d6) [[FAB-9574](https://jira.hyperledger.org/browse/FAB-9574)] overridable Endpoint configs
* [07288147](https://github.com/hyperledger/fabric-sdk-go/commit/07288147) [[FAB-9602](https://jira.hyperledger.org/browse/FAB-9602)] entity matcher refactoring
* [92ada513](https://github.com/hyperledger/fabric-sdk-go/commit/92ada513) [[FAB-9597](https://jira.hyperledger.org/browse/FAB-9597)] Discovery Service Integration Tests
* [637b6556](https://github.com/hyperledger/fabric-sdk-go/commit/637b6556) [[FAB-9555](https://jira.hyperledger.org/browse/FAB-9555)] Local Discovery Provider
* [69158473](https://github.com/hyperledger/fabric-sdk-go/commit/69158473) [[FAB-9554](https://jira.hyperledger.org/browse/FAB-9554)] Dynamic Discovery Provider
* [65a56e23](https://github.com/hyperledger/fabric-sdk-go/commit/65a56e23) [[FAB-9601](https://jira.hyperledger.org/browse/FAB-9601)] Load system cert pool at config init
* [8b4777e6](https://github.com/hyperledger/fabric-sdk-go/commit/8b4777e6) [[FAB-8802](https://jira.hyperledger.org/browse/FAB-8802)] Integrate discovery client
* [ed467a4f](https://github.com/hyperledger/fabric-sdk-go/commit/ed467a4f) [[FAB-9602](https://jira.hyperledger.org/browse/FAB-9602)] Fix for recursive calling of entitymatchers
* [f7ec0aad](https://github.com/hyperledger/fabric-sdk-go/commit/f7ec0aad) [[FAB-9540](https://jira.hyperledger.org/browse/FAB-9540)] bccsp provider type - case insensitive
* [8235b213](https://github.com/hyperledger/fabric-sdk-go/commit/8235b213) [[FAB-9601](https://jira.hyperledger.org/browse/FAB-9601)] Optimize system trust store loading
* [50f07dd9](https://github.com/hyperledger/fabric-sdk-go/commit/50f07dd9) [[FAB-9598](https://jira.hyperledger.org/browse/FAB-9598)] Store config block number
* [acb200e5](https://github.com/hyperledger/fabric-sdk-go/commit/acb200e5) [[FAB-9596](https://jira.hyperledger.org/browse/FAB-9596)] ChClientOptions vs ResMgmtOptions
* [95c85744](https://github.com/hyperledger/fabric-sdk-go/commit/95c85744) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] Identity Config - Resolving paths/pems
* [f4e828f8](https://github.com/hyperledger/fabric-sdk-go/commit/f4e828f8) [[FAB-9543](https://jira.hyperledger.org/browse/FAB-9543)] Remove ephemeral from config
* [91dfe653](https://github.com/hyperledger/fabric-sdk-go/commit/91dfe653) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [7b5d38e2](https://github.com/hyperledger/fabric-sdk-go/commit/7b5d38e2) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [dfc3d5c2](https://github.com/hyperledger/fabric-sdk-go/commit/dfc3d5c2) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [6c1b8bce](https://github.com/hyperledger/fabric-sdk-go/commit/6c1b8bce) [[FAB-9537](https://jira.hyperledger.org/browse/FAB-9537)] Import Discovery Client Local Peers
* [d38bc9c9](https://github.com/hyperledger/fabric-sdk-go/commit/d38bc9c9) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] PeerChannelConfig loading logic refactoring
* [ea94ad20](https://github.com/hyperledger/fabric-sdk-go/commit/ea94ad20) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [ad4e489e](https://github.com/hyperledger/fabric-sdk-go/commit/ad4e489e) [[FAB-9509](https://jira.hyperledger.org/browse/FAB-9509)] JoinChannel: genesis block retrieval failed
* [c8911f35](https://github.com/hyperledger/fabric-sdk-go/commit/c8911f35) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [f8fb6c2a](https://github.com/hyperledger/fabric-sdk-go/commit/f8fb6c2a) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] revoke test to use config yaml
* [66ac24ed](https://github.com/hyperledger/fabric-sdk-go/commit/66ac24ed) [[FAB-9490](https://jira.hyperledger.org/browse/FAB-9490)] Integration: Use require instead of assert
* [6b90b5fe](https://github.com/hyperledger/fabric-sdk-go/commit/6b90b5fe) [[FAB-9471](https://jira.hyperledger.org/browse/FAB-9471)] Set attributes during user registration
* [a5c9ea33](https://github.com/hyperledger/fabric-sdk-go/commit/a5c9ea33) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [b16df6df](https://github.com/hyperledger/fabric-sdk-go/commit/b16df6df) [[FAB-9476](https://jira.hyperledger.org/browse/FAB-9476)]: WithTargets() should check for nil targets
* [f332abab](https://github.com/hyperledger/fabric-sdk-go/commit/f332abab) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] local int tests to use configbackend
* [1ed6f7fc](https://github.com/hyperledger/fabric-sdk-go/commit/1ed6f7fc) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [d2abffb4](https://github.com/hyperledger/fabric-sdk-go/commit/d2abffb4) [[FAB-9445](https://jira.hyperledger.org/browse/FAB-9445)] Endpoint Options: cc query, endorsing peer
* [0842fffb](https://github.com/hyperledger/fabric-sdk-go/commit/0842fffb) [faB-9312] Resolve metalinter warnings
* [65c0afe6](https://github.com/hyperledger/fabric-sdk-go/commit/65c0afe6) [[FAB-9450](https://jira.hyperledger.org/browse/FAB-9450)] Simplify path substitution
* [0bab3c91](https://github.com/hyperledger/fabric-sdk-go/commit/0bab3c91) [faB-9312] Resolve metalinter warnings
* [9160ae3e](https://github.com/hyperledger/fabric-sdk-go/commit/9160ae3e) [[FAB-9398](https://jira.hyperledger.org/browse/FAB-9398)] Mapping for channel config
* [03bf99b1](https://github.com/hyperledger/fabric-sdk-go/commit/03bf99b1) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [3a2cb7c9](https://github.com/hyperledger/fabric-sdk-go/commit/3a2cb7c9) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] config backend in integration tests
* [9f82f8ec](https://github.com/hyperledger/fabric-sdk-go/commit/9f82f8ec) [[FAB-9412](https://jira.hyperledger.org/browse/FAB-9412)] Remove stream from connection
* [9056f46f](https://github.com/hyperledger/fabric-sdk-go/commit/9056f46f) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] config backend in unit-tests
* [fe28d902](https://github.com/hyperledger/fabric-sdk-go/commit/fe28d902) [[FAB-9411](https://jira.hyperledger.org/browse/FAB-9411)] LazyRef finalizer should provide value
* [db394dc0](https://github.com/hyperledger/fabric-sdk-go/commit/db394dc0) [[FAB-9410](https://jira.hyperledger.org/browse/FAB-9410)] Import Discovery Client
* [0eaf2174](https://github.com/hyperledger/fabric-sdk-go/commit/0eaf2174) [[FAB-9077](https://jira.hyperledger.org/browse/FAB-9077)] Update go version to 1.10
* [926bca6e](https://github.com/hyperledger/fabric-sdk-go/commit/926bca6e) [[FAB-9060](https://jira.hyperledger.org/browse/FAB-9060)]Expiry cert check on TLS connection
* [b57bd108](https://github.com/hyperledger/fabric-sdk-go/commit/b57bd108) [[FAB-9381](https://jira.hyperledger.org/browse/FAB-9381)]: Endpoints: Ledger Query Option
* [d6f75dfe](https://github.com/hyperledger/fabric-sdk-go/commit/d6f75dfe) [[FAB-9377](https://jira.hyperledger.org/browse/FAB-9377)] Define chaincode status group
* [b3960887](https://github.com/hyperledger/fabric-sdk-go/commit/b3960887) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [601dcc50](https://github.com/hyperledger/fabric-sdk-go/commit/601dcc50) [[FAB-9367](https://jira.hyperledger.org/browse/FAB-9367)] Fix constructor of discovery services
* [1231625d](https://github.com/hyperledger/fabric-sdk-go/commit/1231625d) [[FAB-9338](https://jira.hyperledger.org/browse/FAB-9338)] Increase def timeout with empty config
* [4b15d651](https://github.com/hyperledger/fabric-sdk-go/commit/4b15d651) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] Removed unmarshal lookup options
* [33b74d45](https://github.com/hyperledger/fabric-sdk-go/commit/33b74d45) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [582c2fcc](https://github.com/hyperledger/fabric-sdk-go/commit/582c2fcc) [[FAB-9325](https://jira.hyperledger.org/browse/FAB-9325)] make CA matcher section consistent
* [449a24db](https://github.com/hyperledger/fabric-sdk-go/commit/449a24db) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] refactoring config implementations
* [fbae5b75](https://github.com/hyperledger/fabric-sdk-go/commit/fbae5b75) [[FAB-9312](https://jira.hyperledger.org/browse/FAB-9312)] Resolve metalinter warnings
* [1796d395](https://github.com/hyperledger/fabric-sdk-go/commit/1796d395) [[FAB-9321](https://jira.hyperledger.org/browse/FAB-9321)] Panic in caching connector
* [5e8a2a17](https://github.com/hyperledger/fabric-sdk-go/commit/5e8a2a17) [[FAB-9304](https://jira.hyperledger.org/browse/FAB-9304)]: Event Client GoDoc
* [0cb0cf98](https://github.com/hyperledger/fabric-sdk-go/commit/0cb0cf98) [[FAB-9241](https://jira.hyperledger.org/browse/FAB-9241)] Import discovery client
* [64203484](https://github.com/hyperledger/fabric-sdk-go/commit/64203484) [[FAB-9238](https://jira.hyperledger.org/browse/FAB-9238)] refactoring configs to multiple interfaces
* [71d1cde9](https://github.com/hyperledger/fabric-sdk-go/commit/71d1cde9) [[FAB-9229](https://jira.hyperledger.org/browse/FAB-9229)] Restore devstable tag
* [429e1ef2](https://github.com/hyperledger/fabric-sdk-go/commit/429e1ef2) [[FAB-9229](https://jira.hyperledger.org/browse/FAB-9229)] Enable CI target for Fabric 1.2
* [7b5bc2e7](https://github.com/hyperledger/fabric-sdk-go/commit/7b5bc2e7) [[FAB-9195](https://jira.hyperledger.org/browse/FAB-9195)]: Event Client
* [213899c5](https://github.com/hyperledger/fabric-sdk-go/commit/213899c5) [[FAB-9200](https://jira.hyperledger.org/browse/FAB-9200)] Added retries to ResMgmt client
* [d177ebb6](https://github.com/hyperledger/fabric-sdk-go/commit/d177ebb6) [[FAB-9169](https://jira.hyperledger.org/browse/FAB-9169)] remove unused options in config
* [b02be7ab](https://github.com/hyperledger/fabric-sdk-go/commit/b02be7ab) [[FAB-9188](https://jira.hyperledger.org/browse/FAB-9188)] Add retries to integration tests
* [6880d843](https://github.com/hyperledger/fabric-sdk-go/commit/6880d843) [[FAB-9170](https://jira.hyperledger.org/browse/FAB-9170)] Define one WithBlockEvents option
* [84bff1ae](https://github.com/hyperledger/fabric-sdk-go/commit/84bff1ae) [[FAB-9121](https://jira.hyperledger.org/browse/FAB-9121)] Retry on premature execution error
* [8a888320](https://github.com/hyperledger/fabric-sdk-go/commit/8a888320) [[FAB-9095](https://jira.hyperledger.org/browse/FAB-9095)] Panic when no channel policies defined
* [a33102a6](https://github.com/hyperledger/fabric-sdk-go/commit/a33102a6) [[FAB-9094](https://jira.hyperledger.org/browse/FAB-9094)] Set correct status group for cc error
* [14892a64](https://github.com/hyperledger/fabric-sdk-go/commit/14892a64) [[FAB-9092](https://jira.hyperledger.org/browse/FAB-9092)] Move response payload check
* [02c85b37](https://github.com/hyperledger/fabric-sdk-go/commit/02c85b37) [[FAB-9062](https://jira.hyperledger.org/browse/FAB-9062)] set the response status on chaincode success
* [b7343fe2](https://github.com/hyperledger/fabric-sdk-go/commit/b7343fe2) [[FAB-9058](https://jira.hyperledger.org/browse/FAB-9058)] add missing global cache entries
* [7983364a](https://github.com/hyperledger/fabric-sdk-go/commit/7983364a) [[FAB-9038](https://jira.hyperledger.org/browse/FAB-9038)] Wrong tag for couchdb Docker image
* [82ece9c0](https://github.com/hyperledger/fabric-sdk-go/commit/82ece9c0) [[FAB-9054](https://jira.hyperledger.org/browse/FAB-9054)] Ledger Client: Unit Tests
* [a8e2b45d](https://github.com/hyperledger/fabric-sdk-go/commit/a8e2b45d) [[FAB-8900](https://jira.hyperledger.org/browse/FAB-8900)] retry options in ChConfig.Query
* [35387b83](https://github.com/hyperledger/fabric-sdk-go/commit/35387b83) [[FAB-9053](https://jira.hyperledger.org/browse/FAB-9053)] Move util/errors/multi to common/
* [a8567870](https://github.com/hyperledger/fabric-sdk-go/commit/a8567870) [[FAB-9051](https://jira.hyperledger.org/browse/FAB-9051)] Return InstallCC errors when applicable
* [37f0a1a1](https://github.com/hyperledger/fabric-sdk-go/commit/37f0a1a1) [[FAB-9023](https://jira.hyperledger.org/browse/FAB-9023)] SaveChannel Response Struct
* [2a697ced](https://github.com/hyperledger/fabric-sdk-go/commit/2a697ced) [[FAB-9009](https://jira.hyperledger.org/browse/FAB-9009)]Cert expiry check
* [1230546f](https://github.com/hyperledger/fabric-sdk-go/commit/1230546f) [[FAB-7975](https://jira.hyperledger.org/browse/FAB-7975)] added test case for multi error
* [1b6dbd48](https://github.com/hyperledger/fabric-sdk-go/commit/1b6dbd48) [[FAB-9036](https://jira.hyperledger.org/browse/FAB-9036)] Source URL and block num in events
* [39fafe97](https://github.com/hyperledger/fabric-sdk-go/commit/39fafe97) [[FAB-8944](https://jira.hyperledger.org/browse/FAB-8944)] Refresh membership only on config update
* [2061a932](https://github.com/hyperledger/fabric-sdk-go/commit/2061a932) [[FAB-8944](https://jira.hyperledger.org/browse/FAB-8944)] Add more unit tests for channel config cache
* [b73b8942](https://github.com/hyperledger/fabric-sdk-go/commit/b73b8942) [[FAB-8900](https://jira.hyperledger.org/browse/FAB-8900)] policies section in SDK config
* [9880e072](https://github.com/hyperledger/fabric-sdk-go/commit/9880e072) [[FAB-9023](https://jira.hyperledger.org/browse/FAB-9023)] return TransactionID for transactions
* [15453a31](https://github.com/hyperledger/fabric-sdk-go/commit/15453a31) [[FAB-9031](https://jira.hyperledger.org/browse/FAB-9031)] Log success upon reading key from config

## v1.0.0-alpha3
Tue 20 Mar 2018 17:52:25 EDT

* [ccecff43](https://github.com/hyperledger/fabric-sdk-go/commit/ccecff43) [[FAB-7546](https://jira.hyperledger.org/browse/FAB-7546)] Release v1.0.0-alpha3
* [aaeebdd3](https://github.com/hyperledger/fabric-sdk-go/commit/aaeebdd3) [[FAB-8983](https://jira.hyperledger.org/browse/FAB-8983)] make org names consistent in small case
* [d7c5c6a0](https://github.com/hyperledger/fabric-sdk-go/commit/d7c5c6a0) [[FAB-7975](https://jira.hyperledger.org/browse/FAB-7975)] return status code for chaincode error
* [8b74992f](https://github.com/hyperledger/fabric-sdk-go/commit/8b74992f) [[FAB-8912](https://jira.hyperledger.org/browse/FAB-8912)] Update stable target to 1.1.0
* [79b343ba](https://github.com/hyperledger/fabric-sdk-go/commit/79b343ba) [[FAB-8995](https://jira.hyperledger.org/browse/FAB-8995)] Event client should read channel config
* [e3276ec1](https://github.com/hyperledger/fabric-sdk-go/commit/e3276ec1) [[FAB-8985](https://jira.hyperledger.org/browse/FAB-8985)] client,common,fabsdk/... metalinter
* [f46da4c6](https://github.com/hyperledger/fabric-sdk-go/commit/f46da4c6) [[FAB-8982](https://jira.hyperledger.org/browse/FAB-8982)] Remove binary output in debug log
* [05a4b001](https://github.com/hyperledger/fabric-sdk-go/commit/05a4b001) [[FAB-8979](https://jira.hyperledger.org/browse/FAB-8979)] EventURL is no longer required
* [7fe0a78c](https://github.com/hyperledger/fabric-sdk-go/commit/7fe0a78c) [[FAB-8974](https://jira.hyperledger.org/browse/FAB-8974)] Move errors packages from util to common
* [cce14eaf](https://github.com/hyperledger/fabric-sdk-go/commit/cce14eaf) [[FAB-7943](https://jira.hyperledger.org/browse/FAB-7943)] Match Go packager file extensions to Fabric
* [73c53c47](https://github.com/hyperledger/fabric-sdk-go/commit/73c53c47) [[FAB-8964](https://jira.hyperledger.org/browse/FAB-8964)]Log Warning instead of panic
* [4498a415](https://github.com/hyperledger/fabric-sdk-go/commit/4498a415) [[FAB-8943](https://jira.hyperledger.org/browse/FAB-8943)] error codes for entity matchers
* [3ce407e8](https://github.com/hyperledger/fabric-sdk-go/commit/3ce407e8) [[FAB-8939](https://jira.hyperledger.org/browse/FAB-8939)] Organize generated mock directories
* [9a32e554](https://github.com/hyperledger/fabric-sdk-go/commit/9a32e554) [[FAB-8968](https://jira.hyperledger.org/browse/FAB-8968)] Ledger Client: QueryBlockByTxID
* [ce374633](https://github.com/hyperledger/fabric-sdk-go/commit/ce374633) [[FAB-8965](https://jira.hyperledger.org/browse/FAB-8965)] Rename verifiers to verifier
* [86817d76](https://github.com/hyperledger/fabric-sdk-go/commit/86817d76) [[FAB-8965](https://jira.hyperledger.org/browse/FAB-8965)] Resource Mgmt: Verify signature
* [bc115910](https://github.com/hyperledger/fabric-sdk-go/commit/bc115910) [[FAB-8966](https://jira.hyperledger.org/browse/FAB-8966)] Cleanup user data in tests
* [20358d59](https://github.com/hyperledger/fabric-sdk-go/commit/20358d59) [[FAB-8945](https://jira.hyperledger.org/browse/FAB-8945)] Reduce lazyref debug line
* [f1fb12c4](https://github.com/hyperledger/fabric-sdk-go/commit/f1fb12c4) [[FAB-8954](https://jira.hyperledger.org/browse/FAB-8954)] Fix index out of range panic
* [19b1c512](https://github.com/hyperledger/fabric-sdk-go/commit/19b1c512) [[FAB-8963](https://jira.hyperledger.org/browse/FAB-8963)] Instantiate and Upgrade default timeouts
* [944cecc8](https://github.com/hyperledger/fabric-sdk-go/commit/944cecc8) [[FAB-8954](https://jira.hyperledger.org/browse/FAB-8954)] Fix index out of range panic
* [ef2ffa7b](https://github.com/hyperledger/fabric-sdk-go/commit/ef2ffa7b) [[FAB-8956](https://jira.hyperledger.org/browse/FAB-8956)] Reference correct config sections
* [6c342a02](https://github.com/hyperledger/fabric-sdk-go/commit/6c342a02) [[FAB-8900](https://jira.hyperledger.org/browse/FAB-8900)] Reduce default number of peers
* [6ab71379](https://github.com/hyperledger/fabric-sdk-go/commit/6ab71379) [[FAB-8949](https://jira.hyperledger.org/browse/FAB-8949)] Add Close() to discovery, selection services
* [ecb7b037](https://github.com/hyperledger/fabric-sdk-go/commit/ecb7b037) [[FAB-8944](https://jira.hyperledger.org/browse/FAB-8944)] Channel configuration cache refresh
* [9c67a795](https://github.com/hyperledger/fabric-sdk-go/commit/9c67a795) [[FAB-8946](https://jira.hyperledger.org/browse/FAB-8946)] Dynamic selection caching bug
* [ff9763c2](https://github.com/hyperledger/fabric-sdk-go/commit/ff9763c2) [[FAB-8945](https://jira.hyperledger.org/browse/FAB-8945)] Fix lazyref bugs
* [6fac3aab](https://github.com/hyperledger/fabric-sdk-go/commit/6fac3aab) [[FAB-8943](https://jira.hyperledger.org/browse/FAB-8943)] client status code
* [c570198d](https://github.com/hyperledger/fabric-sdk-go/commit/c570198d) [[FAB-8936](https://jira.hyperledger.org/browse/FAB-8936)] making cert pool threadsafe
* [6763ec79](https://github.com/hyperledger/fabric-sdk-go/commit/6763ec79) [[FAB-8941](https://jira.hyperledger.org/browse/FAB-8941)]- extract matching logic from queryconfig
* [79f2f4ee](https://github.com/hyperledger/fabric-sdk-go/commit/79f2f4ee) [[FAB-8940](https://jira.hyperledger.org/browse/FAB-8940)] Refactor pkg.common.msp.Providers interface
* [b54994aa](https://github.com/hyperledger/fabric-sdk-go/commit/b54994aa) [[FAB-8938](https://jira.hyperledger.org/browse/FAB-8938)] Ledger Client: Signature validation
* [cc16a4ad](https://github.com/hyperledger/fabric-sdk-go/commit/cc16a4ad) [[FAB-8935](https://jira.hyperledger.org/browse/FAB-8935)] Update TargetFilter in resmgmt to fab import
* [b2c76cdb](https://github.com/hyperledger/fabric-sdk-go/commit/b2c76cdb) [[FAB-8935](https://jira.hyperledger.org/browse/FAB-8935)] Update mockgen
* [c8b483b7](https://github.com/hyperledger/fabric-sdk-go/commit/c8b483b7) [[FAB-8935](https://jira.hyperledger.org/browse/FAB-8935)] Cleanup folder structure
* [af69efe9](https://github.com/hyperledger/fabric-sdk-go/commit/af69efe9) [[FAB-8732](https://jira.hyperledger.org/browse/FAB-8732)]Negative Peer Test
* [f58af974](https://github.com/hyperledger/fabric-sdk-go/commit/f58af974) [[FAB-8912](https://jira.hyperledger.org/browse/FAB-8912)] update to 1.1.0
* [b15f6d33](https://github.com/hyperledger/fabric-sdk-go/commit/b15f6d33) [[FAB-8902](https://jira.hyperledger.org/browse/FAB-8902)] Use random port for unit tests
* [b05338fe](https://github.com/hyperledger/fabric-sdk-go/commit/b05338fe) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] timeout error message refactoring
* [461e4d71](https://github.com/hyperledger/fabric-sdk-go/commit/461e4d71) [[FAB-8900](https://jira.hyperledger.org/browse/FAB-8900)] random max targets for config block
* [e0e34132](https://github.com/hyperledger/fabric-sdk-go/commit/e0e34132) [[FAB-8910](https://jira.hyperledger.org/browse/FAB-8910)] Add payload for non-filtered CC events
* [523348bb](https://github.com/hyperledger/fabric-sdk-go/commit/523348bb) [[FAB-8899](https://jira.hyperledger.org/browse/FAB-8899)] Dynamic selection service caching peers
* [5346c29c](https://github.com/hyperledger/fabric-sdk-go/commit/5346c29c) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] timeout refactoring in chclient,txnhandler
* [2f551a21](https://github.com/hyperledger/fabric-sdk-go/commit/2f551a21) [DEV-6168] improve debug logging
* [7582442a](https://github.com/hyperledger/fabric-sdk-go/commit/7582442a) [[FAB-8874](https://jira.hyperledger.org/browse/FAB-8874)] Refactor Identity interface
* [6cd78cfc](https://github.com/hyperledger/fabric-sdk-go/commit/6cd78cfc) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] removing redundant time.After
* [59fa8aa0](https://github.com/hyperledger/fabric-sdk-go/commit/59fa8aa0) [[FAB-8872](https://jira.hyperledger.org/browse/FAB-8872)] copy GRPC options in matchers
* [05c52e8b](https://github.com/hyperledger/fabric-sdk-go/commit/05c52e8b) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] parent context opts integration test
* [13097042](https://github.com/hyperledger/fabric-sdk-go/commit/13097042) [[FAB-8866](https://jira.hyperledger.org/browse/FAB-8866)] Handle old config naming for file store
* [9f6e5f6f](https://github.com/hyperledger/fabric-sdk-go/commit/9f6e5f6f) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] request context background and timeouts
* [96fd4012](https://github.com/hyperledger/fabric-sdk-go/commit/96fd4012) [[FAB-8866](https://jira.hyperledger.org/browse/FAB-8866)] Rename UserName to Username
* [f6628507](https://github.com/hyperledger/fabric-sdk-go/commit/f6628507) [[FAB-8866](https://jira.hyperledger.org/browse/FAB-8866)] Rename MspID to MSPID
* [a854819d](https://github.com/hyperledger/fabric-sdk-go/commit/a854819d) [[FAB-8866](https://jira.hyperledger.org/browse/FAB-8866)] Cleanup top level interfaces
* [8241d5c4](https://github.com/hyperledger/fabric-sdk-go/commit/8241d5c4) [[FAB-8865](https://jira.hyperledger.org/browse/FAB-8865)] Clenaup pkg/client/msp
* [cb2f56a0](https://github.com/hyperledger/fabric-sdk-go/commit/cb2f56a0) [[FAB-8860](https://jira.hyperledger.org/browse/FAB-8860)] Remove unused config lines
* [2d729eec](https://github.com/hyperledger/fabric-sdk-go/commit/2d729eec) [[FAB-8862](https://jira.hyperledger.org/browse/FAB-8862)] Basic docs - introduction
* [267c094b](https://github.com/hyperledger/fabric-sdk-go/commit/267c094b) [[FAB-8858](https://jira.hyperledger.org/browse/FAB-8858)] Set max message size to be same as fabric
* [ea3acdbc](https://github.com/hyperledger/fabric-sdk-go/commit/ea3acdbc) [[FAB-8852](https://jira.hyperledger.org/browse/FAB-8852)] Create Peer and Orderer from factory
* [063fd0bc](https://github.com/hyperledger/fabric-sdk-go/commit/063fd0bc) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] request context by higherlevel clients
* [60757942](https://github.com/hyperledger/fabric-sdk-go/commit/60757942) [[FAB-8851](https://jira.hyperledger.org/browse/FAB-8851)] Make options consistent across client pkgs
* [d84447e9](https://github.com/hyperledger/fabric-sdk-go/commit/d84447e9) [[FAB-8847](https://jira.hyperledger.org/browse/FAB-8847)] Add ChannelConfigPath opt for SaveChannel
* [1b231dbc](https://github.com/hyperledger/fabric-sdk-go/commit/1b231dbc) [[FAB-8847](https://jira.hyperledger.org/browse/FAB-8847)] Use Reader in SaveChannel
* [cc128484](https://github.com/hyperledger/fabric-sdk-go/commit/cc128484) [[FAB-8845](https://jira.hyperledger.org/browse/FAB-8845)] Update to usage of WithTargetURLs
* [5fa56968](https://github.com/hyperledger/fabric-sdk-go/commit/5fa56968) [[FAB-8846](https://jira.hyperledger.org/browse/FAB-8846)] Improved key and cert management
* [5c378d38](https://github.com/hyperledger/fabric-sdk-go/commit/5c378d38) [[FAB-8442](https://jira.hyperledger.org/browse/FAB-8442)] map CAConfig to static host
* [92a48d09](https://github.com/hyperledger/fabric-sdk-go/commit/92a48d09) [[FAB-8845](https://jira.hyperledger.org/browse/FAB-8845)] WithTargetURLs option
* [8f5d6f49](https://github.com/hyperledger/fabric-sdk-go/commit/8f5d6f49) [[FAB-8839](https://jira.hyperledger.org/browse/FAB-8839)] Use connection cache in event client
* [cc87976e](https://github.com/hyperledger/fabric-sdk-go/commit/cc87976e) [[FAB-8781](https://jira.hyperledger.org/browse/FAB-8781)] Remove incorrect error log line
* [3cc50db2](https://github.com/hyperledger/fabric-sdk-go/commit/3cc50db2) [[FAB-8828](https://jira.hyperledger.org/browse/FAB-8828)] Cleanup enrollment test cases
* [00d068aa](https://github.com/hyperledger/fabric-sdk-go/commit/00d068aa) [[FAB-8783](https://jira.hyperledger.org/browse/FAB-8783)] Enrollment API Refactoring
* [b1908240](https://github.com/hyperledger/fabric-sdk-go/commit/b1908240) [[FAB-8442](https://jira.hyperledger.org/browse/FAB-8442)] map CAConfig to static host
* [c1724a6e](https://github.com/hyperledger/fabric-sdk-go/commit/c1724a6e) [[FAB-8821](https://jira.hyperledger.org/browse/FAB-8821)] Documentation corrections
* [6e67766f](https://github.com/hyperledger/fabric-sdk-go/commit/6e67766f) [[FAB-8813](https://jira.hyperledger.org/browse/FAB-8813)] Pass GRPC options to event client
* [7bb665cc](https://github.com/hyperledger/fabric-sdk-go/commit/7bb665cc) [[FAB-8811](https://jira.hyperledger.org/browse/FAB-8811)] Rename WithOrdererID to WithOrdererURL
* [f2c267ac](https://github.com/hyperledger/fabric-sdk-go/commit/f2c267ac) [[FAB-8782](https://jira.hyperledger.org/browse/FAB-8782)] Check correct err in connection
* [96a926da](https://github.com/hyperledger/fabric-sdk-go/commit/96a926da) [[FAB-8782](https://jira.hyperledger.org/browse/FAB-8782)] Resolve linter in peer/ord/comm
* [91d09d4d](https://github.com/hyperledger/fabric-sdk-go/commit/91d09d4d) [[FAB-8781](https://jira.hyperledger.org/browse/FAB-8781)] orderer needs to call context ReleaseConn
* [5d117af9](https://github.com/hyperledger/fabric-sdk-go/commit/5d117af9) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] ledger to use channelcontext
* [40ef627d](https://github.com/hyperledger/fabric-sdk-go/commit/40ef627d) [[FAB-8780](https://jira.hyperledger.org/browse/FAB-8780)] Recover a cached conn from server HUP
* [e73e227f](https://github.com/hyperledger/fabric-sdk-go/commit/e73e227f) [[FAB-8778](https://jira.hyperledger.org/browse/FAB-8778)] Move Identity interface to msp
* [a9902679](https://github.com/hyperledger/fabric-sdk-go/commit/a9902679) [[FAB-8776](https://jira.hyperledger.org/browse/FAB-8776)] Recv should be setup prior to Send
* [f0b58959](https://github.com/hyperledger/fabric-sdk-go/commit/f0b58959) [[FAB-8769](https://jira.hyperledger.org/browse/FAB-8769)] Remove channel.Close
* [95febafb](https://github.com/hyperledger/fabric-sdk-go/commit/95febafb) [[FAB-8725](https://jira.hyperledger.org/browse/FAB-8725)] JoinChannel should obey WithOrdererID
* [60d97eea](https://github.com/hyperledger/fabric-sdk-go/commit/60d97eea) [[FAB-8766](https://jira.hyperledger.org/browse/FAB-8766)] Rename NewManager to NewIdentityManager
* [a56cd3a3](https://github.com/hyperledger/fabric-sdk-go/commit/a56cd3a3) [[FAB-8766](https://jira.hyperledger.org/browse/FAB-8766)] Move identitymgr implementation to pkg/msp
* [c2a05226](https://github.com/hyperledger/fabric-sdk-go/commit/c2a05226) [[FAB-8773](https://jira.hyperledger.org/browse/FAB-8773)] Channel Config - Rename Name
* [c826e2f6](https://github.com/hyperledger/fabric-sdk-go/commit/c826e2f6) [[FAB-8772](https://jira.hyperledger.org/browse/FAB-8772)] Remove peer and user setters
* [fa735b92](https://github.com/hyperledger/fabric-sdk-go/commit/fa735b92) [[FAB-8769](https://jira.hyperledger.org/browse/FAB-8769)] Remove all remnants of EventHub
* [431d46c1](https://github.com/hyperledger/fabric-sdk-go/commit/431d46c1) [[FAB-8500](https://jira.hyperledger.org/browse/FAB-8500)] Event Client Integration Test
* [69d7191d](https://github.com/hyperledger/fabric-sdk-go/commit/69d7191d) [[FAB-8768](https://jira.hyperledger.org/browse/FAB-8768)] Option for block events in InfraProvider
* [aec90b3f](https://github.com/hyperledger/fabric-sdk-go/commit/aec90b3f) [[FAB-8762](https://jira.hyperledger.org/browse/FAB-8762)] Enable mutual tls for rc1 test
* [41836c51](https://github.com/hyperledger/fabric-sdk-go/commit/41836c51) [[FAB-8771](https://jira.hyperledger.org/browse/FAB-8771)] disable deprecated tests in CI
* [9a989531](https://github.com/hyperledger/fabric-sdk-go/commit/9a989531) [[FAB-8767](https://jira.hyperledger.org/browse/FAB-8767)] Move targets into its own argument
* [2bfab1c4](https://github.com/hyperledger/fabric-sdk-go/commit/2bfab1c4) [[FAB-8770](https://jira.hyperledger.org/browse/FAB-8770)] Remove extra SDK instances for int tests
* [05201534](https://github.com/hyperledger/fabric-sdk-go/commit/05201534) [[FAB-8765](https://jira.hyperledger.org/browse/FAB-8765)] Identity API - move to msp
* [80b04c7b](https://github.com/hyperledger/fabric-sdk-go/commit/80b04c7b) [[FAB-8707](https://jira.hyperledger.org/browse/FAB-8707)]Remove deprecated channel
* [0c63fda7](https://github.com/hyperledger/fabric-sdk-go/commit/0c63fda7) [[FAB-8756](https://jira.hyperledger.org/browse/FAB-8756)] Remove Event Hub
* [0c30596b](https://github.com/hyperledger/fabric-sdk-go/commit/0c30596b) [[FAB-8684](https://jira.hyperledger.org/browse/FAB-8684)] Split IdentityManager and CA Client impl
* [c9bd65a8](https://github.com/hyperledger/fabric-sdk-go/commit/c9bd65a8) [[FAB-8141](https://jira.hyperledger.org/browse/FAB-8141)] moved out impl from api
* [78c631ea](https://github.com/hyperledger/fabric-sdk-go/commit/78c631ea) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] making fab/resource functions package level
* [18e615eb](https://github.com/hyperledger/fabric-sdk-go/commit/18e615eb) [[FAB-8683](https://jira.hyperledger.org/browse/FAB-8683)] Split IdentityManager and CAClient
* [93b132c2](https://github.com/hyperledger/fabric-sdk-go/commit/93b132c2) [[FAB-8758](https://jira.hyperledger.org/browse/FAB-8758)] Add CloseIfIdle
* [53a4bd0d](https://github.com/hyperledger/fabric-sdk-go/commit/53a4bd0d) [[FAB-8610](https://jira.hyperledger.org/browse/FAB-8610)] Integrate with latest Client Context
* [8e5aa5d4](https://github.com/hyperledger/fabric-sdk-go/commit/8e5aa5d4) [[FAB-8755](https://jira.hyperledger.org/browse/FAB-8755)] Event Client Cleanup
* [9e02cf42](https://github.com/hyperledger/fabric-sdk-go/commit/9e02cf42) [[FAB-8754](https://jira.hyperledger.org/browse/FAB-8754)] Stop event stream when unregister fails
* [9f6f6428](https://github.com/hyperledger/fabric-sdk-go/commit/9f6f6428) [[FAB-8753](https://jira.hyperledger.org/browse/FAB-8753)] fetch peer config by url
* [79d6417c](https://github.com/hyperledger/fabric-sdk-go/commit/79d6417c) [[FAB-8752](https://jira.hyperledger.org/browse/FAB-8752)] Add golang.org context to third_party protos
* [d94bd307](https://github.com/hyperledger/fabric-sdk-go/commit/d94bd307) [[FAB-8717](https://jira.hyperledger.org/browse/FAB-8717)] Request Context
* [4664f803](https://github.com/hyperledger/fabric-sdk-go/commit/4664f803) [[FAB-8571](https://jira.hyperledger.org/browse/FAB-8571)] reverting grpc fallback
* [e8fa23b8](https://github.com/hyperledger/fabric-sdk-go/commit/e8fa23b8) [[FAB-8442](https://jira.hyperledger.org/browse/FAB-8442)] map network hosts config with static-config
* [d92bc587](https://github.com/hyperledger/fabric-sdk-go/commit/d92bc587) [[FAB-8482](https://jira.hyperledger.org/browse/FAB-8482)] using contexts in ccpolicyprovider
* [77ce82b7](https://github.com/hyperledger/fabric-sdk-go/commit/77ce82b7) [[FAB-8719](https://jira.hyperledger.org/browse/FAB-8719)] Rename New Selection/Discovery to Create
* [61c4f1aa](https://github.com/hyperledger/fabric-sdk-go/commit/61c4f1aa) [[FAB-8716](https://jira.hyperledger.org/browse/FAB-8716)] Simplify SDK options
* [9394d3cb](https://github.com/hyperledger/fabric-sdk-go/commit/9394d3cb) [[FAB-8711](https://jira.hyperledger.org/browse/FAB-8711)] update timeout config labels
* [7f5d5530](https://github.com/hyperledger/fabric-sdk-go/commit/7f5d5530) [[FAB-8709](https://jira.hyperledger.org/browse/FAB-8709)] Move low level code to fab (tests)
* [aae553da](https://github.com/hyperledger/fabric-sdk-go/commit/aae553da) [[FAB-8702](https://jira.hyperledger.org/browse/FAB-8702)] Peer dialer methods context parameter
* [eaa71c43](https://github.com/hyperledger/fabric-sdk-go/commit/eaa71c43) [[FAB-8482](https://jira.hyperledger.org/browse/FAB-8482)] Context and ChannelContext integration tests
* [b3bab383](https://github.com/hyperledger/fabric-sdk-go/commit/b3bab383) [[FAB-8699](https://jira.hyperledger.org/browse/FAB-8699)] Orderer custom connector
* [6607947a](https://github.com/hyperledger/fabric-sdk-go/commit/6607947a) [[FAB-8681](https://jira.hyperledger.org/browse/FAB-8681)] SaveChannel: Use multiple signing identities
* [9d194d54](https://github.com/hyperledger/fabric-sdk-go/commit/9d194d54) [[FAB-8667](https://jira.hyperledger.org/browse/FAB-8667)] Remove GetLogger method / Rename InitLogger
* [9606a4c4](https://github.com/hyperledger/fabric-sdk-go/commit/9606a4c4) [[FAB-8667](https://jira.hyperledger.org/browse/FAB-8667)] Levels in logging pkg
* [4889a927](https://github.com/hyperledger/fabric-sdk-go/commit/4889a927) [[FAB-8482](https://jira.hyperledger.org/browse/FAB-8482)] Context function implementation for SDK
* [14915595](https://github.com/hyperledger/fabric-sdk-go/commit/14915595) [[FAB-8863](https://jira.hyperledger.org/browse/FAB-8863)]: Rename WithProposalProcessor to WithTargets
* [d7c04c5e](https://github.com/hyperledger/fabric-sdk-go/commit/d7c04c5e) [[FAB-8659](https://jira.hyperledger.org/browse/FAB-8659)]Remove deprecated code all but channel
* [19065b85](https://github.com/hyperledger/fabric-sdk-go/commit/19065b85) [[FAB-8656](https://jira.hyperledger.org/browse/FAB-8656)] Initial logging modules
* [c13b6d31](https://github.com/hyperledger/fabric-sdk-go/commit/c13b6d31) [[FAB-8657](https://jira.hyperledger.org/browse/FAB-8657)]Remove deprecated code for orderer
* [aefe5a58](https://github.com/hyperledger/fabric-sdk-go/commit/aefe5a58) [[FAB-8324](https://jira.hyperledger.org/browse/FAB-8324)]: Resource Mgmt: Query Config
* [3c009b4b](https://github.com/hyperledger/fabric-sdk-go/commit/3c009b4b) [[FAB-8656](https://jira.hyperledger.org/browse/FAB-8656)] Initial logging modules
* [c5e1bc35](https://github.com/hyperledger/fabric-sdk-go/commit/c5e1bc35) [[FAB-8627](https://jira.hyperledger.org/browse/FAB-8627)] Add clientChannel timeout conf
* [4b576fda](https://github.com/hyperledger/fabric-sdk-go/commit/4b576fda) [[FAB-8578](https://jira.hyperledger.org/browse/FAB-8578)] Connection Caching
* [abe90446](https://github.com/hyperledger/fabric-sdk-go/commit/abe90446) [[FAB-8482](https://jira.hyperledger.org/browse/FAB-8482)] refactoring context and fabsdk package
* [1d43fc80](https://github.com/hyperledger/fabric-sdk-go/commit/1d43fc80) [[FAB-8321](https://jira.hyperledger.org/browse/FAB-8321)] Resource Mgmt: Query Instantiated CCs
* [48163647](https://github.com/hyperledger/fabric-sdk-go/commit/48163647) [[FAB-8639](https://jira.hyperledger.org/browse/FAB-8639)] Create peers via factory
* [77daffee](https://github.com/hyperledger/fabric-sdk-go/commit/77daffee) [[FAB-8609](https://jira.hyperledger.org/browse/FAB-8609)] User refactoring
* [3e334a01](https://github.com/hyperledger/fabric-sdk-go/commit/3e334a01) [[FAB-8625](https://jira.hyperledger.org/browse/FAB-8625)] update to v1.1.0-rc1
* [eaac2fa5](https://github.com/hyperledger/fabric-sdk-go/commit/eaac2fa5) [[FAB-8617](https://jira.hyperledger.org/browse/FAB-8617)] Deprecate Channel interface
* [b8eb2d99](https://github.com/hyperledger/fabric-sdk-go/commit/b8eb2d99) [[FAB-8620](https://jira.hyperledger.org/browse/FAB-8620)] Close for fabsdk
* [096e8c6a](https://github.com/hyperledger/fabric-sdk-go/commit/096e8c6a) [[FAB-8616](https://jira.hyperledger.org/browse/FAB-8616)]: Resource Mgmt: Move resource from context
* [713dd0e1](https://github.com/hyperledger/fabric-sdk-go/commit/713dd0e1) [[FAB-8615](https://jira.hyperledger.org/browse/FAB-8615)] Remove Channel dependency from ChannelConfig
* [b45b4e9f](https://github.com/hyperledger/fabric-sdk-go/commit/b45b4e9f) [[FAB-8607](https://jira.hyperledger.org/browse/FAB-8607)] Update README to point to ledger client
* [c3bbc2c0](https://github.com/hyperledger/fabric-sdk-go/commit/c3bbc2c0) [[FAB-8607](https://jira.hyperledger.org/browse/FAB-8607)]: Ledger Client
* [e7a77920](https://github.com/hyperledger/fabric-sdk-go/commit/e7a77920) [[FAB-8597](https://jira.hyperledger.org/browse/FAB-8597)] Simplify Enroll and Reenroll
* [82e18d87](https://github.com/hyperledger/fabric-sdk-go/commit/82e18d87) [[FAB-8583](https://jira.hyperledger.org/browse/FAB-8583)] Move IdentityManager interface
* [42c30019](https://github.com/hyperledger/fabric-sdk-go/commit/42c30019) [[FAB-8582](https://jira.hyperledger.org/browse/FAB-8582)] cleaning up NewPreEnrolledUser
* [e81d3ed0](https://github.com/hyperledger/fabric-sdk-go/commit/e81d3ed0) [[FAB-8566](https://jira.hyperledger.org/browse/FAB-8566)] Load CA Registrar from configuration
* [54c4ba01](https://github.com/hyperledger/fabric-sdk-go/commit/54c4ba01) [[FAB-8300](https://jira.hyperledger.org/browse/FAB-8300)] Remove unused Channel from ChannelService
* [6154281f](https://github.com/hyperledger/fabric-sdk-go/commit/6154281f) [[FAB-8553](https://jira.hyperledger.org/browse/FAB-8553)] Drop CredentialManager interface
* [60253b99](https://github.com/hyperledger/fabric-sdk-go/commit/60253b99) [[FAB-8558](https://jira.hyperledger.org/browse/FAB-8558)] Remove embedded err from TransactionResponse
* [79f443e7](https://github.com/hyperledger/fabric-sdk-go/commit/79f443e7) [[FAB-8300](https://jira.hyperledger.org/browse/FAB-8300)] Expose member mgmt through ChannelService
* [4dff6ab4](https://github.com/hyperledger/fabric-sdk-go/commit/4dff6ab4) [[FAB-8546](https://jira.hyperledger.org/browse/FAB-8546)] Fixed race condition in event client
* [b8174ca4](https://github.com/hyperledger/fabric-sdk-go/commit/b8174ca4) [[FAB-8542](https://jira.hyperledger.org/browse/FAB-8542)] Split invoke package from channel client
* [c2d4afc1](https://github.com/hyperledger/fabric-sdk-go/commit/c2d4afc1) [[FAB-8526](https://jira.hyperledger.org/browse/FAB-8526)] Embed CredentialManager into IdentityManager
* [081c0ff7](https://github.com/hyperledger/fabric-sdk-go/commit/081c0ff7) [[FAB-8537](https://jira.hyperledger.org/browse/FAB-8537)] Make event hub and deliver opts consistent
* [5e519834](https://github.com/hyperledger/fabric-sdk-go/commit/5e519834) [[FAB-8525](https://jira.hyperledger.org/browse/FAB-8525)] Refactor scc chaincode invocation
* [3258e6f2](https://github.com/hyperledger/fabric-sdk-go/commit/3258e6f2) [[FAB-8514](https://jira.hyperledger.org/browse/FAB-8514)] update baseimage version to 0.4.6
* [07c145dd](https://github.com/hyperledger/fabric-sdk-go/commit/07c145dd) [[FAB-7849](https://jira.hyperledger.org/browse/FAB-7849)] Hide deprecated methods behind build tag
* [5cddfd09](https://github.com/hyperledger/fabric-sdk-go/commit/5cddfd09) [[FAB-8398](https://jira.hyperledger.org/browse/FAB-8398)] Deliver Client Implementation
* [e9ea8dfe](https://github.com/hyperledger/fabric-sdk-go/commit/e9ea8dfe) [[FAB-8397](https://jira.hyperledger.org/browse/FAB-8397)] Event Hub Client Implementation
* [983e95a9](https://github.com/hyperledger/fabric-sdk-go/commit/983e95a9) [[FAB-8513](https://jira.hyperledger.org/browse/FAB-8513)] Identity Management refactoring
* [e8b01130](https://github.com/hyperledger/fabric-sdk-go/commit/e8b01130) [FAB-8509](https://jira.hyperledger.org/browse/FAB-8509) update go version to 1.9.2
* [69d59c18](https://github.com/hyperledger/fabric-sdk-go/commit/69d59c18) [[FAB-8512](https://jira.hyperledger.org/browse/FAB-8512)] Rename FabricCAClient to IdentityManager
* [e213fb6e](https://github.com/hyperledger/fabric-sdk-go/commit/e213fb6e) [[FAB-8491](https://jira.hyperledger.org/browse/FAB-8491)] remove def folder
* [834bdb30](https://github.com/hyperledger/fabric-sdk-go/commit/834bdb30) [[FAB-8490](https://jira.hyperledger.org/browse/FAB-8490)] Rename factory methods to Create
* [58e97f4b](https://github.com/hyperledger/fabric-sdk-go/commit/58e97f4b) [[FAB-8391](https://jira.hyperledger.org/browse/FAB-8391)] Change Transaction ID to Header
* [36f1d9d3](https://github.com/hyperledger/fabric-sdk-go/commit/36f1d9d3) [[FAB-8396](https://jira.hyperledger.org/browse/FAB-8396)] Abstract Event Client
* [bbc1dd0b](https://github.com/hyperledger/fabric-sdk-go/commit/bbc1dd0b) [[FAB-7874](https://jira.hyperledger.org/browse/FAB-7874)] refactoring pkg/logging
* [7fbc960a](https://github.com/hyperledger/fabric-sdk-go/commit/7fbc960a) [[FAB-8320](https://jira.hyperledger.org/browse/FAB-8320)] Resource Mgmt: Query Channels for Peer
* [948f2fdd](https://github.com/hyperledger/fabric-sdk-go/commit/948f2fdd) [[FAB-8395](https://jira.hyperledger.org/browse/FAB-8395)] Abstract Event Service
* [907b2633](https://github.com/hyperledger/fabric-sdk-go/commit/907b2633) [[FAB-8441](https://jira.hyperledger.org/browse/FAB-8441)] Save user to store on Enroll()
* [05b5ab9b](https://github.com/hyperledger/fabric-sdk-go/commit/05b5ab9b) [[FAB-8474](https://jira.hyperledger.org/browse/FAB-8474)] Set fail-fast false for CI
* [1a42cd93](https://github.com/hyperledger/fabric-sdk-go/commit/1a42cd93) [[FAB-8466](https://jira.hyperledger.org/browse/FAB-8466)] Use stdlib Context
* [86c95fab](https://github.com/hyperledger/fabric-sdk-go/commit/86c95fab) [[FAB-8464](https://jira.hyperledger.org/browse/FAB-8464)] Organize core pkg folder
* [76be1859](https://github.com/hyperledger/fabric-sdk-go/commit/76be1859) [[FAB-8463](https://jira.hyperledger.org/browse/FAB-8463)] Organize pkg/client folder
* [1170db5b](https://github.com/hyperledger/fabric-sdk-go/commit/1170db5b) [[FAB-8462](https://jira.hyperledger.org/browse/FAB-8462)] Rename fabric-client folder
* [350f5938](https://github.com/hyperledger/fabric-sdk-go/commit/350f5938) [[FAB-7874](https://jira.hyperledger.org/browse/FAB-7874)] refactoring api package
* [fc7d3249](https://github.com/hyperledger/fabric-sdk-go/commit/fc7d3249) [[FAB-8319](https://jira.hyperledger.org/browse/FAB-8319)] Resource Mgmt: Query Installed Chaincodes
* [68f94f92](https://github.com/hyperledger/fabric-sdk-go/commit/68f94f92) [[FAB-8437](https://jira.hyperledger.org/browse/FAB-8437)]High-level client New(context,options)
* [783a09ba](https://github.com/hyperledger/fabric-sdk-go/commit/783a09ba) [[FAB-8428](https://jira.hyperledger.org/browse/FAB-8428)] UserStore
* [eb9db605](https://github.com/hyperledger/fabric-sdk-go/commit/eb9db605) [[FAB-8422](https://jira.hyperledger.org/browse/FAB-8422)] Refactor broadcast to sign proposal
* [20ce0d3e](https://github.com/hyperledger/fabric-sdk-go/commit/20ce0d3e) [[FAB-8427](https://jira.hyperledger.org/browse/FAB-8427)] Seed rand only once
* [0a84288d](https://github.com/hyperledger/fabric-sdk-go/commit/0a84288d) [[FAB-8421](https://jira.hyperledger.org/browse/FAB-8421)] Speed up tests by disabling insecure retries
* [830f8ca9](https://github.com/hyperledger/fabric-sdk-go/commit/830f8ca9) [[FAB-8390](https://jira.hyperledger.org/browse/FAB-8390)] Refactor proposal creation to inject txn ID
* [75c84d20](https://github.com/hyperledger/fabric-sdk-go/commit/75c84d20) [[FAB-8389](https://jira.hyperledger.org/browse/FAB-8389)] Refactor txn ID creation and config sigs
* [42a1d820](https://github.com/hyperledger/fabric-sdk-go/commit/42a1d820) [[FAB-8376](https://jira.hyperledger.org/browse/FAB-8376)] Refactor deployment spec creation
* [54fe2705](https://github.com/hyperledger/fabric-sdk-go/commit/54fe2705) [[FAB-7800](https://jira.hyperledger.org/browse/FAB-7800)] making config pem path consistent
* [396f43e2](https://github.com/hyperledger/fabric-sdk-go/commit/396f43e2) [[FAB-8352](https://jira.hyperledger.org/browse/FAB-8352)] Go 1.10 compatibility
* [20be9f03](https://github.com/hyperledger/fabric-sdk-go/commit/20be9f03) [[FAB-5511](https://jira.hyperledger.org/browse/FAB-5511)]Combine management client and resource client
* [8eaa4fcb](https://github.com/hyperledger/fabric-sdk-go/commit/8eaa4fcb) [[FAB-8344](https://jira.hyperledger.org/browse/FAB-8344)] Remove Channel from integration tests
* [a1037f5c](https://github.com/hyperledger/fabric-sdk-go/commit/a1037f5c) [[FAB-8343](https://jira.hyperledger.org/browse/FAB-8343)] Remove metadata from block comparison
* [ff9b6bf6](https://github.com/hyperledger/fabric-sdk-go/commit/ff9b6bf6) [[FAB-8054](https://jira.hyperledger.org/browse/FAB-8054)] Split transactor from Channel
* [89112b82](https://github.com/hyperledger/fabric-sdk-go/commit/89112b82) [[FAB-8330](https://jira.hyperledger.org/browse/FAB-8330)] grpcs fallback to grpc when failed
* [614551a7](https://github.com/hyperledger/fabric-sdk-go/commit/614551a7) [[FAB-8299](https://jira.hyperledger.org/browse/FAB-8299)] Rename ChannelConfig to Config
* [ebb750ac](https://github.com/hyperledger/fabric-sdk-go/commit/ebb750ac) [[FAB-7512](https://jira.hyperledger.org/browse/FAB-7512)]Expose GRPC(keep-alive and failfast)
* [39a42500](https://github.com/hyperledger/fabric-sdk-go/commit/39a42500) [[FAB-8261](https://jira.hyperledger.org/browse/FAB-8261)] Introduce Multi Errors type
* [ca6fe202](https://github.com/hyperledger/fabric-sdk-go/commit/ca6fe202) [FAB-8262](https://jira.hyperledger.org/browse/FAB-8262) Clean up base_test_setup.go Initialize
* [c981b55c](https://github.com/hyperledger/fabric-sdk-go/commit/c981b55c) [[FAB-8247](https://jira.hyperledger.org/browse/FAB-8247)] Improve instantiated chaincodes test
* [07808ad2](https://github.com/hyperledger/fabric-sdk-go/commit/07808ad2) [[FAB-7978](https://jira.hyperledger.org/browse/FAB-7978)] ResourceMgmt: Join channel with filter only
* [1fd4d4d0](https://github.com/hyperledger/fabric-sdk-go/commit/1fd4d4d0) [[FAB-8203](https://jira.hyperledger.org/browse/FAB-8203)] Channel Client - EventHub not connected
* [9a8e8f5f](https://github.com/hyperledger/fabric-sdk-go/commit/9a8e8f5f) [[FAB-8189](https://jira.hyperledger.org/browse/FAB-8189)] no orderer config test refactoring
* [34108c5c](https://github.com/hyperledger/fabric-sdk-go/commit/34108c5c) [[FAB-8212](https://jira.hyperledger.org/browse/FAB-8212)] Fixed GetSigningIdentity()
* [031b888f](https://github.com/hyperledger/fabric-sdk-go/commit/031b888f) [[FAB-8211](https://jira.hyperledger.org/browse/FAB-8211)] Expose ledger from channel svc
* [ac89b893](https://github.com/hyperledger/fabric-sdk-go/commit/ac89b893) [[FAB-8195](https://jira.hyperledger.org/browse/FAB-8195)] Loading embedded certs is broken
* [bc269e64](https://github.com/hyperledger/fabric-sdk-go/commit/bc269e64) [[FAB-8191](https://jira.hyperledger.org/browse/FAB-8191)] Split ledger queries from channel
* [24c520b4](https://github.com/hyperledger/fabric-sdk-go/commit/24c520b4) [[FAB-6459](https://jira.hyperledger.org/browse/FAB-6459)] Validate endorser signature on response
* [f2b1c3be](https://github.com/hyperledger/fabric-sdk-go/commit/f2b1c3be) [[FAB-8189](https://jira.hyperledger.org/browse/FAB-8189)] using orderer config from channel block
* [95e3a32e](https://github.com/hyperledger/fabric-sdk-go/commit/95e3a32e) [[FAB-8151](https://jira.hyperledger.org/browse/FAB-8151)] Enable latest mock-gen
* [115de641](https://github.com/hyperledger/fabric-sdk-go/commit/115de641) [[FAB-8107](https://jira.hyperledger.org/browse/FAB-8107)] Move api/apicore to pkg/fabsdk/api
* [966fa037](https://github.com/hyperledger/fabric-sdk-go/commit/966fa037) [[FAB-8127](https://jira.hyperledger.org/browse/FAB-8127)] FabricProvider based on context or cfg
* [110c5e6a](https://github.com/hyperledger/fabric-sdk-go/commit/110c5e6a) [[FAB-8150](https://jira.hyperledger.org/browse/FAB-8150)] ChannelService with System Channel
* [d477b43b](https://github.com/hyperledger/fabric-sdk-go/commit/d477b43b) [[FAB-8143](https://jira.hyperledger.org/browse/FAB-8143)] Clear error between retry attempts
* [37e6436d](https://github.com/hyperledger/fabric-sdk-go/commit/37e6436d) [[FAB-8145](https://jira.hyperledger.org/browse/FAB-8145)] Cache configuration on creation
* [0d235e06](https://github.com/hyperledger/fabric-sdk-go/commit/0d235e06) [[FAB-8146](https://jira.hyperledger.org/browse/FAB-8146)]Add unit test for EndorsementValidation
* [a0c4e3e0](https://github.com/hyperledger/fabric-sdk-go/commit/a0c4e3e0) [[FAB-8134](https://jira.hyperledger.org/browse/FAB-8134)] Integrate channel config
* [dbcc63a7](https://github.com/hyperledger/fabric-sdk-go/commit/dbcc63a7) [[FAB-8124](https://jira.hyperledger.org/browse/FAB-8124)] Move genesis block + joinchan to resource
* [578506b8](https://github.com/hyperledger/fabric-sdk-go/commit/578506b8) [FAB-8101](https://jira.hyperledger.org/browse/FAB-8101) - .json file extension with filekeyvaluestore
* [e50cd255](https://github.com/hyperledger/fabric-sdk-go/commit/e50cd255) [[FAB-7513](https://jira.hyperledger.org/browse/FAB-7513)] MSP-like key and cert storage
* [5f54093a](https://github.com/hyperledger/fabric-sdk-go/commit/5f54093a) [[FAB-7508](https://jira.hyperledger.org/browse/FAB-7508)] Discovery greylist filter
* [4c641946](https://github.com/hyperledger/fabric-sdk-go/commit/4c641946) [[FAB-5209](https://jira.hyperledger.org/browse/FAB-5209)] Remove dependency on channel from resource
* [42d6b169](https://github.com/hyperledger/fabric-sdk-go/commit/42d6b169) [[FAB-8050](https://jira.hyperledger.org/browse/FAB-8050)] Remove internal packages from client
* [d4c8e3c7](https://github.com/hyperledger/fabric-sdk-go/commit/d4c8e3c7) [[FAB-8049](https://jira.hyperledger.org/browse/FAB-8049)] Split txn out of channel
* [b8f2be74](https://github.com/hyperledger/fabric-sdk-go/commit/b8f2be74) [[FAB-8039](https://jira.hyperledger.org/browse/FAB-8039)] Split Endorsement Handler
* [913ef174](https://github.com/hyperledger/fabric-sdk-go/commit/913ef174) [[FAB-7998](https://jira.hyperledger.org/browse/FAB-7998)] Move client APIs to apifabclient
* [cb5e5811](https://github.com/hyperledger/fabric-sdk-go/commit/cb5e5811) [[FAB-8004](https://jira.hyperledger.org/browse/FAB-8004)] Remove errors wrapper
* [a0a94304](https://github.com/hyperledger/fabric-sdk-go/commit/a0a94304) [[FAB-8005](https://jira.hyperledger.org/browse/FAB-8005)] Fix CI
* [3c5f08a9](https://github.com/hyperledger/fabric-sdk-go/commit/3c5f08a9) [[FAB-8024](https://jira.hyperledger.org/browse/FAB-8024)] Apply correct timeout option
* [cdb34d50](https://github.com/hyperledger/fabric-sdk-go/commit/cdb34d50) [[FAB-8023](https://jira.hyperledger.org/browse/FAB-8023)] Channel Config from orderer
* [ec2110a0](https://github.com/hyperledger/fabric-sdk-go/commit/ec2110a0) [[FAB-8005](https://jira.hyperledger.org/browse/FAB-8005)] Fix CI
* [e8a79bb2](https://github.com/hyperledger/fabric-sdk-go/commit/e8a79bb2) [[FAB-8008](https://jira.hyperledger.org/browse/FAB-8008)] Update to versioned protobuf
* [bc8e5ef8](https://github.com/hyperledger/fabric-sdk-go/commit/bc8e5ef8) [[FAB-8003](https://jira.hyperledger.org/browse/FAB-8003)] Use stdlib sync map
* [cf9810f1](https://github.com/hyperledger/fabric-sdk-go/commit/cf9810f1) [[FAB-7984](https://jira.hyperledger.org/browse/FAB-7984)] Add InvokeHandler to ChannelClient
* [d6546c82](https://github.com/hyperledger/fabric-sdk-go/commit/d6546c82) [[FAB-7508](https://jira.hyperledger.org/browse/FAB-7508)] Set grpc-go min version constraint to 1.8.0
* [4fc40fa6](https://github.com/hyperledger/fabric-sdk-go/commit/4fc40fa6) [[FAB-7980](https://jira.hyperledger.org/browse/FAB-7980)] Fix ChannelService under ClientContext
* [45297021](https://github.com/hyperledger/fabric-sdk-go/commit/45297021) [[FAB-7972](https://jira.hyperledger.org/browse/FAB-7972)] Whitelist discovery filter
* [f3eca564](https://github.com/hyperledger/fabric-sdk-go/commit/f3eca564) [[FAB-7508](https://jira.hyperledger.org/browse/FAB-7508)] Move status, retry to pkg/errors
* [35b6af70](https://github.com/hyperledger/fabric-sdk-go/commit/35b6af70) [[FAB-7968](https://jira.hyperledger.org/browse/FAB-7968)] Return error from function in ChannelClient
* [47e68c44](https://github.com/hyperledger/fabric-sdk-go/commit/47e68c44) [[FAB-7508](https://jira.hyperledger.org/browse/FAB-7508)] Add retries to channel client
* [7ccbf184](https://github.com/hyperledger/fabric-sdk-go/commit/7ccbf184) [[FAB-7960](https://jira.hyperledger.org/browse/FAB-7960)] Read Channel Configuration from Peer
* [e985ca49](https://github.com/hyperledger/fabric-sdk-go/commit/e985ca49) [[FAB-7968](https://jira.hyperledger.org/browse/FAB-7968)] Remove channel client async option
* [b0efb7e9](https://github.com/hyperledger/fabric-sdk-go/commit/b0efb7e9) [[FAB-7948](https://jira.hyperledger.org/browse/FAB-7948)] Removing *WithOpts functions in resgmtclient
* [0855cd76](https://github.com/hyperledger/fabric-sdk-go/commit/0855cd76) [[FAB-7831](https://jira.hyperledger.org/browse/FAB-7831)] Refactor SDK context
* [99efc28f](https://github.com/hyperledger/fabric-sdk-go/commit/99efc28f) [[FAB-7948](https://jira.hyperledger.org/browse/FAB-7948)] Removing *WithOpts functions in chmgmtclient
* [6af43635](https://github.com/hyperledger/fabric-sdk-go/commit/6af43635) [[FAB-7931](https://jira.hyperledger.org/browse/FAB-7931)] Update to fabric CA v1.1.0-alpha
* [335b2b52](https://github.com/hyperledger/fabric-sdk-go/commit/335b2b52) [[FAB-7935](https://jira.hyperledger.org/browse/FAB-7935)] Make context func names consistent
* [5f28e69e](https://github.com/hyperledger/fabric-sdk-go/commit/5f28e69e) [[FAB-7931](https://jira.hyperledger.org/browse/FAB-7931)] Update to fabric v1.1.0-alpha
* [edf8ad29](https://github.com/hyperledger/fabric-sdk-go/commit/edf8ad29) [[FAB-7883](https://jira.hyperledger.org/browse/FAB-7883)] step based Query and ExecuteTx
* [10205b28](https://github.com/hyperledger/fabric-sdk-go/commit/10205b28) [[FAB-7917](https://jira.hyperledger.org/browse/FAB-7917)] Move packages from def to pkg/fabsdk
* [ac3975ab](https://github.com/hyperledger/fabric-sdk-go/commit/ac3975ab) [[FAB-7912](https://jira.hyperledger.org/browse/FAB-7912)] Extract Resource methods into new pkg
* [f791ff1c](https://github.com/hyperledger/fabric-sdk-go/commit/f791ff1c) [[FAB-7911](https://jira.hyperledger.org/browse/FAB-7911)] Update test gopkg dependencies
* [d296d35b](https://github.com/hyperledger/fabric-sdk-go/commit/d296d35b) [[FAB-6776](https://jira.hyperledger.org/browse/FAB-6776)] Dep 0.4.1
* [ef806275](https://github.com/hyperledger/fabric-sdk-go/commit/ef806275) [[FAB-7880](https://jira.hyperledger.org/browse/FAB-7880)] add error code to empty cert
* [23e4a061](https://github.com/hyperledger/fabric-sdk-go/commit/23e4a061) [[FAB-7864](https://jira.hyperledger.org/browse/FAB-7864)] fabsdk config err should return wrapped err
* [7c5c1bdf](https://github.com/hyperledger/fabric-sdk-go/commit/7c5c1bdf) [[FAB-7869](https://jira.hyperledger.org/browse/FAB-7869)] event mock should not import testing
* [48a3c934](https://github.com/hyperledger/fabric-sdk-go/commit/48a3c934) [[FAB-7862](https://jira.hyperledger.org/browse/FAB-7862)] Add WithConfig helper method to fabsdk
* [40b253e7](https://github.com/hyperledger/fabric-sdk-go/commit/40b253e7) [[FAB-7826](https://jira.hyperledger.org/browse/FAB-7826)] Fix QueryInstantiatedChaincodes
* [0a3ac6fd](https://github.com/hyperledger/fabric-sdk-go/commit/0a3ac6fd) [[FAB-7850](https://jira.hyperledger.org/browse/FAB-7850)] Collection BTL policy proto
* [21d796df](https://github.com/hyperledger/fabric-sdk-go/commit/21d796df) [[FAB-7831](https://jira.hyperledger.org/browse/FAB-7831)] Move funcs that propagate SDK context
* [6d35bfc4](https://github.com/hyperledger/fabric-sdk-go/commit/6d35bfc4) [[FAB-7763](https://jira.hyperledger.org/browse/FAB-7763)] fixed go fmt spacing
* [94ac20c7](https://github.com/hyperledger/fabric-sdk-go/commit/94ac20c7) [[FAB-7830](https://jira.hyperledger.org/browse/FAB-7830)] Refactor client: delay error propagation
* [199cc9d7](https://github.com/hyperledger/fabric-sdk-go/commit/199cc9d7) [[FAB-7827](https://jira.hyperledger.org/browse/FAB-7827)] Use ConfigProvider interface
* [58439853](https://github.com/hyperledger/fabric-sdk-go/commit/58439853) [[FAB-7821](https://jira.hyperledger.org/browse/FAB-7821)] Update tests to use generic client opts
* [864ac4b2](https://github.com/hyperledger/fabric-sdk-go/commit/864ac4b2) [[FAB-7623](https://jira.hyperledger.org/browse/FAB-7623)] Make channel client options generic
* [9f302399](https://github.com/hyperledger/fabric-sdk-go/commit/9f302399) [[FAB-7807](https://jira.hyperledger.org/browse/FAB-7807)] Return response from ExecuteTx
* [151ff037](https://github.com/hyperledger/fabric-sdk-go/commit/151ff037) [[FAB-7776](https://jira.hyperledger.org/browse/FAB-7776)] Customizable default logger in SDK Logging
* [7f08d83f](https://github.com/hyperledger/fabric-sdk-go/commit/7f08d83f) [[FAB-7806](https://jira.hyperledger.org/browse/FAB-7806)] Move gofilter under src directory
* [535bcd35](https://github.com/hyperledger/fabric-sdk-go/commit/535bcd35) [[FAB-6523](https://jira.hyperledger.org/browse/FAB-6523)] Bump Fabric version
* [e084bafe](https://github.com/hyperledger/fabric-sdk-go/commit/e084bafe) [[FAB-7622](https://jira.hyperledger.org/browse/FAB-7622)] Update tests to use fabsdk.New
* [ed3dc93e](https://github.com/hyperledger/fabric-sdk-go/commit/ed3dc93e) [[FAB-7729](https://jira.hyperledger.org/browse/FAB-7729)] Cleanup SDK options
* [28e9a8bf](https://github.com/hyperledger/fabric-sdk-go/commit/28e9a8bf) [[FAB-7798](https://jira.hyperledger.org/browse/FAB-7798)] Update pkcs11 tests to use internal scripts
* [8d22afbd](https://github.com/hyperledger/fabric-sdk-go/commit/8d22afbd) [[FAB-7771](https://jira.hyperledger.org/browse/FAB-7771)] Refactor config impl constructor
* [89148100](https://github.com/hyperledger/fabric-sdk-go/commit/89148100) [[FAB-7763](https://jira.hyperledger.org/browse/FAB-7763)] Tool to import keys into HSM test fixture
* [4439cf27](https://github.com/hyperledger/fabric-sdk-go/commit/4439cf27) [[FAB-7507](https://jira.hyperledger.org/browse/FAB-7507)] Return status codes
* [a03806bf](https://github.com/hyperledger/fabric-sdk-go/commit/a03806bf) [[FAB-7765](https://jira.hyperledger.org/browse/FAB-7765)] Use CryptoSuite to load private key/signer
* [4c444f79](https://github.com/hyperledger/fabric-sdk-go/commit/4c444f79) [[FAB-7699](https://jira.hyperledger.org/browse/FAB-7699)] Embedded pems and functional params support
* [ceed3390](https://github.com/hyperledger/fabric-sdk-go/commit/ceed3390) [[FAB-7736](https://jira.hyperledger.org/browse/FAB-7736)] Move set default cryptosuite to SDK
* [6662c7f2](https://github.com/hyperledger/fabric-sdk-go/commit/6662c7f2) [[FAB-7728](https://jira.hyperledger.org/browse/FAB-7728)] Check for error in fabsdk option
* [ab16930f](https://github.com/hyperledger/fabric-sdk-go/commit/ab16930f) [[FAB-7661](https://jira.hyperledger.org/browse/FAB-7661)] Functional parameters for fabsdk.New
* [3aaff6e1](https://github.com/hyperledger/fabric-sdk-go/commit/3aaff6e1) [[FAB-7618](https://jira.hyperledger.org/browse/FAB-7618)] Split fabsdk from fabapi into new pkg
* [f5d61a3e](https://github.com/hyperledger/fabric-sdk-go/commit/f5d61a3e) [[FAB-7625](https://jira.hyperledger.org/browse/FAB-7625)] Make core fabric generic in fabapi
* [188110a6](https://github.com/hyperledger/fabric-sdk-go/commit/188110a6) [[FAB-7682](https://jira.hyperledger.org/browse/FAB-7682)] TLS Cert Hash to Event for Mutual TLS
* [d3dc3128](https://github.com/hyperledger/fabric-sdk-go/commit/d3dc3128) [[FAB-6982](https://jira.hyperledger.org/browse/FAB-6982)] - Support Private Data Collection Config
* [5add6df5](https://github.com/hyperledger/fabric-sdk-go/commit/5add6df5) [[FAB-6523](https://jira.hyperledger.org/browse/FAB-6523)] Bump Fabric version
* [2c40482e](https://github.com/hyperledger/fabric-sdk-go/commit/2c40482e) [[FAB-7606](https://jira.hyperledger.org/browse/FAB-7606)] Validate endorser payload and status
* [c0080e22](https://github.com/hyperledger/fabric-sdk-go/commit/c0080e22) [[FAB-7602](https://jira.hyperledger.org/browse/FAB-7602)] Add default txProposalResponseFilter
* [083c917b](https://github.com/hyperledger/fabric-sdk-go/commit/083c917b) [[FAB-7599](https://jira.hyperledger.org/browse/FAB-7599)] Move packages from internal to third_party
* [fdeaed1d](https://github.com/hyperledger/fabric-sdk-go/commit/fdeaed1d) [[FAB-7577](https://jira.hyperledger.org/browse/FAB-7577)] Separate cryptosuite pkgs
* [65c26f43](https://github.com/hyperledger/fabric-sdk-go/commit/65c26f43) [[FAB-7576](https://jira.hyperledger.org/browse/FAB-7576)] Conditional PKCS11 support
* [035e4f9b](https://github.com/hyperledger/fabric-sdk-go/commit/035e4f9b) [[FAB-7452](https://jira.hyperledger.org/browse/FAB-7452)] Allow embedding cryptoconfig in the Config
* [e902e9b2](https://github.com/hyperledger/fabric-sdk-go/commit/e902e9b2) [[FAB-7485](https://jira.hyperledger.org/browse/FAB-7485)] Re-organize integration tests
* [2f7918e1](https://github.com/hyperledger/fabric-sdk-go/commit/2f7918e1) [[FAB-7550](https://jira.hyperledger.org/browse/FAB-7550)] No Mutual TLS integration test case
* [3d663614](https://github.com/hyperledger/fabric-sdk-go/commit/3d663614) [[FAB-7546](https://jira.hyperledger.org/browse/FAB-7546)] Update changelog
* [7c09f2e](https://github.com/hyperledger/fabric-sdk-go/commit/7c09f2e) [[FAB-7536](https://jira.hyperledger.org/browse/FAB-7536)] Provide TLS Cert Hash in Channel Header
* [de1c0d8](https://github.com/hyperledger/fabric-sdk-go/commit/de1c0d8) [[FAB-7531](https://jira.hyperledger.org/browse/FAB-7531)] Fix timestamp in channel header
* [d0a811d](https://github.com/hyperledger/fabric-sdk-go/commit/d0a811d) [[FAB-6258](https://jira.hyperledger.org/browse/FAB-6258)] Bump fabric third_party revision
* [ad6e27e](https://github.com/hyperledger/fabric-sdk-go/commit/ad6e27e) [[FAB-7530](https://jira.hyperledger.org/browse/FAB-7530)] Update version of viper
* [9fd0ebd](https://github.com/hyperledger/fabric-sdk-go/commit/9fd0ebd) [[FAB-7516](https://jira.hyperledger.org/browse/FAB-7516)] refactor embbedded cert/key combo
* [8d39602](https://github.com/hyperledger/fabric-sdk-go/commit/8d39602) [[FAB-7528](https://jira.hyperledger.org/browse/FAB-7528)] Update thirdparty pinning scripts
* [56a7adb](https://github.com/hyperledger/fabric-sdk-go/commit/56a7adb) [[FAB-7516](https://jira.hyperledger.org/browse/FAB-7516)] Cert/Key embed & file path combo
* [78503fe](https://github.com/hyperledger/fabric-sdk-go/commit/78503fe) [[FAB-7498](https://jira.hyperledger.org/browse/FAB-7498)] Rename fixtures ca to be part of example.com
* [24f9ecc](https://github.com/hyperledger/fabric-sdk-go/commit/24f9ecc) [[FAB-7452](https://jira.hyperledger.org/browse/FAB-7452)] Allow embedding cryptoconfig in the Config
* [9a9856d](https://github.com/hyperledger/fabric-sdk-go/commit/9a9856d) [[FAB-7488](https://jira.hyperledger.org/browse/FAB-7488)] Fix integration tests outside docker
* [dafcfbc](https://github.com/hyperledger/fabric-sdk-go/commit/dafcfbc) [[FAB-6805](https://jira.hyperledger.org/browse/FAB-6805)] Mutual TLS added further unit tests
* [b99661a](https://github.com/hyperledger/fabric-sdk-go/commit/b99661a) [[FAB-6805](https://jira.hyperledger.org/browse/FAB-6805)] Mutual TLS
* [a00fd98](https://github.com/hyperledger/fabric-sdk-go/commit/a00fd98) [[FAB-7451](https://jira.hyperledger.org/browse/FAB-7451)] Tx Failures shouldn't create CC Events
* [3651028](https://github.com/hyperledger/fabric-sdk-go/commit/3651028) [[FAB-7392](https://jira.hyperledger.org/browse/FAB-7392)] Enable multiple version testing
* [a062492](https://github.com/hyperledger/fabric-sdk-go/commit/a062492) [[FAB-7387](https://jira.hyperledger.org/browse/FAB-7387)] Ability to load system cert pool
* [5c01120](https://github.com/hyperledger/fabric-sdk-go/commit/5c01120) [[FAB-7388](https://jira.hyperledger.org/browse/FAB-7388)] Fix mockgen version to v1.0.0
* [83c4a87](https://github.com/hyperledger/fabric-sdk-go/commit/83c4a87) [[FAB-7386](https://jira.hyperledger.org/browse/FAB-7386)] Enhance multi organisation test
* [17a18b1](https://github.com/hyperledger/fabric-sdk-go/commit/17a18b1) [[FAB-7307](https://jira.hyperledger.org/browse/FAB-7307)] embed cert in config, test update
* [0bd2dfa](https://github.com/hyperledger/fabric-sdk-go/commit/0bd2dfa) [[FAB-7346](https://jira.hyperledger.org/browse/FAB-7346)] Make IsChaincodeInstalled a private method
* [8dde6e6](https://github.com/hyperledger/fabric-sdk-go/commit/8dde6e6) [[FAB-7323](https://jira.hyperledger.org/browse/FAB-7323)] Updates to ResourceMgmtClient interface
* [f853354](https://github.com/hyperledger/fabric-sdk-go/commit/f853354) [[FAB-7307](https://jira.hyperledger.org/browse/FAB-7307)]Config in []byte & embed cert in config
* [9dad8ae](https://github.com/hyperledger/fabric-sdk-go/commit/9dad8ae) [[FAB-7292](https://jira.hyperledger.org/browse/FAB-7292)] Configure fabric-ca server correctly
* [d3c36d4](https://github.com/hyperledger/fabric-sdk-go/commit/d3c36d4) [[FAB-7281](https://jira.hyperledger.org/browse/FAB-7281)] Resource Mgmt - Instantiate/Upgrade CC
* [ce72cd1](https://github.com/hyperledger/fabric-sdk-go/commit/ce72cd1) [[FAB-7113](https://jira.hyperledger.org/browse/FAB-7113)] Init default logger when not set
* [e5f4954](https://github.com/hyperledger/fabric-sdk-go/commit/e5f4954) [[FAB-6258](https://jira.hyperledger.org/browse/FAB-6258)] Fix import path change in fabric
* [55aac02](https://github.com/hyperledger/fabric-sdk-go/commit/55aac02) [[FAB-6258](https://jira.hyperledger.org/browse/FAB-6258)] Bump fabric third_party revision
* [94de933](https://github.com/hyperledger/fabric-sdk-go/commit/94de933) [[FAB-7183](https://jira.hyperledger.org/browse/FAB-7183)] check_license misses some newly added files
* [0e5f0f6](https://github.com/hyperledger/fabric-sdk-go/commit/0e5f0f6) [[FAB-6983](https://jira.hyperledger.org/browse/FAB-6983)] fabric-ca to reuse sdk cryptosuite
* [7053d2c](https://github.com/hyperledger/fabric-sdk-go/commit/7053d2c) [[FAB-7097](https://jira.hyperledger.org/browse/FAB-7097)] Fix tests when GOPATH isn't set
* [dd63d01](https://github.com/hyperledger/fabric-sdk-go/commit/dd63d01) [[FAB-7101](https://jira.hyperledger.org/browse/FAB-7101)] Resource Management Client - Install CC
* [26b3d2e](https://github.com/hyperledger/fabric-sdk-go/commit/26b3d2e) [[FAB-6983](https://jira.hyperledger.org/browse/FAB-6983)] bccsp import refactoring
* [e9fa53a](https://github.com/hyperledger/fabric-sdk-go/commit/e9fa53a) [[FAB-7047](https://jira.hyperledger.org/browse/FAB-7047)] Resource Mgmt Client - Join Channel
* [59a0a8e](https://github.com/hyperledger/fabric-sdk-go/commit/59a0a8e) [[FAB-6983](https://jira.hyperledger.org/browse/FAB-6983)] Moved bccsp from third_party to internal
* [847bedf](https://github.com/hyperledger/fabric-sdk-go/commit/847bedf) [[FAB-7057](https://jira.hyperledger.org/browse/FAB-7057)] Fix third_party pinning script on MacOS
* [a218800](https://github.com/hyperledger/fabric-sdk-go/commit/a218800) [[FAB-7053](https://jira.hyperledger.org/browse/FAB-7053)] Make ChannelConfig public
* [a5e3c16](https://github.com/hyperledger/fabric-sdk-go/commit/a5e3c16) [[FAB-6983](https://jira.hyperledger.org/browse/FAB-6983)] replacing BCCSP with cryptosuite adaptor
* [4ae6eda](https://github.com/hyperledger/fabric-sdk-go/commit/4ae6eda) [[FAB-7003](https://jira.hyperledger.org/browse/FAB-7003)] Remove block from debug logs
* [231ec6c](https://github.com/hyperledger/fabric-sdk-go/commit/231ec6c) [[FAB-6981](https://jira.hyperledger.org/browse/FAB-6981)] Channel Management Client
* [23a0767](https://github.com/hyperledger/fabric-sdk-go/commit/23a0767) [[FAB-6983](https://jira.hyperledger.org/browse/FAB-6983)] cryptosuite adaptor for bccsp override
* [9d2609c](https://github.com/hyperledger/fabric-sdk-go/commit/9d2609c) [[FAB-6460](https://jira.hyperledger.org/browse/FAB-6460)] Increase async query timeout
* [b930c3b](https://github.com/hyperledger/fabric-sdk-go/commit/b930c3b) [[FAB-6356](https://jira.hyperledger.org/browse/FAB-6356)] Setting logprovider in advance for SDK tests
* [73a1e9b](https://github.com/hyperledger/fabric-sdk-go/commit/73a1e9b) [[FAB-6356](https://jira.hyperledger.org/browse/FAB-6356)] enable/disable callerinfo by module/loglevel
* [ba8d5c9](https://github.com/hyperledger/fabric-sdk-go/commit/ba8d5c9) [[FAB-6945](https://jira.hyperledger.org/browse/FAB-6945)] Config Improvements
* [7731bd8](https://github.com/hyperledger/fabric-sdk-go/commit/7731bd8) [[FAB-6928](https://jira.hyperledger.org/browse/FAB-6928)] Update CI to v1.0.4
* [17351d9](https://github.com/hyperledger/fabric-sdk-go/commit/17351d9) [[FAB-6914](https://jira.hyperledger.org/browse/FAB-6914)] fixing Gopkg.toml for knetic/govaluate
* [3710c33](https://github.com/hyperledger/fabric-sdk-go/commit/3710c33) [[FAB-6356](https://jira.hyperledger.org/browse/FAB-6356)] SDK-logging customizable callerinfo
* [4076bda](https://github.com/hyperledger/fabric-sdk-go/commit/4076bda) [[FAB-6914](https://jira.hyperledger.org/browse/FAB-6914)] Adding policy parser in SDK-Go
* [562ea23](https://github.com/hyperledger/fabric-sdk-go/commit/562ea23) [[FAB-6915](https://jira.hyperledger.org/browse/FAB-6915)] Fix chaincode startup issue in tests
* [2443ac7](https://github.com/hyperledger/fabric-sdk-go/commit/2443ac7) [[FAB-6886](https://jira.hyperledger.org/browse/FAB-6886)] SDK logger new utility function
* [2efa8bf](https://github.com/hyperledger/fabric-sdk-go/commit/2efa8bf) [[FAB-6708](https://jira.hyperledger.org/browse/FAB-6708)] Build flags to disable BCCSP plugins
* [8b685f6](https://github.com/hyperledger/fabric-sdk-go/commit/8b685f6) [[FAB-6860](https://jira.hyperledger.org/browse/FAB-6860)] Update third_party fabric to v1.1.0-preview
* [b36fe41](https://github.com/hyperledger/fabric-sdk-go/commit/b36fe41) [[FAB-6809](https://jira.hyperledger.org/browse/FAB-6809)] Dynamic Selection Service
* [cfaafc8](https://github.com/hyperledger/fabric-sdk-go/commit/cfaafc8) [[FAB-6814](https://jira.hyperledger.org/browse/FAB-6814)] Chaincode Policy Provider
* [25ef379](https://github.com/hyperledger/fabric-sdk-go/commit/25ef379) [[FAB-6767](https://jira.hyperledger.org/browse/FAB-6767)] mock utils for SDK go
* [e3ec402](https://github.com/hyperledger/fabric-sdk-go/commit/e3ec402) [[FAB-6812](https://jira.hyperledger.org/browse/FAB-6812)] Peer Group Resolver
* [2b9159f](https://github.com/hyperledger/fabric-sdk-go/commit/2b9159f) [[FAB-6759](https://jira.hyperledger.org/browse/FAB-6759)] Selection Service (Static)  
* [aafbea2](https://github.com/hyperledger/fabric-sdk-go/commit/aafbea2) [[FAB-6767](https://jira.hyperledger.org/browse/FAB-6767)] mock utils in SDK-GO
* [22e666e](https://github.com/hyperledger/fabric-sdk-go/commit/22e666e) [[FAB-6763](https://jira.hyperledger.org/browse/FAB-6763)] Give protos own namespace
* [23ec481](https://github.com/hyperledger/fabric-sdk-go/commit/23ec481) [[FAB-6695](https://jira.hyperledger.org/browse/FAB-6695)] Fixes for default config and protos
* [199fae9](https://github.com/hyperledger/fabric-sdk-go/commit/199fae9) [[FAB-6523](https://jira.hyperledger.org/browse/FAB-6523)] Bump Fabric version
* [8908090](https://github.com/hyperledger/fabric-sdk-go/commit/8908090) [[FAB-5878](https://jira.hyperledger.org/browse/FAB-5878)] - updated connection-profile
* [ae56873](https://github.com/hyperledger/fabric-sdk-go/commit/ae56873) [[FAB-6484](https://jira.hyperledger.org/browse/FAB-6484)] Cleanup third_party patching
* [cd1547f](https://github.com/hyperledger/fabric-sdk-go/commit/cd1547f) [[FAB-6484](https://jira.hyperledger.org/browse/FAB-6484)] Cleanup third_party patching
* [b2d8843](https://github.com/hyperledger/fabric-sdk-go/commit/b2d8843) [[FAB-6484](https://jira.hyperledger.org/browse/FAB-6484)] Cleanup third_party patching
* [a846c50](https://github.com/hyperledger/fabric-sdk-go/commit/a846c50) [[FAB-5878](https://jira.hyperledger.org/browse/FAB-5878)] Fixed a message typo in the config
* [db3029c](https://github.com/hyperledger/fabric-sdk-go/commit/db3029c) [[FAB-6477](https://jira.hyperledger.org/browse/FAB-6477)] Reduce imported Fabric code
* [e216f82](https://github.com/hyperledger/fabric-sdk-go/commit/e216f82) [[FAB-6290](https://jira.hyperledger.org/browse/FAB-6290)] Reduce imported Fabric CA code
* [7ab244f](https://github.com/hyperledger/fabric-sdk-go/commit/7ab244f) [[FAB-6235](https://jira.hyperledger.org/browse/FAB-6235)] dep v0.3.1
* [f048f16](https://github.com/hyperledger/fabric-sdk-go/commit/f048f16) [[FAB-6423](https://jira.hyperledger.org/browse/FAB-6423)] Go SDK config reused issue
* [f4ddd6f](https://github.com/hyperledger/fabric-sdk-go/commit/f4ddd6f) [[FAB-6258](https://jira.hyperledger.org/browse/FAB-6258)] Bump fabric third_party revision
* [c82f744](https://github.com/hyperledger/fabric-sdk-go/commit/c82f744) [[FAB-6428](https://jira.hyperledger.org/browse/FAB-6428)] Update CI & tools to Fabric 1.0.3
* [7f4bc34](https://github.com/hyperledger/fabric-sdk-go/commit/7f4bc34) [[FAB-6275](https://jira.hyperledger.org/browse/FAB-6275)] Removed tls:enabled from SDK GO config file
* [46ea533](https://github.com/hyperledger/fabric-sdk-go/commit/46ea533) [[FAB-6460](https://jira.hyperledger.org/browse/FAB-6460)] Increase query timeout
* [9c02025](https://github.com/hyperledger/fabric-sdk-go/commit/9c02025) [[FAB-6429](https://jira.hyperledger.org/browse/FAB-6429)] : Convert CC args type
* [4c04c1c](https://github.com/hyperledger/fabric-sdk-go/commit/4c04c1c) [[FAB-6406](https://jira.hyperledger.org/browse/FAB-6406)] Convert to errors package
* [3eae44a](https://github.com/hyperledger/fabric-sdk-go/commit/3eae44a) [[FAB-6424](https://jira.hyperledger.org/browse/FAB-6424)] Add ci.properties file
* [a1cb9c9](https://github.com/hyperledger/fabric-sdk-go/commit/a1cb9c9) [[FAB-6385](https://jira.hyperledger.org/browse/FAB-6385)] Go SDK timeouts config restructure
* [8b4dcfb](https://github.com/hyperledger/fabric-sdk-go/commit/8b4dcfb) [[FAB-6383](https://jira.hyperledger.org/browse/FAB-6383)] Add notice to third_party source code
* [7c71a98](https://github.com/hyperledger/fabric-sdk-go/commit/7c71a98) [[FAB-6391](https://jira.hyperledger.org/browse/FAB-6391)] Remove deprecated fabrictxn functions
* [20ff232](https://github.com/hyperledger/fabric-sdk-go/commit/20ff232) [[FAB-6385](https://jira.hyperledger.org/browse/FAB-6385)] Go SDK Configurable timeouts
* [2578687](https://github.com/hyperledger/fabric-sdk-go/commit/2578687) [[FAB-6382](https://jira.hyperledger.org/browse/FAB-6382)] - Add TxValidationCode to ExecuteTxResponse
* [e89a861](https://github.com/hyperledger/fabric-sdk-go/commit/e89a861) [[FAB-6343](https://jira.hyperledger.org/browse/FAB-6343)] Change Print to Log
* [fc03fd7](https://github.com/hyperledger/fabric-sdk-go/commit/fc03fd7) [[FAB-6358](https://jira.hyperledger.org/browse/FAB-6358)] Expose packages in e2e tests
* [8450c98](https://github.com/hyperledger/fabric-sdk-go/commit/8450c98) [[FAB-6322](https://jira.hyperledger.org/browse/FAB-6322)] Regenerate dep constraint file
* [c299d70](https://github.com/hyperledger/fabric-sdk-go/commit/c299d70) [[FAB-6343](https://jira.hyperledger.org/browse/FAB-6343)] Change Print to Log
* [2fb9484](https://github.com/hyperledger/fabric-sdk-go/commit/2fb9484) [[FAB-6342](https://jira.hyperledger.org/browse/FAB-6342)] DefaultLogger can print logger as caller
* [b5ac164](https://github.com/hyperledger/fabric-sdk-go/commit/b5ac164) [[FAB-6272](https://jira.hyperledger.org/browse/FAB-6272)] Improve log vendoring
* [8305ff0](https://github.com/hyperledger/fabric-sdk-go/commit/8305ff0) [[FAB-6340](https://jira.hyperledger.org/browse/FAB-6340)] Pass go test flags into docker tests
* [64a5e3f](https://github.com/hyperledger/fabric-sdk-go/commit/64a5e3f) [[FAB-6272](https://jira.hyperledger.org/browse/FAB-6272)] fixing golint warnings
* [17ad794](https://github.com/hyperledger/fabric-sdk-go/commit/17ad794) [[FAB-6272](https://jira.hyperledger.org/browse/FAB-6272)] SDK Go logging interface and provider
* [3dc34e5](https://github.com/hyperledger/fabric-sdk-go/commit/3dc34e5) [[FAB-6313](https://jira.hyperledger.org/browse/FAB-6313)] Fix linter warning
* [f1c390e](https://github.com/hyperledger/fabric-sdk-go/commit/f1c390e) [[FAB-6314](https://jira.hyperledger.org/browse/FAB-6314)] Channel Client
* [dd185dd](https://github.com/hyperledger/fabric-sdk-go/commit/dd185dd) [[FAB-6313](https://jira.hyperledger.org/browse/FAB-6313)] Fix linter warning
* [585d166](https://github.com/hyperledger/fabric-sdk-go/commit/585d166) [[FAB-6305](https://jira.hyperledger.org/browse/FAB-6305)] Pass Go Tags to docker integration test
* [293fb2b](https://github.com/hyperledger/fabric-sdk-go/commit/293fb2b) [[FAB-6300](https://jira.hyperledger.org/browse/FAB-6300)] - Move orderer and ledger to third_party
* [f08071e](https://github.com/hyperledger/fabric-sdk-go/commit/f08071e) [[FAB-6300](https://jira.hyperledger.org/browse/FAB-6300)] - Move orderer and ledger to third_party
* [08b4db4](https://github.com/hyperledger/fabric-sdk-go/commit/08b4db4) [[FAB-6284](https://jira.hyperledger.org/browse/FAB-6284)] Add ChangeLog.md file to track SDK GO
* [bf29760](https://github.com/hyperledger/fabric-sdk-go/commit/bf29760) [[FAB-6275](https://jira.hyperledger.org/browse/FAB-6275)] Add Default GO SDK config
* [541f496](https://github.com/hyperledger/fabric-sdk-go/commit/541f496) [[FAB-6285](https://jira.hyperledger.org/browse/FAB-6285)] Fabric-CA vendoring: remove tcerts
* [e58cbee](https://github.com/hyperledger/fabric-sdk-go/commit/e58cbee) [[FAB-6270](https://jira.hyperledger.org/browse/FAB-6270)] cleaning up logger info,error,warning
* [20fb840](https://github.com/hyperledger/fabric-sdk-go/commit/20fb840) [[FAB-3783](https://jira.hyperledger.org/browse/FAB-3783)] go-sdk chaincode upgrade support
* [3b6d77b](https://github.com/hyperledger/fabric-sdk-go/commit/3b6d77b) [[FAB-6252](https://jira.hyperledger.org/browse/FAB-6252)] Cleanup Makefile
* [72a94b4](https://github.com/hyperledger/fabric-sdk-go/commit/72a94b4) [[FAB-6252](https://jira.hyperledger.org/browse/FAB-6252)] Cleanup Makefile
* [a8c601e](https://github.com/hyperledger/fabric-sdk-go/commit/a8c601e) [[FAB-5878](https://jira.hyperledger.org/browse/FAB-5878)] - Implement "Connection Profile" for SDK-GO
* [e60551c](https://github.com/hyperledger/fabric-sdk-go/commit/e60551c) [[FAB-6184](https://jira.hyperledger.org/browse/FAB-6184)] Improve Fabric vendoring (populate)
* [2624e50](https://github.com/hyperledger/fabric-sdk-go/commit/2624e50) [[FAB-6222](https://jira.hyperledger.org/browse/FAB-6222)] Import Fabric protos
* [860a3b5](https://github.com/hyperledger/fabric-sdk-go/commit/860a3b5) [[FAB-6221](https://jira.hyperledger.org/browse/FAB-6221)] Import BCCSP as third_party
* [eff58ec](https://github.com/hyperledger/fabric-sdk-go/commit/eff58ec) [[FAB-6205](https://jira.hyperledger.org/browse/FAB-6205)] Improve log vendoring (populate flogging)
* [ff494fa](https://github.com/hyperledger/fabric-sdk-go/commit/ff494fa) [[FAB-6214](https://jira.hyperledger.org/browse/FAB-6214)] Make target to generate channel artifacts
* [4bfb47e](https://github.com/hyperledger/fabric-sdk-go/commit/4bfb47e) [[FAB-6184](https://jira.hyperledger.org/browse/FAB-6184)] Improve Fabric vendoring (scripts)
* [591cea8](https://github.com/hyperledger/fabric-sdk-go/commit/591cea8) [[FAB-6177](https://jira.hyperledger.org/browse/FAB-6177)] Improve Fabric-CA vendoring (populate 1.0.1)
* [e82eb25](https://github.com/hyperledger/fabric-sdk-go/commit/e82eb25) [[FAB-6177](https://jira.hyperledger.org/browse/FAB-6177)] Improve Fabric-CA vendoring (scripts)
* [dab5a98](https://github.com/hyperledger/fabric-sdk-go/commit/dab5a98) [[FAB-6146](https://jira.hyperledger.org/browse/FAB-6146)] Move fixture chaincode under testdata
* [bbc0200](https://github.com/hyperledger/fabric-sdk-go/commit/bbc0200) [[FAB-6079](https://jira.hyperledger.org/browse/FAB-6079)] Update README
* [7cdef1d](https://github.com/hyperledger/fabric-sdk-go/commit/7cdef1d) [[FAB-6111](https://jira.hyperledger.org/browse/FAB-6111)]Use syncmap in eventhub
* [308a18d](https://github.com/hyperledger/fabric-sdk-go/commit/308a18d) [[FAB-6104](https://jira.hyperledger.org/browse/FAB-6104)]Remove IsSecurityEnabled
* [f6e947e](https://github.com/hyperledger/fabric-sdk-go/commit/f6e947e) [[FAB-6079](https://jira.hyperledger.org/browse/FAB-6079)] Update fabric dependency
* [0f6b2f6](https://github.com/hyperledger/fabric-sdk-go/commit/0f6b2f6) [[FAB-6065](https://jira.hyperledger.org/browse/FAB-6065)]Config interface to expose BCCSP
* [74e5fa8](https://github.com/hyperledger/fabric-sdk-go/commit/74e5fa8) [[FAB-6041](https://jira.hyperledger.org/browse/FAB-6041)] Fabric-SDK-Go doesn't log properly
* [35d5b39](https://github.com/hyperledger/fabric-sdk-go/commit/35d5b39) [[FAB-6013](https://jira.hyperledger.org/browse/FAB-6013)]Race condition in eventhub
* [4213e7e](https://github.com/hyperledger/fabric-sdk-go/commit/4213e7e) [[FAB-5947](https://jira.hyperledger.org/browse/FAB-5947)] Fix compatibility information in README.md
* [9fe0963](https://github.com/hyperledger/fabric-sdk-go/commit/9fe0963) [[FAB-5948](https://jira.hyperledger.org/browse/FAB-5948)] Fix check_license script error
* [7570207](https://github.com/hyperledger/fabric-sdk-go/commit/7570207) [[FAB-5919](https://jira.hyperledger.org/browse/FAB-5919)] Refactor CI to run unit tests without docker
* [6648c81](https://github.com/hyperledger/fabric-sdk-go/commit/6648c81) [[FAB-4893](https://jira.hyperledger.org/browse/FAB-4893)] Update vendoring to dep (lint fix)
* [77082f4](https://github.com/hyperledger/fabric-sdk-go/commit/77082f4) [[FAB-5918](https://jira.hyperledger.org/browse/FAB-5918)]Fix build status on CI
* [1937c69](https://github.com/hyperledger/fabric-sdk-go/commit/1937c69) [[FAB-4893](https://jira.hyperledger.org/browse/FAB-4893)] Update vendoring to dep
* [4bea613](https://github.com/hyperledger/fabric-sdk-go/commit/4bea613) [[FAB-5908](https://jira.hyperledger.org/browse/FAB-5908)] Remove transaction proposal from error
* [08be791](https://github.com/hyperledger/fabric-sdk-go/commit/08be791) [[FAB-5897](https://jira.hyperledger.org/browse/FAB-5897)]Upgrade client to use fabric.v1.0.1
* [110bf21](https://github.com/hyperledger/fabric-sdk-go/commit/110bf21) [[FAB-5792](https://jira.hyperledger.org/browse/FAB-5792)] Use race detector for tests
* [2bbb512](https://github.com/hyperledger/fabric-sdk-go/commit/2bbb512) [[FAB-5143](https://jira.hyperledger.org/browse/FAB-5143)] Fix go-logging data races
* [7fb8ad9](https://github.com/hyperledger/fabric-sdk-go/commit/7fb8ad9) [[FAB-5143](https://jira.hyperledger.org/browse/FAB-5143)] [FAB-5557] Fix eventhub race condition
* [5ac8ff6](https://github.com/hyperledger/fabric-sdk-go/commit/5ac8ff6) [[FAB-5626](https://jira.hyperledger.org/browse/FAB-5626)]PKCS11 support using softhsm2
* [f0c65f3](https://github.com/hyperledger/fabric-sdk-go/commit/f0c65f3) [[FAB-5751](https://jira.hyperledger.org/browse/FAB-5751)] Add .gitreview
* [7392c6e](https://github.com/hyperledger/fabric-sdk-go/commit/7392c6e) [[FAB-5750](https://jira.hyperledger.org/browse/FAB-5750)] Fix invalid assertion in ca integration test
* [6ed4b84](https://github.com/hyperledger/fabric-sdk-go/commit/6ed4b84) [[FAB-5681](https://jira.hyperledger.org/browse/FAB-5681)] Expose ability to SendTransactionProposal
* [77507d9](https://github.com/hyperledger/fabric-sdk-go/commit/77507d9) [[FAB-5523](https://jira.hyperledger.org/browse/FAB-5523)] Configuration timeouts are ignored
* [f405d0f](https://github.com/hyperledger/fabric-sdk-go/commit/f405d0f) [[FAB-5515](https://jira.hyperledger.org/browse/FAB-5515)] Regenerate crypto artifacts
* [c721dcf](https://github.com/hyperledger/fabric-sdk-go/commit/c721dcf) [[FAB-5442](https://jira.hyperledger.org/browse/FAB-5442)] Increase create channel wait time
* [8964c14](https://github.com/hyperledger/fabric-sdk-go/commit/8964c14) [[FAB-5405](https://jira.hyperledger.org/browse/FAB-5405)] OrgContext and Session Structure
* [1e9e2f0](https://github.com/hyperledger/fabric-sdk-go/commit/1e9e2f0) [[FAB-5401](https://jira.hyperledger.org/browse/FAB-5401)]Disconnect Event Hub Asynchronously
* [a95d5fc](https://github.com/hyperledger/fabric-sdk-go/commit/a95d5fc) [[FAB-5388](https://jira.hyperledger.org/browse/FAB-5388)] SDK Go - Improve Broadcast mock method
* [1ee9a93](https://github.com/hyperledger/fabric-sdk-go/commit/1ee9a93) [[FAB-5379](https://jira.hyperledger.org/browse/FAB-5379)] Pass signature policy to Instantiate CC
* [1765d77](https://github.com/hyperledger/fabric-sdk-go/commit/1765d77) [[FAB-5343](https://jira.hyperledger.org/browse/FAB-5343)]NewChannel set nil in the channels array
* [3da99de](https://github.com/hyperledger/fabric-sdk-go/commit/3da99de) [[FAB-5016](https://jira.hyperledger.org/browse/FAB-5016)] Configurable connection timeouts
* [a24a856](https://github.com/hyperledger/fabric-sdk-go/commit/a24a856) [[FAB-5294](https://jira.hyperledger.org/browse/FAB-5294)] Retrieve peer config by name
* [4e0def9](https://github.com/hyperledger/fabric-sdk-go/commit/4e0def9) [[FAB-4633](https://jira.hyperledger.org/browse/FAB-4633)] Support Fabric 1.0.0
* [3f9b996](https://github.com/hyperledger/fabric-sdk-go/commit/3f9b996) [[FAB-5238](https://jira.hyperledger.org/browse/FAB-5238)] Poll results to fix intermittent failure
* [45480e5](https://github.com/hyperledger/fabric-sdk-go/commit/45480e5) [[FAB-5232](https://jira.hyperledger.org/browse/FAB-5232)] Add SDK TLS Cert Pool
* [c2c0c70](https://github.com/hyperledger/fabric-sdk-go/commit/c2c0c70) [[FAB-5215](https://jira.hyperledger.org/browse/FAB-5215)] SDK entry point (part 1)
* [cc9e96a](https://github.com/hyperledger/fabric-sdk-go/commit/cc9e96a) [[FAB-5235](https://jira.hyperledger.org/browse/FAB-5235)] Normalize Getter Names
* [8c34459](https://github.com/hyperledger/fabric-sdk-go/commit/8c34459) [[FAB-5233](https://jira.hyperledger.org/browse/FAB-5233)] Return TransactionID from InvokeChaincode
* [fe06786](https://github.com/hyperledger/fabric-sdk-go/commit/fe06786) [[FAB-5520](https://jira.hyperledger.org/browse/FAB-5520)] Make Broadcast select a random orderer
* [223e728](https://github.com/hyperledger/fabric-sdk-go/commit/223e728) [[FAB-5214](https://jira.hyperledger.org/browse/FAB-5214)] Move Identity from Client to User
* [1d36d9e](https://github.com/hyperledger/fabric-sdk-go/commit/1d36d9e) [[FAB-5211](https://jira.hyperledger.org/browse/FAB-5211)] CreateChannel to use TransactionID
* [6265e33](https://github.com/hyperledger/fabric-sdk-go/commit/6265e33) [[FAB-5201](https://jira.hyperledger.org/browse/FAB-5201)] Refactor SendTransactionProposal
* [ab834b8](https://github.com/hyperledger/fabric-sdk-go/commit/ab834b8) [[FAB-5192](https://jira.hyperledger.org/browse/FAB-5192)] Return struct from implementations
* [f71ee06](https://github.com/hyperledger/fabric-sdk-go/commit/f71ee06) [[FAB-5193](https://jira.hyperledger.org/browse/FAB-5193)] Mock servers should accept port & cleanup
* [358abfb](https://github.com/hyperledger/fabric-sdk-go/commit/358abfb) [[FAB-5173](https://jira.hyperledger.org/browse/FAB-5173)] Orderer refactor and test coverage
* [1f579ed](https://github.com/hyperledger/fabric-sdk-go/commit/1f579ed) [[FAB-5174](https://jira.hyperledger.org/browse/FAB-5174)] Packager refactoring and test coverage
* [f584747](https://github.com/hyperledger/fabric-sdk-go/commit/f584747) [[FAB-5183](https://jira.hyperledger.org/browse/FAB-5183)] Fabric-ca client refactoring
* [3a54d75](https://github.com/hyperledger/fabric-sdk-go/commit/3a54d75) [[FAB-5176](https://jira.hyperledger.org/browse/FAB-5176)] Split and refactor channel package
* [020bc5c](https://github.com/hyperledger/fabric-sdk-go/commit/020bc5c) [[FAB-5172](https://jira.hyperledger.org/browse/FAB-5172)] Keyvaluestore refactor and test coverage..
* [aeef71d](https://github.com/hyperledger/fabric-sdk-go/commit/aeef71d) [[FAB-5119](https://jira.hyperledger.org/browse/FAB-5119)] Split up API interfaces into multiple pkg
* [73df8f6](https://github.com/hyperledger/fabric-sdk-go/commit/73df8f6) [[FAB-5170](https://jira.hyperledger.org/browse/FAB-5170)] Refactor User and improve test coverage
* [5cc3579](https://github.com/hyperledger/fabric-sdk-go/commit/5cc3579) [[FAB-5115](https://jira.hyperledger.org/browse/FAB-5115)] Added more test case coverage to channel
* [a86f35f](https://github.com/hyperledger/fabric-sdk-go/commit/a86f35f) [[FAB-5142](https://jira.hyperledger.org/browse/FAB-5142)] Race condition detection
* [50281db](https://github.com/hyperledger/fabric-sdk-go/commit/50281db) [[FAB-5137](https://jira.hyperledger.org/browse/FAB-5137)] Split txn APIs into their own pkg
* [cb15534](https://github.com/hyperledger/fabric-sdk-go/commit/cb15534) [[FAB-5136](https://jira.hyperledger.org/browse/FAB-5136)] Viper shouldn't be exposed in interface
* [f7b2f4f](https://github.com/hyperledger/fabric-sdk-go/commit/f7b2f4f) [[FAB-5133](https://jira.hyperledger.org/browse/FAB-5133)] Split txn proposal APIs into their own pkg
* [a4ae7c1](https://github.com/hyperledger/fabric-sdk-go/commit/a4ae7c1) [[FAB-5115](https://jira.hyperledger.org/browse/FAB-5115)] Add test case to config.go
* [e72516d](https://github.com/hyperledger/fabric-sdk-go/commit/e72516d) [[FAB-5129](https://jira.hyperledger.org/browse/FAB-5129)] Update fabapi NewUser method
* [191cb6a](https://github.com/hyperledger/fabric-sdk-go/commit/191cb6a) [[FAB-5116](https://jira.hyperledger.org/browse/FAB-5116)] Improve Events test coverage
* [13a35d0](https://github.com/hyperledger/fabric-sdk-go/commit/13a35d0) [[FAB-5118](https://jira.hyperledger.org/browse/FAB-5118)] Split fabric-txn into pkg and default impl
* [95ef635](https://github.com/hyperledger/fabric-sdk-go/commit/95ef635) [[FAB-5115](https://jira.hyperledger.org/browse/FAB-5115)] Test coverage for config unit test
* [f0f2ef5](https://github.com/hyperledger/fabric-sdk-go/commit/f0f2ef5) [[FAB-4637](https://jira.hyperledger.org/browse/FAB-4637)] Add multi-org integration test
* [db358ac](https://github.com/hyperledger/fabric-sdk-go/commit/db358ac) [[FAB-5115](https://jira.hyperledger.org/browse/FAB-5115)] Refactoring config and channel get methods.
* [180a594](https://github.com/hyperledger/fabric-sdk-go/commit/180a594) [[FAB-5110](https://jira.hyperledger.org/browse/FAB-5110)] Added eventhub disconnect in InvokeChaincode
* [1876110](https://github.com/hyperledger/fabric-sdk-go/commit/1876110) [[FAB-5109](https://jira.hyperledger.org/browse/FAB-5109)] Extracted primary peer joined logic
* [536c637](https://github.com/hyperledger/fabric-sdk-go/commit/536c637) [[FAB-4637](https://jira.hyperledger.org/browse/FAB-4637)] Organization based configuration structure
* [4bb3690](https://github.com/hyperledger/fabric-sdk-go/commit/4bb3690) [[FAB-4890](https://jira.hyperledger.org/browse/FAB-4890)] Fabric Txn API and removed utils
* [85fa310](https://github.com/hyperledger/fabric-sdk-go/commit/85fa310) [[FAB-5007](https://jira.hyperledger.org/browse/FAB-5007)] Using a Normal user by default in e2e tests.
* [9e10dd3](https://github.com/hyperledger/fabric-sdk-go/commit/9e10dd3) [[FAB-4694](https://jira.hyperledger.org/browse/FAB-4694)] Improve test coverage
* [b164375](https://github.com/hyperledger/fabric-sdk-go/commit/b164375) [[FAB-4632](https://jira.hyperledger.org/browse/FAB-4632)] Upgrade to support 1.0.0 RC
* [965d4fb](https://github.com/hyperledger/fabric-sdk-go/commit/965d4fb) [[FAB-4934](https://jira.hyperledger.org/browse/FAB-4934)] SDK Go - Integration test fails locally
* [19a90a0](https://github.com/hyperledger/fabric-sdk-go/commit/19a90a0) [[FAB-4934](https://jira.hyperledger.org/browse/FAB-4934)] Event integration test failures
* [e62ea36](https://github.com/hyperledger/fabric-sdk-go/commit/e62ea36) [[FAB-4942](https://jira.hyperledger.org/browse/FAB-4942)] Enable Mac tests on Go 1.7
* [941469b](https://github.com/hyperledger/fabric-sdk-go/commit/941469b) [[FAB-4932](https://jira.hyperledger.org/browse/FAB-4932)] Fix README
* [1a2d803](https://github.com/hyperledger/fabric-sdk-go/commit/1a2d803) [[FAB-4638](https://jira.hyperledger.org/browse/FAB-4638)] SDK Go - Join channel response parsing
* [ea51b78](https://github.com/hyperledger/fabric-sdk-go/commit/ea51b78) [[FAB-4939](https://jira.hyperledger.org/browse/FAB-4939)] SDK Go - Removed unused variable
* [dafbe28](https://github.com/hyperledger/fabric-sdk-go/commit/dafbe28) [[FAB-4629](https://jira.hyperledger.org/browse/FAB-4629)] SDK Go - Change folder structure
* [4cb70e8](https://github.com/hyperledger/fabric-sdk-go/commit/4cb70e8) [[FAB-4754](https://jira.hyperledger.org/browse/FAB-4754)] Set correct arch for test containers
* [5923ca4](https://github.com/hyperledger/fabric-sdk-go/commit/5923ca4) [[FAB-4634](https://jira.hyperledger.org/browse/FAB-4634)] Update license headers
* [70e7a26](https://github.com/hyperledger/fabric-sdk-go/commit/70e7a26) [[FAB-4634](https://jira.hyperledger.org/browse/FAB-4634)] Update license headers
* [ab0ba8f](https://github.com/hyperledger/fabric-sdk-go/commit/ab0ba8f) [[FAB-4634](https://jira.hyperledger.org/browse/FAB-4634)] Update license headers
* [4b031d6](https://github.com/hyperledger/fabric-sdk-go/commit/4b031d6) [[FAB-4634](https://jira.hyperledger.org/browse/FAB-4634)] Update license headers
* [1f2c757](https://github.com/hyperledger/fabric-sdk-go/commit/1f2c757) [[FAB-4630](https://jira.hyperledger.org/browse/FAB-4630)] Update README
* [c0c3cb0](https://github.com/hyperledger/fabric-sdk-go/commit/c0c3cb0) [[FAB-4630](https://jira.hyperledger.org/browse/FAB-4630)] SDK Go -Upgrade to support 1.0.0 beta

## v1.0.0-alpha2
Thu 15 Jun 2017 12:16:05 EDT

* [04b9327](https://github.com/hyperledger/fabric-sdk-go/commit/04b9327) [[FAB-4645](https://jira.hyperledger.org/browse/FAB-4645)] Remove commit level script
* [fb21cd2](https://github.com/hyperledger/fabric-sdk-go/commit/fb21cd2) [[FAB-4645](https://jira.hyperledger.org/browse/FAB-4645)] Update integration test scripts
* [8f1d587](https://github.com/hyperledger/fabric-sdk-go/commit/8f1d587) [[FAB-3162](https://jira.hyperledger.org/browse/FAB-3162)] Add optional secret param
* [7939eab](https://github.com/hyperledger/fabric-sdk-go/commit/7939eab) [[FAB-3757](https://jira.hyperledger.org/browse/FAB-3757)] Update to 1.0.0-alpha2

## v1.0.0-alpha
Wed 31 May 2017 12:00:00 EDT

* [2f30561](https://github.com/hyperledger/fabric-sdk-go/commit/2f30561) [[FAB-4432](https://jira.hyperledger.org/browse/FAB-4432)] Allow access to chain anchor peers
* [b288855](https://github.com/hyperledger/fabric-sdk-go/commit/b288855) [[FAB-4375](https://jira.hyperledger.org/browse/FAB-4375)] Update MockEndorserServer
* [9c0f71b](https://github.com/hyperledger/fabric-sdk-go/commit/9c0f71b) [[FAB-4329](https://jira.hyperledger.org/browse/FAB-4329)] SDK Go - externalize CA root in mock test
* [270f57b](https://github.com/hyperledger/fabric-sdk-go/commit/270f57b) [[FAB-4147](https://jira.hyperledger.org/browse/FAB-4147)] Add links to example projects
* [9204ed8](https://github.com/hyperledger/fabric-sdk-go/commit/9204ed8) [[FAB-4023](https://jira.hyperledger.org/browse/FAB-4023)] SDK Go - apply CC interest after Connect
* [b9f8928](https://github.com/hyperledger/fabric-sdk-go/commit/b9f8928) [[FAB-4027](https://jira.hyperledger.org/browse/FAB-4027)] Update build instructions
* [d55fe9b](https://github.com/hyperledger/fabric-sdk-go/commit/d55fe9b) [[FAB-4027](https://jira.hyperledger.org/browse/FAB-4027)] Update build instructions
* [b8813fe](https://github.com/hyperledger/fabric-sdk-go/commit/b8813fe) [[FAB-4035](https://jira.hyperledger.org/browse/FAB-4035)] Improve handling error in orderer
* [d6d3fe0](https://github.com/hyperledger/fabric-sdk-go/commit/d6d3fe0) [[FAB-3999](https://jira.hyperledger.org/browse/FAB-3999)] EventHub client reconnect fix
* [139fe8f](https://github.com/hyperledger/fabric-sdk-go/commit/139fe8f) [[FAB-4006](https://jira.hyperledger.org/browse/FAB-4006)] SDK Go - Update mock broadcast server
* [061e0e0](https://github.com/hyperledger/fabric-sdk-go/commit/061e0e0) [[FAB-3994](https://jira.hyperledger.org/browse/FAB-3994)] Cleanup test suite runs
* [1b4f9fd](https://github.com/hyperledger/fabric-sdk-go/commit/1b4f9fd) [[FAB-3933](https://jira.hyperledger.org/browse/FAB-3933)] Use docker hub fabric images in test suite
* [84f596f](https://github.com/hyperledger/fabric-sdk-go/commit/84f596f) [[FAB-3965](https://jira.hyperledger.org/browse/FAB-3965)] Review go lint suggestions
* [93f5da5](https://github.com/hyperledger/fabric-sdk-go/commit/93f5da5) [[FAB-3966](https://jira.hyperledger.org/browse/FAB-3966)] Add go linting checks
* [4521257](https://github.com/hyperledger/fabric-sdk-go/commit/4521257) [[FAB-3962](https://jira.hyperledger.org/browse/FAB-3962)] Add license & spelling checks.
* [716acd3](https://github.com/hyperledger/fabric-sdk-go/commit/716acd3) [[FAB-3952](https://jira.hyperledger.org/browse/FAB-3952)] Exported identifiers
* [eb9b94b](https://github.com/hyperledger/fabric-sdk-go/commit/eb9b94b) [[FAB-3902](https://jira.hyperledger.org/browse/FAB-3902)] RegisterTxEvent to return error code
* [9bc35ce](https://github.com/hyperledger/fabric-sdk-go/commit/9bc35ce) [[FAB-3670](https://jira.hyperledger.org/browse/FAB-3670)] Close client connection
* [78b6da1](https://github.com/hyperledger/fabric-sdk-go/commit/78b6da1) Return correct exit code from test script
* [44e57b3](https://github.com/hyperledger/fabric-sdk-go/commit/44e57b3) Fixed integration tests
* [b6f4099](https://github.com/hyperledger/fabric-sdk-go/commit/b6f4099) [[FAB-3558](https://jira.hyperledger.org/browse/FAB-3558)] Allow specific env variable prefix for SDK
* [f55b585](https://github.com/hyperledger/fabric-sdk-go/commit/f55b585) [[FAB-3435](https://jira.hyperledger.org/browse/FAB-3435)] Test case to validate transient data
* [b76434a](https://github.com/hyperledger/fabric-sdk-go/commit/b76434a) [ [FAB-3432](https://jira.hyperledger.org/browse/FAB-3432)] ChaincodeEvent should contain channel ID
* [7346b61](https://github.com/hyperledger/fabric-sdk-go/commit/7346b61) [[FAB-3421](https://jira.hyperledger.org/browse/FAB-3421)] SDK Go - Disconnect and RegisterBlockEvent
* [254b040](https://github.com/hyperledger/fabric-sdk-go/commit/254b040) Updated README
* [5d2c9c8](https://github.com/hyperledger/fabric-sdk-go/commit/5d2c9c8) [[FAB-3409](https://jira.hyperledger.org/browse/FAB-3409)] Add support for env variables in config
* [dbdac6e](https://github.com/hyperledger/fabric-sdk-go/commit/dbdac6e) [[FAB-3423](https://jira.hyperledger.org/browse/FAB-3423)] Remove unsafe type assertions in config.go
* [797da09](https://github.com/hyperledger/fabric-sdk-go/commit/797da09) [[FAB-3421](https://jira.hyperledger.org/browse/FAB-3421)] SDK Go - Disconnect, RegisterBlockEvent
* [ec9d2ef](https://github.com/hyperledger/fabric-sdk-go/commit/ec9d2ef) [[FAB-3313](https://jira.hyperledger.org/browse/FAB-3313)] Move APIs to appropriate class
* [47fffbe](https://github.com/hyperledger/fabric-sdk-go/commit/47fffbe) [[FAB-3217](https://jira.hyperledger.org/browse/FAB-3217)] SDK Go - Deadlock in Event Hub
* [eac4440](https://github.com/hyperledger/fabric-sdk-go/commit/eac4440) [[FAB-3255](https://jira.hyperledger.org/browse/FAB-3255)] Added creating orderer with root CAs
* [6d3307f](https://github.com/hyperledger/fabric-sdk-go/commit/6d3307f) [[FAB-3231](https://jira.hyperledger.org/browse/FAB-3231)] Renamed user accessor methods
* [b920301](https://github.com/hyperledger/fabric-sdk-go/commit/b920301) [[FAB-3324](https://jira.hyperledger.org/browse/FAB-3324)] Get organization units
* [115b0db](https://github.com/hyperledger/fabric-sdk-go/commit/115b0db) [[FAB-3128](https://jira.hyperledger.org/browse/FAB-3128)] Added re-enroll
* [7ab1f3f](https://github.com/hyperledger/fabric-sdk-go/commit/7ab1f3f) Updated README
* [04f6d7d](https://github.com/hyperledger/fabric-sdk-go/commit/04f6d7d) [[FAB-3059](https://jira.hyperledger.org/browse/FAB-3059)] SDK Go - Complete Initialize chain
* [7a66106](https://github.com/hyperledger/fabric-sdk-go/commit/7a66106) [[FAB-3073](https://jira.hyperledger.org/browse/FAB-3073)] SDK Go - Move utility functions
* [c47bbf2](https://github.com/hyperledger/fabric-sdk-go/commit/c47bbf2) [[FAB-3072](https://jira.hyperledger.org/browse/FAB-3072)] SDK Go - Add failed transaction events test
* [b63b116](https://github.com/hyperledger/fabric-sdk-go/commit/b63b116) [[FAB-3059](https://jira.hyperledger.org/browse/FAB-3059)] SDK Go - Initialize chain
* [65db7ec](https://github.com/hyperledger/fabric-sdk-go/commit/65db7ec) Update README
* [10712fe](https://github.com/hyperledger/fabric-sdk-go/commit/10712fe) [[FAB-3027](https://jira.hyperledger.org/browse/FAB-3027)] Updated CI scripts
* [1ec6bf1](https://github.com/hyperledger/fabric-sdk-go/commit/1ec6bf1) Update README
* [4614970](https://github.com/hyperledger/fabric-sdk-go/commit/4614970) [[FAB-3027](https://jira.hyperledger.org/browse/FAB-3027)] Added all target to Makefile
* [818aac3](https://github.com/hyperledger/fabric-sdk-go/commit/818aac3) [[FAB-3027](https://jira.hyperledger.org/browse/FAB-3027)] Added Makefile
* [ca82dda](https://github.com/hyperledger/fabric-sdk-go/commit/ca82dda) Fix EventHub event source for tests
* [388838f](https://github.com/hyperledger/fabric-sdk-go/commit/388838f) [[FAB-3021](https://jira.hyperledger.org/browse/FAB-3021)] SDK Go - Get chaincode events from Block
* [5c04433](https://github.com/hyperledger/fabric-sdk-go/commit/5c04433) [[FAB-3022](https://jira.hyperledger.org/browse/FAB-3022)]fix GetQueryValue interface define
* [d15cb07](https://github.com/hyperledger/fabric-sdk-go/commit/d15cb07) [[FAB-3018](https://jira.hyperledger.org/browse/FAB-3018)]return detailed event hub connection error
* [55bdb74](https://github.com/hyperledger/fabric-sdk-go/commit/55bdb74) Fixed property access
* [c153bac](https://github.com/hyperledger/fabric-sdk-go/commit/c153bac) [[FAB-3003](https://jira.hyperledger.org/browse/FAB-3003)] Updated README
* [6a0dd50](https://github.com/hyperledger/fabric-sdk-go/commit/6a0dd50) Test initialized with testing config
* [f731064](https://github.com/hyperledger/fabric-sdk-go/commit/f731064) [[FAB-3003](https://jira.hyperledger.org/browse/FAB-3003)] Added join channel functionality
* [ae71223](https://github.com/hyperledger/fabric-sdk-go/commit/ae71223) Test initialized with testing config
* [17cd4e0](https://github.com/hyperledger/fabric-sdk-go/commit/17cd4e0) Revert back server port - required for testing
* [2ecb4a5](https://github.com/hyperledger/fabric-sdk-go/commit/2ecb4a5) [[FAB-2979](https://jira.hyperledger.org/browse/FAB-2979)]Fixed TLS Config for fabric CA client
* [14370d1](https://github.com/hyperledger/fabric-sdk-go/commit/14370d1) [[FAB-2975](https://jira.hyperledger.org/browse/FAB-2975)] SDK Go - Chain Query LCCC, SCCC Support
* [d36e7eb](https://github.com/hyperledger/fabric-sdk-go/commit/d36e7eb) Adding data access functions
* [fe98dcb](https://github.com/hyperledger/fabric-sdk-go/commit/fe98dcb) Adding Extension interfaces to Chain and EventHub.
* [dacc9ca](https://github.com/hyperledger/fabric-sdk-go/commit/dacc9ca) Disabled gossip TLS handshake for integration tests
* [b29b07d](https://github.com/hyperledger/fabric-sdk-go/commit/b29b07d) Fix instantiate chaincode test case
* [583a90a](https://github.com/hyperledger/fabric-sdk-go/commit/583a90a) Fix issues reported from goreportcard
* [9b8d2c0](https://github.com/hyperledger/fabric-sdk-go/commit/9b8d2c0) [[FAB-2899](https://jira.hyperledger.org/browse/FAB-2899)] Added create channel functionality
* [4e3f737](https://github.com/hyperledger/fabric-sdk-go/commit/4e3f737) [[FAB-2894](https://jira.hyperledger.org/browse/FAB-2894)] Chaincode deployment - integration tests
* [d27a83e](https://github.com/hyperledger/fabric-sdk-go/commit/d27a83e) Update readme with project links
* [ca69dcd](https://github.com/hyperledger/fabric-sdk-go/commit/ca69dcd) [[FAB-2898](https://jira.hyperledger.org/browse/FAB-2898)] SDK Go - Query SCC Support
* [405b815](https://github.com/hyperledger/fabric-sdk-go/commit/405b815) [[FAB-2895](https://jira.hyperledger.org/browse/FAB-2895)] Fix TLS Certificate location
* [f171bef](https://github.com/hyperledger/fabric-sdk-go/commit/f171bef) Update README file
* [da93bdc](https://github.com/hyperledger/fabric-sdk-go/commit/da93bdc) [[FAB-2882](https://jira.hyperledger.org/browse/FAB-2882)] Update SDK Go to work with v1.0.0-alpha:
* [c426926](https://github.com/hyperledger/fabric-sdk-go/commit/c426926) Added user registration and revocation functionality
* [a2ff332](https://github.com/hyperledger/fabric-sdk-go/commit/a2ff332) Initial commit


---
This document is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.
