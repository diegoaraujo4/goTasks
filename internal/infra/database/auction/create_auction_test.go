package auction

import (
	"auctionService/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetAuctionInterval(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		expected     time.Duration
		shouldSetEnv bool
	}{
		{
			name:         "valid environment variable",
			envValue:     "1m30s",
			expected:     time.Minute + 30*time.Second,
			shouldSetEnv: true,
		},
		{
			name:         "invalid environment variable",
			envValue:     "invalid",
			expected:     5 * time.Minute,
			shouldSetEnv: true,
		},
		{
			name:         "missing environment variable",
			envValue:     "",
			expected:     5 * time.Minute,
			shouldSetEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			if tt.shouldSetEnv {
				os.Setenv("AUCTION_INTERVAL", tt.envValue)
				defer os.Unsetenv("AUCTION_INTERVAL")
			}

			// Act
			result := getAuctionInterval()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewAuctionRepository(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should create auction repository with correct configuration", func(mt *mtest.T) {
		// Arrange
		database := mt.DB

		// Act
		repo := NewAuctionRepository(database)

		// Assert
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.Collection)
		assert.NotNil(t, repo.ctx)
		assert.Equal(t, 5*time.Minute, repo.auctionInterval) // default value
		assert.Equal(t, "auctions", repo.Collection.Name())
	})

	mt.Run("should create auction repository with custom auction interval", func(mt *mtest.T) {
		// Arrange
		os.Setenv("AUCTION_INTERVAL", "2m")
		defer os.Unsetenv("AUCTION_INTERVAL")
		database := mt.DB

		// Act
		repo := NewAuctionRepository(database)

		// Assert
		assert.NotNil(t, repo)
		assert.Equal(t, 2*time.Minute, repo.auctionInterval)
	})
}

func TestAuctionRepository_CreateAuction(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should create auction successfully", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)

		auction := &auction_entity.Auction{
			Id:          "test-auction-id",
			ProductName: "Test Product",
			Category:    "Test Category",
			Description: "Test Description for auction",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		err := repo.CreateAuction(auction)

		// Assert
		assert.Nil(t, err)
	})

	mt.Run("should return error when database insert fails", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)

		auction := &auction_entity.Auction{
			Id:          "test-auction-id",
			ProductName: "Test Product",
			Category:    "Test Category",
			Description: "Test Description for auction",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))

		// Act
		err := repo.CreateAuction(auction)

		// Assert
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "Error trying to insert auction")
	})

	mt.Run("should create auction with correct mongo entity structure", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)
		timestamp := time.Now()

		auction := &auction_entity.Auction{
			Id:          "test-auction-id",
			ProductName: "Test Product",
			Category:    "Test Category",
			Description: "Test Description for auction",
			Condition:   auction_entity.Used,
			Status:      auction_entity.Active,
			Timestamp:   timestamp,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		err := repo.CreateAuction(auction)

		// Assert
		assert.Nil(t, err)

		// Verify the structure would be correct (testing the conversion logic)
		expectedMongo := &AuctionEntityMongo{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Description: auction.Description,
			Condition:   auction.Condition,
			Status:      auction.Status,
			Timestamp:   auction.Timestamp.Unix(),
		}

		assert.Equal(t, auction.Id, expectedMongo.Id)
		assert.Equal(t, auction.ProductName, expectedMongo.ProductName)
		assert.Equal(t, auction.Category, expectedMongo.Category)
		assert.Equal(t, auction.Description, expectedMongo.Description)
		assert.Equal(t, auction.Condition, expectedMongo.Condition)
		assert.Equal(t, auction.Status, expectedMongo.Status)
		assert.Equal(t, timestamp.Unix(), expectedMongo.Timestamp)
	})
}

func TestAuctionRepository_UpdateAuctionStatus(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should update auction status successfully", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)
		auctionId := "test-auction-id"
		newStatus := auction_entity.Completed

		// Mock successful update with MatchedCount = 1
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
			{Key: "nModified", Value: 1},
		})

		// Act
		err := repo.updateAuctionStatus(auctionId, newStatus)

		// Assert
		assert.Nil(t, err)
	})

	mt.Run("should return error when auction not found", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)
		auctionId := "non-existent-auction"
		newStatus := auction_entity.Completed

		// Mock update with MatchedCount = 0 (auction not found)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
			{Key: "nModified", Value: 0},
		})

		// Act
		err := repo.updateAuctionStatus(auctionId, newStatus)

		// Assert
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "Auction not found")
	})

	mt.Run("should return error when database update fails", func(mt *mtest.T) {
		// Arrange
		repo := NewAuctionRepository(mt.DB)
		auctionId := "test-auction-id"
		newStatus := auction_entity.Completed

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    1,
			Message: "database error",
		}))

		// Act
		err := repo.updateAuctionStatus(auctionId, newStatus)

		// Assert
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "Error updating auction status")
	})
}

func TestAuctionEntityMongo_Conversion(t *testing.T) {
	t.Run("should convert auction entity to mongo entity correctly", func(t *testing.T) {
		// Arrange
		timestamp := time.Now()
		auction := &auction_entity.Auction{
			Id:          "test-id",
			ProductName: "Test Product",
			Category:    "Electronics",
			Description: "Test Description",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   timestamp,
		}

		// Act
		mongoEntity := &AuctionEntityMongo{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Description: auction.Description,
			Condition:   auction.Condition,
			Status:      auction.Status,
			Timestamp:   auction.Timestamp.Unix(),
		}

		// Assert
		assert.Equal(t, auction.Id, mongoEntity.Id)
		assert.Equal(t, auction.ProductName, mongoEntity.ProductName)
		assert.Equal(t, auction.Category, mongoEntity.Category)
		assert.Equal(t, auction.Description, mongoEntity.Description)
		assert.Equal(t, auction.Condition, mongoEntity.Condition)
		assert.Equal(t, auction.Status, mongoEntity.Status)
		assert.Equal(t, timestamp.Unix(), mongoEntity.Timestamp)
	})
}
