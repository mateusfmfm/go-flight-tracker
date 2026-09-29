package graph

import (
	"strings"

	"go-flight-tracker/graph/model"
	"go-flight-tracker/internal/flight"
)

// hasValidPosition reports whether OpenSky provided a usable lat/lon.
// time_position is null in the API when no position is available.
func hasValidPosition(a *flight.Aircraft) bool {
	if a == nil {
		return false
	}
	return a.TimePosition != 0
}

// MatchesFilter returns true when the aircraft passes the optional FlightFilter.
// Aircraft without a valid position are excluded when any geo constraint is set
// (bbox). Non-geo filters still require a match on the other fields.
func MatchesFilter(a *flight.Aircraft, f *model.FlightFilter) bool {
	if a == nil {
		return false
	}
	if f == nil {
		return true
	}

	if f.OriginCountry != nil && *f.OriginCountry != "" {
		if !strings.EqualFold(a.OriginCountry, *f.OriginCountry) {
			return false
		}
	}

	if f.CallsignPrefix != nil && *f.CallsignPrefix != "" {
		if !strings.HasPrefix(strings.ToUpper(a.Callsign), strings.ToUpper(*f.CallsignPrefix)) {
			return false
		}
	}

	if f.MinAltitude != nil && a.BaroAltitude < *f.MinAltitude {
		return false
	}
	if f.MaxAltitude != nil && a.BaroAltitude > *f.MaxAltitude {
		return false
	}

	if hasBBox(f) {
		if !hasValidPosition(a) {
			return false
		}
		if a.Latitude < *f.Lamin || a.Latitude > *f.Lamax {
			return false
		}
		if a.Longitude < *f.Lomin || a.Longitude > *f.Lomax {
			return false
		}
	}

	return true
}

func hasBBox(f *model.FlightFilter) bool {
	return f != nil &&
		f.Lamin != nil && f.Lomin != nil &&
		f.Lamax != nil && f.Lomax != nil
}

func filterAircrafts(list []*flight.Aircraft, f *model.FlightFilter) []*model.Aircraft {
	out := make([]*model.Aircraft, 0, len(list))
	for _, a := range list {
		if !MatchesFilter(a, f) {
			continue
		}
		if m := mapToModel(a); m != nil {
			out = append(out, m)
		}
	}
	return out
}

func mapToModel(a *flight.Aircraft) *model.Aircraft {
	if a == nil {
		return nil
	}
	lastPos := int32(a.TimePosition)
	lastContact := int32(a.LastContact)
	squawk := a.Squawk

	m := &model.Aircraft{
		Icao24:             a.Icao24,
		Callsign:           a.Callsign,
		OriginCountry:      a.OriginCountry,
		BaroAltitude:       &a.BaroAltitude,
		OnGround:           a.OnGround,
		Velocity:           &a.Velocity,
		TrueTrack:          &a.TrueTrack,
		VerticalRate:       &a.VerticalRate,
		Squawk:             &squawk,
		Spi:                a.Spi,
		PositionSource:     int32(a.PositionSource),
		Category:           int32(a.Category),
		LastPositionUpdate: &lastPos,
		LastContact:        &lastContact,
	}

	if hasValidPosition(a) {
		lon := a.Longitude
		lat := a.Latitude
		m.Longitude = &lon
		m.Latitude = &lat
	}

	return m
}
