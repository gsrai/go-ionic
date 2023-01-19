package types

import (
	"reflect"
	"testing"
)

func TestToSliceMethod(t *testing.T) {
	in := OutputCSVRecord{
		Address:  "0x40ec5b33f54e0e8a33a975908c5ba1c14e5bbbdf",
		Pumps:    2,
		SumTotal: 3210.88124,
		Trades:   4,
		Coins:    []string{"CRO", "ALCX"},
	}
	want := []string{"0x40ec5b33f54e0e8a33a975908c5ba1c14e5bbbdf", "2", `[CRO ALCX]`, "4", "3210.88"}

	got := in.ToSlice()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.ToSlice() == %v, want %v", in, got, want)
	}
}
