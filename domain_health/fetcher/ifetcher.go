package fetcher

type Fetcher interface {
	Fetch() (records []string, err error)
}
