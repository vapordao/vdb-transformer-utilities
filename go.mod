module github.com/makerdao/vdb-transformer-utilities

go 1.15

require (
	github.com/ethereum/go-ethereum v1.9.22
	github.com/makerdao/vulcanizedb v0.0.15-rc.1.0.20200923220430-893edc1b439b
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/rs/cors v1.7.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/viper v1.7.1
)

replace github.com/ethereum/go-ethereum => github.com/makerdao/go-ethereum v1.9.21-rc1
