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

Function auKey(id Uint64, key String) String
10 RETURN "au_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("dero_fee", 5)
50 STORE("au_ctr", 0)
60 initCommit()
70 RETURN 0
End Function

Function CreateAuction(sellAssetId String, bidAssetId String, startAmount Uint64, minBidAmount Uint64, startTimestamp Uint64, duration Uint64) Uint64
10 DIM auId, sellAmount as Uint64
20 LET auId = LOAD("au_ctr")
30 LET sellAmount = ASSETVALUE(HEXDECODE(sellAssetId))
40 IF sellAmount >= 1 THEN GOTO 60
50 RETURN 1
60 IF startTimestamp > 0 THEN GOTO 80
70 LET startTimestamp = BLOCK_TIMESTAMP()
80 beginCommit()
90 storeStateInt(auKey(auId, "startAmount"), startAmount)
100 storeStateString(auKey(auId, "sellAssetId"), sellAssetId)
110 storeStateInt(auKey(auId, "sellAmount"), sellAmount)
120 storeStateInt(auKey(auId, "startTimestamp"), startTimestamp)
130 storeStateInt(auKey(auId, "duration"), duration)
140 storeStateString(auKey(auId, "seller"), ADDRESS_STRING(SIGNER()))
150 storeStateString(auKey(auId, "bidAssetId"), bidAssetId)
160 storeStateInt(auKey(auId, "minBidAmount"), minBidAmount)
170 storeStateInt(auKey(auId, "bidSum"), 0)
180 storeStateInt(auKey(auId, "bidCount"), 0)
190 storeStateInt(auKey(auId, "timestamp"), BLOCK_TIMESTAMP())
200 storeStateInt(auKey(auId, "complete"), 0)
210 endCommit()
220 STORE("au_ctr", auId + 1)
230 RETURN 0
End Function

Function SetAuctionMinBid(auId Uint64, amount Uint64) Uint64
10 IF loadStateString(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 storeStateInt(auKey(auId, "minBidAmount"), amount)
40 RETURN 0
End Function

Function CancelAuction(auId Uint64) Uint64
10 IF loadStateString(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 IF loadStateInt(auKey(auId, "bidCount")) == 0 THEN GOTO 50
40 RETURN 1
50 SEND_ASSET_TO_ADDRESS(SIGNER(), loadStateInt(auKey(auId, "sellAmount")), HEXDECODE(loadStateString(auKey(auId, "sellAssetId"))))
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
170 deleteState(auKey(auId, "sellAmount"))
180 endCommit()
190 RETURN 0
End Function

Function Bid(auId Uint64) Uint64
10 DIM minBidAmount, bidAmount, bidCount, lockedAmount, startAmount, bidSum, startTimestamp, duration, timestamp as Uint64
20 DIM bidAssetId, signerString as String
30 LET signerString = ADDRESS_STRING(SIGNER())
40 LET minBidAmount = loadStateInt(auKey(auId, "minBidAmount"))
50 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
60 LET bidCount = loadStateInt(auKey(auId, "bidCount"))
70 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
80 LET duration = loadStateInt(auKey(auId, "duration"))
90 LET bidAmount = ASSETVALUE(HEXDECODE(bidAssetId))
100 LET startAmount = loadStateInt(auKey(auId, "startAmount"))
110 LET bidSum = loadStateInt(auKey(auId, "bidSum"))
120 LET timestamp = BLOCK_TIMESTAMP()
130 LET lockedAmount = 0
140 IF stateExists(auKey(auId, "bid_" + signerString + "_lockedAmount")) == 0 THEN GOTO 160
150 LET lockedAmount = loadStateInt(auKey(auId, "bid_" + signerString + "_lockedAmount"))
160 IF timestamp <= startTimestamp + duration THEN GOTO 180
170 RETURN 1
180 IF bidAmount >= minBidAmount THEN GOTO 200
190 RETURN 1
200 IF bidSum > 0 THEN GOTO 230
210 IF bidAmount >= startAmount + minBidAmount THEN GOTO 230
220 RETURN 1
230 IF lockedAmount + bidAmount > bidSum THEN GOTO 250
240 RETURN 1
250 beginCommit()
260 storeStateInt(auKey(auId, "bid_" + signerString + "_lockedAmount"), lockedAmount + bidAmount)
270 storeStateInt(auKey(auId, "bid_" + signerString + "_timestamp"), timestamp)
280 storeStateInt(auKey(auId, "bidSum"), bidSum + bidAmount)
290 storeStateInt(auKey(auId, "bidCount"), bidCount + 1)
300 storeStateString(auKey(auId, "lastBidder"), signerString)
310 endCommit()
320 RETURN 0
End Function

Function CheckoutAuction(auId Uint64) Uint64
10 DIM sellAssetId, bidAssetId, seller, winner as String
20 DIM bidSum, amount, startTimestamp, duration, sellAmount as Uint64
30 LET sellAssetId = loadStateString(auKey(auId, "sellAssetId"))
40 LET sellAmount = loadStateInt(auKey(auId, "sellAmount"))
50 LET bidSum = loadStateInt(auKey(auId, "bidSum"))
60 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
70 LET seller = loadStateString(auKey(auId, "seller"))
80 LET winner = loadStateString(auKey(auId, "lastBidder"))
90 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
100 LET duration = loadStateInt(auKey(auId, "duration"))
110 IF loadStateInt(auKey(auId, "complete")) == 0 THEN GOTO 130
120 RETURN 1
130 IF BLOCK_TIMESTAMP() > startTimestamp + duration THEN GOTO 150
140 RETURN 1
150 beginCommit()
160 storeStateInt(auKey(auId, "complete"), 1)
170 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(winner), sellAmount, HEXDECODE(sellAssetId))
180 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), bidSum, HEXDECODE(bidAssetId))
190 endCommit()
200 RETURN 0
End Function

Function RetrieveLockedFunds(auId Uint64) Uint64
10 DIM startTimestamp, duration, lockedAmount, bidSum as Uint64
20 DIM bidAssetId, signerString, winner as String
30 LET startTimestamp = loadStateInt(auKey(auId, "startTimestamp"))
40 LET duration = loadStateInt(auKey(auId, "duration"))
50 LET bidAssetId = loadStateString(auKey(auId, "bidAssetId"))
60 LET bidSum = loadStateInt(auKey(auId, "bidSum"))
70 LET signerString = ADDRESS_STRING(SIGNER())
80 LET winner = loadStateString(auKey(auId, "lastBidder"))
90 IF loadStateInt(auKey(auId, "complete")) == 1 THEN GOTO 110
100 RETURN 1
110 IF winner != signerString THEN GOTO 130
120 RETURN 1
130 LET lockedAmount = loadStateInt(auKey(auId, "bid_" + signerString + "_lockedAmount"))
140 beginCommit()
150 SEND_ASSET_TO_ADDRESS(SIGNER(), lockedAmount, HEXDECODE(bidAssetId))
160 storeStateInt(auKey(auId, "bid_" + signerString + "_lockedAmount"), 0)
170 endCommit()
180 RETURN 0
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
