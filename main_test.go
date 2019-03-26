package main

import (
	"fmt"
	"testing"
)

func TestAvailabilityToString(t *testing.T) {
	av1 := Availability{Days: [7]DailyAvailability{{Halfhours: [16]bool{true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, true}}, {Halfhours: [16]bool{true}}}}
	str, err := av1.availabilityToString()
	fmt.Print(str)
	if err != nil || str != "c001:8000:0:0:0:0:0" {
		t.Error(err)
	}
}

func TestStringToAvailability(t *testing.T) {
	av2 := "c001:8000:0:0:0:0:0"
	availExpected := Availability{Days: [7]DailyAvailability{{Halfhours: [16]bool{true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, true}}, {Halfhours: [16]bool{true}}}}

	avail, err := stringToAvailability(av2)
	fmt.Print(avail)
	if err != nil || *avail != availExpected {
		t.Error(err)
	}
}
