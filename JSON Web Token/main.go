package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
)

const secret = "rahasia"

type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type RegisteredClaims struct {
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	RegisteredClaims
}

func createJWT(userId int64, email string) (string, error) {

	// create the header
	header := Header{
		Algorithm: "HS256",
		Type:      "JWT",
	}

	// encode the header as a JSON string
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	// base64 encode the header
	base64UrlHeader := base64UrlEncode(headerBytes)

	// create claims for jwt
	claims := Claims{
		UserID: userId,
		Email:  email,
		RegisteredClaims: RegisteredClaims{
			Issuer:    "zlfzx",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	// encode the claims as a JSON string
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// base64 encode the claims
	base64UrlClaims := base64UrlEncode(claimsBytes)

	// create signature
	signature := hmac.New(sha256.New, []byte(secret))
	signature.Write([]byte(base64UrlHeader + "." + base64UrlClaims))

	// base64 encode the signature
	base64UrlSignature := base64UrlEncode(signature.Sum(nil))

	// construct the JWT
	token := base64UrlHeader + "." + base64UrlClaims + "." + base64UrlSignature

	return token, nil
}

func base64UrlEncode(data []byte) string {

	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64UrlDecode(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

func verifyJWT(token string) (Claims, error) {
	var header Header
	var claims Claims

	jwt := strings.Split(token, ".")
	if len(jwt) != 3 {
		return claims, errors.New("invalid JWT format!")
	}

	headerEncode, claimsEncode, signatureEncode := jwt[0], jwt[1], jwt[2]

	headerBytes, err := base64UrlDecode(headerEncode)
	if err != nil {
		return claims, err
	}
	if err = json.Unmarshal(headerBytes, &header); err != nil {
		return claims, err
	}

	claimsBytes, err := base64UrlDecode(jwt[1])
	if err != nil {
		return claims, err
	}
	if err = json.Unmarshal(claimsBytes, &claims); err != nil {
		return claims, err
	}

	// create signature
	signature := hmac.New(sha256.New, []byte(secret))
	signature.Write([]byte(headerEncode + "." + claimsEncode))

	expectedSignature := base64UrlEncode(signature.Sum(nil))

	if signatureEncode != expectedSignature {
		return claims, errors.New("invalid signature!")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return claims, errors.New("token expired!")
	}

	return claims, nil
}

func main() {
	// token, _ := createJWT(10, "user@email.com")

	// fmt.Println(token)
	// fmt.Println("")

	// claims, _ := verifyJWT(token)
	// fmt.Println(claims)

	var option, token string
	flag.StringVar(&option, "option", "", "Options")
	flag.StringVar(&token, "token", "", "JWT token")
	flag.Parse()

	if strings.Contains(option, "create") {
		token, err := createJWT(10, "user@email.com")
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(token)
	} else if strings.Contains(option, "verify") {
		if token != "" {
			claims, err := verifyJWT(token)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(claims)
		} else {
			fmt.Println("needs an arguments: -token")
		}
	}
}
