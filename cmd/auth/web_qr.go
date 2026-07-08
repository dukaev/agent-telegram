package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"time"

	"github.com/gotd/td/telegram/auth/qrlogin"
	"rsc.io/qr"
)

func qrRefreshDelay(expiresAt time.Time) int {
	if expiresAt.IsZero() {
		return 1
	}
	seconds := int(time.Until(expiresAt).Seconds())
	if seconds <= 0 {
		return 1
	}
	if seconds > 3 {
		return 3
	}
	return seconds + 1
}

func qrCodeImage(tokenURL string) (string, error) {
	if _, err := qrlogin.ParseTokenURL(tokenURL); err != nil {
		return "", fmt.Errorf("parse QR token URL: %w", err)
	}
	code, err := qr.Encode(tokenURL, qr.M)
	if err != nil {
		return "", fmt.Errorf("render QR image: %w", err)
	}

	const (
		quietZone = 4
		scale     = 12
	)
	side := (code.Size + quietZone*2) * scale
	img := image.NewGray(image.Rect(0, 0, side, side))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}

	black := color.Gray{Y: 0x00}
	for y := range code.Size {
		for x := range code.Size {
			if !code.Black(x, y) {
				continue
			}
			startX := (x + quietZone) * scale
			startY := (y + quietZone) * scale
			for yy := startY; yy < startY+scale; yy++ {
				for xx := startX; xx < startX+scale; xx++ {
					img.SetGray(xx, yy, black)
				}
			}
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return "", fmt.Errorf("encode QR image: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes()), nil
}
