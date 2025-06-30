package addresses

// AddressResponse is the JSON representation of a saved customer address.
type AddressResponse struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city"`
	State     string `json:"state,omitempty"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code,omitempty"`
	Phone     string `json:"phone,omitempty"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
}

type CreateAddressRequest struct {
	Label     string `json:"label"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Address1  string `json:"address1" binding:"required"`
	Address2  string `json:"address2"`
	City      string `json:"city" binding:"required"`
	State     string `json:"state"`
	Country   string `json:"country" binding:"required"`
	ZipCode   string `json:"zip_code"`
	Phone     string `json:"phone"`
	IsDefault bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
	Label     string `json:"label"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code"`
	Phone     string `json:"phone"`
}
