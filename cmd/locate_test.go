package cmd

import "testing"

func Test_getLatLon(t *testing.T) {
	type args struct {
		rectangle []float64
	}
	tests := []struct {
		name    string
		args    args
		wantLon float64
		wantLat float64
	}{{
		name:    "1",
		args:    args{rectangle: []float64{114.1890407, 35.99351112, 114.5033097, 36.20674544}},
		wantLon: 0,
		wantLat: 0,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLon, gotLat := getLatLon(tt.args.rectangle)
			if gotLon != tt.wantLon {
				t.Errorf("getLatLon() gotLon = %v, want %v", gotLon, tt.wantLon)
			}
			if gotLat != tt.wantLat {
				t.Errorf("getLatLon() gotLat = %v, want %v", gotLat, tt.wantLat)
			}
		})
	}
}
