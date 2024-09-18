module github.com/rkennedy/magehelper/examples/mockgen

go 1.22

toolchain go1.22.1

require (
	github.com/golang/mock v1.6.0
	github.com/magefile/mage v1.15.0
	github.com/rkennedy/magehelper v0.0.0-20240801182534-77163c3dd6cb
)

require (
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/onsi/ginkgo/v2 v2.19.0 // indirect
	github.com/onsi/gomega v1.34.1 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rkennedy/magehelper => ../../..
