package sms

import "context"

type IOTPDelivery interface {
	SendOTP(
		ctx context.Context,
		phone, code string,
	) error
}
