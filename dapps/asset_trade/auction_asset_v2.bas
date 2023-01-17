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
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 50) // 5%
50 STORE("au_ctr", 100)
60 initStore()
70 RETURN 0
End Function

Function CreateAuction(sellAssetId String, bidAssetId String, startAmount Uint64, minBidAmount Uint64, startTimestamp Uint64, duration Uint64) Uint64
10 DIM auId as Uint64
20 LET auId = LOAD("au_ctr")
30 IF startTimestamp > 0 THEN GOTO 50
40 LET startTimestamp = BLOCK_TIMESTAMP()
50 IF EXISTS("fee_" + bidAssetId) == 0 THEN GOTO 230
60 beginStore()
70 storeUint64(auKey(auId, "startAmount"), startAmount)
80 storeString(auKey(auId, "sellAssetId"), sellAssetId)
90 storeUint64(auKey(auId, "sellAmount"), ASSETVALUE(HEXDECODE(sellAssetId)))
100 storeUint64(auKey(auId, "startTimestamp"), startTimestamp)
110 storeUint64(auKey(auId, "duration"), duration)
120 storeString(auKey(auId, "seller"), ADDRESS_STRING(SIGNER()))
130 storeString(auKey(auId, "bidAssetId"), bidAssetId)
140 storeUint64(auKey(auId, "minBidAmount"), minBidAmount)
150 storeUint64(auKey(auId, "currentBid"), 0)
160 storeUint64(auKey(auId, "bidCount"), 0)
170 storeUint64(auKey(auId, "timestamp"), BLOCK_TIMESTAMP())
180 storeUint64(auKey(auId, "close"), 0)
190 storeString(auKey(auId, "txId"), HEX(TXID()))
200 endStore()
210 STORE("au_ctr", auId + 1)
220 RETURN 0
230 RETURN 1
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
10 IF LOAD(auKey(auId, "seller")) != ADDRESS_STRING(SIGNER()) THEN GOTO 90
20 IF LOAD(auKey(auId, "bidCount")) > 0 THEN GOTO 90
30 IF LOAD(auKey(auId, "close")) == 1 THEN GOTO 90
40 SEND_ASSET_TO_ADDRESS(SIGNER(), LOAD(auKey(auId, "sellAmount")), HEXDECODE(LOAD(auKey(auId, "sellAssetId"))))
50 beginStore()
60 storeUint64(auKey(auId, "close"), 1)
70 endStore()
80 RETURN 0
90 RETURN 1
End Function

Function isAuctionFinished (auId Uint64) Uint64
10 DIM startTimestamp, duration, timestamp as Uint64
20 LET startTimestamp = LOAD(auKey(auId, "startTimestamp"))
30 LET duration = LOAD(auKey(auId, "duration"))
40 LET timestamp = BLOCK_TIMESTAMP()
50 IF timestamp <= startTimestamp + duration THEN GOTO 70
60 RETURN 1
70 RETURN 0
End Function

Function Bid(auId Uint64) Uint64
10 DIM minBidAmount, bidAmount, bidCount, lockedAmount, startAmount, currentBid, timestamp, bdrBidCount as Uint64
20 DIM bidAssetId, signerString as String
30 LET signerString = ADDRESS_STRING(SIGNER())
40 LET minBidAmount = LOAD(auKey(auId, "minBidAmount"))
50 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
60 LET bidCount = LOAD(auKey(auId, "bidCount"))
70 LET bidAmount = ASSETVALUE(HEXDECODE(bidAssetId))
80 LET startAmount = LOAD(auKey(auId, "startAmount"))
90 LET currentBid = LOAD(auKey(auId, "currentBid"))
95 LET timestamp = BLOCK_TIMESTAMP()
100 LET lockedAmount = 0
110 IF EXISTS(auKey(auId, "bdr_" + signerString + "_lockedAmount")) == 0 THEN GOTO 130
120 LET lockedAmount = LOAD(auKey(auId, "bdr_" + signerString + "_lockedAmount"))
130 LET lockedAmount = lockedAmount + bidAmount
135 IF LOAD(auKey(auId, "close")) == 1 THEN GOTO 150
136 IF timestamp < LOAD(auKey(auId, "startTimestamp")) THEN GOTO 150
140 IF isAuctionFinished(auId) == 0 THEN GOTO 160
150 RETURN 1
160 IF currentBid > 0 THEN GOTO 200
170 IF bidAmount >= startAmount THEN GOTO 220
190 RETURN 1
200 IF lockedAmount >= currentBid + minBidAmount THEN GOTO 220
210 RETURN 1
220 beginStore()
230 LET currentBid = currentBid + (lockedAmount - currentBid)
240 LET bdrBidCount = 0
250 IF EXISTS(auKey(auId, "bdr_" + signerString + "_bidCount")) == 0 THEN GOTO 270
260 LET bdrBidCount = LOAD(auKey(auId, "bdr_" + signerString + "_bidCount"))
270 storeUint64(auKey(auId, "bdr_" + signerString + "_lockedAmount"), lockedAmount)
280 storeUint64(auKey(auId, "bdr_" + signerString + "_bidAmount"), lockedAmount)
290 storeUint64(auKey(auId, "bdr_" + signerString + "_timestamp"), timestamp)
300 storeUint64(auKey(auId, "bdr_" + signerString + "_bidCount"), bdrBidCount + 1)
310 storeUint64(auKey(auId, "currentBid"), currentBid)
320 storeUint64(auKey(auId, "bidCount"), bidCount + 1)
330 storeString(auKey(auId, "lastBidder"), signerString)
340 endStore()
350 RETURN 0
End Function

