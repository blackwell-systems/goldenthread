// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package ecommerce demonstrates a complete e-commerce API schema with:
// - Nested objects (Address, Payment)
// - Arrays with validation (OrderItems, Tags)
// - Enums (OrderStatus, PaymentMethod)
// - Maps (Metadata)
// - Optional fields (pointers)
// - Embedded structs (Timestamps)
// - Multiple validation rules
package ecommerce

// Timestamps contains common audit fields embedded in all entities.
type Timestamps struct {
	// CreatedAt is when the entity was created
	CreatedAt string `json:"created_at" gt:"datetime,required"`

	// UpdatedAt is when the entity was last modified
	UpdatedAt string `json:"updated_at" gt:"datetime,required"`
}

// Address represents a shipping or billing address.
type Address struct {
	// Street address line 1
	Street string `json:"street" gt:"required,len:1..100"`

	// City name
	City string `json:"city" gt:"required,len:1..50"`

	// State or province code
	State string `json:"state" gt:"required,len:2..50"`

	// Postal/ZIP code
	PostalCode string `json:"postal_code" gt:"required,pattern:^[A-Z0-9]{3,10}$"`

	// ISO 3166-1 alpha-2 country code
	Country string `json:"country" gt:"required,len:2..2"`
}

// PaymentMethod represents accepted payment types.
type PaymentMethod struct {
	// Type of payment method
	Type string `json:"type" gt:"enum:credit_card,debit_card,paypal,apple_pay,google_pay"`

	// Last 4 digits of card (for card payments)
	Last4 *string `json:"last_4" gt:"pattern:^[0-9]{4}$"`

	// Card brand (for card payments)
	Brand *string `json:"brand" gt:"enum:visa,mastercard,amex,discover"`

	// PayPal email (for PayPal payments)
	Email *string `json:"email" gt:"email"`
}

// Customer represents a registered user account.
type Customer struct {
	Timestamps

	// ID is the unique customer identifier
	ID string `json:"id" gt:"uuid,required"`

	// Email address (unique)
	Email string `json:"email" gt:"email,required"`

	// Display name
	Name string `json:"name" gt:"required,len:1..100"`

	// Phone number (E.164 format)
	Phone *string `json:"phone" gt:"pattern:^\\+[1-9]\\d{1,14}$"`

	// Billing address
	BillingAddress Address `json:"billing_address" gt:"required"`

	// Shipping addresses (multiple allowed)
	ShippingAddresses []Address `json:"shipping_addresses" gt:"min:0,max:5"`

	// Saved payment methods
	PaymentMethods []PaymentMethod `json:"payment_methods" gt:"min:0,max:3"`

	// Customer notes (admin only)
	Notes *string `json:"notes" gt:"len:0..500"`

	// Custom metadata (arbitrary key-value pairs)
	Metadata map[string]string `json:"metadata"`
}

// Product represents an item in the catalog.
type Product struct {
	Timestamps

	// ID is the unique product identifier
	ID string `json:"id" gt:"uuid,required"`

	// SKU is the stock keeping unit (unique)
	SKU string `json:"sku" gt:"required,pattern:^[A-Z0-9-]{8,20}$"`

	// Product name
	Name string `json:"name" gt:"required,len:1..200"`

	// Product description
	Description string `json:"description" gt:"required,len:1..2000"`

	// Price in cents (avoid float precision issues)
	PriceCents int `json:"price_cents" gt:"required,min:0,max:10000000"`

	// Currency code (ISO 4217)
	Currency string `json:"currency" gt:"required,len:3..3"`

	// Stock quantity
	StockQuantity int `json:"stock_quantity" gt:"min:0"`

	// Product category
	Category string `json:"category" gt:"required,enum:electronics,clothing,home,toys,books,sports"`

	// Product tags for search
	Tags []string `json:"tags" gt:"min:0,max:10"`

	// Image URLs
	Images []string `json:"images" gt:"min:1,max:10"`

	// Product specifications (key-value pairs)
	Specifications map[string]string `json:"specifications"`

	// Whether product is active
	Active bool `json:"active"`
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	// Product ID
	ProductID string `json:"product_id" gt:"uuid,required"`

	// Product SKU at time of order
	SKU string `json:"sku" gt:"required"`

	// Product name at time of order
	Name string `json:"name" gt:"required"`

	// Quantity ordered
	Quantity int `json:"quantity" gt:"required,min:1,max:100"`

	// Price per unit in cents
	UnitPriceCents int `json:"unit_price_cents" gt:"required,min:0"`

	// Total price for this line (quantity * unit price)
	TotalCents int `json:"total_cents" gt:"required,min:0"`
}

