package txt

type options struct {
	skipEmptyLines bool
}

type Option func(*options)
