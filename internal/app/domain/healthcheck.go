package domain

type HealthcheckDep struct {
	Name   string
	Ok     bool
	Status string
}

func NewHealthcheckDep(
	name string,
	ok bool,
	status string,
) HealthcheckDep {
	return HealthcheckDep{
		Name:   name,
		Ok:     ok,
		Status: status,
	}
}
