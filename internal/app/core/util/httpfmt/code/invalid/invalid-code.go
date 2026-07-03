package invalid

type Code string

const (
	ShouldDate   Code = "date"
	ShouldNumber Code = "number"
	ShouldString Code = "string"

	ShouldEnum Code = "enum"

	MaxLength Code = "max-length-exceeded"
	Required  Code = "required"
	Unique    Code = "unique"
)
