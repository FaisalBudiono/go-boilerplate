package invalid

type Code string

const (
	ShouldString Code = "string"

	ShouldDate     Code = "date"
	ShouldDateOnly Code = "date-only"

	ShouldNumber Code = "number"
	ShouldUnique Code = "unique"

	ShouldEnum Code = "enum"
	Luhn       Code = "luhn"

	Required Code = "required"

	MaxLength Code = "max-length-exceeded"
)
