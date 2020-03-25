package action

import(
	"log"

	sdk "CocosSDK"
)

func DeploySmartContract(contractName, contractPath string) {
	authKey := sdk.Wallet.CreateKey().GetPublicKey().ToBase58String()

	log.Printf("Deploying smart contract... (contract name: %v; contract path: %v; auth key: %v)", contractName, contractPath, authKey)
	txHash, err := sdk.CreateContractByFile(contractName, authKey, contractPath)

	if err != nil {
		log.Fatalf("Smart contract deploy failed error - %v", err)
	} else {
		log.Printf("Smart contract deployed with auth key - %s, transaction hash - %s", authKey, txHash)
	}
}
