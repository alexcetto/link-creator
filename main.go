package main

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	ecies "github.com/ecies/go/v2"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/lytics/base62"
)

var (
	activityStore = make(map[string][]byte)

	// Apsis redirection domain
	domain = "localhost:8067"

	mode string

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

	mode = os.Getenv("MODE")

	r := mux.NewRouter()
	r.Path("/{sendingID}/{shortenedURL}").HandlerFunc(redirect)
	r.Path("/{sendingID}").HandlerFunc(createNewID)
	fmt.Print(http.ListenAndServe(domain, r))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sendingID := vars["sendingID"]
	shortenedURL := vars["shortenedURL"]
	l.Info("Path", "sendingId", sendingID, "shortenedURL", shortenedURL, "keyvalue", string(activityStore[sendingID]))

	var dec string
	if mode == "aes" {
		dec = aesdecypher(shortenedURL, activityStore[sendingID])
	} else if mode == "rsa" {
		dec = rsaDecipher(shortenedURL, rsa.PrivateKey{})
	} else if mode == "ec" {
		dec = ecDecipher(shortenedURL, activityStore[sendingID])
	} else {
		panic("chose a mode")
	}

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

	var finalURL string
	if mode == "aes" {
		finalURL = useAES([]byte(link.URL), sendingID)
	} else if mode == "rsa" {
		finalURL = useRSA([]byte(link.URL), sendingID)
	} else if mode == "ec" {
		finalURL = useEC([]byte(link.URL), sendingID)
	} else {
		panic("chose a mode")
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(finalURL))
	return
}

func useAES(URL []byte, sendingID string) string {
	// generate a new key for this sending
	secretSendingKey := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, secretSendingKey); err != nil {
		panic(err.Error())
	}
	enc := aescypher(URL, secretSendingKey)
	l.Info("MyURL ciphered", "ciphered", enc)

	finalURL := fmt.Sprintf("http://" + domain + "/" + sendingID + "/" + enc)
	l.Info("FinalURL ", "value", finalURL)

	// store the key to decrypt the URL later
	activityStore[sendingID] = secretSendingKey
	return finalURL
}

func aescypher(plaintext, key []byte) string {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return base62.StdEncoding.EncodeToString(cipherText)
}

func aesdecypher(message string, key []byte) string {
	cipherText, err := base62.StdEncoding.DecodeString(message)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("invalid ciphertext block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText)
}

func useEC(originalURL []byte, sendingID string) string {
	// The GenerateKey method takes in a reader that returns random bits, and
	// the number of bits
	privateKey, err := ecies.GenerateKey()
	if err != nil {
		panic(err)
	}

	ciphertext, err := ecies.Encrypt(privateKey.PublicKey, originalURL)
	if err != nil {
		panic(err)
	}
	l.Info("plaintext encrypted ", "value", string(ciphertext))

	enc := base62.StdEncoding.EncodeToString(ciphertext)

	finalURL := fmt.Sprintf("http://" + domain + "/" + sendingID + "/" + enc)
	l.Info("FinalURL ", "value", finalURL)
	// store the key to decrypt the URL later
	activityStore[sendingID] = privateKey.Bytes()
	return finalURL
}

func ecDecipher(msg string, key []byte) string {
	cy, err := base62.StdEncoding.DecodeString(msg)
	if err != nil {
		panic(err)
	}
	priv := ecies.NewPrivateKeyFromBytes(key)
	plaintext, err := ecies.Decrypt(priv, cy)
	if err != nil {
		panic(err)
	}
	return string(plaintext)
}

func useRSA(originalURL []byte, sendingID string) string {
	// The GenerateKey method takes in a reader that returns random bits, and
	// the number of bits
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// The public key is a part of the *rsa.PrivateKey struct
	publicKey := privateKey.PublicKey

	enc := rsaCipher(originalURL, publicKey)

	finalURL := fmt.Sprintf("http://" + domain + "/" + sendingID + "/" + enc)
	l.Info("FinalURL ", "value", finalURL)
	// store the key to decrypt the URL later
	//activityStore[sendingID] = privateKey
	return finalURL
}

func rsaCipher(msg []byte, pubKey rsa.PublicKey) string {
	enc, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&pubKey,
		msg,
		nil,
	)
	if err != nil {
		panic(err)
	}

	return base62.StdEncoding.EncodeToString(enc)
}

func rsaDecipher(msg string, key rsa.PrivateKey) string {
	b, err := base62.StdEncoding.DecodeString(msg)
	if err != nil {
		panic(err)
	}
	decryptedBytes, err := key.Decrypt(nil, b, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		panic(err)
	}

	return string(decryptedBytes)
}
