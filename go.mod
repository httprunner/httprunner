module github.com/httprunner/httprunner/v4

go 1.16

require (
	github.com/andybalholm/brotli v1.0.4
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/fatih/color v1.13.0
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-openapi/spec v0.20.6
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.1
	github.com/httprunner/funplugin v0.4.8
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
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	gopkg.in/yaml.v3 v3.0.0
)

// replace github.com/httprunner/funplugin => ../funplugin
