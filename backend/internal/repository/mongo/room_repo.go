package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"zametka/internal/domain"
)

const roomsCollection = "rooms"

type RoomRepository struct {
	coll *mongo.Collection
}

func NewRoomRepository(db *mongo.Database) *RoomRepository {
	return &RoomRepository{coll: db.Collection(roomsCollection)}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.coll.InsertOne(ctx, room)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrCodeTaken
		}
		return fmt.Errorf("insert room: %w", err)
	}
	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	var room domain.Room
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("find room by id: %w", err)
	}
	return &room, nil
}

func (r *RoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	var room domain.Room
	err := r.coll.FindOne(ctx, bson.M{"code": code}).Decode(&room)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("find room by code: %w", err)
	}
	return &room, nil
}

func (r *RoomRepository) AddMember(ctx context.Context, code string, m domain.Member) (*domain.Room, error) {
	filter := bson.M{
		"code": code,
		"$expr": bson.M{
			"$lt": bson.A{
				bson.M{"$size": "$members"},
				2,
			},
		},
	}
	update := bson.M{
		"$push": bson.M{"members": m},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var room domain.Room
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&room)
	if err == nil {
		return &room, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("add member: %w", err)
	}

	existing, getErr := r.GetByCode(ctx, code)
	if getErr != nil {
		return nil, getErr
	}
	if len(existing.Members) >= 2 {
		return nil, domain.ErrRoomFull
	}
	return nil, domain.ErrRoomNotFound
}
