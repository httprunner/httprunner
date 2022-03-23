module github.com/httprunner/hrp

go 1.16

require (
	github.com/andybalholm/brotli v1.0.4
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/google/uuid v1.3.0
	github.com/httprunner/funplugin v0.4.0
	github.com/jinzhu/copier v0.3.2
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/maja42/goval v1.2.1
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.26.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// replace github.com/httprunner/funplugin => ../funplugin
