package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/spf13/cobra"
)

// Version of this tool
var Version = "1.0.1"

// BuildDate of this tool
var BuildDate = "2018-08-08 08:08:08"

// Borrow from https://github.com/btcsuite/btcd/tree/master/btcec

func generate() ([2]string, error) {
	fmt.Printf("generate key")
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2d4f87" +
		"20ee63e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return [2]string{"", ""}, err
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)
	return [2]string{string(privKey.Serialize()), string(pubKey.SerializeUncompressed())}, nil
}

func verify(prikey string, pubkey string) {
	//
	fmt.Printf("private key: %s\n", prikey)
	fmt.Printf("public  key: %s\n", pubkey)
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2d4f87" +
		"20ee63e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)

	// Sign a message using the private key.
	message := "test message"
	messageHash := chainhash.DoubleHashB([]byte(message))
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Serialize and display the signature.
	fmt.Printf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	fmt.Printf("Signature Verified? %v\n", verified)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "keypair",
		Short:   "keypair tool",
		Long:    "keypair generate/verify private/public key pair",
		Version: fmt.Sprintf("[%s @ %s]", Version, BuildDate),
	}
	//rootCmd.SetVersionTemplate("Version: `Version`, build: `BuildDate`")

	generateCmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g"},
		Short:   "generate key",
		Long:    "generate private/public key pair",
		Run: func(cmd *cobra.Command, args []string) {
			generate()
		},
	}
	rootCmd.AddCommand(generateCmd)

	verifyCmd := &cobra.Command{
		Use:     "verify",
		Aliases: []string{"v"},
		Short:   "verify key",
		Long:    "verify private/public key pair",
		Run: func(cmd *cobra.Command, args []string) {

			verify(cmd.Flag("prikey").Value.String(), cmd.Flag("pubkey").Value.String())
		},
	}
	verifyCmd.Flags().StringP("prikey", "i", "", "private key")
	verifyCmd.Flags().StringP("pubkey", "u", "", "public key")

	verifyCmd.MarkFlagRequired("prikey")
	verifyCmd.MarkFlagRequired("pubkey")

	rootCmd.AddCommand(verifyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
