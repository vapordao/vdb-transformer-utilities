package constants

import (
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/makerdao/vulcanizedb/pkg/eth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var initialized = false

func initConfig() {
	if initialized {
		return
	}

	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		panic(fmt.Sprintf("Could not find environment file: %v", err))
	}
	initialized = true
}

func getEnvironmentString(key string) string {
	initConfig()
	value := viper.GetString(key)
	if value == "" {
		logrus.Fatalf("No environment configuration variable set for key: \"%v\"", key)
	}
	return value
}

/* Returns all contract config names from transformer configuration:
[exporter.vow_file]
	path = "transformers/events/vow_file/initializer"
	type = "eth_event"
	repository = "github.com/makerdao/vdb-mcd-transformers"
	migrations = "db/migrations"
	contracts = ["MCD_VOW"]   <----
	rank = "0"
*/
func GetTransformerContractNames(transformerLabel string) []string {
	initConfig()
	configKey := "exporter." + transformerLabel + ".contracts"
	contracts := viper.GetStringSlice(configKey)
	if len(contracts) == 0 {
		logrus.Fatalf("No contracts configured for transformer: \"%v\"", transformerLabel)
	}
	return contracts
}

// GetContractABI returns the ABI for the given contract name
func GetContractABI(contractName string) string {
	initConfig()
	configKey := "contract." + contractName + ".abi"
	contractABI := viper.GetString(configKey)
	if contractABI == "" {
		logrus.Fatalf("No ABI configured for contract: \"%v\"", contractName)
	}
	return contractABI
}

// GetABIFromContractsWithMatchingABI gets the ABI for multiple contracts from config
// Makes sure the ABI matches for all, since a single transformer may run against many contracts.
func GetABIFromContractsWithMatchingABI(contractNames []string) string {
	if len(contractNames) < 1 {
		logrus.Fatalf("No contracts to get ABI for")
	}
	initConfig()
	contractABI := GetContractABI(contractNames[0])
	parsedABI, parseErr := eth.ParseAbi(contractABI)
	if parseErr != nil {
		panic(fmt.Sprintf("unable to parse ABI for %s", contractNames[0]))
	}
	for _, contractName := range contractNames[1:] {
		nextABI := GetContractABI(contractName)
		nextParsedABI, nextParseErr := eth.ParseAbi(nextABI)
		if nextParseErr != nil {
			panic(fmt.Sprintf("unable to parse ABI for %s", contractName))
		}
		if compareErr := CompareContractABI(parsedABI, nextParsedABI); compareErr != nil {
			panic(fmt.Errorf("ABIs don't match for contracts: %s and %s. Reason: %w", contractNames[0], contractName, compareErr))
		}
	}
	return contractABI
}

// GetFirstABI returns the ABI from the first contract in a collection in config
func GetFirstABI(contractNames []string) string {
	if len(contractNames) < 1 {
		logrus.Fatalf("No contracts to get ABI for")
	}
	initConfig()
	return GetContractABI(contractNames[0])
}

// Get the minimum deployment block for multiple contracts from config
func GetMinDeploymentBlock(contractNames []string) int64 {
	if len(contractNames) < 1 {
		logrus.Fatalf("No contracts supplied")
	}
	initConfig()
	minBlock := int64(math.MaxInt64)
	for _, c := range contractNames {
		deployed := getDeploymentBlock(c)
		if deployed < minBlock {
			minBlock = deployed
		}
	}
	return minBlock
}

func getDeploymentBlock(contractName string) int64 {
	configKey := "contract." + contractName + ".deployed"
	value := viper.GetInt64(configKey)
	if value == -1 {
		logrus.Infof("No deployment block configured for contract \"%v\", defaulting to 0.", contractName)
		return 0
	}
	return value
}

// Get the addresses for multiple contracts from config
func GetContractAddresses(contractNames []string) (addresses []string) {
	if len(contractNames) < 1 {
		logrus.Fatalf("No contracts supplied")
	}
	initConfig()
	for _, contractName := range contractNames {
		addresses = append(addresses, GetContractAddress(contractName))
	}
	return
}

func GetContractAddress(contract string) string {
	return getEnvironmentString("contract." + contract + ".address")
}

type MismatchedConstructorsError struct {
	constructorOne, constructorTwo string
}

func NewMismatchedConstructorsError(a, b abi.ABI) error {
	return &MismatchedConstructorsError{
		constructorOne: a.Constructor.String(),
		constructorTwo: b.Constructor.String(),
	}
}

func (err *MismatchedConstructorsError) Error() string {
	return fmt.Sprintf("constructors don't match constructorOne: %s, constructorTwo: %s", err.constructorOne, err.constructorTwo)
}

type MismatchedOuterMethodsError struct {
	method, methodOne, methodTwo string
}

func NewMismatchedOuterMethodsError(a, b abi.ABI, method string) error {
	return &MismatchedOuterMethodsError{
		method:    method,
		methodOne: a.Methods[method].String(),
		methodTwo: b.Methods[method].String(),
	}
}

func (err *MismatchedOuterMethodsError) Error() string {
	return fmt.Sprintf("outer methods don't match for method %s, method one: %s, method two: %s", err.method, err.methodOne, err.methodTwo)
}

type MismatchedOuterEventsError struct {
	event, eventOne, eventTwo string
}

func NewMismatchedOuterEventsError(a, b abi.ABI, event string) error {
	return &MismatchedOuterEventsError{
		eventOne: a.Events[event].String(),
		eventTwo: a.Events[event].String(),
		event:    event,
	}
}

func (err *MismatchedOuterEventsError) Error() string {
	return fmt.Sprintf("outer events don't match for event %s, event one: %s, event two: %s", err.event, err.eventOne, err.eventTwo)
}

func CompareContractABI(a, b abi.ABI) error {
	if a.Constructor.String() != b.Constructor.String() {
		return NewMismatchedConstructorsError(a, b)
	}
	for ak, av := range a.Methods {
		if av.String() != b.Methods[ak].String() {
			return NewMismatchedOuterMethodsError(a, b, ak)
		}
	}
	for ak, av := range a.Events {
		if av.String() != b.Events[ak].String() {
			return NewMismatchedOuterEventsError(a, b, ak)
		}
	}
	return nil
}
