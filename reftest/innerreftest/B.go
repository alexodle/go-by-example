package innerreftest

type B struct {
}

func (a *B) F1(b, c string) (string, error) {
	return "", nil
}
