package models

import (
	"time"
)

type Contact struct {
	ID          string    `json:"id" db:"id"`
	FirstName   string    `json:"firstName" db:"first_name"`
	LastName    string    `json:"lastName" db:"last_name"`
	Email       string    `json:"email" db:"email"`
	PhoneNumber *string   `json:"phoneNumber,omitempty" db:"phone_number"`
	Address     *Address  `json:"address,omitempty"`
	Company     *string   `json:"company,omitempty" db:"company"`
	JobTitle    *string   `json:"jobTitle,omitempty" db:"job_title"`
	Tags        []string  `json:"tags"`
	Notes       *string   `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Address struct {
	Street     *string `json:"street,omitempty"`
	City       *string `json:"city,omitempty"`
	State      *string `json:"state,omitempty"`
	PostalCode *string `json:"postalCode,omitempty"`
	Country    *string `json:"country,omitempty"`
}

type ContactInput struct {
	FirstName   string   `json:"firstName" binding:"required,min=1,max=50"`
	LastName    string   `json:"lastName" binding:"required,min=1,max=50"`
	Email       string   `json:"email" binding:"required,email"`
	PhoneNumber *string  `json:"phoneNumber,omitempty"`
	Address     *Address `json:"address,omitempty"`
	Company     *string  `json:"company,omitempty"`
	JobTitle    *string  `json:"jobTitle,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

type ListOptions struct {
	Page   int
	Limit  int
	Search string
	Tag    string
}
