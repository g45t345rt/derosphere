Function initStore()
10 STORE("commit_ctr", 0)
20 RETURN
End Function

Function beginStore()
10 MAPSTORE("commit", "{")
20 MAPSTORE("commit_ctr", LOAD("commit_ctr"))
30 RETURN
End Function

Function appendStore(key String, value String)
10 DIM commit as String
20 LET commit = MAPGET("commit")
30 IF commit == "{" THEN GOTO 50
40 LET commit = commit + ","
50 MAPSTORE("commit", commit + "\"" + key + "\":" + value)
60 RETURN
End Function

Function storeString(key String, value String)
10 DIM commit as String
20 STORE(key, value)
30 appendStore(key, "\"" + value + "\"")
40 RETURN
End Function

Function storeUint64(key String, value Uint64)
10 DIM commit as String
20 STORE(key, value)
30 appendStore(key, "" + value + "")
40 RETURN
End Function

Function deleteKey(key String)
10 DIM commit as String
20 DELETE(key)
30 appendStore(key, "-1")
40 RETURN
End Function

Function endStore()
10 DIM ctr as Uint64
20 LET ctr = MAPGET("commit_ctr")
30 STORE("commit_" + ctr, MAPGET("commit") + "}")
40 STORE("commit_ctr", ctr + 1)
50 RETURN
End Function

Function auKey(id Uint64, key String) String
10 RETURN "au_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 2)
50 STORE("au_ctr", 0)
60 initStore()
70 RETURN 0
End Function

Function CreateAuction(sellAssetId String, bidAssetId String, startAmount Uint64, minBidAmount Uint64, startTimestamp Uint64, duration Uint64) Uint64
10 DIM auId, sellAmount as Uint64
20 LET auId = LOAD("au_ctr")
30 LET sellAmount = ASSETVALUE(HEXDECODE(sellAssetId))
40 IF sellAmount > 0 THEN GOTO 60
50 RETURN 1
60 IF startTimestamp > 0 THEN GOTO 80
70 LET startTimestamp = BLOCK_TIMESTAMP()
80 beginStore()
90 storeUint64(auKey(auId, "startAmount"), startAmount)
100 storeString(auKey(auId, "sellAssetId"), sellAssetId)
110 storeUint64(auKey(auId, "sellAmount"), sellAmount)
120 storeUint64(auKey(auId, "startTimestamp"), startTimestamp)
130 storeUint64(auKey(auId, "duration"), duration)
140 storeString(auKey(auId, "seller"), ADDRESS_STRING(SIGNER()))
150 storeString(auKey(auId, "bidAssetId"), bidAssetId)
160 storeUint64(auKey(auId, "minBidAmount"), minBidAmount)
170 storeUint64(auKey(auId, "bidSum"), 0)
180 storeUint64(auKey(auId, "bidCount"), 0)
190 storeUint64(auKey(auId, "timestamp"), BLOCK_TIMESTAMP())
200 storeUint64(auKey(auId, "close"), 0)
210 endStore()
220 STORE("au_ctr", auId + 1)
230 RETURN 0
End Function

