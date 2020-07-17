package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type VideoPageData struct {
	VideoURL string
	PosterURL string
}

func handler(w http.ResponseWriter, r *http.Request) {
	keyID := "APKAIJ6FUGQSVTWYI4QQ"
	fileKey, err := ioutil.ReadFile("priv.key")
	if err != nil {
		log.Fatal(err)
	}
	rawKey := string(fileKey)
	fmt.Fprintf(w, "Hello World!")
	privKey, _ := ParseRsaPrivateKeyFromPemStr(rawKey)
	s := sign.NewCookieSigner(keyID, privKey)
	//// Get Signed cookies for a resource that will expire in 1 hour
	//cookies, err := s.Sign("*", time.Now().Add(1 * time.Hour))
	//if err != nil {
	//	fmt.Println("failed to create signed cookies", err)
	//	return
	//}

	// Or get Signed cookies for a resource that will expire in 1 hour
	// and set path and domain of cookies
	cookies, err := s.Sign("*", time.Now().Add(1 * time.Hour), func(o *sign.CookieOptions) {
		o.Path = "/video"
		o.Domain = "cdn.honeybadgers.tech"
	})
	if err != nil {
		fmt.Println("failed to create signed cookies", err)
		return
	}

	// Server Response via http.ResponseWriter
	for _, c := range cookies {
		http.SetCookie(w, c)
	}
	http.FileServer(http.Dir("/")).ServeHTTP(w, r)
}

func main() {


	//
	//
	//tmpl := template.Must(template.ParseFiles("video.html"))
	//http.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
	//	video := r.URL.Query().Get("video")
	//	rawURL := "https://cdn.honeybadgers.tech/video/" + video
	//	data := generateVideoPageData(keyID, rawKey, rawURL, coverURL)
	//	tmpl.Execute(w, data)
	//})
	//http.HandleFunc("/getVideo", func(w http.ResponseWriter, r *http.Request) {
	//	video := r.URL.Query().Get("video")
	//	rawURL := "https://cdn.honeybadgers.tech/video/" + video
	//	data := generateVideoPageData(keyID, rawKey, rawURL, coverURL)
	//
	//	js, err := json.Marshal(data)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//
	//	w.Header().Set("Content-Type", "application/json")
	//	w.Write(js)
	//})



	//// Client request via the cookie jar
	//if client.CookieJar != nil {
	//	for _, c := range cookies {
	//		client.Cookie(w, c)
	//	}
	//}


	http.HandleFunc("/", handler)

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

