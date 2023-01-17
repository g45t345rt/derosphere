Function initStore()
10 STORE("commit_ctr", 0)
20 RETURN
End Function

Function beginStore()
10 MAPSTORE("commit", "{")
20 MAPSTORE("commit_ctr", LOAD("commit_ctr"))
30 RETURN
End Function

Function appendCommit(key String, value String)
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
30 appendCommit(key, "\"" + value + "\"")
40 RETURN
End Function

Function storeUint64(key String, value Uint64)
10 DIM commit as String
20 STORE(key, value)
30 appendCommit(key, "" + value + "")
40 RETURN
End Function

Function deleteKey(key String)
10 DIM commit as String
20 DELETE(key)
30 appendCommit(key, "-1")
40 RETURN
End Function

Function endStore()
10 DIM ctr as Uint64
20 LET ctr = MAPGET("commit_ctr")
30 STORE("commit_" + ctr, MAPGET("commit") + "}")
40 STORE("commit_ctr", ctr + 1)
50 RETURN
End Function

Function gbKey(id Uint64, key String) String
10 RETURN "gb_" + id + "_" + key
End Function

Function storeAssets(gbId Uint64, assetList String) Uint64
DIM maxAssets, index, amount as Uint64
DIM asset, indexCard as String
LET maxAssets = STRLEN(assetList) / 64
LET index = LOAD(gbKey(gbId, "assetInCount"))
LET asset = SUBSTR(assetList, index, index + 64)
LET amount = ASSETVALUE(HEXDECODE(asset))
storeString(gbKey(gbId, "dp_" + index + "_asset"), asset)
storeUint64(gbKey(gbId, "dp_" + index + "_amount"), amount)
LET index = index + 1
IF index < maxAssets THEN GOTO
storeUint64(gbKey(gbId, "assetInCount"), index)
RETURN index
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 50) // 5%
50 STORE("gb_ctr", 0)
60 initStore()
70 RETURN 0
End Function

Function Create(price Uint64, priceAssetId String, assetList String, dispenseCooldown Uint64, oneDispensePerTx Uint64, freeze Uint64) Uint64
LET gbId = LOAD("gb_ctr")
beginStore()
storeString(gbKey(gbId, "creator"), ADDRESS_STRING(SIGNER()))
storeUint64(gbKey(gbId, "price"), price)
storeString(gbKey(gbId, "priceAssetId"), priceAssetId)
storeUint64(gbKey(gbId, "dispenseCooldown"), dispenseCooldown)
storeUint64(gbKey(gbId, "oneDispensePerTx"), oneDispensePerTx)
storeUint64(gbKey(gbId, "assetInCount"), 0)
storeUint64(gbKey(gbId, "assetOutCount"), 0)
storeUint64(gbKey(gbId, "freeze"), freeze)
storeString(odKey(odId, "txId"), HEX(TXID()))
storeAssets(gbId, assetList)
STORE("gb_ctr", gbId + 1)
endStore()
RETURN 0
End Function

Function Edit(gbId Uint64, price Uint64, dispenseCooldown Uint64, oneDispensePerTx Uint64, freeze Uint64) Uint64
IF LOAD(gbKey(gbId, "creator")) != ADDRESS_STRING(SIGNER()) THEN GOTO
IF LOAD(gbKey(gbId, "freeze") > 0) THEN GOTO
RETURN 0
RETURN 1
End Function

Function internalDispense(signer String, assetInCount Uint64, assetOutCount Uint64) Uint64
DIM asset as String
DIM nbr as Uint64
LET index = RANDOM() * assetInCount
LET asset = LOAD(gbKey(gbId, "dp_" + index + "_asset"))
LET amount = LOAD(gbKey(gbId, "dp_" + index + "_amount"))
IF amount = 0 THEN GOTO
SEND_ASSET_TO_ADDRESS(signer, amount, asset)
storeUint64(gbKey(gbId, "dp_" + index + "_amount"), 0)
storeUint64(gbKey(gbId, "assetOutCount"), assetOutCount-1)
storeUint64(gbKey(gbId, "dp_" + index + "_txId"), HEX(TXID()))

LET index = index + 1
IF index > assetInCount THEN

End Function

Function Dispense(gbId Uint64, amount Uint64) Uint64
DIM totalAssets, oneDispensePerTx as Uint64
LET signer = SIGNER()
LET assetInCount = LOAD(gbKey(gbId, "assetInCount"))
LET assetOutCount = LOAD(gbKey(gbId, "assetOutCount"))
LET totalAssets = assetInCount - assetOutCount
LET oneDispensePerTx = LOAD(gbKey(gbId, "oneDispensePerTx"))
IF amount > totalAssets THEN GOTO
RETURN 0
RETURN 1
End Function

Function Load(gbId Uint64, assetList String) Uint64
IF LOAD(gbKey(gbId, "creator")) != ADDRESS_STRING(SIGNER()) THEN GOTO
beginStore()
storeAssets(gbId, assetList)
endStore()
RETURN 0
RETURN 1
End Function

Function Delete(gbId Uint64) Uint64
IF LOAD(gbKey(gbId, "creator")) != ADDRESS_STRING(SIGNER()) THEN GOTO
RETURN 0
RETURN 1
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