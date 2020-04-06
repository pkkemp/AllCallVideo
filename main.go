package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type VideoPageData struct {
	VideoURL string
	PosterURL string
}

func main() {
	keyID := "APKAIJ6FUGQSVTWYI4QQ"
	coverURL := "https://cdn.honeybadgers.tech/video/Honeybadger-Dark.jpg"
	fileKey, err := ioutil.ReadFile("priv.key")
	if err != nil {
		log.Fatal(err)
	}
	rawKey := string(fileKey)

	tmpl := template.Must(template.ParseFiles("video.html"))
	http.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		video := r.URL.Query().Get("video")
		rawURL := "https://cdn.honeybadgers.tech/video/" + video
		data := generateVideoPageData(keyID, rawKey, rawURL, coverURL)
		tmpl.Execute(w, data)
	})
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("./dist"))))
	http.ListenAndServe(":8000", nil)
}

func generateVideoPageData (keyID string, rawKey string, rawURL string, coverURL string) VideoPageData {
	privKey, _ := ParseRsaPrivateKeyFromPemStr(rawKey)
	// Sign URL to be valid for 1 hour from now.
	signer := sign.NewURLSigner(keyID, privKey)
	signedURL, err := signer.Sign(rawURL, time.Now().Add(1*time.Hour))
	signedCoverURL, err := signer.Sign(coverURL, time.Now().Add(1*time.Hour))
	if err != nil {
		log.Fatalf("Failed to sign url, err: %s\n", err.Error())
	}
	data := VideoPageData{
		VideoURL: signedURL,
		PosterURL: signedCoverURL,
	}
	return data
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

