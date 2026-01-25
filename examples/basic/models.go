// Copyright 2025 Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package models

// User represents a user account in the system.
type User struct {
	// ID is the unique identifier
	ID string `json:"id" gt:"uuid,required"`

	// Username is the unique handle
	Username string `json:"username" gt:"required,len:3..20"`

	// Email is the primary contact address
	Email string `json:"email" gt:"email"`

	// Age must be at least 13 (COPPA compliance)
	Age int `json:"age" gt:"min:13,max:130"`
}

// Product represents a product in the catalog.
type Product struct {
	SKU   string  `json:"sku" gt:"required,pattern:^[A-Z0-9]{8}$"`
	Name  string  `json:"name" gt:"required,len:1..100"`
	Price float64 `json:"price" gt:"required,min:0"`
}
