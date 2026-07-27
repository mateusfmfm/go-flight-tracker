package store

import (
	"fmt"
	"go-flight-tracker/internal/flight"
	"sync"
	"testing"
)

func TestAircraftStoreConcurrentAccess(t *testing.T) {
	st := NewAircraftStore()

	var wg sync.WaitGroup

	//Simulate Poller updating the store continuously (Write)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			aircrafts := []*flight.Aircraft{
				{Icao24: "a123", Callsign: fmt.Sprintf("FLIGHT-%d", i)},
				{Icao24: "b456", Callsign: "AZUL-100"},
			}
			st.Update(aircrafts)
		}
	}()

	//Simulate multiple readers (GetAll and GetByICAO)
	for r := 0; r < 10; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = st.GetAll()
				_, _ = st.GetByICAO("a123")
			}
		}()
	}
	wg.Wait()
	if len(st.GetAll()) == 0 {
		t.Errorf("expected aircrafts in store, got empty")
	}

}
