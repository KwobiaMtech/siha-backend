// MongoDB initialization script for wallets
// Run with: mongosh healthypay init_wallets.js

// Create default wallets for existing users
db.users.find().forEach(function(user) {
    // Create wallet if doesn't exist
    const existingWallet = db.wallets.findOne({user_id: user._id});
    if (!existingWallet) {
        db.wallets.insertOne({
            user_id: user._id,
            balance: 2540.50,
            currency: "GHS",
            created_at: new Date(),
            updated_at: new Date()
        });
        print("Created wallet for user: " + user.email);
    }
    
    // Create mobile money wallet if doesn't exist
    const existingMobile = db.mobile_money_wallets.findOne({user_id: user._id});
    if (!existingMobile) {
        db.mobile_money_wallets.insertOne({
            user_id: user._id,
            provider: "MTN",
            phone_number: "0244123456",
            balance: 850.00,
            is_active: true,
            created_at: new Date(),
            updated_at: new Date()
        });
        print("Created mobile wallet for user: " + user.email);
    }
});

print("Wallet initialization completed!");
