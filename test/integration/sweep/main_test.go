package main

import "testing"

func TestIsTestObjectName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "sdk-test-tenant", want: true},
		{name: "sdk-key-disk-20260102", want: true},
		{name: "sdk-networking-test-network", want: true},
		{name: "sdk-test.local", want: true},
		{name: "goVergeOS-test-1.txt", want: true},
		{name: "test-view", want: false},
		{name: "test-wg0", want: false},
		{name: "test-peer", want: false},
		{name: "production-sdk-test", want: false},
		{name: "", want: false},
	}
	for _, tc := range cases {
		if got := isTestObjectName(tc.name); got != tc.want {
			t.Errorf("isTestObjectName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCertificateListsTestDomain(t *testing.T) {
	if !certificateListsTestDomain("example.com, sdk-test.local") {
		t.Fatal("expected sdk-test.local in the domain list to match")
	}
	if certificateListsTestDomain("example.com, test-view.local") {
		t.Fatal("expected an unrelated domain list not to match")
	}
}
