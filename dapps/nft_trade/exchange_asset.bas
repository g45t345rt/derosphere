Function itemKey(id Uint64, key String) String
10 RETURN "item_" + id + "_" + key
End Function

Function deleteItem(id Uint64)
10 DELETE(itemKey(id, "amount"))
20 DELETE(itemKey(id, "assetId"))
30 DELETE(itemKey(id, "forAssetId"))
40 DELETE(itemKey(id, "forAmount"))
50 DELETE(itemKey(id, "seller"))
End Function

Function Initialize() Uint64
10 STORE("owner", SIGNER())
20 STORE("dero_fee", 5)
30 STORE("item_count", 0)
40 STORE("sold_items", 0)
50 RETURN 0
End Function

Function Sell(assetId String, forAssetId String, forAmount Uint64) Uint64
10 DIM id, amount as Uint64
20 LET id = LOAD("item_count")
30 LET amount = ASSETVALUE(assetId)
40 IF amount > 0 THEN GOTO 60
50 RETURN 1
60 STORE(itemKey(id, "amount") amount)
70 STORE(itemKey(id, "assetId"), assetId)
80 STORE(itemKey(id, "forAssetId"), forAssetId)
90 STORE(itemKey(id, "forAmount"), forAmount)
100 STORE(itemKey(id, "seller"), SIGNER())
110 STORE("item_count", id + 1)
120 RETURN 0
End Function

Function CancelSell(id String) Uint64
10 IF LOAD(itemKey(id, "seller")) == SIGNER() THEN GOTO 30
20 RETURN 1
30 deleteItem(id)
40 RETURN 0
End Function

Function Buy(id String) Uint64
10 DIM assetId, forAssetId, seller as String
20 DIM amount, forAmount, deroCut as Uint64
30 LET assetId = loadStateString(itemKey(id, "assetId"))
40 LET amount = loadStateInt(itemKey(id, "amount"))
50 LET forAssetId = loadStateString(itemKey(id, "forAssetId"))
60 LET forAmount = loadStateInt(itemKey(id, "forAmount"))
70 LET seller = loadStateString(itemKey(id, "seller"))
80 IF toAssetId != "" THEN GOTO 130
90 IF DEROVALUE() != forAmount THEN GOTO 200
100 LET deroCut = forAmount * LOAD("dero_fee") / 100
110 SEND_DERO_TO_ADDRESS(seller, forAmount - deroCut)
120 SEND_DERO_TO_ADDRESS(LOAD("owner"), deroCut)
130 IF toAssetId == "" THEN GOTO 160
140 IF ASSETVALUE(toAssetId) != forAmount THEN GOTO 200
150 SEND_ASSET_TO_ADDRESS(seller, forAmount, toAssetId)
160 SEND_ASSET_TO_ADDRESS(SIGNER(), amount, assetId)
170 deleteItem(id)
180 STORE("sold_items", LOAD("sold_items") + 1)
190 RETURN 0
200 RETURN 1
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
