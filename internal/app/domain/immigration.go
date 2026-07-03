package domain

import (
	"slices"
	"time"
)

type Gender string

func (g Gender) IsValid() bool     { return slices.Contains(genders, g) }
func (g Gender) Options() []Gender { return genders }
func (g Gender) String() string    { return string(g) }

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

var genders = []Gender{GenderMale, GenderFemale}

type ImmigrationClearance struct {
	FirstName string
	LastName  string
	Gender    Gender

	PlaceOfBirth string
	DateOfBirth  time.Time

	PassportNumber string
	CountryCode    CountryCode

	IsInIndonesia bool

	IssuedAt  time.Time
	ExpiredAt time.Time
}

func NewImmigrationClearance(
	firstName string,
	lastName string,
	gender Gender,
	placeOfBirth string,
	dateOfBirth time.Time,
	passportNumber string,
	countryCode CountryCode,
	isInIndonesia bool,
	issuedAt time.Time,
	expiredAt time.Time,
) ImmigrationClearance {
	return ImmigrationClearance{
		FirstName: firstName,
		LastName:  lastName,
		Gender:    gender,

		PlaceOfBirth: placeOfBirth,
		DateOfBirth:  dateOfBirth,

		PassportNumber: passportNumber,
		CountryCode:    countryCode,

		IsInIndonesia: isInIndonesia,

		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}
}

const StatusCodeErrInternal int64 = -1

type ImmigrationClearanceLog struct {
	ID          string
	RawResponse []byte
	StatusCode  int64

	FirstName *string
	LastName  *string
	Gender    *Gender

	PlaceOfBirth *string
	DateOfBirth  *time.Time

	PassportNumber *string
	CountryCode    *CountryCode

	IsInIndonesia *bool

	IssuedAt  *time.Time
	ExpiredAt *time.Time

	ActorID *string
	IMEI    IMEI

	CreatedAt time.Time
}

func NewImmigrationClearanceLog(
	id string,
	rawResponse []byte,
	statusCode int64,
	firstName *string,
	lastName *string,
	gender *Gender,
	placeOfBirth *string,
	dateOfBirth *time.Time,
	passportNumber *string,
	countryCode *CountryCode,
	isInIndonesia *bool,
	issuedAt *time.Time,
	expiredAt *time.Time,
	actorID *string,
	imei IMEI,
	createdAt time.Time,
) ImmigrationClearanceLog {
	return ImmigrationClearanceLog{
		ID:          id,
		RawResponse: rawResponse,
		StatusCode:  statusCode,

		FirstName: firstName,
		LastName:  lastName,
		Gender:    gender,

		PlaceOfBirth: placeOfBirth,
		DateOfBirth:  dateOfBirth,

		PassportNumber: passportNumber,
		CountryCode:    countryCode,

		IsInIndonesia: isInIndonesia,

		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,

		ActorID: actorID,
		IMEI:    imei,

		CreatedAt: createdAt,
	}
}

type ImmigrationClearanceLogData struct {
	RawResponse []byte
	StatusCode  int64

	FirstName *string
	LastName  *string
	Gender    *Gender

	PlaceOfBirth *string
	DateOfBirth  *time.Time

	PassportNumber *string
	CountryCode    *CountryCode

	IsInIndonesia *bool

	IssuedAt  *time.Time
	ExpiredAt *time.Time

	ActorID *string
	IMEI    IMEI
}

func NewImmigrationClearanceLogData(
	rawResponse []byte,
	statusCode int64,
	firstName *string,
	lastName *string,
	gender *Gender,
	placeOfBirth *string,
	dateOfBirth *time.Time,
	passportNumber *string,
	countryCode *CountryCode,
	isInIndonesia *bool,
	issuedAt *time.Time,
	expiredAt *time.Time,
	actorID *string,
	imei IMEI,
) ImmigrationClearanceLogData {
	return ImmigrationClearanceLogData{
		RawResponse: rawResponse,
		StatusCode:  statusCode,

		FirstName: firstName,
		LastName:  lastName,
		Gender:    gender,

		PlaceOfBirth: placeOfBirth,
		DateOfBirth:  dateOfBirth,

		PassportNumber: passportNumber,
		CountryCode:    countryCode,

		IsInIndonesia: isInIndonesia,

		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,

		ActorID: actorID,
		IMEI:    imei,
	}
}

type ImmigrationClearanceLogEagerLoad struct {
	ICL ImmigrationClearanceLog

	Actor *UserEagerLoad
}

func NewImmigrationClearanceLogEagerLoad(
	icl ImmigrationClearanceLog,
	actor *UserEagerLoad,
) ImmigrationClearanceLogEagerLoad {
	return ImmigrationClearanceLogEagerLoad{
		ICL:   icl,
		Actor: actor,
	}
}
