package fetch

type Fetcher interface {
	Fetch() (records []string, err error)
}
