package main

import (
	"reflect"
	"testing"
)

func Test_lookupFullTZ(t *testing.T) {
	type args struct {
		abbrev    string
		utcOffset int
	}
	tests := []struct {
		name    string
		args    args
		want    TZLabel
		wantErr bool
	}{
		{"double abbrev off by 1 hour", args{"CDT", -6}, TZLabel{"CDT", "Central Daylight Time (North America)", -5}, false},
		{"double abbrev", args{"CDT", -5}, TZLabel{"CDT", "Central Daylight Time (North America)", -5}, false},
		{"double abbrev cuba", args{"CDT", -4}, TZLabel{"CDT", "Cuba Daylight Time", -4}, false},
		{"pst filipino", args{"PST", 8}, TZLabel{"PST", "Philippine Standard Time", 8}, false},
		{"pst pacific", args{"PST", -8}, TZLabel{"PST", "Pacific Standard Time (North America)", -8}, false},
		{"invalid abbrev", args{"LMSKTK", -5}, TZLabel{}, true},
		{"invalid abbrev 2", args{"", -5}, TZLabel{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lookupFullTZ(tt.args.abbrev, tt.args.utcOffset)
			if (err != nil) != tt.wantErr {
				t.Errorf("lookupFullTZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lookupFullTZ() got = %v, want %v", got, tt.want)
			}
		})
	}
}
