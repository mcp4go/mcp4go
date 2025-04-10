package main

import (
	"context"
	"testing"
)

func Test_simpleTime_getTimezoneTime(t *testing.T) {
	type args struct {
		in0      context.Context
		timeZone string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{
				in0:      context.Background(),
				timeZone: "",
			},
		},
		{
			name: "2",
			args: args{
				in0:      context.Background(),
				timeZone: "America/Los_Angeles",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &simpleTime{}
			got := x.getTimezoneTime(tt.args.in0, tt.args.timeZone)
			t.Logf("got: %v", got)
		})
	}
}
