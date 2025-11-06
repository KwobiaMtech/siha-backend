// MongoDB script to create a verified test user
use healthypay

// Delete existing test user
db.users.deleteOne({email: "test@example.com"})

// Create new verified test user
db.users.insertOne({
  email: "test@example.com",
  password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password123
  first_name: "Test",
  last_name: "User",
  pin: "", // No PIN set
  payment_method: "", // No payment method set
  is_verified: true,
  kyc_status: "pending",
  created_at: new Date(),
  updated_at: new Date()
})

print("Test user created successfully")
