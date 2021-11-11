module github.com/httprunner/hrp

go 1.16

require (
	github.com/debugtalk/boomer v1.6.1-0.20211111125723-e54e33eea42f
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/jmespath/go-jmespath v0.4.0
	github.com/maja42/goval v1.2.1
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20211110154304-99a53858aa08 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6
