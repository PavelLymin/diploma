package jwtmanager

import "github.com/golang-jwt/jwt/v5"

type claims struct {
	TokenUse string `json:"token_use"`
	FamilyID string `json:"sid"`
	jwt.RegisteredClaims
}
