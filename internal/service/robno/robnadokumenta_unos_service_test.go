package robno

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"helia/internal/common"
)

// TestRobnoUnosPostojiID checks the optional key of a šifarnik (the org. jedinica and the mesto troška of the robna
// dokumenta): no key (empty or "-") is no error, a given key must be a whole number and exist.
func TestRobnoUnosPostojiID(t *testing.T) {
	postoji := func(_ context.Context, id int64) (bool, error) {
		switch id {
		case 5:
			return true, nil
		case 9:
			return false, errors.New("db down")
		}
		return false, nil
	}
	for _, tt := range []struct {
		name  string
		value string
		want  []string
	}{
		{name: "empty", value: "", want: nil},
		{name: "dash", value: " - ", want: nil},
		{name: "existing key", value: "5", want: nil},
		{name: "missing key", value: "7", want: []string{robnoUnosMsgNePostoji}},
		{name: "not a number", value: "abc", want: []string{robnoUnosMsgNePostoji}},
		{name: "read error", value: "9", want: []string{common.ErrMsgReadData}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			(&RobnoDokumentaResource{}).robnoUnosPostojiID(context.Background(), tt.value, "idorgjed", postoji, func(field, message string) {
				if field != "idorgjed" {
					t.Errorf("field = %q, want idorgjed", field)
				}
				got = append(got, message)
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("messages = %q, want %q", got, tt.want)
			}
		})
	}
}
