# E-Commerce API Example

This example demonstrates a complete e-commerce API schema using goldenthread to maintain type safety between Go backend and TypeScript frontend.

## Scenario

An e-commerce platform with:
- Customer accounts with multiple addresses and payment methods
- Product catalog with categories, tags, and specifications
- Shopping cart system with expiration
- Order processing with status tracking
- Product review system with moderation

## What This Example Demonstrates

### Complex Types
- **Nested objects**: `Address`, `PaymentMethod` nested in `Customer` and `Order`
- **Arrays with constraints**: `Items []OrderItem` with min:1, max:50
- **Maps**: `Specifications map[string]string` for flexible key-value data
- **Enums**: `OrderStatus`, `PaymentMethod.Type`, `Product.Category`
- **Optional fields**: `TrackingNumber *string`, `CustomerNotes *string`

### Real-World Validation
- **Business constraints**: Age min:13 (COPPA compliance), max price, max cart items
- **Format validators**: UUID for IDs, email for contacts, datetime for timestamps
- **Regex patterns**: SKU format, phone numbers (E.164), order numbers
- **Relational constraints**: Order contains CustomerID (foreign key pattern)

### Advanced Features
- **Embedded structs**: `Timestamps` fields flattened into `Customer`, `Order`, `Product`, `Review`
- **Collision prevention**: Multiple structs embed `Timestamps` without conflict
- **Type references**: `ShippingAddress Address` generates correct TypeScript reference
- **Array element types**: `[]OrderItem` generates `z.array(OrderItemSchema)`

## Generate Schemas

```bash
# From examples/ecommerce directory
goldenthread generate .

# Or from repository root
goldenthread generate ./examples/ecommerce --out ./examples/ecommerce/gen
```

## Generated Output

Creates 7 TypeScript files:

```
gen/
├── timestamps.ts       # Embedded timestamp fields
├── address.ts          # Shipping/billing addresses
├── payment-method.ts   # Payment types
├── customer.ts         # Customer accounts (uses Address, PaymentMethod)
├── product.ts          # Product catalog
├── order-item.ts       # Order line items
├── order.ts            # Orders (uses OrderItem, Address, PaymentMethod)
├── cart.ts             # Shopping carts
├── review.ts           # Product reviews
└── .goldenthread.json  # Drift detection metadata
```

## Example Usage in TypeScript

```typescript
import { OrderSchema, Order } from './gen/order'
import { CustomerSchema } from './gen/customer'

// Validate API response
const response = await fetch('/api/orders/123')
const data = await response.json()

const result = OrderSchema.safeParse(data)

if (!result.success) {
  // Type-safe error handling
  console.error('Validation failed:', result.error.issues)
  return
}

// Type-safe access
const order: Order = result.data
console.log(order.order_number)        // Type: string
console.log(order.items[0].quantity)   // Type: number
console.log(order.tracking_number)     // Type: string | undefined

// Validate request body
async function createOrder(data: unknown) {
  const result = OrderSchema.safeParse(data)
  
  if (!result.success) {
    return { error: result.error }
  }
  
  // result.data is now type Order with runtime validation guarantee
  await saveOrder(result.data)
}
```

## Schema Relationships

```
Customer
├── BillingAddress: Address
├── ShippingAddresses: []Address
└── PaymentMethods: []PaymentMethod

Order
├── Items: []OrderItem
├── ShippingAddress: Address
└── PaymentMethod: PaymentMethod

Product
├── Tags: []string
├── Images: []string
└── Specifications: map[string]string

Cart
└── Items: []OrderItem

Review
└── (references Product, Customer, Order via IDs)
```

## Validation Highlights

### Price Handling

All prices use **cents** (integers) to avoid floating-point precision issues:

```go
// Go
PriceCents int `json:"price_cents" gt:"required,min:0,max:10000000"`

// TypeScript
price_cents: z.number().int().min(0).max(10000000)
```

### Enum Constraints

Status fields use enums for type safety:

```go
// Go
Status string `json:"status" gt:"enum:pending,processing,shipped,delivered,cancelled,refunded"`

// TypeScript
status: z.enum(['pending', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded'])
```

### Optional Fields with Validation

Optional fields still have constraints:

```go
// Go
TrackingNumber *string `json:"tracking_number" gt:"len:10..50"`

// TypeScript
tracking_number: z.string().min(10).max(50).optional()
```

### Array Constraints

Prevent abuse with reasonable limits:

```go
// Go
Items []OrderItem `json:"items" gt:"min:1,max:50"`

// TypeScript
items: z.array(OrderItemSchema).min(1).max(50)
```

### Embedded Timestamps

All entities automatically get `created_at` and `updated_at`:

```go
// Go
type Customer struct {
    Timestamps  // Fields flattened
    ID string `json:"id" gt:"uuid,required"`
    // ...
}

// TypeScript (flattened)
const CustomerSchema = z.object({
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  id: z.string().uuid(),
  // ...
})
```

## Drift Detection in CI

Add to your CI pipeline:

```yaml
- name: Generate schemas
  run: goldenthread generate ./models --out ./frontend/src/schemas

- name: Check for drift
  run: goldenthread check ./models

- name: Commit generated schemas
  if: failure()
  run: |
    echo "::error::Generated schemas are out of sync with Go structs"
    echo "Run: goldenthread generate ./models"
    exit 1
```

This ensures frontend schemas stay synchronized with backend changes.

## Key Takeaways

This example shows goldenthread handling a real-world API with:
- 7 interconnected domain models
- 60+ validated fields
- Nested relationships
- Multiple data types (primitives, enums, arrays, maps, objects)
- Complex validation rules (patterns, formats, bounds)
- Embedded struct composition

All generated from Go struct tags. Zero manual synchronization. Type-safe end-to-end.
