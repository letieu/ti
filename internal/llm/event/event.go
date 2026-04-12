package event

type Event interface{}

type Start struct{}
type End struct{}

type TextStart struct{}
type TextDelta struct {
	Delta string
}
type TextEnd struct{}

type ThinkingStart struct{}
type ThinkingDelta struct {
	Delta string
}
type ThinkingEnd struct{}

type FunctionStart struct {
	Id               string
	Name             string
	Args             map[string]any
	ThoughtSignature string
}
type FunctionDelta struct {
	Delta string
}
type FunctionEnd struct{}

type Error struct {
	Type string
	Msg  string
	Code string
}
