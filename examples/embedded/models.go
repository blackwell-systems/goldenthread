// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package embedded

// Base contains common fields for all entities.
type Base struct {
	// ID is the unique identifier
	ID string `json:"id" gt:"uuid,required"`

	// CreatedAt is the creation timestamp
	CreatedAt string `json:"created_at" gt:"datetime,required"`
}

// User embeds Base and adds user-specific fields.
type User struct {
	Base

	// Username is the unique handle
	Username string `json:"username" gt:"required,len:3..20"`

	// Email is the primary contact
	Email string `json:"email" gt:"email"`
}

// Product embeds Base and adds product-specific fields.
type Product struct {
	Base

	// Name is the product name
	Name string `json:"name" gt:"required,len:1..100"`

	// Price is the product price
	Price float64 `json:"price" gt:"required,min:0"`
}
