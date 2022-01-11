package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"testing"
)

func Test_createNewID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creating new ids",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createNewID()
		})
	}
}

func Benchmark_createNewID(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	for n := 0; n < b.N; n++ {
		createNewID()
	}
}

func Test_cypher(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := cypher(tt.args.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cypher() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("cypher() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_redirect(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_shorten(t *testing.T) {
	type args struct {
		a []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shorten(tt.args.a); got != tt.want {
				t.Errorf("shorten() = %v, want %v", got, tt.want)
			}
		})
	}
}
