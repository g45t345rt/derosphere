Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("originalOwner", SIGNER())
50 STORE("type", "G45-NFT-COLLECTION")
60 STORE("frozenCollection", 0)
70 STORE("frozenMetadata", 0)
80 STORE("metadata", "")
90 STORE("nftCount", 0)
100 STORE("timestamp", BLOCK_TIMESTAMP())
110 RETURN 0
End Function

Function Freeze(collection Uint64, metadata Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF collection == 0 THEN GOTO 50
40 STORE("frozenCollection", 1)
50 IF metadata == 0 THEN GOTO 70
60 STORE("frozenMetadata", 1)
70 RETURN 0
End Function

Function SetNft(nft String, index Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenCollection") == 0 THEN GOTO 50
40 RETURN 1
50 IF EXISTS("nft_" + nft) == 1 THEN GOTO 70
60 STORE("nftCount", LOAD("nftCount") + 1)
70 STORE("nft_" + nft, index)
80 RETURN 0
End Function

Function DelNft(nft String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenCollection") == 0 THEN GOTO 50
40 RETURN 1
50 IF EXISTS("nft_" + nft) == 1 THEN GOTO 70
60 RETURN 1
70 DELETE("nft_" + nft)
80 STORE("nftCount", LOAD("nftCount") - 1)
90 RETURN 0
End Function

Function SetMetadata(metadata String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenMetadata") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("metadata", metadata)
60 RETURN 0
End Function

Function TransferOwnership(newOwner string) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("tempOwner", ADDRESS_RAW(newOwner))
40 RETURN 0
End Function

Function CancelTransferOwnership() Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 DELETE("tempOwner")
40 RETURN 0
End Function

Function ClaimOwnership() Uint64
10 IF LOAD("tempOwner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 DELETE("tempOwner")
50 RETURN 0
End Function