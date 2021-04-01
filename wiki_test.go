package main

import (
	"testing"
)

func TestInterPageLink(t *testing.T) {
	tcs := []struct {
		target []byte
		want   []byte
	}{
		{
			target: []byte("[!InterPageLink]"),
			want:   []byte("<a href=\"/view/InterPageLink\">InterPageLink</a>"),
		},
		{
			target: []byte("[!Inter-page link]"),
			want:   []byte("<a href=\"/view/Inter-page-link\">Inter-page link</a>"),
		},
	}
	for _, tc := range tcs {
		got := interPageLink(tc.target)
		if string(got) != string(tc.want) {
			t.Errorf("\nwant: %s\ngot: %s", tc.want, got)
		}
	}
}
