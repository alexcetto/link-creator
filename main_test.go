package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/glogr"
)

func init() {
	l = glogr.New()
}

func Test_useAES(t *testing.T) {
	type args struct {
		URL       []byte
		sendingID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "create a url",
			args: args{
				URL:       []byte("https://github.com/ApsisInternational/email-devdb"),
				sendingID: "team-email-new-sendingID1234",
			},
			want: "http://localhost:8067/team-email-new-sendingID1234/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useAES(tt.args.URL, tt.args.sendingID); !strings.Contains(got, tt.want) {

				t.Errorf("useAES() = %v, want %v", got, tt.want)
			} else {
				t.Log(got)
			}
		})
	}
}

var result string

func Benchmark_useAES(b *testing.B) {
	r := ""
	for i := 0; i < b.N; i++ {
		for _, url := range randomURLs() {
			r = useAES(url, "team-email-new-sendingID1234")
		}
		result = r
	}
}

func Benchmark_useEC(b *testing.B) {
	r := ""
	for i := 0; i < b.N; i++ {
		for _, url := range randomURLs() {
			r = useEC(url, "team-email-new-sendingID1234")
		}
		result = r
	}
}

func Test_useAES_statistics(t *testing.T) {
	sID := "sending1234"

	avgFullLenRatio := .0
	avgTransLenRatio := .0
	allURLs := randomURLs()
	totURLs := float64(len(allURLs))

	for _, u := range allURLs {
		originalLen := len(u)
		if originalLen == 0 {
			continue
		}
		r := useAES(u, sID)

		fullLen := len(r)
		avg := float64(fullLen) / float64(originalLen)
		avgFullLenRatio += avg

		transformLen := len(strings.Replace(strings.Replace(r, "http://"+domain, "", -1), sID, "", -1))
		avg = float64(transformLen) / float64(originalLen)
		avgTransLenRatio += avg
	}

	avgFullLenRatio = avgFullLenRatio / totURLs
	avgTransLenRatio = avgTransLenRatio / totURLs
	fmt.Printf("Average length ratio between original and fully transformed %f %% \n", avgFullLenRatio*100)
	fmt.Printf("Average length ratio between original and transformed %f %% \n", avgTransLenRatio*100)
}

func Test_useEC_statistics(t *testing.T) {
	sID := "sending1234"

	avgFullLenRatio := .0
	avgTransLenRatio := .0
	allURLs := randomURLs()
	totURLs := float64(len(allURLs))

	for _, u := range allURLs {
		originalLen := len(u)
		if originalLen == 0 {
			continue
		}
		r := useEC(u, sID)

		fullLen := len(r)
		avg := float64(fullLen) / float64(originalLen)
		avgFullLenRatio += avg

		transformLen := len(strings.Replace(strings.Replace(r, "http://"+domain, "", -1), sID, "", -1))
		avg = float64(transformLen) / float64(originalLen)
		avgTransLenRatio += avg
	}

	avgFullLenRatio = avgFullLenRatio / totURLs
	avgTransLenRatio = avgTransLenRatio / totURLs
	fmt.Printf("Average length ratio between original and fully transformed %f %% \n", avgFullLenRatio*100)
	fmt.Printf("Average length ratio between original and transformed %f %% \n", avgTransLenRatio*100)
}

func randomURLs() [][]byte {
	f, _ := os.Open("urls.txt")
	s, _ := f.Stat()
	data := make([]byte, s.Size())
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	return bytes.Split(data, []byte("\n"))
}

func Test_useRSA(t *testing.T) {
	type args struct {
		originalURL []byte
		sendingID   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should create a new shortened URL",
			args: args{
				originalURL: []byte("https://github.com/ApsisInternational/email-devdb"),
				sendingID:   "sending1234",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useRSA(tt.args.originalURL, tt.args.sendingID); got != tt.want {
				t.Errorf("useRSA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_useEC(t *testing.T) {
	type args struct {
		originalURL []byte
		sendingID   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should create a new url using EC",
			args: args{
				originalURL: []byte("https://github.com/ApsisInternational/email-devdb"),
				sendingID:   "sending1234",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useEC(tt.args.originalURL, tt.args.sendingID); got != tt.want {
				t.Errorf("useEC() = %v, want %v", got, tt.want)
			}
		})
	}
}
