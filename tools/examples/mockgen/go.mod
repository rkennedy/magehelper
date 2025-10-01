module github.com/rkennedy/magehelper/examples/mockgen

go 1.25

require (
	github.com/logrusorgru/aurora/v3 v3.0.0
	github.com/magefile/mage v1.15.0
	github.com/rkennedy/magehelper v0.0.0-20240929185338-3185725b6dfb
	go.uber.org/mock v0.6.0
)

require (
	github.com/deckarep/golang-set/v2 v2.8.0 // indirect
	github.com/onsi/ginkgo/v2 v2.19.0 // indirect
	github.com/onsi/gomega v1.34.1 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rkennedy/magehelper => ../../..
