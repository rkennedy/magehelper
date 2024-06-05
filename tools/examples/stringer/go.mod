module github.com/rkennedy/magehelper/examples/stringer

go 1.22

toolchain go1.22.1

require (
	github.com/magefile/mage v1.15.0
	github.com/onsi/gomega v1.33.1
	github.com/rkennedy/magehelper v0.0.0-20240515031727-7ac4d753137e
	golang.org/x/tools v0.20.0
)

require (
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rkennedy/magehelper => ../../..
