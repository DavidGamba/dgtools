package tz

default_group: "work"

_miami: {city: "Miami", admin1: "Florida", country_code: "US"}
_calgary: {city: "Calgary", country_code: "CA"}
_san_jose: {city: "San Jose", admin1: "San Jose", country_code: "CR"}
_palo_alto: {city: "Palo Alto", country_code: "US"}
_detroit: {city: "Detroit", country_code: "US"}

group: "work": {
	member: "John": _palo_alto
	member: "Alan": _palo_alto
	member: "Mike": {city: "Rochester", country_code: "gb"}
	member: "Steve": {time_zone: "Europe/London"}
	member: "Patrick": {city: "Flint", time_zone: "America/Detroit"}

	member: "Palo_Alto": _palo_alto
	member: "Palo_Alto": {type: "city"}
	member: "Detroit": _detroit
	member: "Detroit": {type: "city"}
	member: "London": {type: "city", time_zone: "Europe/London"}
}

group: "family": {
	member: "Brother": _miami
	member: "David": _calgary
	member: "Parents": _san_jose
	member: "Bogota": {city: "Bogota", country_code: "CO", type: "city"}
}
