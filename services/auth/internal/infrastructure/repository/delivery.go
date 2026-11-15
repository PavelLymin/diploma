package repository

import "context"

type OTPDelivery struct {
}

func New() *OTPDelivery {
	return &OTPDelivery{}
}

func (d *OTPDelivery) SendOTP(ctx context.Context, normalizedPhone, code string) error {
	return nil
}
