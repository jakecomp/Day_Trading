db = db.getSiblingDB("admin") 
db.auth("admin","admin") 
db = db.getSiblingDB("day_trading") 

db.createUser({ 
    user: "user", 
    pwd: "user", 
    roles: [ 
        { 
            role: "readWrite", 
            db: "day_trading"
        }
    ]
}); 

db.createCollection("users")