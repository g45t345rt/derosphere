package config

import "fmt"

var DATA_FOLDER = "./data"
var DB_WALLETS_FILEPATH = fmt.Sprintf("%s/wallets.db", DATA_FOLDER)
var WALLET_FOLDER_PATH = fmt.Sprintf("%s/wallets", DATA_FOLDER)
