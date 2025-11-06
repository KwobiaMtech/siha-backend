// MongoDB script to create a verified test user for Stellar testing

// Hash for password "password123" (bcrypt)
const hashedPassword = "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi";

// Create test user
const testUser = {
  email: "stellar_test@example.com",
  password: hashedPassword,
  first_name: "Stellar",
  last_name: "Test",
  phone_number: "+1234567890",
  is_verified: true,
  kyc_status: "approved",
  created_at: new Date(),
  updated_at: new Date()
};

// Remove existing test user if exists
db.users.deleteOne({email: "stellar_test@example.com"});

// Insert new test user
const result = db.users.insertOne(testUser);
print("Test user created with ID:", result.insertedId);

// Verify user was created
const user = db.users.findOne({email: "stellar_test@example.com"});
print("User verification:", user ? "SUCCESS" : "FAILED");
