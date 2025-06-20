package main

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"strings"
// 	"testing"
// 	"time"
// 	"github.com/Gerard-007/ajor_app/internal/models"
// 	"github.com/Gerard-007/ajor_app/internal/routes"
// 	"github.com/Gerard-007/ajor_app/pkg/payment"
// 	"github.com/Gerard-007/ajor_app/pkg/utils"
// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"golang.org/x/crypto/bcrypt"
// )

// // MockMongoCollection is a mock for mongo.Collection
// type MockMongoCollection struct {
// 	mock.Mock
// }

// func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
// 	args := m.Called(ctx, filter)
// 	result := &mongo.SingleResult{}
// 	if args.Get(0) != nil {
// 		result = args.Get(0).(*mongo.SingleResult)
// 	}
// 	if args.Error(1) != nil {
// 		result.Err = args.Error(1)
// 	}
// 	return result
// }

// func (m *MockMongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
// 	args := m.Called(ctx, document)
// 	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
// }

// func (m *MockMongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*mongo.UpdateOptions) (*mongo.UpdateResult, error) {
// 	args := m.Called(ctx, filter, update, opts)
// 	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
// }

// func (m *MockMongoCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
// 	args := m.Called(ctx, filter)
// 	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
// }

// func (m *MockMongoCollection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
// 	args := m.Called(ctx, filter)
// 	return args.Get(0).(int64), args.Error(1)
// }

// // MockMongoDatabase is a mock for mongo.Database
// type MockMongoDatabase struct {
// 	mock.Mock
// }

// func (m *MockMongoDatabase) Collection(name string) *mongo.Collection {
// 	args := m.Called(name)
// 	return args.Get(0).(*mongo.Collection)
// }

// // MockPaymentGateway is a mock for payment.PaymentGateway
// type MockPaymentGateway struct {
// 	mock.Mock
// }

// func (m *MockPaymentGateway) CreateVirtualAccount(ctx context.Context, ownerID primitive.ObjectID, email, phone, narration string, isPermanent bool, bvn string, amount float64) (*payment.VirtualAccount, error) {
// 	args := m.Called(ctx, ownerID, email, phone, narration, isPermanent, bvn, amount)
// 	return args.Get(0).(*payment.VirtualAccount), args.Error(1)
// }

// func (m *MockPaymentGateway) GetVirtualAccount(ctx context.Context, accountID string) (*payment.VirtualAccount, error) {
// 	args := m.Called(ctx, accountID)
// 	return args.Get(0).(*payment.VirtualAccount), args.Error(1)
// }

// func (m *MockPaymentGateway) DeactivateVirtualAccount(ctx context.Context, accountID string) error {
// 	args := m.Called(ctx, accountID)
// 	return args.Error(0)
// }

// func (m *MockPaymentGateway) FundVirtualAccount(ctx context.Context, accountID string, req payment.FundingRequest) (*payment.TransactionResponse, error) {
// 	args := m.Called(ctx, accountID, req)
// 	return args.Get(0).(*payment.TransactionResponse), args.Error(1)
// }

// func (m *MockPaymentGateway) VerifyTransaction(ctx context.Context, transactionID string) (*payment.TransactionResponse, error) {
// 	args := m.Called(ctx, transactionID)
// 	return args.Get(0).(*payment.TransactionResponse), args.Error(1)
// }

// func setupRouter(db *mongo.Database, pg payment.PaymentGateway) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	router := gin.New()
// 	routes.InitRoutes(router, db, pg)
// 	return router
// }

// func generateTestToken(t *testing.T, username, email string, userID primitive.ObjectID, isAdmin bool) string {
// 	os.Setenv("JWT_SECRET", "test-secret-key-1234567890123456")
// 	token, err := utils.GenerateToken(username, email, userID, isAdmin)
// 	assert.NoError(t, err)
// 	return token
// }

// func TestRoutes(t *testing.T) {
// 	// Setup mock database and payment gateway
// 	mockDB := new(MockMongoDatabase)
// 	mockUsersCollection := new(MockMongoCollection)
// 	mockProfilesCollection := new(MockMongoCollection)
// 	mockWalletsCollection := new(MockMongoCollection)
// 	mockBlacklistCollection := new(MockMongoCollection)
// 	mockTransactionsCollection := new(MockMongoCollection)
// 	mockPG := new(MockPaymentGateway)

// 	mockDB.On("Collection", "users").Return(mockUsersCollection)
// 	mockDB.On("Collection", "profiles").Return(mockProfilesCollection)
// 	mockDB.On("Collection", "wallets").Return(mockWalletsCollection)
// 	mockDB.On("Collection", "blacklisted_tokens").Return(mockBlacklistCollection)
// 	mockDB.On("Collection", "transactions").Return(mockTransactionsCollection)

// 	// Setup router
// 	router := setupRouter(mockDB, mockPG)

