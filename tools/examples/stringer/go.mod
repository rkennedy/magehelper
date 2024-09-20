module github.com/rkennedy/magehelper/examples/stringer

go 1.23

toolchain go1.23.1

require (
	github.com/magefile/mage v1.15.0
	github.com/onsi/gomega v1.34.2
	github.com/rkennedy/magehelper v0.0.0-20240814151936-35a7f2a0f33a
	golang.org/x/tools v0.25.0
)

require (
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rkennedy/magehelper => ../../..
