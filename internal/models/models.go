package models

type MediaInfo struct {
	Title        string
	Year         int
	Season       int
	Episode      int
	IsMovie      bool
	Quality      string
	Source       string
	Language     string
	OriginalName string // Para debugging
}

type ParseResult struct {
	MediaInfo  *MediaInfo
	Errors     []error
	Warnings   []string
	Confidence float32 // 0.0 - 1.0
}

type TestCase struct {
	Input       string
	Expected    *MediaInfo
	ShouldError bool
}
