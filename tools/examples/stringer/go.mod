module github.com/rkennedy/magehelper/examples/stringer

go 1.23

toolchain go1.23.1

require (
	github.com/magefile/mage v1.15.0
	github.com/onsi/gomega v1.34.1
	github.com/rkennedy/magehelper v0.0.0-20240801182534-77163c3dd6cb
	golang.org/x/tools v0.23.0
)

require (
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rkennedy/magehelper => ../../..
