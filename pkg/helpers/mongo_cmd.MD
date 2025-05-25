MongoDB Help for ajor_app
This guide provides essential MongoDB commands to interact with the ajor_app_db database used by the ajor_app project. These commands are executed in the MongoDB shell (mongosh) after connecting to your Dockerized MongoDB instance (mongo:latest) using:
`docker exec -it ajor-mongo mongosh -u admin -p password`

Prerequisites

MongoDB container is running (docker ps to verify).
You have the correct credentials (admin/password) set in your Docker environment variables.
The ajor_app_db database and collections (users, profiles) are populated via your application’s API.

Accessing the MongoDB Shell

Connect to the MongoDB container:
`docker exec -it ajor-mongo mongosh -u admin -p password`


Switch to the ajor_app_db database:
`use ajor_app_db`



Common Commands
1. List Available Collections
View all collections in the ajor_app_db database (e.g., users, profiles):
`show collections`

Expected Output:
`profiles`
`users`

Note: If no collections appear, no data has been inserted yet. Use your API (e.g., POST /api/register) to add data.
2. View All Data in a Collection
Retrieve all documents in a collection.

Users Collection:
`db.users.find()`

For better readability:
`db.users.find().pretty()`

Example Output:
```{
  "_id": UUID("550e8400-e29b-41d4-a716-446655440000"),
  "email": "gerard@example.com",
  "password": "$2a$10$...",
  "username": "Gerard",
  "phone": "1234567890",
  "verified": false,
  "is_admin": true,
  "created_at": ISODate("2025-05-22T15:00:00Z"),
  "updated_at": ISODate("2025-05-22T15:00:00Z")
}```


Profiles Collection:
`db.profiles.find().pretty()`

Example Output:
```{
  "_id": UUID("550e8400-e29b-41d4-a716-446655440001"),
  "user_id": UUID("550e8400-e29b-41d4-a716-446655440000"),
  "bio": "",
  "location": "",
  "website": "",
  "social_links": [],
  "profile_pic": "",
  "cover_pic": "",
  "created_at": ISODate("2025-05-22T15:00:00Z"),
  "updated_at": ISODate("2025-05-22T15:00:00Z")
}```


3. Count Documents
Check the number of documents in a collection:
db.users.countDocuments()
db.profiles.countDocuments()

Example Output:
3

4. Query Specific Documents
Find documents matching specific criteria.

By Email (Users):
db.users.find({ email: "gerard@example.com" }).pretty()


By User ID (UUID):
db.users.find({ _id: UUID("550e8400-e29b-41d4-a716-446655440000") }).pretty()


By User ID (Profiles):
db.profiles.find({ user_id: UUID("550e8400-e29b-41d4-a716-446655440000") }).pretty()



5. View Database Information
Get statistics about the ajor_app_db database:
db.stats()

Example Output:
{
  "db": "ajor_app_db",
  "collections": 2,
  "objects": 5,
  ...
}

6. List All Databases
View all databases in the MongoDB instance:
show dbs

Note: If ajor_app_db doesn’t appear, no data has been inserted yet.
7. Delete a Document
Remove a specific document from a collection:
db.users.deleteOne({ email: "gerard@example.com" })

Example Output:
`{ "acknowledged": true, "deletedCount": 1 }`

8. Drop a Collection
Delete an entire collection (use with caution):
`db.users.drop()`

Example Output:
true

9. Exit the Shell
Exit the mongosh shell:
exit

Troubleshooting
No Data Appears

Cause: No documents have been inserted.
Solution:
Use your API to insert data:curl -X POST -H "Content-Type: application/json" -d '{"email":"gerard@example.com","password":"password123","username":"Gerard","phone":"1234567890","is_admin":true}' http://localhost:8080/api/register


Retry db.users.find().pretty().



Collection Not Found

Cause: No data has been inserted, so the collection hasn’t been created.
Solution: Insert data via the API or manually:db.users.insertOne({ email: "test@example.com", username: "Test" })



Authentication Error

Cause: Incorrect username or password.
Solution:
Verify credentials match MONGO_INITDB_ROOT_USERNAME and MONGO_INITDB_ROOT_PASSWORD set in your Docker command:docker run -d --name ajor-mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password mongo:latest


Reconnect with correct credentials.



Connection Issues

Cause: MongoDB container is not running.
Solution:
Check container status:docker ps


Restart if needed:docker start ajor-mongo





Additional Tools

MongoDB Compass: A GUI for browsing data.
Connect to: mongodb://admin:password@localhost:27017
Select ajor_app_db and explore users and profiles.


Studio 3T: Another GUI option for advanced querying.

Notes

These commands are specific to the ajor_app_db database with users and profiles collections, as defined in your Go application.
Use these commands to verify data after interacting with your API (e.g., registering users, updating profiles).
For donation group features (e.g., group savings, funding requests), you’ll have additional collections (e.g., groups, contributions). Use similar find() commands to inspect them.