// 	// Test user data
// 	userID := primitive.NewObjectID()
// 	adminID := primitive.NewObjectID()
// 	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securepassword123"), bcrypt.DefaultCost)
// 	testUser := models.User{
// 		ID:        userID,
// 		Username:  "user1",
// 		Email:     "user1@example.com",
// 		Password:  string(hashedPassword),
// 		Phone:     "+2348062134747",
// 		BVN:       "11234567897",
// 		IsAdmin:   false,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}
// 	testProfile := models.Profile{
// 		ID:         primitive.NewObjectID(),
// 		UserID:     userID,
// 		Bio:        "",
// 		Location:   "",
// 		ProfilePic: "",
// 		CreatedAt:  time.Now(),
// 		UpdatedAt:  time.Now(),
// 	}
// 	testWallet := models.Wallet{
// 		ID:                   primitive.NewObjectID(),
// 		OwnerID:             userID,
// 		Type:                models.WalletTypeUser,
// 		Balance:             0.0,
// 		VirtualAccountID:    "VA123",
// 		VirtualAccountNumber: "1234567890",
// 		VirtualBankName:     "Test Bank",
// 		CreatedAt:           time.Now(),
// 		UpdatedAt:           time.Now(),
// 	}
// 	testVirtualAccount := &payment.VirtualAccount{
// 		AccountNumber: "1234567890",
// 		AccountID:     "VA123",
// 		BankName:      "Test Bank",
// 	}

// 	// Generate tokens
// 	userToken := generateTestToken(t, "user1", "user1@example.com", userID, false)
// 	adminToken := generateTestToken(t, "admin", "admin@example.com", adminID, true)

// 	// 1. Test POST /register
// 	t.Run("Register_Success", func(t *testing.T) {
// 		payload := `{"email":"user1@example.com","password":"securepassword123","phone":"+2348062134747","bvn":"11234567897"}`
// 		req, _ := http.NewRequest("POST", "/register", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		w := httptest.NewRecorder()

// 		// Mock database responses
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"email": "user1@example.com"}).Return(nil, mongo.ErrNoDocuments).Once()
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"username": "user1"}).Return(nil, mongo.ErrNoDocuments).Once()
// 		mockUsersCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: userID}, nil).Once()
// 		mockProfilesCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: testProfile.ID}, nil).Once()
// 		mockWalletsCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: testWallet.ID}, nil).Once()
// 		mockPG.On("CreateVirtualAccount", mock.Anything, userID, "user1@example.com", "+2348062134747", "Wallet for user1", true, "11234567897", 0.0).Return(testVirtualAccount, nil).Once()
// 		mockWalletsCollection.On("UpdateOne", mock.Anything, bson.M{"_id": testWallet.ID}, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusCreated, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "User registered successfully", response["message"])
// 	})

// 	t.Run("Register_DuplicateEmail", func(t *testing.T) {
// 		payload := `{"email":"user1@example.com","password":"securepassword123","phone":"+2348062134747","bvn":"11234567897"}`
// 		req, _ := http.NewRequest("POST", "/register", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		w := httptest.NewRecorder()

// 		// Mock duplicate email
// 		singleResult := &mongo.SingleResult{}
// 		singleResult.Err = nil
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"email": "user1@example.com"}).Return(singleResult, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "email already exists", response["error"])
// 	})

// 	// 2. Test POST /login
// 	t.Run("Login_Success", func(t *testing.T) {
// 		payload := `{"email":"user1@example.com","password":"securepassword123"}`
// 		req, _ := http.NewRequest("POST", "/login", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		w := httptest.NewRecorder()

// 		// Mock user lookup
// 		singleResult := &mongo.SingleResult{}
// 		singleResult.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.User) = testUser
// 			return nil
// 		}
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"email": "user1@example.com"}).Return(singleResult, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.NotEmpty(t, response["token"])
// 	})

// 	t.Run("Login_InvalidCredentials", func(t *testing.T) {
// 		payload := `{"email":"user1@example.com","password":"wrongpassword"}`
// 		req, _ := http.NewRequest("POST", "/login", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		w := httptest.NewRecorder()

// 		// Mock user lookup
// 		singleResult := &mongo.SingleResult{}
// 		singleResult.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.User) = testUser
// 			return nil
// 		}
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"email": "user1@example.com"}).Return(singleResult, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusUnauthorized, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "Invalid credentials", response["error"])
// 	})

// 	// 3. Test POST /logout
// 	t.Run("Logout_Success", func(t *testing.T) {
// 		req, _ := http.NewRequest("POST", "/logout", nil)
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock blacklist
// 		mockBlacklistCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "Logged out successfully", response["message"])
// 	})

// 	// 4. Test GET /users/:id
// 	t.Run("GetUserById_Success", func(t *testing.T) {
// 		req, _ := http.NewRequest("GET", "/users/"+userID.Hex(), nil)
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock user, profile, and wallet lookup
// 		singleResultUser := &mongo.SingleResult{}
// 		singleResultUser.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.User) = testUser
// 			return nil
// 		}
// 		singleResultProfile := &mongo.SingleResult{}
// 		singleResultProfile.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.Profile) = testProfile
// 			return nil
// 		}
// 		singleResultWallet := &mongo.SingleResult{}
// 		singleResultWallet.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.Wallet) = testWallet
// 			return nil
// 		}
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"_id": userID}).Return(singleResultUser, nil).Once()
// 		mockProfilesCollection.On("FindOne", mock.Anything, bson.M{"user_id": userID}).Return(singleResultProfile, nil).Once()
// 		mockWalletsCollection.On("FindOne", mock.Anything, bson.M{"owner_id": userID}).Return(singleResultWallet, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response models.UserResponse
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, testUser.Username, response.Username)
// 	})

