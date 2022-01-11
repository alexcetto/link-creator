package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/lytics/base62"
)

var (
	activityStore = make(map[string][]byte)

	// Key created for each sending
	mySuperSecretKey = ""
	// Apsis redirection domain
	domain = "localhost:8067"

	l logr.Logger
)

var glogInit = sync.Once{}

func initGlog() {
	glogInit.Do(func() {
		_ = flag.Set("v", "1")
		_ = flag.Set("logtostderr", "true")
		flag.Parse()
	})
	os.Stderr = os.Stdout
}

func main() {
	initGlog()
	l = glogr.New()
	mySuperSecretKey = os.Getenv("MY_SECRET")
	r := mux.NewRouter()
	r.Path("/{sendingID}/{shortenedURL}").HandlerFunc(redirect)
	r.Path("/{sendingID}").HandlerFunc(createNewID)
	fmt.Print(http.ListenAndServe(domain, r))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sendingID := vars["sendingID"]
	shortenedURL := vars["shortenedURL"]
	l.Info("Path", "sendingId", sendingID, "shortenedURL", shortenedURL, "value", string(activityStore[sendingID]))

	lengthened := lengthen(shortenedURL)
	l.Info("lenghtened url", "lengthened", string(lengthened))
	dec := decypher(lengthened, activityStore[sendingID])
	l.Info("deciphered", "url", dec)

	http.Redirect(w, r, dec, http.StatusTemporaryRedirect)
}

type LinkReq struct {
	URL string `json:"url"`
}

func createNewID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sendingID := vars["sendingID"]
	l.Info("new request", "sendingId", sendingID)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Error(err, "reading broke")
		return
	}
	var link LinkReq
	if err := json.Unmarshal(data, &link); err != nil {
		l.Error(err, "reading broke")
		return
	}
	l.Info("MyURL no transfo", "url", link.URL)

	enc, nonce := cypher(link.URL)
	l.Info("MyURL ciphered", "ciphered", string(enc))
	short := shorten(enc)
	l.Info("MyURL shortened", "short", short)

	finalURL := fmt.Sprintf("http://" + domain + "/" + sendingID + "/" + short)
	l.Info("FinalURL ", "value", finalURL)
	activityStore[sendingID] = nonce

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(finalURL))
	return

}

func shorten(a []byte) string {
	return base62.StdEncoding.EncodeToString(a)
}

func lengthen(a string) []byte {
	byteUrl, _ := base62.StdEncoding.DecodeString(a)
	return byteUrl
}

func cypher(s string) ([]byte, []byte) {
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	key, _ := hex.DecodeString(mySuperSecretKey)
	plaintext := []byte(s)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce
}

func decypher(c []byte, snonce []byte) string {
	key, _ := hex.DecodeString(mySuperSecretKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, snonce, c, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}
