package hrp

type StepType string

const (
	stepTypeRequest     StepType = "request"
	stepTypeAPI         StepType = "api"
	stepTypeTestCase    StepType = "testcase"
	stepTypeTransaction StepType = "transaction"
	stepTypeRendezvous  StepType = "rendezvous"
	stepTypeThinkTime   StepType = "thinktime"
	stepTypeWebSocket   StepType = "websocket"
	stepTypeAndroid     StepType = "android"
	stepTypeHarmony     StepType = "harmony"
	stepTypeIOS         StepType = "ios"
	stepTypeShell       StepType = "shell"

	stepTypeSuffixExtraction StepType = "_extraction"
	stepTypeSuffixValidation StepType = "_validation"
)
