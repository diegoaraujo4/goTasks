package auction

import (
	"auctionService/configuration/logger"
	"auctionService/internal/entity/auction_entity"
	"auctionService/internal/internal_error"
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection      *mongo.Collection
	ctx             context.Context
	auctionInterval time.Duration
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	ctx := context.Background()
	return &AuctionRepository{
		Collection:      database.Collection("auctions"),
		ctx:             ctx,
		auctionInterval: getAuctionInterval(),
	}
}

func (ar *AuctionRepository) CreateAuction(
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ar.ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go func() {
		select {
		case <-time.After(ar.auctionInterval):
			ar.updateAuctionStatus(auctionEntityMongo.Id, auction_entity.Completed)
			if err != nil {
				logger.Error("Error trying to update auction status to completed", err)
				return
			}
		case <-ar.ctx.Done():
			logger.Error("Context cancelled while waiting for auction expiry", ar.ctx.Err())
			return
		}
	}()

	return nil
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		logger.Error("Error parsing AUCTION_INTERVAL, using default 5 minutes", err)
		return time.Minute * 5
	}
	return duration
}

func (ar *AuctionRepository) updateAuctionStatus(auctionId string, status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}

	result, err := ar.Collection.UpdateOne(ar.ctx, filter, update)
	if err != nil {
		logger.Error("Error updating auction status in database", err)
		return internal_error.NewInternalServerError("Error updating auction status")
	}

	if result.MatchedCount == 0 {
		logger.Error("Auction not found for status update", nil, zap.String("auction_id", auctionId))
		return internal_error.NewNotFoundError("Auction not found")
	}

	return nil
}
