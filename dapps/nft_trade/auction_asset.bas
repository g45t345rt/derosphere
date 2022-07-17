Function storeCommitString(action String, key String, value String)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function storeCommitInt(action String, key String, value Uint64)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function initCommit()
10 STORE("commit_count", 0)
20 RETURN
End Function

Function beginCommit()
10 MAPSTORE("commit_count", LOAD("commit_count"))
20 RETURN
End Function

Function endCommit()
10 STORE("commit_count", MAPGET("commit_count"))
20 RETURN
End Function

Function storeStateString(key String, value String)
10 STORE("state_" + key, value)
20 storeCommitString("S", "state_" + key, value)
30 RETURN
End Function

Function storeStateInt(key String, value Uint64)
10 STORE("state_" + key, value)
20 storeCommitInt("S", "state_" + key, value)
30 RETURN
End Function

Function deleteState(key String)
10 DELETE("state_" + key)
20 storeCommitInt("D", "state_" + key, 0)
30 RETURN
End Function

Function loadStateString(key String) String
10 RETURN LOAD("state_" + key)
End Function

Function loadStateInt(key String) Uint64
10 RETURN LOAD("state_" + key)
End Function

Function stateExists(key String) Uint64
10 RETURN EXISTS("state_" + key)
End Function

Function storeTX()
10 STORE("txid_" + HEX(TXID()), 1)
20 RETURN
End Function

Function auKey(id Uint64, key String) String
10 RETURN "au_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("dero_fee", 5)
50 STORE("au_ctr", 0)
60 RETURN 0
End Function

Function CreateAuction(sellAssetId String, bidAssetId String, startAmount Uint64, minBidAmount Uint64, startTimestamp Uint64, duration Uint64) Uint64
10 DIM auId, start_timestamp as Uint64
20 LET auId = LOAD("au_ctr")
30 IF ASSETVALUE(HEXDECODE(sellAssetId)) == 1 THEN GOTO
40 RETURN 1
50 beginCommit()
60 storeStateInt(auKey(auId, "startAmount"), startAmount)
70 storeStateString(auKey(auId, "sellAssetId"), sellAssetId)
80 storeStateInt(auKey(auId, "startTimestamp"), startTimestamp)
90 storeStateInt(auKey(auId, "duration"), duration)
100 storeStateString(auKey(auId, "seller"), ADDRESS_STRING(SIGNER()))
110 storeStateInt(auKey(auId, "bidAssetId"), bidAssetId)
120 storeStateInt(auKey(auId, "minBidAmount"), minBidAmount)
130 storeStateInt(auKey(auId, "bidSum"), 0)
140 storeStateInt(auKey(auId, "bidCount"), 0)
150 storeStateInt(auKey(auId, "timestamp"), BLOCK_TIMESTAMP())
160 endCommit()
170 STORE("au_ctr", auId + 1)
180 storeTX()
190 RETURN 0
End Function

Function SetAuctionMinBid(auId Uint64, amount Uint64) Uint64
10 IF loadStateString(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 storeStateInt(auKey(auId, "minBidAmount"), amount)
40 RETURN 0
End Function

Function CancelAuction(auId Uint64) Uint64
10 IF loadStateString(auKey(id, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 IF loadStateInt(auKey(auId, "bidCount")) == 0 THEN GOTO 50
40 RETURN 1
50 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, HEXDECODE(loadStateString(auKey(id, "assetId"))))
60 beginCommit()
70 deleteState(auKey(auId, "startAmount"))
80 deleteState(auKey(auId, "sellAssetId"))
90 deleteState(auKey(auId, "startTimestamp"))
100 deleteState(auKey(auId, "duration"))
110 deleteState(auKey(auId, "seller"))
120 deleteState(auKey(auId, "minBidAmount"))
130 deleteState(auKey(auId, "bidAssetId"))
140 deleteState(auKey(auId, "bidSum"))
150 deleteState(auKey(auId, "bidCount"))
160 deleteState(auKey(auId, "timestamp"))
170 endCommit()
180 storeTX()
190 RETURN 0
End Function

Function Bid(auId Uint64) Uint64
10 DIM minBidAmount, bidAmount, bidCount, lockedAmount as Uint64
20 DIM bidAssetId, signerString as String
30 LET signerString = ADDRESS_RAW(SIGNER())
40 LET minBidAmount = loadStateInt(auKey(auId, "minBidAmount"))
50 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
60 LET bidCount = loadStateInt(auKey(auId, "bidCount"))
70 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
80 LET duration = loadStateInt(auKey(auId, "duration"))
90 LET bidAmount = ASSETVALUE(HEXDECODE(bidAssetId))
100 LET lockedAmount = 0
120 IF stateExists(aubey(auId, "bid_" + signerString + "_lockedAmount")) == 0 THEN GOTO 140
130 LET lockedAmount = loadStateInt(aubey(auId, "bid_" + signerString + "_lockedAmount"))
140 IF lockedAmount + bidAmount >= minBidAmount + startAmount THEN GOTO 160
150 RETURN 1
160 IF BLOCK_TIMESTAMP() <= startTimestamp + duration THEN GOTO 180
170 RETURN 1
180 beginCommit()
190 LET lockedAmount = lockedAmount + bidAmount
200 storeStateInt(auKey(auId, "bid_" + signerString + "_lockedAmount"), lockedAmount)
210 storeStateInt(auKey(auId, "bid_" + signerString + "_timestamp"), BLOCK_TIMESTAMP())
220 storeStateInt(auKey(auId, "bidSum"), lockedAmount)
230 storeStateInt(auKey(auId, "bidCount"), bidCount + 1)
240 endCommit()
250 storeTX()
260 RETURN 0
End Function

Function CheckoutAuction(auId Uint64) Uint64
10 DIM sellAssetId, bidAssetId, seller as String
20 DIM bidSum, amount, startTimestamp, duration as Uint64
30 LET sellAssetId = loadStateString(auKey(auId, "sellAssetId"))
40 LET amount = loadStateInt(auKey(auId, "amount"))
50 LET bidSum = loadStateInt(auKey(auId, "bidSum"))
60 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
70 LET seller = loadStateString(auKey(auId, "seller"))
80 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
90 LET duration = loadStateInt(auKey(auId, "duration"))
100 IF BLOCK_TIMESTAMP() > startTimestamp + duration THEN GOTO
110 RETURN 1
120 beginCommit()
130 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(signer), amount, HEXDECODE(sellAssetId))
140 endCommit()
150 storeTX()
160 RETURN 0
End Function

Function RetrieveLockedFunds(auId String, signer String)
10 DIM startTimestamp, duration, signerAmount, bidSum as Uint64
20 DIM bidAssetId as String
30 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
40 LET duration = loadStateInt(auKey(auId, "duration"))
50 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
60 LET bidSum = loadStateInt(auKey(auId, "bidSum"))
70 IF BLOCK_TIMESTAMP() > startTimestamp + duration THEN GOTO 90
80 RETURN 1
90 LET signerAmount = loadStateInt(auKey(auId, "bid_" + signer + "_lockedAmount"))
100 IF signerAmount < bidSum THEN GOTO 120
110 RETURN 1
120 beginCommit()
130 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(signer), signerAmount, HEXDECODE(bidAssetId))
140 storeStateInt(auKey(auId, "bid_" + signer + "_lockedAmount"), 0)
150 endCommit()
160 storeTX()
170 RETURN 0
End Function

Function SetDeroFee(fee Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF fee <= 100 THEN GOTO 50
40 RETURN 1
50 STORE("dero_fee", fee)
60 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function