// Order represents a customer purchase.
type Order struct {
	Timestamps

	// ID is the unique order identifier
	ID string `json:"id" gt:"uuid,required"`

	// Order number (human-friendly)
	OrderNumber string `json:"order_number" gt:"required,pattern:^ORD-[0-9]{8}$"`

	// Customer ID
	CustomerID string `json:"customer_id" gt:"uuid,required"`

	// Order status
	Status string `json:"status" gt:"enum:pending,processing,shipped,delivered,cancelled,refunded"`

	// Line items
	Items []OrderItem `json:"items" gt:"min:1,max:50"`

	// Subtotal in cents (sum of all items)
	SubtotalCents int `json:"subtotal_cents" gt:"required,min:0"`

	// Shipping cost in cents
	ShippingCents int `json:"shipping_cents" gt:"required,min:0"`

	// Tax in cents
	TaxCents int `json:"tax_cents" gt:"required,min:0"`

	// Total in cents (subtotal + shipping + tax)
	TotalCents int `json:"total_cents" gt:"required,min:0"`

	// Currency code
	Currency string `json:"currency" gt:"required,len:3..3"`

	// Shipping address
	ShippingAddress Address `json:"shipping_address" gt:"required"`

	// Payment method used
	PaymentMethod PaymentMethod `json:"payment_method" gt:"required"`

	// Tracking number (populated when shipped)
	TrackingNumber *string `json:"tracking_number" gt:"len:10..50"`

	// Estimated delivery date (ISO 8601)
	EstimatedDelivery *string `json:"estimated_delivery" gt:"date"`

	// Customer notes for this order
	CustomerNotes *string `json:"customer_notes" gt:"len:0..500"`

	// Internal admin notes
	AdminNotes *string `json:"admin_notes" gt:"len:0..1000"`
}

// Cart represents a shopping cart (temporary order).
type Cart struct {
	// ID is the unique cart identifier
	ID string `json:"id" gt:"uuid,required"`

	// Customer ID (optional for guest carts)
	CustomerID *string `json:"customer_id" gt:"uuid"`

	// Cart items
	Items []OrderItem `json:"items" gt:"min:0,max:50"`

	// Subtotal in cents
	SubtotalCents int `json:"subtotal_cents" gt:"min:0"`

	// Expires at timestamp
	ExpiresAt string `json:"expires_at" gt:"datetime,required"`
}

// Review represents a product review.
type Review struct {
	Timestamps

	// ID is the unique review identifier
	ID string `json:"id" gt:"uuid,required"`

	// Product ID
	ProductID string `json:"product_id" gt:"uuid,required"`

	// Customer ID
	CustomerID string `json:"customer_id" gt:"uuid,required"`

	// Order ID (verified purchase)
	OrderID string `json:"order_id" gt:"uuid,required"`

	// Star rating (1-5)
	Rating int `json:"rating" gt:"required,min:1,max:5"`

	// Review title
	Title string `json:"title" gt:"required,len:1..100"`

	// Review body
	Body string `json:"body" gt:"required,len:1..2000"`

	// Whether review is verified purchase
	VerifiedPurchase bool `json:"verified_purchase"`

	// Whether review is approved (moderation)
	Approved bool `json:"approved"`

	// Helpful votes count
	HelpfulCount int `json:"helpful_count" gt:"min:0"`
}
