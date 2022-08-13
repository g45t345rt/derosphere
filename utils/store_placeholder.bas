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

Function Initialize() Uint64
10 initStore()
20 RETURN 0
End Function

Function StoreTest() Uint64
10 beginStore()
20 storeString("teststring", "hello")
30 storeUint64("testuint64", 12345)
40 endStore()
50 RETURN 0
End Function

Function StoreValue(value String) Uint64
10 beginStore()
20 storeString("value", value)
30 endStore()
40 RETURN 0
End Function

Function DeleteTest() Uint64
10 beginStore()
20 deleteKey("teststring")
30 endStore()
40 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 UPDATE_SC_CODE(code)
20 RETURN 0
End Function