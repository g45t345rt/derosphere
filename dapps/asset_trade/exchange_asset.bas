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

Function odKey(id Uint64, key String) String
10 RETURN "od_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 25) // 2.5%
50 STORE("od_ctr", 0)
60 initStore()
70 RETURN 0
End Function

Function CreateOrder(odType String, lAssetId String, rAssetId String, lRateAmount Uint64, rRateAmount Uint64, expireTimestamp Uint64, oneTxOnly Uint64) Uint64
10 DIM odId, amount, unitPrice as Uint64
20 LET odId = LOAD("od_ctr")
30 IF oneTxOnly > 1 THEN GOTO 300
40 IF odType != "sell" THEN GOTO 70
50 LET amount = ASSETVALUE(HEXDECODE(lAssetId))
60 IF odType != "buy" THEN GOTO 90
70 LET amount = ASSETVALUE(HEXDECODE(rAssetId))
80 IF amount == 0 THEN GOTO 300
90 LET unitPrice = rRateAmount / lRateAmount
100 IF unitPrice == 0 THEN GOTO 300
110 beginStore()
120 storeUint64(odKey(odId, "amountSent"), 0)
130 storeUint64(odKey(odId, "amount"), amount)
140 storeString(odKey(odId, "lAssetId"), lAssetId)
150 storeString(odKey(odId, "type"), odType)
160 storeString(odKey(odId, "rAssetId"), rAssetId)
170 storeUint64(odKey(odId, "lRateAmount"), lRateAmount)
180 storeUint64(odKey(odId, "rRateAmount"), rRateAmount)
190 storeUint64(odKey(odId, "unitPrice"), unitPrice)
200 storeString(odKey(odId, "creator"), ADDRESS_STRING(SIGNER()))
210 storeUint64(odKey(odId, "close"), 0)
220 storeUint64(odKey(odId, "timestamp"), BLOCK_TIMESTAMP())
230 storeUint64(odKey(odId, "expireTimestamp"), expireTimestamp)
240 storeUint64(odKey(odId, "oneTxOnly"), oneTxOnly)
250 storeUint64(odKey(odId, "txCount"), 0)
260 storeString(odKey(odId, "txId"), HEX(TXID()))
270 endStore()
280 STORE("od_ctr", odId + 1)
290 RETURN 0
300 RETURN 1
End Function

Function CloseOrder(odId Uint64) Uint64
5 DIM retrieveFunds as Uint64
10 DIM signer, odType as String
20 LET signer = SIGNER()
30 LET odType = LOAD(odKey(odId, "type"))
40 IF ADDRESS_STRING(signer) != LOAD(odKey(odId, "creator")) THEN GOTO 150
50 IF LOAD(odKey(odId, "close")) == 1 THEN GOTO 150
60 beginStore()
70 LET retrieveFunds = LOAD(odKey(odId, "amount")) - LOAD(odKey(odId, "amountSent"))
80 IF odType != "sell" THEN GOTO 100
90 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "lAssetId"))))
100 IF odType != "buy" THEN GOTO 120
110 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "rAssetId"))))
120 storeUint64(odKey(odId, "close"), 1)
130 endStore()
140 RETURN 0
150 RETURN 1
End Function

Function Transact(odId Uint64) Uint64
10 DIM lAssetId, rAssetId, creator, signer, odType as String
20 DIM ctr, txAmountSent, lRateAmount, rRateAmount, txAmountReceived, priceCut, timestamp, amount, expireTimestamp, amountSent as Uint64
30 IF LOAD(odKey(odId, "close")) == 1 THEN GOTO 600
40 LET timestamp = BLOCK_TIMESTAMP()
50 LET expireTimestamp = LOAD(odKey(odId, "expireTimestamp"))
60 IF expireTimestamp == 0 THEN GOTO 80
70 IF timestamp > expireTimestamp THEN GOTO 600
80 LET lAssetId = LOAD(odKey(odId, "lAssetId"))
90 LET amount = LOAD(odKey(odId, "amount"))
100 LET rAssetId = LOAD(odKey(odId, "rAssetId"))
110 LET odType = LOAD(odKey(odId, "type"))
120 LET amountSent = LOAD(odKey(odId, "amountSent"))
130 LET lRateAmount = LOAD(odKey(odId, "lRateAmount"))
140 LET rRateAmount = LOAD(odKey(odId, "rRateAmount"))
150 LET creator = LOAD(odKey(odId, "creator"))
160 LET signer = SIGNER()
170 LET priceCut = 0
180 LET txAmountSent = 0
190 LET txAmountReceived = 0
200 beginStore()
210 IF odType != "sell" THEN GOTO 330
220 LET txAmountSent = ASSETVALUE(HEXDECODE(rAssetId))
230 IF txAmountSent == 0 THEN GOTO 600
240 LET txAmountReceived = txAmountSent * lRateAmount / rRateAmount
250 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 270
260 IF txAmountReceived != amount THEN GOTO 600
270 IF txAmountReceived == 0 THEN GOTO 600
280 IF txAmountReceived > amount - amountSent THEN GOTO 600
290 IF EXISTS("fee_" + rAssetId) == 0 THEN GOTO 310
300 LET priceCut = txAmountSent * LOAD("fee_" + rAssetId) / 1000
310 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent - priceCut, HEXDECODE(rAssetId))
320 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived, HEXDECODE(lAssetId))
330 IF odType != "buy" THEN GOTO 450
340 LET txAmountSent = ASSETVALUE(HEXDECODE(lAssetId))
350 IF txAmountSent == 0 THEN GOTO 600
360 LET txAmountReceived = txAmountSent * rRateAmount / lRateAmount
370 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 390
380 IF txAmountSent != amount THEN GOTO 600
390 IF txAmountReceived == 0 THEN GOTO 600
400 IF txAmountReceived > amount - amountSent THEN GOTO 600
410 IF EXISTS("fee_" + rAssetId) == 0 THEN GOTO 430
420 LET priceCut = txAmountReceived * LOAD("fee_" + rAssetId) / 1000
430 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent, HEXDECODE(lAssetId))
440 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived - priceCut, HEXDECODE(rAssetId))
450 LET amountSent = amountSent + txAmountReceived
460 storeUint64(odKey(odId, "amountSent"), amountSent)
470 IF amountSent < amount THEN GOTO 490
480 storeUint64(odKey(odId, "close"), 1)
490 SEND_ASSET_TO_ADDRESS(LOAD("owner"), priceCut, HEXDECODE(rAssetId))
500 LET ctr = LOAD(odKey(odId, "txCount"))
510 storeString(odKey(odId, "tx_" + ctr + "_sender"), ADDRESS_STRING(signer))
520 storeUint64(odKey(odId, "tx_" + ctr + "_amountSent"), txAmountSent)
530 storeUint64(odKey(odId, "tx_" + ctr + "_amountReceived"), txAmountReceived)
540 storeUint64(odKey(odId, "tx_" + ctr + "_timestamp"), timestamp)
550 storeString(odKey(odId, "tx_" + ctr + "_txId"), HEX(TXID()))
560 storeUint64(odKey(odId, "tx_" + ctr + "_fee"), priceCut)
570 storeUint64(odKey(odId, "txCount"), ctr + 1)
580 endStore()
590 RETURN 0
600 RETURN 1
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 50
20 IF fee > 100 THEN GOTO 50
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