// 	// 5. Test GET /profile/:id
// 	t.Run("GetProfile_Success", func(t *testing.T) {
// 		req, _ := http.NewRequest("GET", "/profile/"+userID.Hex(), nil)
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock profile lookup
// 		singleResult := &mongo.SingleResult{}
// 		singleResult.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.Profile) = testProfile
// 			return nil
// 		}
// 		mockProfilesCollection.On("FindOne", mock.Anything, bson.M{"user_id": userID}).Return(singleResult, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response models.Profile
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, testProfile.UserID, response.UserID)
// 	})

// 	// 6. Test GET /wallet
// 	t.Run("GetWallet_Success", func(t *testing.T) {
// 		req, _ := http.NewRequest("GET", "/wallet", nil)
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock wallet lookup
// 		singleResult := &mongo.SingleResult{}
// 		singleResult.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.Wallet) = testWallet
// 			return nil
// 		}
// 		mockWalletsCollection.On("FindOne", mock.Anything, bson.M{"owner_id": userID}).Return(singleResult, nil).Once()
// 		mockPG.On("GetVirtualAccount", mock.Anything, "VA123").Return(testVirtualAccount, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response models.Wallet
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, testWallet.ID, response.ID)
// 	})

// 	// 7. Test POST /wallet/fund
// 	t.Run("FundWallet_Success", func(t *testing.T) {
// 		payload := `{"amount":1000}`
// 		req, _ := http.NewRequest("POST", "/wallet/fund", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock user lookup
// 		singleResultUser := &mongo.SingleResult{}
// 		singleResultUser.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.User) = testUser
// 			return nil
// 		}
// 		mockUsersCollection.On("FindOne", mock.Anything, bson.M{"_id": userID}).Return(singleResultUser, nil).Once()

// 		// Mock wallet lookup
// 		singleResultWallet := &mongo.SingleResult{}
// 		singleResultWallet.DecodeFunc = func(v interface{}) error {
// 			*v.(*models.Wallet) = testWallet
// 			return nil
// 		}
// 		mockWalletsCollection.On("FindOne", mock.Anything, bson.M{"owner_id": userID}).Return(singleResultWallet, nil).Once()

// 		// Mock payment gateway
// 		txResponse := &payment.TransactionResponse{
// 			TransactionID: "TX123",
// 			Status:        "pending",
// 			Amount:        1000,
// 		}
// 		verifiedTx := &payment.TransactionResponse{
// 			TransactionID: "TX123",
// 			Status:        "success",
// 			Amount:        1000,
// 		}
// 		mockPG.On("FundVirtualAccount", mock.Anything, "VA123", mock.Anything).Return(txResponse, nil).Once()
// 		mockPG.On("VerifyTransaction", mock.Anything, "TX123").Return(verifiedTx, nil).Once()

// 		// Mock transaction creation and updates
// 		transactionID := primitive.NewObjectID()
// 		mockTransactionsCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: transactionID}, nil).Once()
// 		mockWalletsCollection.On("UpdateOne", mock.Anything, bson.M{"_id": testWallet.ID}, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()
// 		mockTransactionsCollection.On("UpdateOne", mock.Anything, bson.M{"_id": transactionID}, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "Wallet funding initiated successfully", response["message"])
// 	})

// 	t.Run("FundWallet_InvalidAmount", func(t *testing.T) {
// 		payload := `{"amount":0}`
// 		req, _ := http.NewRequest("POST", "/wallet/fund", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Contains(t, response["error"], "Invalid input")
// 	})

// 	// Add similar tests for other routes (e.g., contributions, approvals)
// 	// Example for POST /contributions
// 	t.Run("CreateContribution_Success", func(t *testing.T) {
// 		payload := `{"name":"Test Contribution","description":"A test group","amount":1000,"cycle":"monthly"}`
// 		req, _ := http.NewRequest("POST", "/contributions", strings.NewReader(payload))
// 		req.Header.Set("Content-Type", "application/json")
// 		req.Header.Set("Authorization", "Bearer "+userToken)
// 		w := httptest.NewRecorder()

// 		// Mock contribution creation
// 		contributionID := primitive.NewObjectID()
// 		mockDB.On("Collection", "contributions").Return(mockUsersCollection).Once()
// 		mockUsersCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: contributionID}, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusCreated, w.Code)
// 		var response map[string]string
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "Contribution created successfully", response["message"])
// 	})

// 	// Clean up mocks
// 	mockDB.AssertExpectations(t)
// 	mockUsersCollection.AssertExpectations(t)
// 	mockProfilesCollection.AssertExpectations(t)
// 	mockWalletsCollection.AssertExpectations(t)
// 	mockBlacklistCollection.AssertExpectations(t)
// 	mockTransactionsCollection.AssertExpectations(t)
// 	mockPG.AssertExpectations(t)
// }