Function CheckoutAuction(auId Uint64) Uint64
10 DIM sellAssetId, bidAssetId, seller, winner as String
20 DIM currentBid, sellAmount, auctionCut as Uint64
30 LET sellAssetId = LOAD(auKey(auId, "sellAssetId"))
40 LET sellAmount = LOAD(auKey(auId, "sellAmount"))
50 LET currentBid = LOAD(auKey(auId, "currentBid"))
60 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
70 LET seller = LOAD(auKey(auId, "seller"))
80 LET winner = LOAD(auKey(auId, "lastBidder"))
90 LET auctionCut = 0
100 IF LOAD(auKey(auId, "close")) == 0 THEN GOTO 120
110 RETURN 1
120 IF isAuctionFinished(auId) == 1 THEN GOTO 140
130 RETURN 1
140 IF EXISTS("fee_" + bidAssetId) == 0 THEN GOTO 170
150 LET auctionCut = currentBid * LOAD("fee_" + bidAssetId) / 1000
160 LET currentBid = currentBid - auctionCut
170 beginStore()
180 storeUint64(auKey(auId, "close"), 1)
190 storeUint64(auKey(auId, "bdr_" + winner + "_lockedAmount"), 0)
200 storeUint64(auKey(auId, "checkoutFee"), auctionCut)
210 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(winner), sellAmount, HEXDECODE(sellAssetId))
220 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), currentBid, HEXDECODE(bidAssetId))
230 SEND_ASSET_TO_ADDRESS(LOAD("owner"), auctionCut, HEXDECODE(bidAssetId))
240 endStore()
250 RETURN 0
End Function

Function RetrieveLockedFunds(auId Uint64) Uint64
10 DIM lockedAmount as Uint64
20 DIM bidAssetId, signerString, lastBidder as String
30 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
40 LET signerString = ADDRESS_STRING(SIGNER())
50 LET lastBidder = LOAD(auKey(auId, "lastBidder"))
60 IF lastBidder != signerString THEN GOTO 80
70 RETURN 1
80 LET lockedAmount = LOAD(auKey(auId, "bdr_" + signerString + "_lockedAmount"))
90 IF lockedAmount > 0 THEN GOTO 110
100 RETURN 1
110 beginStore()
120 SEND_ASSET_TO_ADDRESS(SIGNER(), lockedAmount, HEXDECODE(bidAssetId))
130 storeUint64(auKey(auId, "bdr_" + signerString + "_lockedAmount"), 0)
140 endStore()
150 RETURN 0
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 50
20 IF fee > 1000 THEN GOTO 50
30 STORE("fee_" + assetId, fee)
40 RETURN 0
50 RETURN 1
End Function

Function DelAssetFee(assetId String) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 DELETE("fee_" + assetId)
30 RETURN 0
40 RETURN 1
End Function

Function TransferOwnership(newMinter string) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 STORE("tempOwner", ADDRESS_RAW(newMinter))
30 RETURN 0
40 RETURN 1
End Function

Function CancelTransferOwnership() Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 DELETE("tempOwner")
30 RETURN 0
40 RETURN 1
End Function

Function ClaimOwnership() Uint64
10 IF LOAD("tempOwner") != SIGNER() THEN GOTO 50
20 STORE("owner", SIGNER())
30 DELETE("tempOwner")
40 RETURN 0
50 RETURN 1
End Function