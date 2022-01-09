package models

type ValueCompare struct {
	LeftName      string
	LeftOriginal  string
	LeftValue     string
	LeftBasePath  string
	RightName     string
	RightOriginal string
	RightValue    string
	RightBasePath string
	Different     bool
}
