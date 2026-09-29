package graph

import (
	"testing"

	"go-flight-tracker/graph/model"
	"go-flight-tracker/internal/flight"
)

func ptr[T any](v T) *T { return &v }

func TestMatchesFilter_BBox(t *testing.T) {
	a := &flight.Aircraft{
		Icao24:       "e48b00",
		Callsign:     "GLO1045",
		Latitude:     -23.5,
		Longitude:    -46.6,
		TimePosition: 1,
		BaroAltitude: 10000,
	}

	f := &model.FlightFilter{
		Lamin: ptr(-25.0),
		Lomin: ptr(-48.0),
		Lamax: ptr(-22.0),
		Lomax: ptr(-45.0),
	}
	if !MatchesFilter(a, f) {
		t.Fatal("expected aircraft inside bbox to match")
	}

	a.Latitude = -10
	if MatchesFilter(a, f) {
		t.Fatal("expected aircraft outside bbox to be rejected")
	}
}

func TestMatchesFilter_NoPositionWithBBox(t *testing.T) {
	a := &flight.Aircraft{Icao24: "abc", Latitude: 0, Longitude: 0, TimePosition: 0}
	f := &model.FlightFilter{
		Lamin: ptr(-25.0),
		Lomin: ptr(-48.0),
		Lamax: ptr(-22.0),
		Lomax: ptr(-45.0),
	}
	if MatchesFilter(a, f) {
		t.Fatal("aircraft without position must not match bbox filter")
	}
}

func TestMatchesFilter_CallsignAndAltitude(t *testing.T) {
	a := &flight.Aircraft{
		Icao24:       "e48b00",
		Callsign:     "GLO1045",
		TimePosition: 1,
		BaroAltitude: 5000,
		Latitude:     -23,
		Longitude:    -46,
	}
	f := &model.FlightFilter{
		CallsignPrefix: ptr("glo"),
		MinAltitude:    ptr(4000.0),
		MaxAltitude:    ptr(6000.0),
	}
	if !MatchesFilter(a, f) {
		t.Fatal("expected match")
	}
	f.CallsignPrefix = ptr("TAM")
	if MatchesFilter(a, f) {
		t.Fatal("expected callsign mismatch")
	}
}
