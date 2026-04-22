package dto

type CreateGeneratorRequest struct {
	ExternalID string   `json:"external_id"`
	Name       string   `json:"name" binding:"required"`
	CNPJ       string   `json:"cnpj"`
	Address    string   `json:"address"`
	Zipcode    string   `json:"zipcode"`
	CityID     *uint    `json:"city_id"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
}

type UpdateGeneratorRequest struct {
	Name      *string  `json:"name"`
	CNPJ      *string  `json:"cnpj"`
	Address   *string  `json:"address"`
	Zipcode   *string  `json:"zipcode"`
	CityID    *uint    `json:"city_id"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Active    *bool    `json:"active"`
}

type CreateReceiverRequest struct {
	ExternalID    string   `json:"external_id"`
	Name          string   `json:"name" binding:"required"`
	CNPJ          string   `json:"cnpj"`
	Address       string   `json:"address"`
	Zipcode       string   `json:"zipcode"`
	CityID        *uint    `json:"city_id"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	LicenseNumber string   `json:"license_number"`
	LicenseExpiry *string  `json:"license_expiry"`
}

type UpdateReceiverRequest struct {
	Name          *string  `json:"name"`
	CNPJ          *string  `json:"cnpj"`
	Address       *string  `json:"address"`
	Zipcode       *string  `json:"zipcode"`
	CityID        *uint    `json:"city_id"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	LicenseNumber *string  `json:"license_number"`
	LicenseExpiry *string  `json:"license_expiry"`
	Active        *bool    `json:"active"`
}
