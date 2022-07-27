package config

import "fmt"

var DATA_FOLDER = "./data"
var WALLET_FOLDER_PATH = fmt.Sprintf("%s/wallets", DATA_FOLDER)
var START_ENV = "simulator"

func GetCountFilename(env string) string {
	return fmt.Sprintf("%s/%s_counts.json", DATA_FOLDER, env)
}
