package action

import (
	"log"
	"fmt"

	"github.com/spf13/viper"
	sdk "CocosSDK"
	"github.com/siddontang/go-mysql/client"
)

var dbConn *client.Conn; 
var dbConnPools []*client.Conn;

func Init() {
	InitBCXSdk()
	InitDB()
}

func Close() {
	if dbConn != nil {
		dbConn.Close()
	}

	for _, con := range dbConnPools {
		con.Close()
	}
}

func InitBCXSdk() {
	host := viper.GetString("cocosbcx.node.host")
	port := viper.GetInt("cocosbcx.node.port")
	ssl := viper.GetBool("cocosbcx.node.use_ssl")

	//init cocos sdk
	log.Printf("Initializing CocosBCX sdk host - %#v; port - %#v; use ssl - %#v\n", host, port, ssl)
	sdk.InitSDK(host, ssl, port)
}

func InitDB() { 
	host := viper.GetString("mysql.host")
	port := viper.GetInt("mysql.port")
    username := viper.GetString("mysql.username")
	password := viper.GetString("mysql.password")
	database := viper.GetString("mysql.database")

	// init db
	log.Printf("Initializing MySQL db host - %#v; port - %#v; username - %#v; password - %#v; database - %#v\n", host, port, username, password, database)
	var err error

	dbUrl := fmt.Sprintf("%v:%v", host, port)
	dbConn, err = client.Connect(dbUrl, username, password, database)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		con, err := client.Connect(dbUrl, username, password, database)
		if err != nil {
			log.Fatal(err)
		}
		dbConnPools = append(dbConnPools, con)
	}
}

func InitBCXWallet() error {
	path := viper.GetString("cocosbcx.wallet.wallet_path")
	password := viper.GetString("cocosbcx.wallet.wallet_pwd")
	prikey := viper.GetString("cocosbcx.wallet.wif_prk")
	bcxAccount := viper.GetString("cocosbcx.wallet.bcx_account")

	err := sdk.Wallet.LoadWallet(path)
	if err != nil {
		err = sdk.Wallet.AddAccountByPrivateKey(prikey, password)
		if err == nil {
			sdk.Wallet.SaveAs(path)
		}
	}

	if err == nil {
		return sdk.Wallet.SetDefaultAccount(bcxAccount, password)
	}

	return err
}