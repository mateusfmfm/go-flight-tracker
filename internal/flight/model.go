package flight

import (
	"strings"
)

type OpenSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

type Aircraft struct {
	Icao24         string  `json:"icao24"`
	Callsign       string  `json:"callsign"`
	OriginCountry  string  `json:"origin_country"`
	TimePosition   int64   `json:"time_position"`
	LastContact    int64   `json:"last_contact"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	BaroAltitude   float64 `json:"baro_altitude"`
	OnGround       bool    `json:"on_ground"`
	Velocity       float64 `json:"velocity"`
	TrueTrack      float64 `json:"true_track"`
	VerticalRate   float64 `json:"vertical_rate"`
	Sensors        []int   `json:"sensors"`
	GeoAltitude    float64 `json:"geo_altitude"`
	Squawk         string  `json:"squawk"`
	Spi            bool    `json:"spi"`
	PositionSource int     `json:"position_source"`
	Category       int     `json:"category"`
}

// ParseOpenSkyResponse converte toda a resposta da OpenSky para uma lista de Aircrafts
func ParseOpenSkyResponse(resp *OpenSkyResponse) []*Aircraft {
	if resp == nil || len(resp.States) == 0 {
		return nil
	}

	aircrafts := make([]*Aircraft, 0, len(resp.States))
	for _, vector := range resp.States {
		if aircraft := parseStateVector(vector); aircraft != nil {
			aircrafts = append(aircrafts, aircraft)
		}
	}

	return aircrafts
}

// parseStateVector converte um vetor da matriz OpenSky ([]interface{}) para Aircraft
func parseStateVector(vector []interface{}) *Aircraft {
	if len(vector) < 18 {
		return nil
	}

	a := &Aircraft{}

	// [0] icao24 (string)
	if val, ok := vector[0].(string); ok {
		a.Icao24 = val
	}

	// [1] callsign (string)
	if val, ok := vector[1].(string); ok {
		a.Callsign = strings.TrimSpace(val)
	}

	// [2] origin_country (string)
	if val, ok := vector[2].(string); ok {
		a.OriginCountry = val
	}

	// [3] time_position (int64)
	if vector[3] != nil {
		if val, ok := vector[3].(float64); ok {
			a.TimePosition = int64(val)
		}
	}

	// [4] last_contact (int64)
	if vector[4] != nil {
		if val, ok := vector[4].(float64); ok {
			a.LastContact = int64(val)
		}
	}

	// [5] longitude (float64)
	if vector[5] != nil {
		if val, ok := vector[5].(float64); ok {
			a.Longitude = val
		}
	}

	// [6] latitude (float64)
	if vector[6] != nil {
		if val, ok := vector[6].(float64); ok {
			a.Latitude = val
		}
	}

	// [7] baro_altitude (float64)
	if vector[7] != nil {
		if val, ok := vector[7].(float64); ok {
			a.BaroAltitude = val
		}
	}

	// [8] on_ground (bool)
	if vector[8] != nil {
		if val, ok := vector[8].(bool); ok {
			a.OnGround = val
		}
	}

	// [9] velocity (float64)
	if vector[9] != nil {
		if val, ok := vector[9].(float64); ok {
			a.Velocity = val
		}
	}

	// [10] true_track (float64)
	if vector[10] != nil {
		if val, ok := vector[10].(float64); ok {
			a.TrueTrack = val
		}
	}

	// [11] vertical_rate (float64)
	if vector[11] != nil {
		if val, ok := vector[11].(float64); ok {
			a.VerticalRate = val
		}
	}

	// [13] geo_altitude (float64)
	if vector[13] != nil {
		if val, ok := vector[13].(float64); ok {
			a.GeoAltitude = val
		}
	}

	// [14] squawk (string)
	if vector[14] != nil {
		if val, ok := vector[14].(string); ok {
			a.Squawk = val
		}
	}

	// [15] spi (bool)
	if vector[15] != nil {
		if val, ok := vector[15].(bool); ok {
			a.Spi = val
		}
	}

	// [16] position_source (int)
	if vector[16] != nil {
		if val, ok := vector[16].(float64); ok {
			a.PositionSource = int(val)
		}
	}

	// [17] category (int)
	if vector[17] != nil {
		if val, ok := vector[17].(float64); ok {
			a.Category = int(val)
		}
	}

	return a
}
