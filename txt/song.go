package txt

type Line struct {
	Spec   rune
	Fields []string
}

type Song struct {
	Header Header

	LinesP1 []Line
	LinesP2 []Line
}
