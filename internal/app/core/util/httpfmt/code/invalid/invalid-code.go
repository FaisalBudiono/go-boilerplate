package invalid

type Code string

const (
	ShouldDate   Code = "date"
	ShouldNumber Code = "number"
	ShouldString Code = "string"
	ShouldArray  Code = "array"

	ShouldEnum Code = "enum"

	Email Code = "email"

	MaxLength Code = "max-length-exceeded"
	Required  Code = "required"
	Unique    Code = "unique"
)
