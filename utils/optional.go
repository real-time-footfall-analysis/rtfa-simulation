package utils

type OptionalFloat64 struct {
	hasValue bool
	value    float64
}

func (optional OptionalFloat64) Value() (float64, bool) {
	return optional.value, optional.hasValue
}

func OptionalFloat64WithValue(v float64) OptionalFloat64 {
	return OptionalFloat64{
		hasValue: true,
		value:    v,
	}
}

func OptionalFloat64WithEmptyValue() OptionalFloat64 {
	return OptionalFloat64{
		hasValue: false,
		value:    0,
	}
}