Function SetAuctionMinBid(auId Uint64, amount Uint64) Uint64
10 IF LOAD(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 beginStore()
40 storeUint64(auKey(auId, "minBidAmount"), amount)
50 endStore()
60 RETURN 0
End Function

Function CloseAuction(auId Uint64) Uint64
10 IF LOAD(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 IF LOAD(auKey(auId, "bidCount")) == 0 THEN GOTO 50
40 RETURN 1
50 SEND_ASSET_TO_ADDRESS(SIGNER(), LOAD(auKey(auId, "sellAmount")), HEXDECODE(LOAD(auKey(auId, "sellAssetId"))))
60 beginStore()
70 storeUint64(auKey(auId, "close"), 1)
80 endStore()
90 RETURN 0
End Function

Function Bid(auId Uint64) Uint64
10 DIM minBidAmount, bidAmount, bidCount, lockedAmount, startAmount, bidSum, startTimestamp, duration, timestamp as Uint64
20 DIM bidAssetId, signerString as String
30 LET signerString = ADDRESS_STRING(SIGNER())
40 LET minBidAmount = LOAD(auKey(auId, "minBidAmount"))
50 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
60 LET bidCount = LOAD(auKey(auId, "bidCount"))
70 LET startTimestamp = LOAD(auKey(auId, "startTimestamp"))
80 LET duration = LOAD(auKey(auId, "duration"))
90 LET bidAmount = ASSETVALUE(HEXDECODE(bidAssetId))
100 LET startAmount = LOAD(auKey(auId, "startAmount"))
110 LET bidSum = LOAD(auKey(auId, "bidSum"))
120 LET timestamp = BLOCK_TIMESTAMP()
130 LET lockedAmount = 0
140 IF EXISTS(auKey(auId, "bid_" + signerString + "_lockedAmount")) == 0 THEN GOTO 160
150 LET lockedAmount = LOAD(auKey(auId, "bid_" + signerString + "_lockedAmount"))
160 IF timestamp <= startTimestamp + duration THEN GOTO 180
170 RETURN 1
180 IF bidAmount >= minBidAmount THEN GOTO 200
190 RETURN 1
200 IF bidSum > 0 THEN GOTO 230
210 IF bidAmount >= startAmount + minBidAmount THEN GOTO 230
220 RETURN 1
230 IF lockedAmount + bidAmount > bidSum THEN GOTO 250
240 RETURN 1
250 beginStore()
260 storeUint64(auKey(auId, "bid_" + signerString + "_lockedAmount"), lockedAmount + bidAmount)
270 storeUint64(auKey(auId, "bid_" + signerString + "_timestamp"), timestamp)
280 storeUint64(auKey(auId, "bidSum"), bidSum + bidAmount)
290 storeUint64(auKey(auId, "bidCount"), bidCount + 1)
300 storeString(auKey(auId, "lastBidder"), signerString)
310 endStore()
320 RETURN 0
End Function

Function CheckoutAuction(auId Uint64) Uint64
10 DIM sellAssetId, bidAssetId, seller, winner as String
20 DIM bidSum, amount, startTimestamp, duration, sellAmount as Uint64
30 LET sellAssetId = LOAD(auKey(auId, "sellAssetId"))
40 LET sellAmount = LOAD(auKey(auId, "sellAmount"))
50 LET bidSum = LOAD(auKey(auId, "bidSum"))
60 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
70 LET seller = LOAD(auKey(auId, "seller"))
80 LET winner = LOAD(auKey(auId, "lastBidder"))
90 LET startTimestamp = LOAD(auKey(auId, "startTimestamp"))
100 LET duration = LOAD(auKey(auId, "duration"))
110 IF LOAD(auKey(auId, "close")) == 0 THEN GOTO 130
120 RETURN 1
130 IF BLOCK_TIMESTAMP() > startTimestamp + duration THEN GOTO 150
140 RETURN 1
150 IF EXISTS("fee_" + bidAssetId) == 0 THEN GOTO 290
160 LET auctionCut = amountToReceive * LOAD("fee_" + bidAssetId) / 100
170 LET bidSum = bidSum - auctionCut
180 beginStore()
190 storeUint64(auKey(auId, "close"), 1)
200 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(winner), sellAmount, HEXDECODE(sellAssetId))
210 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), bidSum, HEXDECODE(bidAssetId))
220 SEND_ASSET_TO_ADDRESS(LOAD("owner"), auctionCut, HEXDECODE(bidAssetId))
230 endStore()
240 RETURN 0
End Function

Function RetrieveLockedFunds(auId Uint64) Uint64
10 DIM startTimestamp, duration, lockedAmount, bidSum as Uint64
20 DIM bidAssetId, signerString, winner as String
30 LET startTimestamp = LOAD(auKey(auId, "startTimestamp"))
40 LET duration = LOAD(auKey(auId, "duration"))
50 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
60 LET bidSum = LOAD(auKey(auId, "bidSum"))
70 LET signerString = ADDRESS_STRING(SIGNER())
80 LET winner = LOAD(auKey(auId, "lastBidder"))
90 IF LOAD(auKey(auId, "close")) == 1 THEN GOTO 110
100 RETURN 1
110 IF winner != signerString THEN GOTO 130
120 RETURN 1
130 LET lockedAmount = LOAD(auKey(auId, "bid_" + signerString + "_lockedAmount"))
140 beginStore()
150 SEND_ASSET_TO_ADDRESS(SIGNER(), lockedAmount, HEXDECODE(bidAssetId))
160 storeUint64(auKey(auId, "bid_" + signerString + "_lockedAmount"), 0)
170 endStore()
180 RETURN 0
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF fee <= 100 THEN GOTO 50
40 RETURN 1
50 STORE("fee" + assetId, fee)
60 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function
